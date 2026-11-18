package main

import "testing"

func TestParseADXOReadsAnnouncedOperations(t *testing.T) {
	feed := `<?xml version="1.0" ?><rss version="2.0"><channel><title>ADXO</title>
<item><title>Guinea: Nov 11-24, 2026 -- 3X4U -- QSL via: M0URX OQRS</title>
<description>
Nov 11-24, 2026 --
Guinea --
3X4U --
QSL: M0URX OQRS --
By 10 ON/PA ops; CW SSB FT8 RTTY
</description><link>https://www.qrz.com/db/3X4U</link></item>
<item><title>Rodrigues I: Oct 20-Nov 28, 2026 -- 3B9 -- QSL via: LoTW</title><description>x</description><link>javascript:alert(1)</link></item>
<item><title>formato non previsto</title><description>x</description><link>https://example.org</link></item>
</channel></rss>`
	entries, err := parseADXO([]byte(feed))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 valid entries, got %d: %+v", len(entries), entries)
	}
	first := entries[0]
	if first.Call != "3X4U" || first.Entity != "Guinea" || first.Dates != "Nov 11-24, 2026" || first.QSL != "M0URX OQRS" || first.Link != "https://www.qrz.com/db/3X4U" {
		t.Fatalf("first entry: %+v", first)
	}
	if first.Info == "" || len(first.Info) > 705 {
		t.Fatalf("info: %q", first.Info)
	}
	if entries[1].Link != adxoSource {
		t.Fatalf("an unsafe link must be replaced by the source page: %q", entries[1].Link)
	}
}
