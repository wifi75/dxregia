package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDailySummarySchedule(t *testing.T) {
	day := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	if summaryDue(summaryState{}, day.Add(5*time.Hour)) {
		t.Fatal("summary due before 06:00 UTC")
	}
	if !summaryDue(summaryState{}, day.Add(7*time.Hour)) {
		t.Fatal("summary not due after 06:00 UTC")
	}
	if summaryDue(summaryState{Date: "2026-11-22"}, day.Add(9*time.Hour)) {
		t.Fatal("summary sent twice in one day")
	}
	if !summaryDue(summaryState{Date: "2026-11-21"}, day.Add(9*time.Hour)) {
		t.Fatal("summary not due on the next day")
	}
	if summaryDue(summaryState{RetryAfter: day.Add(8 * time.Hour)}, day.Add(7*time.Hour)) {
		t.Fatal("retry delay ignored")
	}
}

func TestDailySummaryCounts(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 11, 22, 8, 0, 0, 0, time.UTC)
	debug := filepath.Join(root, "local_data", "debug", "2026")
	if err := os.MkdirAll(debug, 0700); err != nil {
		t.Fatal(err)
	}
	recent, old := now.Add(-time.Hour).Unix(), now.Add(-30*time.Hour).Unix()
	lines := strings.Join([]string{
		"%d^(*) PCPROT: PC61 192.168.0.40 is a localhost address or in RFC1918, dropped",
		"%d^(err) DXChannel DH8WR-1: too many errors (11 > 10), disconnecting",
		"%d^(*) KM3T on 203.0.113.9 is in badspotter, disconnected",
		"%d^(*) PCPROT: PC61 192.168.0.40 is a localhost address or in RFC1918, dropped",
	}, "\n")
	body := ""
	for i, line := range strings.Split(lines, "\n") {
		at := recent
		if i == 3 {
			at = old
		}
		body += strings.Replace(line, "%d", itoa(at), 1) + "\n"
	}
	if err := os.WriteFile(filepath.Join(debug, "326.dat"), []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	startup := itoa(recent) + "^DXRegia watchdog restart\n" + itoa(recent) + "^DXRegia watchdog giveup\n"
	if err := os.WriteFile(filepath.Join(root, "local_data", "dxregia-startup.log"), []byte(startup), 0600); err != nil {
		t.Fatal(err)
	}
	counts := countSummaryEvents(root, now)
	if counts.Routing != 1 || counts.Accessi != 2 || counts.Riavvii != 1 {
		t.Fatalf("unexpected counts %+v", counts)
	}
}

func itoa(value int64) string { return strconv.FormatInt(value, 10) }
