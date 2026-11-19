package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

const consoleIdle = 45 * time.Second

var liveSpotLine = regexp.MustCompile(`(?i)^[ \t]*DX de [A-Z0-9/-]+:`)
var livePrompt = regexp.MustCompile(`(?i)^[ \t]*[A-Z0-9/-]+ de [A-Z0-9/-]+ [^\r\n]*? dxspider >[ \t]*`)

type consoleLine struct {
	Text string
	Spot bool
}

type liveConsole struct {
	conn      net.Conn
	owner, id string
	mu        sync.Mutex
	lines     []consoleLine
	closed    bool
	dropped   bool
}

func (s *liveConsole) read(reader *bufio.Reader) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 64<<10)
	for scanner.Scan() {
		line := livePrompt.ReplaceAllString(scanner.Text(), "")
		if strings.TrimSpace(line) == "" {
			continue
		}
		s.mu.Lock()
		if len(s.lines) == 240 {
			copy(s.lines, s.lines[1:])
			s.lines = s.lines[:239]
			s.dropped = true
		}
		s.lines = append(s.lines, consoleLine{Text: line, Spot: liveSpotLine.MatchString(line)})
		s.mu.Unlock()
	}
	s.conn.Close()
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
}

func registerLiveTelnetConsole(mux *http.ServeMux, auth *sysopAuth, target func() string) {
	var mu sync.Mutex
	var current *liveConsole
	handle := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		cookie, _ := r.Cookie(sysopCookie)
		owner := sessionHash(cookie.Value)
		if r.Method != http.MethodGet && r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return
		}

		id := r.Header.Get("X-Console-Session")
		mu.Lock()
		defer mu.Unlock()
		if r.Method == http.MethodPost && id == "" {
			var input struct{ Call, Password string }
			if !decodeConsoleInput(w, r, &input) {
				return
			}
			call := strings.ToUpper(strings.TrimSpace(input.Call))
			if len(call) > 12 || !callsign.MatchString(call) || len(input.Password) > 72 || strings.ContainsAny(input.Password, "\r\n\x00\xff") {
				http.Error(w, "Nominativo o password Telnet non validi.", http.StatusBadRequest)
				return
			}
			address := target()
			host, port, err := net.SplitHostPort(address)
			if ip := net.ParseIP(host); err != nil || ip == nil || !ip.IsLoopback() || port == "" {
				http.Error(w, "Il listener Telnet locale non è disponibile.", http.StatusServiceUnavailable)
				return
			}
			var token [24]byte
			if _, err := rand.Read(token[:]); err != nil {
				http.Error(w, "Console non disponibile.", http.StatusServiceUnavailable)
				return
			}
			if current != nil {
				current.conn.Close()
				current = nil
			}
			dialer := net.Dialer{Timeout: 2 * time.Second}
			conn, err := dialer.DialContext(r.Context(), "tcp", address)
			if err != nil {
				http.Error(w, "Il listener Telnet locale non risponde.", http.StatusServiceUnavailable)
				return
			}
			reader := bufio.NewReader(conn)
			if err := conn.SetDeadline(time.Now().Add(8 * time.Second)); err != nil || loginConsole(conn, reader, call, input.Password) != nil {
				conn.Close()
				http.Error(w, "DXSpider non ha accettato l’accesso. Controlla nominativo e password Telnet.", http.StatusBadGateway)
				return
			}
			conn.SetReadDeadline(time.Now().Add(consoleIdle))
			session := &liveConsole{conn: conn, owner: owner, id: hex.EncodeToString(token[:])}
			current = session
			go session.read(reader)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(struct{ ID string }{session.id})
			return
		}
		session := current
		if session == nil || session.owner != owner || session.id != id {
			http.Error(w, "Sessione Telnet terminata.", http.StatusGone)
			return
		}
		switch r.Method {
		case http.MethodGet:
			session.mu.Lock()
			output := struct {
				Lines           []consoleLine
				Closed, Dropped bool
			}{session.lines, session.closed, session.dropped}
			session.lines, session.dropped = nil, false
			session.mu.Unlock()
			if !output.Closed {
				session.conn.SetReadDeadline(time.Now().Add(consoleIdle))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(output)
		case http.MethodPost:
			var input struct{ Command string }
			if !decodeConsoleInput(w, r, &input) {
				return
			}
			fields := strings.Fields(input.Command)
			if len(fields) > 0 {
				fields[0] = strings.ToLower(fields[0])
				for i := 1; i < len(fields); i++ {
					fields[i] = strings.ToUpper(fields[i])
				}
			}
			line := strings.Join(fields, " ")
			if !consoleCommand.MatchString(line) {
				http.Error(w, "Comando non disponibile nella console web.", http.StatusBadRequest)
				return
			}
			session.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if _, err := io.WriteString(session.conn, line+"\r\n"); err != nil {
				http.Error(w, "Sessione Telnet terminata.", http.StatusGone)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			session.conn.Close()
			current = nil
			w.WriteHeader(http.StatusNoContent)
		}
	}
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		mux.HandleFunc(method+" /api/console/session", handle)
	}
}

func decodeConsoleInput(w http.ResponseWriter, r *http.Request, input any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	decoder.DisallowUnknownFields()
	if decoder.Decode(input) != nil || decoder.Decode(new(any)) != io.EOF {
		http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
		return false
	}
	return true
}
