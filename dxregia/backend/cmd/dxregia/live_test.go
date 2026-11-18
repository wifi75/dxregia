package main

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLiveStreamSignalsChangesToSignedInSysopOnly(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "local_data"), 0700); err != nil {
		t.Fatal(err)
	}
	presenceFile := filepath.Join(root, "local_data", "dxregia-presence.json")
	if err := os.WriteFile(presenceFile, []byte(`{"At":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	auth := &sysopAuth{session: "token", expires: time.Now().Add(time.Hour), factorOK: true}
	mux := http.NewServeMux()
	registerLive(mux, root, auth)
	server := httptest.NewServer(mux)
	defer server.Close()

	anonymous, err := http.Get(server.URL + "/api/live")
	if err != nil {
		t.Fatal(err)
	}
	anonymous.Body.Close()
	if anonymous.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anonymous stream allowed: %d", anonymous.StatusCode)
	}

	request, _ := http.NewRequest(http.MethodGet, server.URL+"/api/live", nil)
	request.AddCookie(&http.Cookie{Name: sysopCookie, Value: "token"})
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatalf("content type %q", response.Header.Get("Content-Type"))
	}
	lines := make(chan string, 16)
	go func() {
		scanner := bufio.NewScanner(response.Body)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()
	waitFor := func(want string) {
		t.Helper()
		deadline := time.After(5 * time.Second)
		for {
			select {
			case line, ok := <-lines:
				if !ok {
					t.Fatalf("stream closed before %q", want)
				}
				if line == want {
					return
				}
			case <-deadline:
				t.Fatalf("no %q in time", want)
			}
		}
	}
	waitFor("data: hello")
	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile(presenceFile, []byte(`{"At":2,"Channels":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	waitFor("data: change")
}
