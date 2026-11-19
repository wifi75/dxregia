package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Impostazioni di DXSpider che si cambiavano a mano nei file: script di avvio, liste di blocco,
// hop del protocollo e filtri. Ogni gruppo ha azioni esplicite con parametri validati: nessun testo
// libero arriva a Perl. Le righe che il pannello non sa interpretare si conservano senza toccarle.

var (
	startupSimple = regexp.MustCompile(`^(?:(?:set|unset)/[a-z0-9]+|load/[a-z0-9]+)(?: [A-Za-z0-9_.:/=,+\- ]{1,150})?$`)
	startupVar    = regexp.MustCompile(`^set/var \$[A-Za-z_][A-Za-z0-9_]*(?:::[A-Za-z_][A-Za-z0-9_]*)* = (?:-?[0-9]+(?:\.[0-9]+)?|'[A-Za-z0-9 _.:/\-]{0,60}')$`)
	startupLink   = regexp.MustCompile(`^connect [A-Z0-9]{3,10}(?:-[0-9]{1,2})?$`)
	badWord       = regexp.MustCompile(`^[A-Z0-9]{2,30}$`)
	badCall       = regexp.MustCompile(`^[A-Z0-9]{1,8}(?:[/-][A-Z0-9]{1,8}){0,2}$`)
	filterSort    = map[string]bool{"ann": true, "route": true, "spots": true, "wcy": true, "wwv": true}
	filterName    = regexp.MustCompile(`^(?:in_)?(?:[A-Za-z0-9\-]{1,15}|node_default|user_default)\.pl$`)
	dumperEntry   = regexp.MustCompile(`'?([A-Z0-9/\-]{1,24})'? => ([0-9]{1,12})`)
)

func validStartupCommand(line string) bool {
	return startupSimple.MatchString(line) && !strings.HasPrefix(line, "set/var") ||
		startupVar.MatchString(line) || startupLink.MatchString(line)
}

// splitManaged separa le righe che il pannello gestisce da quelle da conservare (commenti e altro).
func splitManaged(text string, managed func(string) bool) (items, extra []string) {
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line = strings.TrimRight(line, " \t")
		switch {
		case strings.TrimSpace(line) == "":
		case managed(strings.TrimSpace(line)):
			items = append(items, strings.TrimSpace(line))
		default:
			extra = append(extra, line)
		}
	}
	return
}

func validIPEntry(line string) bool {
	if _, err := netip.ParsePrefix(line); err == nil {
		return true
	}
	_, err := netip.ParseAddr(line)
	return err == nil
}

// parseBadDX legge il file di DXSpider (un hash Perl «bless( {...}, 'DXHash' )»).
func parseBadDX(text string) map[string]int64 {
	out := map[string]int64{}
	for _, m := range dumperEntry.FindAllStringSubmatch(text, -1) {
		if m[1] == "NAME" || !badCall.MatchString(m[1]) {
			continue
		}
		at, _ := strconv.ParseInt(m[2], 10, 64)
		out[m[1]] = at
	}
	return out
}

func renderBadDX(entries map[string]int64) string {
	calls := make([]string, 0, len(entries))
	for call := range entries {
		calls = append(calls, call)
	}
	sort.Strings(calls)
	var b strings.Builder
	b.WriteString("bless( {\n  name => 'baddx',\n")
	for _, call := range calls {
		fmt.Fprintf(&b, "  '%s' => %d,\n", call, entries[call])
	}
	b.WriteString("}, 'DXHash' )\n")
	return b.String()
}

type hopsConfig struct {
	Default int
	Table   map[string]int
}

func (h hopsConfig) valid() bool {
	if h.Default < 1 || h.Default > 99 || len(h.Table) > 80 {
		return false
	}
	for pc, hops := range h.Table {
		number, err := strconv.Atoi(pc)
		if err != nil || number < 1 || number > 999 || strconv.Itoa(number) != pc || hops < 1 || hops > 99 {
			return false
		}
	}
	return true
}

