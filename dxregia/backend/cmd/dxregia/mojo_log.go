package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type mojoLogEvent struct {
	At      int64
	Kind    string
	Message string
}

var mojoConnectionLine = regexp.MustCompile(`^([0-9]{9,12})\^\(connect\) (?:Connected [0-9]+ to ([A-Za-z0-9.:-]+) ([0-9]{1,5})|\*\*\*Connect [0-9]+ Failed to ([A-Za-z0-9.:-]+) ([0-9]{1,5})(?: .*?)?|connect [0-9]+: timed out after ([0-9]+) seconds)\s*$`)

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
		base := filepath.Join(root, "local_data")
		now := time.Now().UTC()
		for day := 0; day < 2 && len(events) < 80; day++ {
			date := now.AddDate(0, 0, -day)
			path := filepath.Join(base, "debug", fmt.Sprintf("%04d", date.Year()), fmt.Sprintf("%03d.dat", date.YearDay()))
			events = appendMojoEvents(events, logTail(path), parseMojoConnection)
		}
		events = appendMojoEvents(events, logTail(filepath.Join(base, "dxregia-startup.log")), parseMojoStartup)
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
	const maxTail = 128 << 10
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
		return mojoLogEvent{at, "Collegamenti", "Connessione TCP aperta verso " + fields[2] + ":" + fields[3]}, true
	case fields[4] != "":
		return mojoLogEvent{at, "Collegamenti", "Connessione TCP non riuscita verso " + fields[4] + ":" + fields[5]}, true
	default:
		return mojoLogEvent{at, "Collegamenti", "Timeout del collegamento dopo " + fields[6] + " secondi"}, true
	}
}

func parseMojoStartup(line string) (mojoLogEvent, bool) {
	const kind = "Avvio"
	switch {
	case strings.Contains(line, "Can't locate") && strings.Contains(line, "in @INC"):
		return mojoLogEvent{0, kind, "Modulo Perl mancante"}, true
	case strings.Contains(line, "Compilation failed") || strings.Contains(line, "BEGIN failed"):
		return mojoLogEvent{0, kind, "Errore di compilazione Perl"}, true
	case strings.Contains(line, "Permission denied"):
		return mojoLogEvent{0, kind, "Permesso negato durante l’avvio"}, true
	case strings.Contains(line, "Address already in use"):
		return mojoLogEvent{0, kind, "Porta già in uso"}, true
	default:
		return mojoLogEvent{}, false
	}
}
