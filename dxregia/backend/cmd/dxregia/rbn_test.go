package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRBNFeedAuthorizationAndVisibility(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"local", "local_data"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	auth := &sysopAuth{session: sessionHash("token"), expires: now.Add(time.Hour), factorOK: true}
	mux := http.NewServeMux()
	registerRBN(mux, root, auth)
	request := func(method, body string, signedIn bool, status int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "/api/rbn", strings.NewReader(body))
		if method == http.MethodPut {
			r.Header.Set("Content-Type", "application/json")
		}
		if signedIn {
			r.AddCookie(&http.Cookie{Name: sysopCookie, Value: "token"})
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s: %d %s", method, body, w.Code, w.Body)
		}
		return w
	}
	decode := func(w *httptest.ResponseRecorder) (bool, rbnSnapshot) {
		t.Helper()
		var feed struct {
			Enabled bool
			rbnSnapshot
		}
		if err := json.Unmarshal(w.Body.Bytes(), &feed); err != nil {
			t.Fatal(err)
		}
		return feed.Enabled, feed.rbnSnapshot
	}

	request(http.MethodPut, `{"Enabled":true}`, false, http.StatusUnauthorized)
	request(http.MethodPut, `{"Enabled":null}`, true, http.StatusBadRequest)
	request(http.MethodPut, `{"Enabled":true,"Host":"other"}`, true, http.StatusBadRequest)
	if enabled, feed := decode(request(http.MethodGet, "", false, http.StatusOK)); enabled || feed.State != "disabled" {
		t.Fatal("the feed was enabled by default")
	}
	if enabled, feed := decode(request(http.MethodPut, `{"Enabled":true}`, true, http.StatusOK)); !enabled || feed.State != "waiting" {
		t.Fatal("the feed did not start waiting for the engine")
	}
	config, err := readRBNConfig(filepath.Join(root, "local", "dxregia-rbn.json"))
	if err != nil || !config.Enabled || len(config.InboundHash) != 60 {
		t.Fatal("activation did not persist the protected local account settings")
	}
	path := filepath.Join(root, "local_data", "dxregia-rbn.json")
	snapshot := rbnSnapshot{At: now.Unix(), State: "connected", Spots: []rbnSpot{
		{Call: "I1ZZZ", Skimmer: "DL1ZZZ", Mode: "CW", At: now.Unix(), Frequency: 14020.5, SNR: 18, Quality: 3},
		{Call: "I2ZZZ", Mode: "RTTY", At: now.Add(-time.Hour).Unix()},
		{Call: "I3ZZZ", Mode: "FT8", At: now.Unix()},
	}}
	if err := writePrivateJSON(path, snapshot); err != nil {
		t.Fatal(err)
	}
	public := request(http.MethodGet, "", false, http.StatusOK)
	if strings.Contains(public.Body.String(), config.InboundHash) {
		t.Fatal("the public API exposed the local password hash")
	}
	if enabled, feed := decode(public); !enabled || feed.State != "connected" || len(feed.Spots) != 1 || feed.Spots[0] != snapshot.Spots[0] {
		t.Fatalf("unexpected filtered feed: %+v", feed)
	}
	request(http.MethodPut, `{"Enabled":false}`, true, http.StatusOK)
	if enabled, feed := decode(request(http.MethodGet, "", false, http.StatusOK)); enabled || feed.State != "disabled" || len(feed.Spots) != 0 {
		t.Fatal("the feed leaked spots after being disabled")
	}
}

func TestRBNHookCapturesCuratedCandidates(t *testing.T) {
	perl, err := exec.LookPath("perl")
	if err != nil {
		t.Skip("Perl is unavailable")
	}
	root := t.TempDir()
	for _, dir := range []string{"local", "local_data"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := saveRBNConfig(filepath.Join(root, "local", "dxregia-rbn.json"), rbnConfig{Enabled: true}); err != nil {
		t.Fatal(err)
	}
	hook := filepath.Join(root, "Local.pm")
	if err := os.WriteFile(hook, presenceHook, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(perl, filepath.Join("testdata", "rbn_hook.pl"), root, hook)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("RBN hook: %v\n%s", err, output)
	}
}
