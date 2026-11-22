package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBackupsFromPanel(t *testing.T) {
	root := filepath.Join(t.TempDir(), "mojo")
	if err := os.MkdirAll(filepath.Join(root, "local"), 0700); err != nil {
		t.Fatal(err)
	}
	b := newBackupManager(root, "unused")
	calls := 0
	b.run = func(ctx context.Context, output string) error {
		calls++
		return os.WriteFile(output, []byte("zip"), 0600)
	}
	auth := &sysopAuth{session: sessionHash("token"), expires: time.Now().Add(time.Hour), factorOK: true}
	mux := http.NewServeMux()
	b.routes(mux, auth)
	call := func(method, path, body string, signedIn bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if method != http.MethodGet && method != http.MethodDelete {
			r.Header.Set("Content-Type", "application/json")
		}
		if signedIn {
			r.AddCookie(&http.Cookie{Name: sysopCookie, Value: "token"})
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	if call(http.MethodPut, "/api/backups/schedule", `{"Enabled":true,"Hour":3,"Keep":7}`, false).Code != http.StatusUnauthorized {
		t.Fatal("schedule changed without sign-in")
	}
	if call(http.MethodPut, "/api/backups/schedule", `{"Enabled":true,"Hour":24,"Keep":7}`, true).Code != http.StatusBadRequest {
		t.Fatal("an invalid hour was accepted")
	}
	if call(http.MethodPut, "/api/backups/schedule", `{"Enabled":true,"Hour":4,"Keep":2}`, true).Code != http.StatusNoContent {
		t.Fatal("schedule refused")
	}
	if s := b.schedule(); !s.Enabled || s.Hour != 4 || s.Keep != 2 {
		t.Fatalf("unexpected schedule %+v", s)
	}
	times := []time.Time{
		time.Date(2026, 11, 20, 4, 0, 0, 0, time.UTC),
		time.Date(2026, 11, 21, 4, 0, 0, 0, time.UTC),
		time.Date(2026, 11, 22, 4, 0, 0, 0, time.UTC),
	}
	for _, at := range times {
		if err := b.create(context.Background(), at); err != nil {
			t.Fatal(err)
		}
	}
	files := b.list()
	if len(files) != 2 || files[0].Name != "dxregia-backup-20261122-040000.zip" || files[1].Name != "dxregia-backup-20261121-040000.zip" {
		t.Fatalf("retention kept %+v", files)
	}
	if b.due(time.Date(2026, 11, 22, 3, 59, 0, 0, time.UTC)) || b.due(time.Date(2026, 11, 22, 9, 0, 0, 0, time.UTC)) {
		t.Fatal("a backup was due before the hour or twice in a day")
	}
	if !b.due(time.Date(2026, 11, 23, 4, 0, 0, 0, time.UTC)) {
		t.Fatal("no backup due on the next day")
	}
	if call(http.MethodGet, "/api/backups/files/dxregia-backup-20261122-040000.zip", "", false).Code != http.StatusUnauthorized {
		t.Fatal("download without sign-in")
	}
	if w := call(http.MethodGet, "/api/backups/files/dxregia-backup-20261122-040000.zip", "", true); w.Code != http.StatusOK || w.Body.String() != "zip" {
		t.Fatalf("download failed: %d", w.Code)
	}
	if call(http.MethodGet, "/api/backups/files/..%2F..%2Fpasswd", "", true).Code != http.StatusNotFound {
		t.Fatal("a path outside the backups was reachable")
	}
	if call(http.MethodDelete, "/api/backups/files/dxregia-backup-20261121-040000.zip", "", true).Code != http.StatusNoContent || len(b.list()) != 1 {
		t.Fatal("delete failed")
	}
	if w := call(http.MethodPost, "/api/backups/run", `{"Url":"x"}`, true); w.Code != http.StatusBadRequest {
		t.Fatalf("unknown field accepted: %d", w.Code)
	}
	b.run = func(ctx context.Context, output string) error { return os.ErrPermission }
	if err := b.create(context.Background(), times[2].Add(24*time.Hour)); err == nil {
		t.Fatal("a failed backup reported success")
	}
	if _, err := os.Stat(filepath.Join(b.dir, "dxregia-backup-20261123-040000.zip")); err == nil {
		t.Fatal("a partial archive was kept")
	}
	if b.due(times[2].Add(24*time.Hour + 5*time.Minute)) {
		t.Fatal("no wait after a failed backup")
	}
	if !b.due(times[2].Add(24*time.Hour + 40*time.Minute)) {
		t.Fatal("no retry after the wait")
	}
}
