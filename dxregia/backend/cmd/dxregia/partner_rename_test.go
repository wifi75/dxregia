package main

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRenamePartner(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".dxregia-partners.json")
	partners := map[string]partner{
		"Q1QQQ-2": {partnerDetails: partnerDetails{Call: "Q1QQQ-2", Host: "a.example", Port: 7300}},
		"Q2QQQ-2": {partnerDetails: partnerDetails{Call: "Q2QQQ-2", Host: "b.example", Port: 7300}},
		"Q3QQQ-2": {partnerDetails: partnerDetails{Call: "Q3QQQ-2", Host: "c.example", Port: 7300}, Provisioning: "Abilitato in Mojo"},
	}
	if err := os.WriteFile(reconnectPath(root, "Q1QQQ-2"), []byte(`{"Host":"a.example","Port":7300}`), 0600); err != nil {
		t.Fatal(err)
	}
	snap := presence{PartnerConnect: true}
	now := time.Now()

	for _, test := range []struct {
		from, to string
		status   int
	}{
		{"Q1QQQ-2", "Q1QQQ-2", http.StatusBadRequest},
		{"Q1QQQ-2", "no good", http.StatusBadRequest},
		{"Q1QQQ-2", "Q2QQQ-2", http.StatusConflict},
		{"Q3QQQ-2", "Q4QQQ-2", http.StatusConflict},
		{"Q9QQQ-2", "Q4QQQ-2", http.StatusNotFound},
	} {
		if _, status, _ := renamePartner(root, path, partners, test.from, test.to, snap, now); status != test.status {
			t.Fatalf("%s -> %s: got %d, want %d", test.from, test.to, status, test.status)
		}
	}
	entry, status, message := renamePartner(root, path, partners, "Q1QQQ-2", "q5qqq-2", snap, now)
	if status != http.StatusOK || entry.Call != "Q5QQQ-2" {
		t.Fatalf("rename failed: %d %s %+v", status, message, entry)
	}
	if _, ok := partners["Q1QQQ-2"]; ok || partners["Q5QQQ-2"].Host != "a.example" || len(partners["Q5QQQ-2"].Notes) != 1 {
		t.Fatalf("unexpected registry: %+v", partners)
	}
	if _, err := os.Stat(reconnectPath(root, "Q5QQQ-2")); err != nil {
		t.Fatal("credentials not moved")
	}
	if _, err := os.Stat(reconnectPath(root, "Q1QQQ-2")); err == nil {
		t.Fatal("old credentials left behind")
	}
	if err := os.WriteFile(filepath.Join(dir, "Q2QQQ-2.mail.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, status, _ := renamePartner(root, path, partners, "Q2QQQ-2", "Q6QQQ-2", snap, now); status != http.StatusConflict {
		t.Fatalf("rename with connection data: %d", status)
	}
}
