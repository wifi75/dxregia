package main

import (
	"net"
	"testing"
)

func TestIssueCountryRangeEncoding(t *testing.T) {
	for _, fixture := range []struct {
		cidr, first, last string
	}{
		{"198.51.100.0/24", "198.51.100.0", "198.51.100.255"},
		{"2001:db8:5::/48", "2001:db8:5::", "2001:db8:5:ffff:ffff:ffff:ffff:ffff"},
	} {
		_, network, err := net.ParseCIDR(fixture.cidr)
		if err != nil {
			t.Fatal(err)
		}
		rangeBytes := issueRange(network, "it")
		if !net.IP(rangeBytes[:16]).Equal(net.ParseIP(fixture.first)) ||
			!net.IP(rangeBytes[16:32]).Equal(net.ParseIP(fixture.last)) || string(rangeBytes[32:]) != "it" {
			t.Fatalf("incorrect range for %s", fixture.cidr)
		}
	}
}

func TestMemorySampleUsesAvailableAndKeepsZero(t *testing.T) {
	available, total, ok := meminfo("MemTotal: 5000 kB\nMemFree: 400 kB\nMemAvailable: 0 kB\n")
	if !ok || available != 0 || total != 5000*1024 {
		t.Fatalf("unexpected memory sample %d/%d, valid %t", available, total, ok)
	}
	if _, _, ok := meminfo("MemTotal: 5000 kB\nMemFree: 400 kB"); ok {
		t.Fatal("missing MemAvailable should mark the sample unavailable")
	}
}
