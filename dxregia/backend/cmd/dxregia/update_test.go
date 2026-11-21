package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	oldRevision = "1111111111111111111111111111111111111111"
	newRevision = "2222222222222222222222222222222222222222"
)

func TestUpdateFromPanel(t *testing.T) {
	dir := t.TempDir()
	auth := &sysopAuth{session: sessionHash("token"), expires: time.Now().Add(time.Hour), factorOK: true}
	up := &updater{
		revisionFile: filepath.Join(dir, "revision"),
		pathUnit:     filepath.Join(dir, "dxregia-update.path"),
		requestFile:  filepath.Join(dir, "update-request"),
		logFile:      filepath.Join(dir, "update.log"),
		remote: func(context.Context) (string, map[string]string, error) {
			return newRevision, map[string]string{newRevision: "v9.9.9", oldRevision: "v9.9.0"}, nil
		},
		running: func() bool { return false },
	}
	if err := os.WriteFile(up.revisionFile, []byte(oldRevision+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	up.register(mux, auth)
	call := func(method, path, body string, signedIn bool) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if method == http.MethodPost {
			r.Header.Set("Content-Type", "application/json")
		}
		if signedIn {
			r.AddCookie(&http.Cookie{Name: sysopCookie, Value: "token"})
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	status := func() updateInfo {
		t.Helper()
		w := call(http.MethodGet, "/api/update?check=1", "", true)
		if w.Code != http.StatusOK {
			t.Fatalf("status gave %d", w.Code)
		}
		var info updateInfo
		if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
			t.Fatal(err)
		}
		return info
	}
	good := `{}`

	if call(http.MethodGet, "/api/update", "", false).Code != http.StatusUnauthorized {
		t.Fatal("status readable without sign-in")
	}
	if call(http.MethodPost, "/api/update/start", good, false).Code != http.StatusUnauthorized {
		t.Fatal("update started without sign-in")
	}
	info := status()
	if info.Installed || info.Current != oldRevision[:12] || info.Available != newRevision[:12] || info.UpToDate || info.AvailableTag != "v9.9.9" || info.CurrentTag != "v9.9.0" {
		t.Fatalf("unexpected status %+v", info)
	}
	if w := call(http.MethodPost, "/api/update/start", good, true); w.Code != http.StatusConflict {
		t.Fatalf("update started with the systemd units missing: %d", w.Code)
	}

	if err := os.WriteFile(up.pathUnit, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if w := call(http.MethodPost, "/api/update/start", `{"Url":"http://example.test"}`, true); w.Code != http.StatusBadRequest {
		t.Fatalf("unknown field accepted: %d", w.Code)
	}
	if _, err := os.Stat(up.requestFile); err == nil {
		t.Fatal("request written by an invalid call")
	}
	if w := call(http.MethodPost, "/api/update/start", good, true); w.Code != http.StatusAccepted {
		t.Fatalf("update refused: %d %s", w.Code, w.Body)
	}
	if data, err := os.ReadFile(up.requestFile); err != nil || len(data) != 0 {
		t.Fatal("request file missing or carrying data")
	}
	if state := status().State; state != "queued" {
		t.Fatalf("state %q instead of queued", state)
	}
	if w := call(http.MethodPost, "/api/update/start", good, true); w.Code != http.StatusConflict {
		t.Fatalf("second update accepted while one is queued: %d", w.Code)
	}

	os.Remove(up.requestFile)
	up.running = func() bool { return true }
	if state := status().State; state != "running" {
		t.Fatalf("state %q instead of running", state)
	}
	up.running = func() bool { return false }
	log := "\x1b[32m++++++++++\x1b[0m  Sorgente pronto.\n\x1b[32m++++++++++\x1b[0m  DXRegia aggiornato a 222222222222.\n"
	if err := os.WriteFile(up.logFile, []byte(log), 0644); err != nil {
		t.Fatal(err)
	}
	info = status()
	if info.State != "success" || len(info.Log) != 2 || strings.Contains(info.Log[0], "\x1b") {
		t.Fatalf("unexpected result %+v", info)
	}
	os.WriteFile(up.logFile, []byte("[3/5] Compilo\n++++++++++  Errore: Compilazione non riuscita.\n++++++++++  Aggiornamento interrotto. Ripristino della versione precedente.\n"), 0644)
	if state := status().State; state != "failed" {
		t.Fatalf("state %q instead of failed", state)
	}

	os.WriteFile(up.revisionFile, []byte(newRevision), 0644)
	if info := status(); !info.UpToDate {
		t.Fatal("an installation on the latest revision is not up to date")
	}
	if w := call(http.MethodPost, "/api/update/start", good, true); w.Code != http.StatusConflict {
		t.Fatalf("update accepted when already up to date: %d", w.Code)
	}
}
