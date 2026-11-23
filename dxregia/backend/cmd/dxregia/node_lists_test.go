package main

import "testing"

func TestStartupCommandsAreRestricted(t *testing.T) {
	for _, ok := range []string{
		"set/var $Spot::maxspots = 100", "set/var $Foo::bar = 'a b'", "load/badwords", "connect GB7DJK-2", "unset/announce",
	} {
		if !validStartupCommand(ok) {
			t.Errorf("rejected %q", ok)
		}
	}
	for _, bad := range []string{
		"set/var $x = system('id')", "set/var $x = 1; system('id')", "run_cmd(\"x\")", "shutdown", "set/var @a = 1",
		"connect x;rm", "set/foo `id`", "export_users",
	} {
		if validStartupCommand(bad) {
			t.Errorf("accepted %q", bad)
		}
	}
}

func TestBadDXRoundTrip(t *testing.T) {
	text := "bless( {\n  ID10T => 1673336995,\n  name => 'baddx',\n  'K1/T' => 5,\n}, 'DXHash' )\n"
	entries := parseBadDX(text)
	if len(entries) != 2 || entries["ID10T"] != 1673336995 || entries["K1/T"] != 5 {
		t.Fatalf("parsed %+v", entries)
	}
	again := parseBadDX(renderBadDX(entries))
	if len(again) != 2 || again["ID10T"] != 1673336995 {
		t.Fatalf("round trip %+v", again)
	}
}

func TestBadListPlainTextAndHash(t *testing.T) {
	text := "# nodi bloccati\nK1ABC-5 w9zz  # commento\n\nNOT_A_CALL\n"
	entries := parseBadDX(text)
	if len(entries) != 2 || entries["K1ABC-5"] != 0 || entries["W9ZZ"] != 0 {
		t.Fatalf("plain text parsed as %+v", entries)
	}
	legacy := parseBadDX("bless( {\n  name => 'badnode',\n  'K1ABC-5' => 17,\n}, 'DXHash' )\n")
	if len(legacy) != 1 || legacy["K1ABC-5"] != 17 {
		t.Fatalf("hash parsed as %+v", legacy)
	}
}

func TestRenderBadTextKeepsCommentsAndDropsHash(t *testing.T) {
	got := renderBadText("# nodi\nK1ABC-5\n", []string{"W9ZZ", "K1ABC-5"})
	if got != "# nodi\nK1ABC-5\nW9ZZ\n" {
		t.Fatalf("rendered %q", got)
	}
	// Un file nel vecchio formato non porta righe di hash nel nuovo testo.
	got = renderBadText("bless( {\n  name => 'badnode',\n  'K1ABC-5' => 17,\n}, 'DXHash' )\n", []string{"W9ZZ"})
	if got != "W9ZZ\n" {
		t.Fatalf("rendered %q", got)
	}
}

func TestIPEntries(t *testing.T) {
	for _, ok := range []string{"192.0.2.1", "192.0.2.0/24", "2001:db8::/32"} {
		if !validIPEntry(ok) {
			t.Errorf("rejected %q", ok)
		}
	}
	for _, bad := range []string{"999.1.1.1", "abc", "192.0.2.0/99", "1.2.3.4; id"} {
		if validIPEntry(bad) {
			t.Errorf("accepted %q", bad)
		}
	}
}

func TestHopsValidation(t *testing.T) {
	if !defaultHops().valid() {
		t.Fatal("defaults must be valid")
	}
	if (hopsConfig{Default: 0}).valid() || (hopsConfig{Default: 5, Table: map[string]int{"x": 3}}).valid() ||
		(hopsConfig{Default: 5, Table: map[string]int{"16": 100}}).valid() {
		t.Fatal("invalid hops accepted")
	}
}
