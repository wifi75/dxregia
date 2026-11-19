package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oschwald/maxminddb-golang"
)

type mojoLogEvent struct {
	At        int64
	Kind      string
	Message   string
	Call      string `json:",omitempty"`
	Host      string `json:",omitempty"`
	Session   string `json:",omitempty"`
	Direction string `json:",omitempty"`
	Country   string `json:",omitempty"`
}

var mojoConnectionLine = regexp.MustCompile(`^([0-9]{9,12})\^\(connect\) (?:Connected [0-9]+ to ([A-Za-z0-9.:-]+) ([0-9]{1,5})|\*\*\*Connect [0-9]+ Failed to ([A-Za-z0-9.:-]+) ([0-9]{1,5})(?: .*?)?|connect [0-9]+: timed out after ([0-9]+) seconds(?: (incoming|outgoing)(?: ([A-Z0-9-]{1,12}))?)?)\s*$`)

var (
	mojoAcceptLine    = regexp.MustCompile(`^([0-9]{9,12})\^\((?:connect|conn)\) ExtMsg accept ([0-9]+):\S+ from ([0-9A-Fa-f:.]+):[0-9]{1,5}$`)
	mojoLegacyInbound = regexp.MustCompile(`^[0-9]+\^\(connect\) connect ([0-9]+): timed out after ([0-9]{1,5}) seconds incoming$`)
	mojoStageLine     = regexp.MustCompile(`^([0-9]{9,12})\^\(connect\) DXRegia session ([0-9]+) (incoming|outgoing) ([A-Z0-9/-]{1,16}) ([A-Za-z0-9.:-]{1,253}) ([a-z-]+) ([0-9]{1,5})$`)
	mojoPartnerLine   = regexp.MustCompile(`^([0-9]{9,12})\^\(connect\) DXRegia partner ([A-Z0-9-]{1,12}) (.{1,80})$`)
	mojoStageNames    = map[string]string{
		"login-timeout":         "Nominativo non ricevuto entro il tempo previsto",
		"password-timeout":      "Password Telnet non ricevuta entro il tempo previsto",
		"password-rejected":     "Password Telnet rifiutata dal nodo locale",
		"remote-prompt-timeout": "Prompt atteso dal nodo remoto non ricevuto",
		"login-sent":            "Nominativo locale inviato al nodo remoto",
		"password-sent":         "Password inviata al nodo remoto",
		"protocol-start":        "Avviato lo scambio tra nodi DXSpider",
	}
)

func mojoDebugParser(tail string) func(string) (mojoLogEvent, bool) {
	type accepted struct {
		at int64
		ip string
	}
	var peers = make(map[string]accepted)
	for _, line := range strings.Split(tail, "\n") {
		m := mojoAcceptLine.FindStringSubmatch(strings.TrimSpace(line))
		if m != nil && net.ParseIP(m[3]) != nil {
			at, _ := strconv.ParseInt(m[1], 10, 64)
			peers[m[2]] = accepted{at, m[3]}
		}
	}
	return func(line string) (mojoLogEvent, bool) {
		if m := mojoStageLine.FindStringSubmatch(line); m != nil {
			message, ok := mojoStageNames[m[6]]
			if !ok {
				return mojoLogEvent{}, false
			}
			at, _ := strconv.ParseInt(m[1], 10, 64)
			direction := "Ingresso"
			if m[3] == "outgoing" {
				direction = "Uscita"
			}
			if m[7] != "0" {
				message += " (" + m[7] + " s)"
			}
			return mojoLogEvent{At: at, Kind: "Collegamenti", Message: message, Call: strings.Trim(m[4], "-"), Host: strings.Trim(m[5], "-"), Session: m[2], Direction: direction}, true
		}
		if m := mojoPartnerLine.FindStringSubmatch(line); m != nil {
			switch m[3] {
			case "Account Mojo esistente da verificare", "Nodo bloccato in Mojo", "Pronto in uscita", "Abilitato in Mojo", "Tentativo avviato", "Account Mojo non abilitato", "Sessione gia presente", "Parametri non validi", "Avvio non riuscito", "Indicativo riservato", "Richiesta non valida":
				at, _ := strconv.ParseInt(m[1], 10, 64)
				message := m[3]
				if message == "Account Mojo esistente da verificare" {
					message = "Configurazione sospesa: verifica il tipo e le credenziali locali del nodo"
				}
				return mojoLogEvent{At: at, Kind: "Collegamenti", Message: message, Call: m[2]}, true
			}
			return mojoLogEvent{}, false
		}
		event, ok := parseMojoConnection(line)
		if m := mojoLegacyInbound.FindStringSubmatch(line); m != nil {
			event.Session = m[1]
			seconds, _ := strconv.ParseInt(m[2], 10, 64)
			if peer, found := peers[m[1]]; found && event.At >= peer.at && event.At-peer.at <= seconds+10 {
				event.Host = peer.ip
			}
		}
		return event, ok
	}
}

