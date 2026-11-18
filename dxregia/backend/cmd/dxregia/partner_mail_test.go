package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPartnerMailLifecycle(t *testing.T) {
	dir := t.TempDir()
	job := partnerMail{Call: "Q3QQQ-2", Recipient: "operator@example.org", Contact: "Mario",
		Password: "temporary-lab-password", LocalNode: "IU3WJQ-7", Host: "node.example.org", Port: 7373, Lang: "en"}
	var sent []mailContent
	send := func(to string, c mailContent) error {
		sent = append(sent, c)
		return nil
	}
	if advancePartnerMail(dir, send, presence{}, &job) || len(sent) != 0 {
		t.Fatal("sent instructions before confirmation")
	}
	if err := os.WriteFile(filepath.Join(dir, job.Call+".result"), []byte("Abilitato in Mojo"), 0600); err != nil {
		t.Fatal(err)
	}
	if !advancePartnerMail(dir, send, presence{}, &job) || len(sent) != 1 || job.Password != "" || job.WatchUntil.IsZero() {
		t.Fatal("instructions or watch not recorded")
	}
	first := sent[0]
	if !strings.Contains(first.HTML, "7373") || !strings.Contains(first.HTML, "temporary-lab-password") ||
		!strings.Contains(first.Text, "temporary-lab-password") || !strings.Contains(first.Subject, "instructions") {
		t.Fatalf("instructions must be HTML and text in the partner language: %q", first.Subject)
	}
	if advancePartnerMail(dir, send, presence{At: time.Now().Unix()}, &job) || len(sent) != 1 {
		t.Fatal("sent a notice before the deadline")
	}
	job.VerifyAfter = time.Now().Add(-time.Second)
	idle := presence{At: time.Now().Unix()}
	if !advancePartnerMail(dir, send, idle, &job) || len(sent) != 2 || !job.PendingSent {
		t.Fatal("missing pending notice")
	}
	if advancePartnerMail(dir, send, idle, &job) || len(sent) != 2 {
		t.Fatal("pending notice sent twice")
	}
	snapshot := presence{At: time.Now().Unix(), Channels: []presenceChannel{{Call: job.Call, Kind: "node", State: "normal"}}}
	if !advancePartnerMail(dir, send, snapshot, &job) || len(sent) != 3 || !job.WatchUntil.IsZero() ||
		!strings.Contains(sent[2].Subject, "connected") {
		t.Fatal("missing confirmation of a real peer")
	}
	if advancePartnerMail(dir, send, snapshot, &job) || len(sent) != 3 {
		t.Fatal("confirmation sent twice")
	}
}

func TestMailLanguagesAreComplete(t *testing.T) {
	for _, lang := range mailLanguages {
		s := textsFor(lang)
		for name, value := range map[string]string{"InstrSubject": s.InstrSubject, "InstrIntro": s.InstrIntro, "OKText": s.OKText, "NoText": s.NoText,
			"RecvText": s.RecvText, "Hello": s.Hello, "AutoNote": s.AutoNote, "Steps0": s.Steps[0], "Steps3": s.Steps[3], "NoChecks2": s.NoChecks[2]} {
			if strings.TrimSpace(value) == "" {
				t.Errorf("%s: %s is empty", lang, name)
			}
		}
	}
	if normalizeLang("pt-BR") != "pt" || normalizeLang("xx") != "" {
		t.Fatal("language normalization")
	}
	if textsFor("zh").InstrSubject != textsFor("en").InstrSubject || textsFor("").InstrSubject != textsFor("en").InstrSubject {
		t.Fatal("unsupported languages must fall back to English")
	}
}

func TestMailMessageIsMultipart(t *testing.T) {
	c := mailLinkOK("de", partnerMail{Call: "A1B", LocalNode: "IU3WJQ-7"}, "<b>x</b>")
	message := buildMessage("dx@example.org", "a@example.org", "", c)
	if !strings.Contains(message, "multipart/alternative") || !strings.Contains(message, "text/html") || !strings.Contains(message, "text/plain") {
		t.Fatal("not a multipart alternative message")
	}
	if strings.Contains(c.HTML, "<b>x</b>") {
		t.Fatal("dynamic values must be HTML-escaped")
	}
}
