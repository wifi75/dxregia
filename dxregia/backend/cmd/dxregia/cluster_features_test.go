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

func TestSpotFilterEditor(t *testing.T) {
	root := t.TempDir()
	bundled := filepath.Join("..", "..", "..", "..", "engine", "data")
	for _, name := range []string{"bands.pl", "prefix_data.pl", "cty.dat"} {
		content, err := os.ReadFile(filepath.Join(bundled, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := writeNodeFile(filepath.Join(root, "data", name), content); err != nil {
			t.Fatal(err)
		}
	}
	filter := filepath.Join(root, "filter", "spots", "in_Q1LAB.pl")
	original := `{"sort":"spots","name":"in_Q1LAB.pl","filter0":{"reject":{"user":"call K1","asc":"0"}}}`
	if err := writeNodeFile(filter, []byte(original)); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	registerNodeLists(mux, root, &sysopAuth{session: sessionHash("test-session"), expires: time.Now().Add(time.Minute), factorOK: true})
	request := func(method, path, body string, authorized bool, status int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if authorized {
			r.AddCookie(&http.Cookie{Name: sysopCookie, Value: "test-session"})
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s: %d, want %d: %s", method, path, w.Code, status, w.Body)
		}
		return w
	}
	route := "/api/filters/spots/in_Q1LAB.pl/"
	request("PUT", route+"1", "{}", false, 401)
	request("PUT", route+"1", `{"Action":"accept","Call":"X' or 1"}`, true, 400)
	request("PUT", route+"0", `{"Action":"reject","Call":"JA1ABC"}`, true, 409)
	request("PUT", route+"1", `{"Action":"accept","Band":"20m","Mode":"ft8","Continent":"AS","Call":"ja1abc"}`, true, 204)
	marker := filepath.Join(filepath.Dir(filter), ".dxregia-reload")
	before, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	request("PUT", route+"1", `{"Action":"reject","Band":"20m","Mode":"ft8","Call":"JA1ABC","PreviousAction":"accept"}`, true, 204)
	after, err := os.ReadFile(marker)
	if err != nil || len(after) == 0 || string(before) == string(after) {
		t.Fatalf("edit did not signal the running engine: %v", err)
	}
	data, err := os.ReadFile(filter)
	if err != nil {
		t.Fatal(err)
	}
	rows := readSpotFilterSlots(data)
	if len(rows) != 2 || rows[0].Editor != nil || rows[0].Text != "call K1" ||
		rows[1].Editor == nil || rows[1].Action != "reject" || rows[1].Number != 1 ||
		!strings.Contains(string(data), "14074") || !strings.Contains(string(data), "14077") {
		t.Fatalf("unexpected native rules: %s", data)
	}
	changedByConsole := strings.Replace(string(data), "call JA1ABC", "call K2", 1)
	if rows := readSpotFilterSlots([]byte(changedByConsole)); rows[1].Editor != nil {
		t.Fatal("console modification was still marked editable")
	}
	request("DELETE", route+"1?action=reject", "", true, 204)
	data, err = os.ReadFile(filter)
	if err != nil {
		t.Fatal(err)
	}
	if rows := readSpotFilterSlots(data); len(rows) != 1 || rows[0].Text != "call K1" {
		t.Fatalf("console rule lost: %s", data)
	}
}

func TestPublicBulletinArchiveExcludesPrivateMessages(t *testing.T) {
	now := time.Now().UTC()
	root := t.TempDir()
	path := filepath.Join(root, "local_data", "log", fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d.dat", now.Month()))
	log := `1800000000^ann^*^Q1LAB^Public^with caret
1800000000^ann^*^Q1LAB^Public^with caret
1800000001^ann^SYSOP^Q1LAB^Secret admins
1800000001^ann^Q2LAB^Q1LAB^Secret recipient
1800000001^talk^ALL^Q1LAB^Secret talk
1800000001^ann^*^Q1LAB^[to Q1LAB only] Rejected
1800000001^ann^ALL^Q1LAB^Incomplete`
	if err := writeNodeFile(path, []byte(log)); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	registerBulletins(mux, root)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/cluster-bulletins/ann", nil))
	var rows []clusterBulletin
	if err := json.Unmarshal(w.Body.Bytes(), &rows); err != nil || len(rows) != 1 || rows[0].Text != "Public^with caret" {
		t.Fatalf("unexpected public announcements: %s", w.Body)
	}
	wwv, ok := parseBulletin("wwv", "Q1LAB^1800000000^147^6^2^quiet^Q2LAB^71")
	if !ok || wwv.SFI != "147" || wwv.K != "2" || wwv.Text != "quiet" || wwv.Sunspots != "71" {
		t.Fatalf("bad WWV: %+v", wwv)
	}
	wcy, ok := parseBulletin("wcy", "1800000000^147^6^2^3^71^low^quiet^yes^Q1LAB^Q2LAB")
	if !ok || wcy.ExpectedK != "3" || wcy.Geomagnetic != "quiet" || wcy.Aurora != "yes" {
		t.Fatalf("bad WCY: %+v", wcy)
	}
}

func TestSpotQSLDirectoryInvalidatesFeed(t *testing.T) {
	root := t.TempDir()
	now := time.Now().UTC()
	path := spotArchivePaths(root, now)[0]
	if err := writeNodeFile(path, []byte(fmt.Sprintf("14074^JA1ABC^%d^FT8^Q1LAB^95\n", now.Unix()))); err != nil {
		t.Fatal(err)
	}
	directory := &qslDirectory{}
	mux := http.NewServeMux()
	snapshot := registerSpots(mux, root, directory)
	before := snapshot()
	if len(before.Spots) != 1 || before.Spots[0].LoTW || before.Spots[0].EQSL {
		t.Fatalf("unexpected initial spots: %+v", before)
	}
	lotw, err := readQSLDirectory([]byte("JA1ABC,2026-07-27,15:21:03\n"), true)
	if err != nil {
		t.Fatal(err)
	}
	eqsl, err := readQSLDirectory([]byte("List of AG members as of sample\r\nja1abc\r\nJA2ABC/P\r\n"), false)
	if err != nil {
		t.Fatal(err)
	}
	directory.eqsl.Store(&eqsl)
	directory.lotw.Store(&lotw)
	after := snapshot()
	if after.Sequence <= before.Sequence || !after.LoTWAvailable || !after.EQSLAvailable ||
		!after.Spots[0].LoTW || !after.Spots[0].EQSL {
		t.Fatalf("missing QSL enrichment: %+v", after)
	}
	portable := publicSpot{Call: "JA2ABC"}
	directory.snapshot().enrich(&portable)
	if portable.EQSL {
		t.Fatal("portable membership attributed to a different callsign")
	}
}
