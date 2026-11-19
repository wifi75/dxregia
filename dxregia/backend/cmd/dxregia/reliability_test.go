package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func runConcurrentReaders(t *testing.T, viewers, reads int, read func() error) {
	t.Helper()
	var workers sync.WaitGroup
	start := make(chan struct{})
	for range viewers {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			for range reads {
				if err := read(); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	close(start)
	workers.Wait()
}

func TestWebSpotsConcurrentReadersAndRevalidation(t *testing.T) {
	root := t.TempDir()
	now := time.Now().UTC()
	archive := spotArchivePaths(root, now)[0]
	if err := os.MkdirAll(filepath.Dir(archive), 0700); err != nil {
		t.Fatal(err)
	}
	writeRows := func(count int) {
		t.Helper()
		var data strings.Builder
		for i := range count {
			fmt.Fprintf(&data, "14074^K1ABC^%d^FT8^I1ABC^226\n", now.Unix()+int64(i))
		}
		if err := os.WriteFile(archive, []byte(data.String()), 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeRows(100)
	mux := http.NewServeMux()
	snapshot := registerSpots(mux, root, nil)
	server := httptest.NewServer(mux)
	defer server.Close()
	client := &http.Client{Timeout: 10 * time.Second}
	first := snapshot()
	runConcurrentReaders(t, 32, 6, func() error {
		response, err := client.Get(server.URL + "/api/spots")
		if err != nil {
			return err
		}
		defer response.Body.Close()
		var result spotSnapshot
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			return err
		}
		if response.StatusCode != 200 || !result.Available || len(result.Spots) != 100 ||
			result.Sequence != first.Sequence || result.LatestAt != now.Unix()+99 {
			return fmt.Errorf("inconsistent spot snapshot: status %d, version %d, count %d, latest %d",
				response.StatusCode, result.Sequence, len(result.Spots), result.LatestAt)
		}
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/api/spots", nil)
	initial := httptest.NewRecorder()
	mux.ServeHTTP(initial, req)
	tag := initial.Header().Get("ETag")
	if len(tag) < 2 || tag[0] != '"' || tag[len(tag)-1] != '"' {
		t.Fatalf("invalid quoted ETag: %q", tag)
	}
	req.Header.Set("If-None-Match", tag)
	unchanged := httptest.NewRecorder()
	mux.ServeHTTP(unchanged, req)
	if unchanged.Code != http.StatusNotModified || unchanged.Body.Len() != 0 {
		t.Fatalf("unchanged response returned %d with %d bytes", unchanged.Code, unchanged.Body.Len())
	}
	writeRows(101)
	changed := httptest.NewRecorder()
	mux.ServeHTTP(changed, req)
	latest := snapshot()
	var changedBody spotSnapshot
	if err := json.NewDecoder(changed.Body).Decode(&changedBody); err != nil {
		t.Fatal(err)
	}
	if changedBody.Sequence != latest.Sequence || changedBody.LatestAt != latest.LatestAt {
		t.Fatalf("stale JSON after archive update: version %d, latest %d", changedBody.Sequence, changedBody.LatestAt)
	}
	if changed.Code != http.StatusOK || changed.Header().Get("ETag") == initial.Header().Get("ETag") ||
		latest.Sequence <= first.Sequence || latest.LatestAt != now.Unix()+100 {
		t.Fatalf("updated archive did not invalidate cache, status %d, sequence %d, latest %d",
			changed.Code, latest.Sequence, latest.LatestAt)
	}
}

func TestPartnerDirectoryConcurrentReads(t *testing.T) {
	root := t.TempDir()
	const count = 250
	partners := make(map[string]partner, count)
	for i := range count {
		key := fmt.Sprintf("K1A%c%c-7", 'A'+i/26, 'A'+i%26)
		partners[key] = partner{partnerDetails: partnerDetails{Call: key}}
	}
	if err := savePartners(filepath.Join(root, ".dxregia-partners.json"), partners); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mux := http.NewServeMux()
	auth := &sysopAuth{session: sessionHash("test"), expires: time.Now().Add(time.Hour), factorOK: true}
	if err := registerPartners(ctx, mux, root, auth); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(mux)
	defer server.Close()
	client := &http.Client{Timeout: 10 * time.Second}
	runConcurrentReaders(t, 12, 4, func() error {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/api/partners", nil)
		if err != nil {
			return err
		}
		req.AddCookie(&http.Cookie{Name: sysopCookie, Value: "test"})
		response, err := client.Do(req)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		var result map[string]partner
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			return err
		}
		if response.StatusCode != 200 || len(result) != count {
			return fmt.Errorf("partner directory: status %d, entries %d", response.StatusCode, len(result))
		}
		return nil
	})
}

func TestEventRetentionOnlyPrunesRecognizedOldArchives(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 11, 19, 19, 0, 0, 0, time.UTC)
	old := now.AddDate(0, -4, 0)
	files := map[string]bool{
		"debug/" + old.Format("2006/002.dat"): true,
		"log/" + old.Format("2006/01.dat"):    true,
		"dxregia-partner-mail/" + strings.Repeat("1", 20) + "-" + strings.Repeat("a", 48) + ".json": true,
		"debug/" + now.Format("2006/002.dat"):                                                       false,
		"log/" + now.Format("2006/01.dat"):                                                          false,
		"dxregia-partner-mail/notes.json":                                                           false,
		"dxregia-startup.log":                                                                       false,
	}
	for name := range files {
		file := filepath.Join(root, "local_data", name)
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("test"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(file, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if bytes, ok := scanEventStorage(root, now, 0); !ok || bytes != 28 {
		t.Fatalf("initial storage bytes %d, ok %t", bytes, ok)
	}
	if bytes, ok := scanEventStorage(root, now, 30); !ok || bytes != 16 {
		t.Fatalf("remaining storage bytes %d, ok %t", bytes, ok)
	}
	for name, shouldRemove := range files {
		_, err := os.Stat(filepath.Join(root, "local_data", name))
		if shouldRemove && !os.IsNotExist(err) || !shouldRemove && err != nil {
			t.Errorf("unexpected retention result for %s: %v", name, err)
		}
	}
}

func TestPublicStatusOptInAndPrivateDiagnostics(t *testing.T) {
	root := t.TempDir()
	when := time.Now().Unix()
	data := filepath.Join(root, "local_data")
	if err := os.MkdirAll(data, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(data, "dxregia-presence.json")
	if err := os.WriteFile(path, fmt.Appendf(nil, "{\"At\":%d,\"Channels\":[{\"Call\":\"PRIVATE\"}]}", when), 0600); err != nil {
		t.Fatal(err)
	}
	spots := func() spotSnapshot { return spotSnapshot{LatestAt: when} }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	auth := &sysopAuth{session: sessionHash("test"), expires: time.Now().Add(time.Hour), factorOK: true}
	mux := http.NewServeMux()
	t.Setenv("DXREGIA_PUBLIC_STATUS", "")
	registerNodeHealth(ctx, mux, root, auth, spots)
	request := func(mux *http.ServeMux, path string, signed bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if signed {
			req.AddCookie(&http.Cookie{Name: sysopCookie, Value: "test"})
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}
	if status := request(mux, "/stato", false).Code; status != 404 {
		t.Fatalf("public status enabled by default, status %d", status)
	}
	if status := request(mux, "/api/node-health", false).Code; status != 401 {
		t.Fatalf("private diagnostics exposed, status %d", status)
	}
	if rec := request(mux, "/api/node-health", true); rec.Code != 200 || !strings.Contains(rec.Body.String(), "\"DXRegiaPID\":") {
		t.Fatalf("private diagnostics unavailable, status %d", rec.Code)
	}
	t.Setenv("DXREGIA_PUBLIC_STATUS", "on")
	public := http.NewServeMux()
	registerNodeHealth(ctx, public, root, nil, spots)
	rec := request(public, "/stato", false)
	body, err := io.ReadAll(rec.Result().Body)
	if err != nil || rec.Code != 200 || !strings.Contains(string(body), "Operativo") || strings.Contains(string(body), "PRIVATE") {
		t.Fatalf("unexpected public response, status %d, read error %v", rec.Code, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if status := request(public, "/stato", false).Code; status != 503 {
		t.Fatalf("missing presence reported as operational, status %d", status)
	}
}
