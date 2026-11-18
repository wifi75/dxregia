package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Storico degli accessi: chi si è collegato al nodo (utenti Telnet/Web e nodi in ingresso), da dove e per quanto tempo.
// Si ricava dal log di DXSpider (local_data/log/AAAA/MM.dat), che registra ogni collegamento e ogni disconnessione.

type accessSession struct {
	Call    string
	Kind    string // user | node
	Host    string
	Start   int64
	End     int64 `json:",omitempty"`
	Seconds int64 `json:",omitempty"`
	Open    bool  `json:",omitempty"`
}

type accessCaller struct {
	Call  string
	Kind  string
	Count int
	Last  int64
	Host  string
}

var (
	accessConnect    = regexp.MustCompile(`^([A-Z0-9/-]{3,16}) connected from (\S+)`)
	accessDisconnect = regexp.MustCompile(`^([A-Z0-9/-]{3,16}) [Dd]isconnected\b`)
)

// parseAccessLine riconosce le righe di collegamento e disconnessione del log di DXSpider.
func parseAccessLine(line string) (at int64, kind, call, host string, connect, ok bool) {
	parts := strings.SplitN(line, "^", 3)
	if len(parts) != 3 {
		return
	}
	switch parts[1] {
	case "DXCommand":
		kind = "user"
	case "DXProt":
		kind = "node"
	default:
		return
	}
	value, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return
	}
	if m := accessConnect.FindStringSubmatch(parts[2]); m != nil {
		return value, kind, m[1], m[2], true, true
	}
	if m := accessDisconnect.FindStringSubmatch(parts[2]); m != nil {
		return value, kind, m[1], "", false, true
	}
	return
}

// buildAccess abbina ogni collegamento alla disconnessione successiva dello stesso nominativo.
func buildAccess(lines []string, since int64, live map[string]bool) ([]accessSession, []accessCaller) {
	var sessions []accessSession
	openAt := map[string]int{}
	for _, line := range lines {
		at, kind, call, host, connect, ok := parseAccessLine(line)
		if !ok || at < since {
			continue
		}
		key := kind + ":" + call
		if connect {
			sessions = append(sessions, accessSession{Call: call, Kind: kind, Host: host, Start: at})
			openAt[key] = len(sessions) - 1
			continue
		}
		if i, found := openAt[key]; found {
			sessions[i].End = at
			sessions[i].Seconds = max(0, at-sessions[i].Start)
			delete(openAt, key)
		}
	}
	for key, i := range openAt {
		if live[key] {
			sessions[i].Open = true
		}
	}
	callers := map[string]*accessCaller{}
	for _, session := range sessions {
		key := session.Kind + ":" + session.Call
		entry := callers[key]
		if entry == nil {
			entry = &accessCaller{Call: session.Call, Kind: session.Kind}
			callers[key] = entry
		}
		entry.Count++
		if session.Start >= entry.Last {
			entry.Last, entry.Host = session.Start, session.Host
		}
	}
	summary := make([]accessCaller, 0, len(callers))
	for _, entry := range callers {
		summary = append(summary, *entry)
	}
	sort.Slice(summary, func(i, j int) bool { return summary[i].Last > summary[j].Last })
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].Start > sessions[j].Start })
	return sessions, summary
}

func readAccessLines(root string, days int, now time.Time) []string {
	var lines []string
	seen := map[string]bool{}
	for back := days; back >= 0; back-- {
		date := now.UTC().AddDate(0, 0, -back)
		path := filepath.Join(root, "local_data", "log", fmt.Sprintf("%04d", date.Year()), fmt.Sprintf("%02d.dat", int(date.Month())))
		if seen[path] {
			continue
		}
		seen[path] = true
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 64*1024), 1<<20)
		for scanner.Scan() {
			text := scanner.Text()
			if strings.Contains(text, "onnected") && (strings.Contains(text, "^DXCommand^") || strings.Contains(text, "^DXProt^")) {
				lines = append(lines, text)
			}
		}
		file.Close()
	}
	return lines
}

func registerConnections(mux *http.ServeMux, root string, auth *sysopAuth) {
	if root == "" || auth == nil {
		return
	}
	mux.HandleFunc("GET /api/connections", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		days, err := strconv.Atoi(r.URL.Query().Get("days"))
		if err != nil || (days != 1 && days != 7 && days != 30) {
			days = 7
		}
		now := time.Now()
		snapshot := readPresence(filepath.Join(root, "local_data", "dxregia-presence.json"))
		live := map[string]bool{}
		for _, channel := range snapshot.Channels {
			kind := "user"
			if channel.Kind == "node" {
				kind = "node"
			}
			live[kind+":"+channel.Call] = true
		}
		sessions, callers := buildAccess(readAccessLines(root, days, now), now.Add(-time.Duration(days)*24*time.Hour).Unix(), live)
		if len(sessions) > 500 {
			sessions = sessions[:500]
		}
		if len(callers) > 200 {
			callers = callers[:200]
		}
		if sessions == nil {
			sessions = []accessSession{}
		}
		if callers == nil {
			callers = []accessCaller{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"Days": days, "Sessions": sessions, "Callers": callers})
	})
}