func parseMojoSession(line string) (mojoLogEvent, bool) {
	at, kind, call, host, connected, ok := parseAccessLine(line)
	if !ok || kind != "node" {
		return mojoLogEvent{}, false
	}
	message := "Nodo disconnesso"
	if connected {
		message = "Sessione nodo aperta"
	}
	if net.ParseIP(host) == nil && !partnerHost.MatchString(host) {
		host = ""
	}
	return mojoLogEvent{At: at, Kind: "Collegamenti", Message: message, Call: call, Host: host}, true
}

func registerMojoLog(mux *http.ServeMux, auth *sysopAuth, root string) {
	if root == "" || auth == nil {
		return
	}
	mux.HandleFunc("GET /api/mojo-log", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		events := make([]mojoLogEvent, 0, 80)
		query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		showAnonymous := r.URL.Query().Get("anonymous") != "false"
		filter := func(event mojoLogEvent, ok bool) (mojoLogEvent, bool) {
			return event, ok && (showAnonymous || event.Direction != "Ingresso" || event.Call != "") &&
				strings.Contains(strings.ToLower(event.Call+" "+event.Host+" "+event.Session+" "+event.Message), query)
		}
		base := filepath.Join(root, "local_data")
		now := time.Now().UTC()
		for day := 0; day < 2 && len(events) < 80; day++ {
			date := now.AddDate(0, 0, -day)
			path := filepath.Join(base, "debug", fmt.Sprintf("%04d", date.Year()), fmt.Sprintf("%03d.dat", date.YearDay()))
			tail := logTail(path)
			parse := mojoDebugParser(tail)
			events = appendMojoEvents(events, tail, func(line string) (mojoLogEvent, bool) { return filter(parse(line)) })
		}
		month := filepath.Join(base, "log", fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d.dat", now.Month()))
		events = append(events, appendMojoEvents(nil, logTail(month), func(line string) (mojoLogEvent, bool) { return filter(parseMojoSession(line)) })...)
		events = append(events, appendMojoEvents(nil, logTail(filepath.Join(base, "dxregia-startup.log")), func(line string) (mojoLogEvent, bool) { return filter(parseMojoStartup(line)) })...)
		sort.SliceStable(events, func(i, j int) bool { return events[i].At > events[j].At })
		if db, err := maxminddb.Open(filepath.Join(root, "dbip-country.mmdb")); err == nil {
			defer db.Close()
			for i := range events {
				if events[i].Call != "" {
					events[i].Country = accessCountry(db, events[i].Host)
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	})
}

func logTail(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return ""
	}
	const maxTail = 512 << 10
	start := max(0, info.Size()-maxTail)
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(file, maxTail))
	if err != nil {
		return ""
	}
	contents := string(body)
	if start != 0 {
		_, contents, _ = strings.Cut(contents, "\n")
	}
	return contents
}

func appendMojoEvents(events []mojoLogEvent, tail string, parse func(string) (mojoLogEvent, bool)) []mojoLogEvent {
	lines := strings.Split(tail, "\n")
	for index := len(lines) - 1; index >= 0 && len(events) < 100; index-- {
		if event, ok := parse(strings.TrimSpace(lines[index])); ok {
			events = append(events, event)
		}
	}
	return events
}

func parseMojoConnection(line string) (mojoLogEvent, bool) {
	fields := mojoConnectionLine.FindStringSubmatch(line)
	if fields == nil {
		return mojoLogEvent{}, false
	}
	at, _ := strconv.ParseInt(fields[1], 10, 64)
	switch {
	case fields[2] != "":
		return mojoLogEvent{At: at, Kind: "Collegamenti", Message: "Connessione TCP aperta verso " + net.JoinHostPort(fields[2], fields[3]), Host: fields[2], Direction: "Uscita"}, true
	case fields[4] != "":
		return mojoLogEvent{At: at, Kind: "Collegamenti", Message: "Connessione TCP non riuscita verso " + net.JoinHostPort(fields[4], fields[5]), Host: fields[4], Direction: "Uscita"}, true
	default:
		message := "Timeout del collegamento"
		var direction string
		switch fields[7] {
		case "incoming":
			message = "Accesso Telnet in ingresso non completato"
			direction = "Ingresso"
		case "outgoing":
			message = "Collegamento in uscita non completato"
			direction = "Uscita"
			if fields[8] != "" {
				message += " verso " + fields[8]
			}
		}
		return mojoLogEvent{At: at, Kind: "Collegamenti", Message: message + " dopo " + fields[6] + " secondi", Call: fields[8], Direction: direction}, true
	}
}

func parseMojoStartup(line string) (mojoLogEvent, bool) {
	const kind = "Avvio"
	switch {
	case strings.Contains(line, "Can't locate") && strings.Contains(line, "in @INC"):
		return mojoLogEvent{Kind: kind, Message: "Modulo Perl mancante"}, true
	case strings.Contains(line, "Compilation failed") || strings.Contains(line, "BEGIN failed"):
		return mojoLogEvent{Kind: kind, Message: "Errore di compilazione Perl"}, true
	case strings.Contains(line, "Permission denied"):
		return mojoLogEvent{Kind: kind, Message: "Permesso negato durante l’avvio"}, true
	case strings.Contains(line, "Address already in use"):
		return mojoLogEvent{Kind: kind, Message: "Porta già in uso"}, true
	default:
		return mojoLogEvent{}, false
	}
}
