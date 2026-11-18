package main

import "testing"

func TestCrontabRoundTripKeepsUnknownLines(t *testing.T) {
	file := parseCrontab("# nota\n0 3 * * * run_cmd(\"export_users\")\n5 4 * * * Spot::daily()\n1 1 * * * run_cmd(\"x$y\")\n")
	if len(file.Entries) != 1 || file.Entries[0].Command != "export_users" {
		t.Fatalf("entries: %+v", file.Entries)
	}
	if len(file.Extra) != 3 {
		t.Fatalf("extra: %+v", file.Extra)
	}
	if again := parseCrontab(file.render()); len(again.Entries) != 1 || len(again.Extra) != 3 {
		t.Fatalf("round trip: %+v", again)
	}
}

func TestCronEntryRejectsPerlInjection(t *testing.T) {
	good := cronEntry{"0", "3", "*", "*", "*", "export_users"}
	if !good.valid() {
		t.Fatal("good entry rejected")
	}
	for _, bad := range []cronEntry{
		{"0", "3", "*", "*", "*", `x"); system("id"); ("`},
		{"0", "3", "*", "*", "*", "a$b"},
		{"0;", "3", "*", "*", "*", "ok"},
		{"0", "3", "*", "*", "*", ""},
	} {
		if bad.valid() {
			t.Fatalf("accepted %+v", bad)
		}
	}
}

func TestMotdValidation(t *testing.T) {
	if !validMotd("Benvenuti\nin DXRegia") || validMotd("a\x00b") {
		t.Fatal("motd validation")
	}
}
