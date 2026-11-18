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
