package main

import (
	"strings"
	"testing"
)

func TestMojoLogAllowsDiagnosticFactsOnly(t *testing.T) {
	lines := strings.Join([]string{
		"1794995292^(connect) Connected 6 to peer.example 7393",
		"1794995293^(connect) connect 6: got password sending test-secret",
		"1794995294^(connect) ***Connect 7 Failed to peer.example 7393 test-secret",
		"1794995295^(connect) connect 7: timed out after 15 seconds",
	}, "\n")
	got := appendMojoEvents(nil, lines, parseMojoConnection)
	if len(got) != 3 || got[0].At != 1794995295 || !strings.Contains(got[2].Message, "peer.example:7393") {
		t.Fatalf("unexpected diagnostic events: %+v", got)
	}
	for _, event := range got {
		if strings.Contains(event.Message, "test-secret") {
			t.Fatal("a raw diagnostic payload was exposed")
		}
	}
}

func TestMojoTimeoutDirection(t *testing.T) {
	for suffix, want := range map[string]string{
		" incoming":         "Accesso Telnet in ingresso non completato dopo 60 secondi",
		" outgoing EA1AA-6": "Collegamento in uscita non completato verso EA1AA-6 dopo 60 secondi",
		"":                  "Timeout del collegamento dopo 60 secondi",
	} {
		event, ok := parseMojoConnection("1794995295^(connect) connect 7: timed out after 60 seconds" + suffix)
		if !ok || event.Message != want {
			t.Errorf("timeout %q: got %q, want %q", suffix, event.Message, want)
		}
	}
}

func TestMojoLogShowsPeerAndStage(t *testing.T) {
	parse := mojoDebugParser("1794995200^(connect) ExtMsg accept 7:192.0.2.1 from 2001:db8::1:42100")
	legacy, ok := parse("1794995260^(connect) connect 7: timed out after 60 seconds incoming")
	if !ok || legacy.Host != "2001:db8::1" || legacy.Session != "7" || legacy.Call != "" {
		t.Fatalf("unexpected historical event: %+v", legacy)
	}
	stage, ok := parse("1794995310^(connect) DXRegia session 8 outgoing DA0BCC-7 dx.da0bcc.de password-sent 0")
	if !ok || stage.Call != "DA0BCC-7" || stage.Host != "dx.da0bcc.de" || stage.Direction != "Uscita" {
		t.Fatalf("unexpected stage: %+v", stage)
	}
	if _, ok := parse("1794995310^(connect) DXRegia partner DA0BCC-7 secret"); ok {
		t.Fatal("unexpected payload exposed in partner log")
	}
}
