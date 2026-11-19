package main

import (
	"strings"
	"testing"
	"time"
)

func TestPartnerAlertsThresholdReasonAndRecovery(t *testing.T) {
	root := t.TempDir()
	start := time.Now().UTC().Truncate(time.Second)
	entry := partner{partnerDetails: partnerDetails{Call: "DA0BCC-7", Host: "example.org", Port: 7300}, FirstConnectedAt: &start}
	partners := map[string]partner{entry.Call: entry}
	prefs := emailPreferences{PartnerDown: true, PartnerRecovered: true, PartnerDownMinutes: 3}
	snapshot := presence{PartnerConnect: true, LockedPartners: map[string]bool{entry.Call: true}}
	outages := map[string]partnerOutage{}
	var delivered []mailContent
	send := func(content mailContent) error { delivered = append(delivered, content); return nil }
	advancePartnerAlerts(root, snapshot, start, prefs, partners, outages, send)
	advancePartnerAlerts(root, snapshot, start.Add(3*time.Minute-time.Second), prefs, partners, outages, send)
	if len(delivered) != 0 {
		t.Fatal("sent before the configured threshold")
	}
	advancePartnerAlerts(root, snapshot, start.Add(3*time.Minute), prefs, partners, outages, send)
	advancePartnerAlerts(root, snapshot, start.Add(5*time.Minute), prefs, partners, outages, send)
	if len(delivered) != 1 || !strings.Contains(delivered[0].Text, "Nodo bloccato in Mojo") {
		t.Fatal("expected exactly one alert with the recorded reason")
	}
	snapshot.Channels = []presenceChannel{{Call: entry.Call, Kind: "node", State: "normal"}}
	advancePartnerAlerts(root, snapshot, start.Add(6*time.Minute), prefs, partners, outages, send)
	if len(delivered) != 2 || !strings.Contains(delivered[1].Text, "Collegamento ripristinato") || len(outages) != 0 {
		t.Fatal("expected exactly one recovery email and cleared outage")
	}
}
