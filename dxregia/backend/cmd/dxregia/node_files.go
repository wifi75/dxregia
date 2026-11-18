package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// Messaggio del giorno (file «motd») e pianificazione (local_cmd/crontab) di DXSpider.
// Il crontab di DXSpider valuta la colonna del comando come codice Perl: qui si accetta solo la forma
// run_cmd("<comando>") con un testo semplice, mai codice libero.

const motdMaxBytes = 2000

type cronEntry struct {
	Minute, Hour, Day, Month, Weekday, Command string
}

type cronFile struct {
	Entries []cronEntry
	Extra   []string // righe che il pannello non sa interpretare: si conservano così come sono
}

var (
	cronField   = regexp.MustCompile(`^[0-9A-Za-z*/,\-]{1,40}$`)
	cronCommand = regexp.MustCompile(`^[A-Za-z0-9_ ./:=,+\-]{1,200}$`)
	cronLine    = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+run_cmd\("([^"\\$@]*)"\)\s*$`)
)

func (e cronEntry) valid() bool {
	for _, field := range []string{e.Minute, e.Hour, e.Day, e.Month, e.Weekday} {
		if !cronField.MatchString(field) {
			return false
		}
	}
	return cronCommand.MatchString(e.Command) && strings.TrimSpace(e.Command) != ""
}

func parseCrontab(text string) cronFile {
	var file cronFile
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "# Gestito dal pannello DXRegia") {
			continue
		}
		if m := cronLine.FindStringSubmatch(line); m != nil {
			entry := cronEntry{m[1], m[2], m[3], m[4], m[5], m[6]}
			if entry.valid() {
				file.Entries = append(file.Entries, entry)
				continue
			}
		}
		file.Extra = append(file.Extra, line)
	}
	return file
}

func (f cronFile) render() string {
	var b strings.Builder
	b.WriteString("# Gestito dal pannello DXRegia (Messaggio e pianificazione).\n")
	for _, line := range f.Extra {
		b.WriteString(line + "\n")
	}
	for _, e := range f.Entries {
		b.WriteString(e.Minute + " " + e.Hour + " " + e.Day + " " + e.Month + " " + e.Weekday + ` run_cmd("` + e.Command + `")` + "\n")
	}
	return b.String()
}

func validMotd(text string) bool {
	if len(text) > motdMaxBytes {
		return false
	}
	for _, r := range text {
		if r == 0 || (r < 32 && r != '\n' && r != '\r' && r != '\t') {
			return false
		}
	}
	return true
}

func writeNodeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0640); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func registerNodeFiles(mux *http.ServeMux, root string, auth *sysopAuth) {
	if root == "" || auth == nil {
		return
	}
	var lock sync.Mutex
	motdPath := filepath.Join(root, "motd")
	cronPath := filepath.Join(root, "local_cmd", "crontab")
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
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024))
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
	mux.HandleFunc("GET /api/motd", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, false) {
			return
		}
		lock.Lock()
		defer lock.Unlock()
		data, _ := os.ReadFile(motdPath)
		reply(w, map[string]any{"Text": string(data), "MaxBytes": motdMaxBytes})
	})
	mux.HandleFunc("PUT /api/motd", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, true) {
			return
		}
		var change struct{ Text string }
		if !decode(w, r, &change) {
			return
		}
		if !validMotd(change.Text) {
			http.Error(w, "Messaggio non valido: massimo 2000 byte, senza caratteri di controllo.", http.StatusBadRequest)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		if strings.TrimSpace(change.Text) == "" {
			if err := os.Remove(motdPath); err != nil && !os.IsNotExist(err) {
				http.Error(w, "Impossibile rimuovere il messaggio.", http.StatusInternalServerError)
				return
			}
		} else if err := writeNodeFile(motdPath, []byte(strings.TrimRight(change.Text, "\r\n")+"\n")); err != nil {
			http.Error(w, "Impossibile salvare il messaggio.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/crontab", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, false) {
			return
		}
		lock.Lock()
		defer lock.Unlock()
		data, _ := os.ReadFile(cronPath)
		file := parseCrontab(string(data))
		if file.Entries == nil {
			file.Entries = []cronEntry{}
		}
		if file.Extra == nil {
			file.Extra = []string{}
		}
		reply(w, file)
	})
	mux.HandleFunc("PUT /api/crontab", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r, true) {
			return
		}
		var change struct{ Entries []cronEntry }
		if !decode(w, r, &change) {
			return
		}
		if len(change.Entries) > 100 {
			http.Error(w, "Troppe voci (massimo 100).", http.StatusBadRequest)
			return
		}
		for _, entry := range change.Entries {
			if !entry.valid() {
				http.Error(w, "Voce non valida: orario o comando con caratteri non ammessi.", http.StatusBadRequest)
				return
			}
		}
		lock.Lock()
		defer lock.Unlock()
		data, _ := os.ReadFile(cronPath)
		file := parseCrontab(string(data))
		file.Entries = change.Entries
		if err := writeNodeFile(cronPath, []byte(file.render())); err != nil {
			http.Error(w, "Impossibile salvare la pianificazione.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
