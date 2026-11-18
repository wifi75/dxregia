package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPartnerMailWaitsForMojo(t *testing.T) {
	dir := t.TempDir()
	job := partnerMail{Call: "Q3QQQ-2", Recipient: "operator@example.org",
		Password: "temporary-lab-password", Host: "node.example.org", Port: 7373}
	var messages []string
	send := func(to, subject, text string) error {
		messages = append(messages, text)
		return nil
	}
	if advancePartnerMail(dir, send, presence{}, &job) || len(messages) != 0 {
		t.Fatal("sent instructions before confirmation")
	}
	if err := os.WriteFile(filepath.Join(dir, job.Call+".result"), []byte("Abilitato in Mojo"), 0600); err != nil {
		t.Fatal(err)
	}
	if !advancePartnerMail(dir, send, presence{}, &job) || len(messages) != 1 ||
		!strings.Contains(messages[0], "7373") || job.Password != "" || job.VerifyAfter.Before(time.Now()) {
		t.Fatal("instructions or scheduled verification not recorded")
	}
	if advancePartnerMail(dir, send, presence{}, &job) || len(messages) != 1 {
		t.Fatal("sent verification before deadline")
	}
	job.VerifyAfter = time.Now().Add(-time.Second)
	snapshot := presence{At: time.Now().Unix(), Channels: []presenceChannel{{Call: job.Call, Kind: "node", State: "normal"}}}
	if !advancePartnerMail(dir, send, snapshot, &job) || len(messages) != 2 ||
		!strings.Contains(messages[1], "operativo") || !job.VerifyAfter.IsZero() {
		t.Fatal("missing confirmation of a real peer")
	}
}
