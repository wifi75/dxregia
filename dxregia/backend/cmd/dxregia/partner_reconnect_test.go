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
	snapshot.LockedPartners = map[string]bool{entry.Call: true}
	reconnectPartners(root, partners, snapshot, later)
	absent()
	delete(snapshot.LockedPartners, entry.Call)
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

func TestReconnectUsesLiveOutgoingReadinessAfterInboundFailure(t *testing.T) {
	root := t.TempDir()
	now := time.Now()
	entry := partner{
		partnerDetails: partnerDetails{Call: "N4ZKF-1", Host: "node.example.org", Port: 7300},
		Provisioning:   "Account Mojo esistente da verificare",
	}
	plan := reconnectPlan{Host: entry.Host, Port: entry.Port}
	if err := saveReconnect(root, entry.Call, plan); err != nil {
		t.Fatal(err)
	}
	partners := map[string]partner{entry.Call: entry}
	snapshot := presence{PartnerConnect: true, OutgoingReady: map[string]bool{}}
	reconnectPartners(root, partners, snapshot, now)
	path := filepath.Join(root, "local_data", "dxregia-partners", entry.Call+".connect")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("unexpected request without a verified local account: %v", err)
	}
	snapshot.OutgoingReady[entry.Call] = true
	if err := writePrivateJSON(inboundTestPath(root, entry.Call), now.Add(time.Minute).Unix()); err != nil {
		t.Fatal(err)
	}
	reconnectPartners(root, partners, snapshot, now)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("queued an outgoing connection during the inbound test: %v", err)
	}
	reconnectPartners(root, partners, snapshot, now.Add(2*time.Minute))
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "node.example.org\n7300\n" {
		t.Fatalf("expected outgoing retry without password: %q, %v", data, err)
	}
}
