package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecordFirstConnections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partners.json")
	partners := map[string]partner{"Q1QQQ-2": {}, "Q3QQQ-2": {}}
	now := time.Now().Unix()
	recordFirstConnections(path, partners, presence{At: now, Channels: []presenceChannel{
		{Call: "Q1QQQ-2", Kind: "node", State: "normal", Since: now - 30},
		{Call: "Q3QQQ-2", Kind: "node", State: "init", Since: now - 1},
	}})
	recordFirstConnections(path, partners, presence{At: now + 3, Channels: []presenceChannel{
		{Call: "Q1QQQ-2", Kind: "node", State: "normal", Since: now + 1},
	}})
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]partner
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	first := saved["Q1QQQ-2"].FirstConnectedAt
	if first == nil || first.Unix() != now-30 || saved["Q3QQQ-2"].FirstConnectedAt != nil {
		t.Fatalf("unexpected first connections: %+v", saved)
	}
}
