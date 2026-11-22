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

func TestMojoRoutingAndAccessEvents(t *testing.T) {
	tests := []struct{ line, kind, call, want string }{
		{"1795336715^RING: 08:38:31^(*) PCPROT: Bad Spot JH0QEV on 21074.0 by EA9ACF(80.31.5.110)@EA4RCH-5 User EA9ACF not on node EA4RCH-5, WARNING via DA0BCC-7 26 hops",
			"Routing", "DA0BCC-7", "EA9ACF non risulta collegato al nodo EA4RCH-5"},
		{"1795320222^(*) PCPROT: PC61 192.168.0.40 is a localhost address or in RFC1918, dropped", "Routing", "", "non è un indirizzo pubblico"},
		{"1795320222^(*) PCPROT: PC61 192.0.2.1 is not a public Internet address, dropped", "Routing", "", "non è un indirizzo pubblico"},
		{"1795330000^PC92_MALFORMED drop peer=DA0BCC-7 origin=EA4RCH-5 id=12 type=C reason=invalid object: x packet=^PC92^secret", "Routing", "DA0BCC-7", "invalid object: x"},
		{"1795330001^(err) DXChannel DH8WR-1: too many errors (11 > 10), disconnecting", "Accessi", "DH8WR-1", "11 su un massimo di 10"},
		{"1795337428^(*) DL9DBZ on 88.66.49.149 nextlogin $nextlogin > $systime DL9DBZ logging in too quickly, banned for 307 seconds until 08:54:51Z CLOSING", "Accessi", "DL9DBZ", "bloccato per 307 secondi"},
		{"1795337430^(*) KM3T on 203.0.113.9 is in badspotter, disconnected", "Accessi", "KM3T", "badspotter"},
		{"1795337431^(*) IW1ABC has too many connections (13 > 12) disconnected", "Accessi", "IW1ABC", "13 su un massimo di 12"},
	}
	for _, test := range tests {
		event, ok := parseMojoRouting(test.line)
		if !ok || event.Kind != test.kind || event.Call != test.call || !strings.Contains(event.Message, test.want) {
			t.Errorf("%q: got %+v ok=%v", test.line, event, ok)
		}
		if strings.Contains(event.Message, "80.31.5.110") || strings.Contains(event.Message, "secret") {
			t.Errorf("a private detail reached the message: %q", event.Message)
		}
	}
	if _, ok := parseMojoRouting("1795337431^(*) something else entirely"); ok {
		t.Error("an unknown line produced an event")
	}
}
