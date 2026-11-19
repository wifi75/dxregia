package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLiveConsoleKeepsConnectionAndSeparatesSpots(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	result := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			result <- err
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		reader := bufio.NewReader(conn)
		exchange := func(output, expected string) error {
			if _, err := fmt.Fprint(conn, output); err != nil {
				return err
			}
			line, err := reader.ReadString('\n')
			if err == nil && line != expected+"\r\n" {
				err = fmt.Errorf("unexpected input %q, want %q", line, expected)
			}
			return err
		}
		prompt := "IU3WJQ de IU3WJQ-7 19-Nov-2026 1034Z dxspider >"
		for _, step := range []struct{ output, expected string }{
			{"login: ", "IU3WJQ"},
			{"Password: ", "test-password"},
			{prompt, "show/dx 10"},
			{"\r\n  7000.0 IZ7EUB\r\n" + prompt + "DX de N2KI: 14030.0 IU3WJQ\r\n", "show/users"},
		} {
			if err := exchange(step.output, step.expected); err != nil {
				result <- err
				return
			}
		}
		_, err = fmt.Fprint(conn, "Users: 2\r\n")
		result <- err
	}()

	auth := &sysopAuth{session: sessionHash("token"), expires: time.Now().Add(time.Hour), factorOK: true}
	mux := http.NewServeMux()
	mux.Handle("GET /", http.NotFoundHandler())
	registerTelnetConsole(mux, auth, listener.Addr().String)
	request := func(method, id, body string, authenticated bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/console/session", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Console-Session", id)
		if authenticated {
			req.AddCookie(&http.Cookie{Name: sysopCookie, Value: "token"})
		}
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		return response
	}

	response := request(http.MethodPost, "", `{"call":"iu3wjq","password":"test-password"}`, true)
	var opened struct{ ID string }
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &opened) != nil || opened.ID == "" {
		t.Fatalf("open: %d %s", response.Code, response.Body)
	}
	if response := request(http.MethodGet, opened.ID, "", false); response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated read returned %d", response.Code)
	}
	for _, command := range []string{"show/dx 10", "show/users"} {
		response := request(http.MethodPost, opened.ID, fmt.Sprintf(`{"command":%q}`, command), true)
		if response.Code != http.StatusNoContent {
			t.Fatalf("command %q: %d %s", command, response.Code, response.Body)
		}
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	var collected []consoleLine
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); {
		response := request(http.MethodGet, opened.ID, "", true)
		var packet struct {
			Lines  []consoleLine
			Closed bool
		}
		if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &packet) != nil {
			t.Fatalf("read: %d %s", response.Code, response.Body)
		}
		collected = append(collected, packet.Lines...)
		if packet.Closed {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if len(collected) != 3 || collected[0].Spot || !collected[1].Spot || collected[1].Text != "DX de N2KI: 14030.0 IU3WJQ" || collected[2].Spot {
		t.Fatalf("unexpected output: %+v", collected)
	}
}
