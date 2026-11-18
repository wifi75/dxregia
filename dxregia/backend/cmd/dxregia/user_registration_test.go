package main

import (
	"regexp"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestCleanPublicURL(t *testing.T) {
	for raw, want := range map[string]string{
		"https://dxregia.example.org":   "https://dxregia.example.org",
		"https://dxregia.example.org/":  "https://dxregia.example.org",
		"http://localhost:8080":         "http://localhost:8080",
		"  https://x.example.org:8443 ": "https://x.example.org:8443",
	} {
		if got, ok := cleanPublicURL(raw); !ok || got != want {
			t.Errorf("%q → %q %v", raw, got, ok)
		}
	}
	for _, bad := range []string{"", "http://example.org", "https://example.org/path", "javascript:alert(1)", "https://u:p@example.org", "ftp://example.org"} {
		if _, ok := cleanPublicURL(bad); ok {
			t.Errorf("accepted %q", bad)
		}
	}
}

func TestRegistrationInputRules(t *testing.T) {
	for _, call := range []string{"IU3WJQ", "K1ABC", "2E0XYZ"} {
		if !userCall.MatchString(call) {
			t.Errorf("rejected call %s", call)
		}
	}
	for _, call := range []string{"IU3WJQ-7", "IU3/P", "ABC", "TE5T9", "12345", "iu3wjq", "A B1C"} {
		if userCall.MatchString(call) {
			t.Errorf("accepted call %s", call)
		}
	}
	if !userPassword.MatchString("Sup3rS3gret0!") || userPassword.MatchString("short") || userPassword.MatchString("has space 123456") {
		t.Fatal("password rules")
	}
}

// L'hash prodotto dal backend deve avere esattamente la forma che l'hook di Mojo accetta.
func TestPasswordHashMatchesHookPattern(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("Sup3rS3gret0!"), 10)
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^\$2a\$10\$[./A-Za-z0-9]{53}$`).Match(hash) {
		t.Fatalf("unexpected hash shape: %s", hash)
	}
}

func TestConfirmationMail(t *testing.T) {
	for _, lang := range mailLanguages {
		content := mailRegistrationConfirm(lang, "IU3WJQ", "IU3WJQ-7", "https://dxregia.example.org/registrazione/conferma?token=abc")
		if !strings.Contains(content.HTML, "token=abc") || !strings.Contains(content.Text, "token=abc") || strings.Contains(content.Subject, "{") {
			t.Errorf("%s: link or placeholders: %q", lang, content.Subject)
		}
	}
}
