package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReconnectQueuesOnlyDueMatchingPartner(t *testing.T) {
	root := t.TempDir()
	now := time.Now().UTC()
	entry := partner{
		partnerDetails: partnerDetails{Call: "Q3QQQ-2", Host: "node.example.org", Port: 7373},
		Provisioning:   "Abilitato in Mojo",
	}
	partners := map[string]partner{entry.Call: entry}
	plan := reconnectPlan{Host: entry.Host, Port: entry.Port, Password: "testPassword", LastAttempt: now.Add(-time.Minute)}
	if err := saveReconnect(root, entry.Call, plan); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "local_data", "dxregia-partners", entry.Call+".connect")
	snapshot := presence{PartnerConnect: true, At: now.Unix()}
	absent := func() {
		t.Helper()
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("unexpected queued connection")
		}
	}
	reconnectPartners(root, partners, snapshot, now)
	absent()
	later := now.Add(2 * time.Minute)
	changed := entry
	changed.Port = 7400
	partners[entry.Call] = changed
	reconnectPartners(root, partners, snapshot, later)
	absent()
	partners[entry.Call] = entry
	snapshot.Channels = []presenceChannel{{Call: entry.Call, Kind: "node", State: "normal"}}
	reconnectPartners(root, partners, snapshot, later)
	absent()
	snapshot.Channels = nil
	reconnectPartners(root, partners, snapshot, later)
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "node.example.org\n7373\ntestPassword" {
		t.Fatal("expected queue entry for the registered endpoint")
	}
	saved, err := readReconnect(root, entry.Call)
	if err != nil || !saved.LastAttempt.Equal(later) {
		t.Fatal("retry timestamp not persisted")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	reconnectPartners(root, partners, snapshot, later.Add(time.Second))
	absent()
}
