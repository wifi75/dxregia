package main

import "testing"

func TestBuildAccess(t *testing.T) {
	lines := []string{
		"1000^DXCommand^IU3WJQ connected from 192.168.1.2 cols 80",
		"1060^DXCommand^IU3WJQ disconnected",
		"1100^DXProt^N2KI connected from 162.0.184.162",
		"1200^DXProt^N2KI Disconnected",
		"1300^DXCommand^IK1ABC connected from 203.0.113.9 cols 80",
		"1400^DXCommand^IU3WJQ connected from 192.168.1.2 cols 80",
		"1500^ann^ALL^X connected from nowhere",
	}
	sessions, callers := buildAccess(lines, 0, map[string]bool{"user:IK1ABC": true})
	if len(sessions) != 4 {
		t.Fatalf("sessions: %+v", sessions)
	}
	if sessions[0].Call != "IU3WJQ" || sessions[0].Start != 1400 || sessions[0].Open {
		t.Fatalf("newest first: %+v", sessions[0])
	}
	var ik1 *accessSession
	var first *accessSession
	for i := range sessions {
		switch sessions[i].Call {
		case "IK1ABC":
			ik1 = &sessions[i]
		case "IU3WJQ":
			if sessions[i].Start == 1000 {
				first = &sessions[i]
			}
		}
	}
	if ik1 == nil || !ik1.Open || first == nil || first.Seconds != 60 || first.End != 1060 {
		t.Fatalf("pairing failed: %+v %+v", ik1, first)
	}
	if len(callers) != 3 || callers[0].Call != "IU3WJQ" || callers[0].Count != 2 {
		t.Fatalf("callers: %+v", callers)
	}
}

func TestParseAccessLineIgnoresOtherCategories(t *testing.T) {
	if _, _, _, _, _, ok := parseAccessLine("1^ann^ALL^X connected from y"); ok {
		t.Fatal("announcements must be ignored")
	}
	if at, kind, call, host, connect, ok := parseAccessLine("5^DXCommand^IU3WJQ connected from 10.0.0.1 cols 80"); !ok || at != 5 || kind != "user" || call != "IU3WJQ" || host != "10.0.0.1" || !connect {
		t.Fatal("connect line not parsed")
	}
}
