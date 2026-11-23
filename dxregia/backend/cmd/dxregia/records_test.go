package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestRecordsKeepTheHighestCountEver(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dxregia-records.json")
	store := loadRecords(path)
	day := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)

	if !store.observe(31, 4, day) {
		t.Fatal("the first observation must set the records")
	}
	if store.observe(20, 4, day.Add(time.Hour)) {
		t.Fatal("a lower or equal count must not change the records")
	}
	if got := store.snapshot(); got.Telnet.Count != 31 || got.Telnet.At != day.Unix() || got.Web.Count != 4 {
		t.Fatalf("records after a quieter hour: %+v", got)
	}
	later := day.AddDate(0, 0, 40)
	store.observe(50, 3, later)

	reloaded := loadRecords(path)
	got := reloaded.snapshot()
	if got.Telnet.Count != 50 || got.Telnet.At != later.Unix() {
		t.Fatalf("telnet record not kept across restarts or days: %+v", got.Telnet)
	}
	if got.Web.Count != 4 || got.Web.At != day.Unix() {
		t.Fatalf("web record must keep its own date: %+v", got.Web)
	}
}

func TestRecordsIgnoreUnreadableFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dxregia-records.json")
	if err := saveRecords(path, recordsData{Telnet: recordEntry{Count: 7, At: 1}}); err != nil {
		t.Fatal(err)
	}
	if got := loadRecords(path).snapshot(); got.Telnet.Count != 7 {
		t.Fatalf("saved record not read back: %+v", got)
	}
	if got := loadRecords(filepath.Join(t.TempDir(), "missing.json")).snapshot(); got != (recordsData{}) {
		t.Fatalf("a missing file must start from zero: %+v", got)
	}
}

func TestCountSessionsCountsOnlyUsers(t *testing.T) {
	telnet, web := countSessions(presence{Channels: []presenceChannel{
		{Call: "IU3WJQ", Kind: "telnet"}, {Call: "K1ABC", Kind: "telnet"}, {Call: "DL1XYZ", Kind: "web"},
		{Call: "IW3SSD-3", Kind: "node"}, {Call: "EA3CV-2", Kind: "node"},
	}})
	if telnet != 2 || web != 1 {
		t.Fatalf("telnet %d web %d (nodes must not be counted)", telnet, web)
	}
}
