package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSpotArchiveHistoryAndResume(t *testing.T) {
	root := t.TempDir()
	bundled := filepath.Join("..", "..", "..", "..", "engine", "data")
	if err := os.Mkdir(filepath.Join(root, "data"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"prefix_data.pl", "cty.dat"} {
		data, err := os.ReadFile(filepath.Join(bundled, name))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(root, "data", name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	write := func(now time.Time, text string) {
		t.Helper()
		path := spotArchivePaths(root, now)[0]
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	before := fmt.Sprintf("14040^JA1ABC^%d^CW^F1ABC^95^227^F1NODE^45^25\n", yesterday.Unix())
	write(yesterday, before)
	mux := http.NewServeMux()
	snapshot := registerSpots(mux, root)
	first := snapshot()
	if len(first.Spots) != 1 || !spotCursor.MatchString(first.Spots[0].Cursor) {
		t.Fatalf("pause cursor unavailable: %+v", first)
	}
	cursor := first.Spots[0].Cursor
	// New records include two archive bands and a CTY exact-call exception.
	today := fmt.Sprintf("21074^JA1ABC^%d^FT8 PM95AB^I1ABC^95^85^I1NODE^45^25^28^15^^^192.0.2.24\n", now.Unix()) +
		fmt.Sprintf("14074^DX0NE^%d^JN55^JA1ABC^269^95^JA1NODE^50^26\n", now.Unix())
	write(now, today)
	get := func(path string, status int) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != status {
			t.Fatalf("%s status %d, want %d: %s", path, rec.Code, status, rec.Body)
		}
		return rec
	}
	var recovery struct {
		Spots    []publicSpot
		Complete bool
	}
	if err := json.Unmarshal(get("/api/spot-resume?after="+cursor, 200).Body.Bytes(), &recovery); err != nil {
		t.Fatal(err)
	}
	if !recovery.Complete || len(recovery.Spots) != 2 || !recovery.Spots[0].Special {
		t.Fatalf("cross-day recovery: %+v", recovery)
	}
	var catalog []dxccEntry
	if err := json.Unmarshal(get("/api/dxcc", 200).Body.Bytes(), &catalog); err != nil || len(catalog) < 300 {
		t.Fatalf("catalog has %d entries: %v", len(catalog), err)
	}
	var history struct {
		Spots []publicSpot
		Bands map[string]int
	}
	response := get("/api/spot-history?dxcc=95", 200)
	if strings.Contains(response.Body.String(), "192.0.2.24") {
		t.Fatal("archive included private spotter IP")
	}
	if err := json.Unmarshal(response.Body.Bytes(), &history); err != nil {
		t.Fatal(err)
	}
	if len(history.Spots) != 2 || history.Bands["20m"] != 1 || history.Bands["15m"] != 1 {
		t.Fatalf("inconsistent DXCC history: %+v", history)
	}
	row := history.Spots[0]
	if row.CQ != 25 || row.ITU != 45 || row.Continent != "AS" || row.SpotterContinent != "EU" || len(row.Grids) != 1 || row.Grids[0] != "PM95AB" {
		t.Fatalf("unexpected geography: %+v", row)
	}
	get("/api/spot-history?call=JA1ABC&dxcc=95", 400)
	get("/api/spot-resume?after=../../../example", 400)
}
