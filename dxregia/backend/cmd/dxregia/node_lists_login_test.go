package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBadLoginLists(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "local_data")
	if err := os.MkdirAll(data, 0700); err != nil {
		t.Fatal(err)
	}
	base := "bless( {\n  name => 'badnode',\n  'OH8X-10' => 1258730778,\n  XE1FAS => 1276635800,\n}, 'DXHash' )\n"
	if err := os.WriteFile(filepath.Join(data, "badnode"), []byte(base), 0600); err != nil {
		t.Fatal(err)
	}
	auth := &sysopAuth{session: sessionHash("token"), expires: time.Now().Add(time.Hour), factorOK: true}
	mux := http.NewServeMux()
	registerNodeLists(mux, root, auth)
	call := func(method, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/api/badlists", strings.NewReader(body))
		if method == http.MethodPut {
			r.Header.Set("Content-Type", "application/json")
		}
		r.AddCookie(&http.Cookie{Name: sysopCookie, Value: "token"})
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	if w := call(http.MethodPut, `{"Words":[],"IPs":[],"DX":[],"Nodes":["kk3zz-5"],"Spotters":["IW1ABC"]}`); w.Code != http.StatusNoContent {
		t.Fatalf("save refused: %d %s", w.Code, w.Body)
	}
	local, err := os.ReadFile(filepath.Join(data, "badnode.local"))
	if err != nil || string(local) != "KK3ZZ-5\n" {
		t.Fatalf("unexpected node list file: %q %v", local, err)
	}
	if got, _ := os.ReadFile(filepath.Join(data, "badnode")); string(got) != base {
		t.Fatal("the base list was touched")
	}
	w := call(http.MethodGet, "")
	for _, want := range []string{`"Nodes":["KK3ZZ-5"]`, `"Spotters":["IW1ABC"]`, `"BaseNodes":2`, `"BaseSpotters":0`} {
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("missing %s in %s", want, w.Body)
		}
	}
	if w := call(http.MethodPut, `{"Words":[],"IPs":[],"DX":[],"Nodes":["bad call"]}`); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid callsign accepted: %d", w.Code)
	}
	// Without the new fields the lists stay as they are.
	if w := call(http.MethodPut, `{"Words":[],"IPs":[],"DX":[]}`); w.Code != http.StatusNoContent {
		t.Fatalf("old request refused: %d", w.Code)
	}
	if got, _ := os.ReadFile(filepath.Join(data, "badnode.local")); !strings.Contains(string(got), "KK3ZZ-5") {
		t.Fatal("an old request erased the node list")
	}
}
