package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var consoleCommand = regexp.MustCompile(`^(?:show/(?:cluster|users|uptime|version|links)|show/dx(?: [1-9][0-9]?)?|show/(?:prefix|qrz) [A-Z0-9/-]{1,16})$`)

func registerTelnetConsole(mux *http.ServeMux, auth *sysopAuth, target func() string) {
	if auth == nil {
		return
	}
	registerLiveTelnetConsole(mux, auth, target)
	busy := make(chan struct{}, 1)
	mux.HandleFunc("POST /api/console", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return
		}
		var input struct{ Call, Command, Password string }
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return
		}
		call := strings.ToUpper(strings.TrimSpace(input.Call))
		command := strings.Fields(input.Command)
		if len(command) == 0 || len(call) > 12 || !callsign.MatchString(call) {
			http.Error(w, "Inserisci un nominativo valido e un comando.", http.StatusBadRequest)
			return
		}
		command[0] = strings.ToLower(command[0])
		for i := 1; i < len(command); i++ {
			command[i] = strings.ToUpper(command[i])
		}
		line := strings.Join(command, " ")
		if !consoleCommand.MatchString(line) {
			http.Error(w, "Comando non disponibile nella console web.", http.StatusBadRequest)
			return
		}
		host, port, err := net.SplitHostPort(target())
		if ip := net.ParseIP(host); err != nil || ip == nil || !ip.IsLoopback() || port == "" {
			http.Error(w, "Il listener Telnet locale non è disponibile.", http.StatusServiceUnavailable)
			return
		}
		select {
		case busy <- struct{}{}:
			defer func() { <-busy }()
		default:
			http.Error(w, "Console occupata. Riprova tra poco.", http.StatusTooManyRequests)
			return
		}
		dialer := net.Dialer{Timeout: 2 * time.Second}
		conn, err := dialer.DialContext(r.Context(), "tcp", net.JoinHostPort(host, port))
		if err != nil {
			http.Error(w, "Il listener Telnet locale non risponde.", http.StatusServiceUnavailable)
			return
		}
		defer conn.Close()
		if err := conn.SetDeadline(time.Now().Add(8 * time.Second)); err != nil {
			http.Error(w, "Console non disponibile.", http.StatusServiceUnavailable)
			return
		}
		output, err := queryTelnet(conn, call, line, input.Password)
		if err != nil {
			http.Error(w, "DXSpider non ha restituito il prompt. Controlla nominativo, password Telnet e stato del nodo.", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"output": output})
	})
}

func queryTelnet(conn io.ReadWriter, call, command, password string) (string, error) {
	reader := bufio.NewReader(conn)
	if err := loginConsole(conn, reader, call, password); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(conn, "%s\r\n", command); err != nil {
		return "", err
	}
	output, err := readConsolePrompt(reader, "dxspider >")
	if err == nil {
		_, _ = fmt.Fprint(conn, "bye\r\n")
	}
	return strings.TrimSpace(output), err
}

func loginConsole(conn io.Writer, reader *bufio.Reader, call, password string) error {
	if _, err := readConsolePrompt(reader, "login:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(conn, "%s\r\n", call); err != nil {
		return err
	}
	welcome, err := readConsolePrompt(reader, "dxspider >", "password:")
	if err != nil {
		return err
	}
	if strings.HasSuffix(strings.ToLower(welcome), "password:") {
		if password == "" || len(password) > 72 || strings.ContainsAny(password, "\r\n\x00") {
			return errors.New("Telnet credentials required")
		}
		if _, err := fmt.Fprintf(conn, "%s\r\n", password); err != nil {
			return err
		}
		if _, err := readConsolePrompt(reader, "dxspider >"); err != nil {
			return err
		}
	}
	return nil
}

func readConsolePrompt(reader *bufio.Reader, markers ...string) (string, error) {
	var output strings.Builder
	for output.Len() < 64<<10 {
		ch, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		output.WriteByte(ch)
		for _, marker := range markers {
			if output.Len() >= len(marker) && strings.EqualFold(output.String()[output.Len()-len(marker):], marker) {
				return output.String(), nil
			}
		}
	}
	return "", errors.New("Telnet output exceeded its limit")
}
