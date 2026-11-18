package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSMTPSettingsValidate(t *testing.T) {
	valid := smtpSettings{Address: "smtp.example.com:465", From: "dx@example.com", TLS: "implicit"}
	if err := valid.validate(); err != nil {
		t.Fatalf("valid settings rejected: %v", err)
	}
	for name, mutate := range map[string]func(*smtpSettings){
		"no port":       func(s *smtpSettings) { s.Address = "smtp.example.com" },
		"bad sender":    func(s *smtpSettings) { s.From = "not an address" },
		"named sender":  func(s *smtpSettings) { s.From = "DX <dx@example.com>" },
		"bad tls":       func(s *smtpSettings) { s.TLS = "none" },
		"header inject": func(s *smtpSettings) { s.User = "a\r\nBcc: x@example.com" },
	} {
		settings := valid
		mutate(&settings)
		if settings.validate() == nil {
			t.Errorf("%s: invalid settings accepted", name)
		}
	}
}

func TestCurrentMailerPrefersPanelSettings(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DXREGIA_SMTP_ADDRESS", "env.example.com:465")
	t.Setenv("DXREGIA_SMTP_FROM", "env@example.com")
	if m := currentMailer(root); m == nil || m.host != "env.example.com" {
		t.Fatalf("expected the environment mailer, got %+v", m)
	}
	settings := smtpSettings{Address: "panel.example.com:587", From: "panel@example.com", User: "u", Password: "p", TLS: "starttls"}
	if err := writeSMTPSettings(root, settings); err != nil {
		t.Fatal(err)
	}
	m := currentMailer(root)
	if m == nil || m.host != "panel.example.com" || !m.startTLS || m.password != "p" {
		t.Fatalf("expected the panel mailer, got %+v", m)
	}
	if view := smtpViewOf(root); view.Source != "pannello" || !view.PasswordSet {
		t.Fatalf("unexpected view %+v", view)
	}
	if err := os.WriteFile(smtpSettingsPath(root), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if currentMailer(root) != nil {
		t.Fatal("a corrupt panel file must disable mail, not fall back silently")
	}
	if _, err := os.Stat(filepath.Dir(smtpSettingsPath(root))); err != nil {
		t.Fatal(err)
	}
}