// Valori predefiniti di DXProtVars.pm, usati finché il sysop non salva i propri.
func defaultHops() hopsConfig {
	return hopsConfig{Default: 30, Table: map[string]int{"16": 15, "17": 15, "19": 15, "21": 15}}
}

func registerNodeLists(mux *http.ServeMux, root string, auth *sysopAuth) {
	if root == "" || auth == nil {
		return
	}
	var lock sync.Mutex
	startupPath := filepath.Join(root, "scripts", "startup")
	wordsPath := filepath.Join(root, "local_data", "badword.new")
	ipsPath := filepath.Join(root, "local_data", "badip.local")
	dxPath := filepath.Join(root, "local_data", "baddx")
	hopsPath := filepath.Join(root, "local", "dxregia-hops.json")
	filterDir := filepath.Join(root, "filter")

	guard := func(w http.ResponseWriter, r *http.Request, write bool) bool {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return false
		}
		if write && r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return false
		}
		return true
	}
	decode := func(w http.ResponseWriter, r *http.Request, into any) bool {
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512*1024))
		decoder.DisallowUnknownFields()
		if decoder.Decode(into) != nil || decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return false
		}
		return true
	}
	reply := func(w http.ResponseWriter, value any) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(value)
	}
	read := func(path string) string {
		data, _ := os.ReadFile(path)
		return string(data)
	}

	// Script di avvio
	mux.HandleFunc("GET /api/startup", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, false) {
			return
		}
		lock.Lock()
		defer lock.Unlock()
		commands, extra := splitManaged(read(startupPath), validStartupCommand)
		reply(w, map[string]any{"Commands": nonNil(commands), "Extra": nonNil(extra)})
	})
	mux.HandleFunc("PUT /api/startup", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, true) {
			return
		}
		var change struct{ Commands []string }
		if !decode(w, r, &change) {
			return
		}
		if len(change.Commands) > 200 {
			http.Error(w, "Troppi comandi (massimo 200).", http.StatusBadRequest)
			return
		}
		for i, command := range change.Commands {
			change.Commands[i] = strings.TrimSpace(command)
			if !validStartupCommand(change.Commands[i]) {
				http.Error(w, "Comando non ammesso: «"+command+"».", http.StatusBadRequest)
				return
			}
		}
		lock.Lock()
		defer lock.Unlock()
		_, extra := splitManaged(read(startupPath), validStartupCommand)
		text := strings.Join(append(extra, change.Commands...), "\n") + "\n"
		if err := writeNodeFile(startupPath, []byte(text)); err != nil {
			http.Error(w, "Impossibile salvare lo script di avvio.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Liste di blocco
	mux.HandleFunc("GET /api/badlists", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, false) {
			return
		}
		lock.Lock()
		defer lock.Unlock()
		words, wordsExtra := splitManaged(read(wordsPath), func(s string) bool { return badWord.MatchString(strings.ToUpper(s)) })
		ips, ipsExtra := splitManaged(read(ipsPath), validIPEntry)
		dx := []string{}
		for call := range parseBadDX(read(dxPath)) {
			dx = append(dx, call)
		}
		sort.Strings(dx)
		for i := range words {
			words[i] = strings.ToUpper(words[i])
		}
		reply(w, map[string]any{"Words": nonNil(words), "IPs": nonNil(ips), "DX": dx,
			"OtherWords": len(wordsExtra) - commentCount(wordsExtra), "OtherIPs": len(ipsExtra) - commentCount(ipsExtra)})
	})
	mux.HandleFunc("PUT /api/badlists", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, true) {
			return
		}
		var change struct{ Words, IPs, DX []string }
		if !decode(w, r, &change) {
			return
		}
		if len(change.Words) > 5000 || len(change.IPs) > 20000 || len(change.DX) > 5000 {
			http.Error(w, "Liste troppo lunghe.", http.StatusBadRequest)
			return
		}
		words := dedupe(change.Words, strings.ToUpper)
		ips := dedupe(change.IPs, strings.TrimSpace)
		dx := dedupe(change.DX, strings.ToUpper)
		for _, word := range words {
			if !badWord.MatchString(word) {
				http.Error(w, "Parola non valida: «"+word+"» (solo lettere e cifre, 2–30 caratteri).", http.StatusBadRequest)
				return
			}
		}
		for _, ip := range ips {
			if !validIPEntry(ip) {
				http.Error(w, "Indirizzo non valido: «"+ip+"» (usa 192.0.2.1 oppure 192.0.2.0/24).", http.StatusBadRequest)
				return
			}
		}
		for _, call := range dx {
			if !badCall.MatchString(call) {
				http.Error(w, "Nominativo non valido: «"+call+"».", http.StatusBadRequest)
				return
			}
		}
		lock.Lock()
		defer lock.Unlock()
		_, wordsExtra := splitManaged(read(wordsPath), func(s string) bool { return badWord.MatchString(strings.ToUpper(s)) })
		_, ipsExtra := splitManaged(read(ipsPath), validIPEntry)
		old := parseBadDX(read(dxPath))
		now := time.Now().Unix()
		entries := map[string]int64{}
		for _, call := range dx {
			if at, ok := old[call]; ok {
				entries[call] = at
			} else {
				entries[call] = now
			}
		}
		files := []struct {
			path string
			text string
		}{
			{wordsPath, strings.Join(append(wordsExtra, words...), "\n") + "\n"},
			{ipsPath, strings.Join(append(ipsExtra, ips...), "\n") + "\n"},
			{dxPath, renderBadDX(entries)},
		}
		for _, file := range files {
			if err := writeNodeFile(file.path, []byte(file.text)); err != nil {
				http.Error(w, "Impossibile salvare le liste.", http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Hop del protocollo
	mux.HandleFunc("GET /api/hops", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, false) {
			return
		}
		lock.Lock()
		defer lock.Unlock()
		hops, custom := defaultHops(), false
		if data, err := os.ReadFile(hopsPath); err == nil {
			var stored hopsConfig
			if json.Unmarshal(data, &stored) == nil && stored.valid() {
				hops, custom = stored, true
			}
		}
		reply(w, map[string]any{"Default": hops.Default, "Table": hops.Table, "Custom": custom})
	})
	mux.HandleFunc("PUT /api/hops", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, true) {
			return
		}
		var change hopsConfig
		if !decode(w, r, &change) {
			return
		}
		if change.Table == nil {
			change.Table = map[string]int{}
		}
		if !change.valid() {
			http.Error(w, "Hop non validi: valori da 1 a 99 e numeri PC da 1 a 999.", http.StatusBadRequest)
			return
		}
		encoded, _ := json.Marshal(change)
		lock.Lock()
		defer lock.Unlock()
		if err := writeNodeFile(hopsPath, encoded); err != nil {
			http.Error(w, "Impossibile salvare gli hop.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	finishFilterChange := func(w http.ResponseWriter, sortName string) {
		token, err := randomToken()
		if err == nil {
			err = writeNodeFile(filepath.Join(filterDir, sortName, ".dxregia-reload"), []byte(token))
		}
		if err != nil {
			http.Error(w, "Modifica salvata. Il motore applicherà la regola al prossimo collegamento perché non è stato possibile richiedere il ricaricamento.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}

	// Filtri nativi di DXSpider. L’editor modifica una sola regola spot per volta.
	mux.HandleFunc("GET /api/filters/spot-options", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, false) {
			return
		}
		options := map[string][]string{}
		for band, modes := range spotFilterOptions(root) {
			options[band] = []string{}
			for mode := range modes {
				if mode != "band" {
					options[band] = append(options[band], mode)
				}
			}
			sort.Strings(options[band])
		}
		reply(w, options)
	})
	mux.HandleFunc("PUT /api/filters/spots/{name}/{number}", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, true) {
			return
		}
		name, number, ok := readSpotFilterChange(w, r)
		if !ok {
			return
		}
		if name == "in_user_default.pl" {
			http.Error(w, "I filtri in ingresso si applicano ai nodi partner.", http.StatusBadRequest)
			return
		}
		var change struct {
			spotFilterRule
			PreviousAction string
		}
		if !decode(w, r, &change) {
			return
		}
		previous := change.PreviousAction
		if previous != "" && previous != "reject" && previous != "accept" {
			http.Error(w, "Azione precedente non valida.", http.StatusBadRequest)
			return
		}
		side, err := change.spotFilterRule.build(spotFilterOptions(root), loadDXCC(root))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		path := filepath.Join(filterDir, "spots", name)
		if _, err := os.Stat(path); os.IsNotExist(err) && name != strings.ToLower(name) {
			lowerName := strings.ToLower(name)
			lowerPath := filepath.Join(filterDir, "spots", lowerName)
			if _, err := os.Stat(lowerPath); err == nil {
				path, name = lowerPath, lowerName
			}
		}
		if err := changeSpotFilter(path, name, number, change.Action, previous, &side); err != nil {
			http.Error(w, "Impossibile salvare la regola. "+err.Error(), http.StatusConflict)
			return
		}
		finishFilterChange(w, "spots")
	})
	mux.HandleFunc("DELETE /api/filters/spots/{name}/{number}", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, false) {
			return
		}
		name, number, ok := readSpotFilterChange(w, r)
		action := r.URL.Query().Get("action")
		if !ok {
			return
		}
		if action != "accept" && action != "reject" {
			http.Error(w, "Azione del filtro non valida.", http.StatusBadRequest)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		if err := changeSpotFilter(filepath.Join(filterDir, "spots", name), name, number, action, "", nil); err != nil {
			http.Error(w, "Impossibile eliminare la regola. "+err.Error(), http.StatusConflict)
			return
		}
		finishFilterChange(w, "spots")
	})
	mux.HandleFunc("GET /api/filters", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, false) {
			return
		}
		type filterItem struct {
			Sort, Name, Text string
			Size             int64
			Rules            []spotFilterSlot
		}
		items := []filterItem{}
		lock.Lock()
		defer lock.Unlock()
		for sortName := range filterSort {
			entries, _ := os.ReadDir(filepath.Join(filterDir, sortName))
			for _, entry := range entries {
				if entry.IsDir() || !filterName.MatchString(entry.Name()) {
					continue
				}
				info, err := entry.Info()
				if err != nil {
					continue
				}
				text := read(filepath.Join(filterDir, sortName, entry.Name()))
				var rules []spotFilterSlot
				if sortName == "spots" {
					rules = readSpotFilterSlots([]byte(text))
				}
				if len(text) > 4000 {
					text = text[:4000] + "…"
				}
				items = append(items, filterItem{sortName, entry.Name(), text, info.Size(), rules})
			}
		}
		sort.Slice(items, func(i, j int) bool {
			if items[i].Sort != items[j].Sort {
				return items[i].Sort < items[j].Sort
			}
			return items[i].Name < items[j].Name
		})
		reply(w, items)
	})
	mux.HandleFunc("DELETE /api/filters/{sort}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, false) {
			return
		}
		sortName, name := r.PathValue("sort"), r.PathValue("name")
		if !filterSort[sortName] || !filterName.MatchString(name) {
			http.Error(w, "Filtro non valido.", http.StatusBadRequest)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		if err := os.Remove(filepath.Join(filterDir, sortName, name)); err != nil {
			http.Error(w, "Filtro non trovato.", http.StatusNotFound)
			return
		}
		finishFilterChange(w, sortName)
	})
}

func nonNil(items []string) []string {
	if items == nil {
		return []string{}
	}
	return items
}

func commentCount(lines []string) int {
	count := 0
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			count++
		}
	}
	return count
}

func dedupe(items []string, normalize func(string) string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		item = normalize(strings.TrimSpace(item))
		if item != "" && !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	return out
}
