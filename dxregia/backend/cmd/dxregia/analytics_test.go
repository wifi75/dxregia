package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAnalyzeSpotsCountsBandsModesAndEntities(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	dir := filepath.Join(root, "local_data", "spots", "2026")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	line := func(khz float64, call string, at time.Time, comment, spotter string, dxcc int) string {
		return fmt.Sprintf("%.1f^%s^%d^%s^%s^%d^x", khz, call, at.Unix(), comment, spotter, dxcc)
	}
	file := line(14074, "K1ABC", now.Add(-time.Hour), "FT8 -10 dB", "IU3WJQ", 226) + "\n" +
		line(14074, "K1ABC", now.Add(-2*time.Hour), "FT8", "G4XYZ", 226) + "\n" +
		line(7020, "DL1XX", now.Add(-3*time.Hour), "CW up", "IU3WJQ", 230) + "\n" +
		line(14200, "OLD", now.Add(-60*time.Hour), "too old for 1 day", "A1A", 5) + "\n" +
		"rotta\n"
	if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%03d.dat", now.YearDay())), []byte(file), 0600); err != nil {
		t.Fatal(err)
	}
	report := analyzeSpots(root, 1, now, map[int]string{226: "United States", 230: "Germany"})
	if report.Total != 3 {
		t.Fatalf("expected 3 spots in the last 24 hours, got %d", report.Total)
	}
	if len(report.Bands) != 2 || report.Bands[0].Name != "40m" || report.Bands[1].Name != "20m" || report.Bands[1].Count != 2 {
		t.Fatalf("bands: %+v", report.Bands)
	}
	if len(report.Modes) != 2 || report.Modes[0].Name != "FT8" || report.Modes[0].Count != 2 {
		t.Fatalf("modes: %+v", report.Modes)
	}
	if len(report.Entities) != 2 || report.Entities[0].Name != "United States" || report.Entities[0].DXCC != 226 {
		t.Fatalf("entities: %+v", report.Entities)
	}
	if report.TopDX[0].Name != "K1ABC" || report.TopSpotters[0].Name != "IU3WJQ" || report.Hours[11] != 1 {
		t.Fatalf("top lists or hours: %+v %+v %v", report.TopDX, report.TopSpotters, report.Hours)
	}
	for khz, want := range map[float64]string{7074: "FT8", 14080: "FT4", 7129: "SSB", 14027: "CW", 3573: "FT8", 14100: "", 28100: "DIGI"} {
		if got := inferMode(khz); got != want {
			t.Errorf("inferMode(%v) = %q, want %q", khz, got, want)
		}
	}
	week := analyzeSpots(root, 7, now, nil)
	if week.Total != 4 || len(week.Days) != 7 {
		t.Fatalf("week: total %d, days %d", week.Total, len(week.Days))
	}
}

func TestPortalAudienceCountsUniqueVisitorsLanguagesAndDevices(t *testing.T) {
	root := t.TempDir()
	traffic, err := registerTraffic(http.NewServeMux(), root, nil)
	if err != nil {
		t.Fatal(err)
	}
	visit := func(ip, agent, language string) {
		request := httptest.NewRequest(http.MethodGet, "/spots", nil)
		request.RemoteAddr = ip + ":5000"
		request.Header.Set("User-Agent", agent)
		request.Header.Set("Accept-Language", language)
		traffic.view(request)
	}
	phone := "Mozilla/5.0 (iPhone; CPU iPhone OS 17) Mobile Safari"
	desktop := "Mozilla/5.0 (Windows NT 10.0) Chrome"
	visit("10.0.0.1", phone, "de-DE,de;q=0.9")
	visit("10.0.0.1", phone, "de-DE,de;q=0.9")
	visit("10.0.0.2", desktop, "zh-CN")
	visit("10.0.0.3", "Googlebot/2.1", "")
	report := traffic.report(1, time.Now())
	if report.Views != 4 || report.Visitors != 2 {
		t.Fatalf("views %d, unique visitors %d (a repeat visit and a bot must not add visitors)", report.Views, report.Visitors)
	}
	names := map[string]uint64{}
	for _, item := range report.Languages {
		names[item.Name] = item.Count
	}
	if names["de"] != 2 || names["altre"] != 2 {
		t.Fatalf("languages: %+v", report.Languages)
	}
	devices := map[string]uint64{}
	for _, item := range report.Devices {
		devices[item.Name] = item.Count
	}
	if devices["Telefono"] != 2 || devices["Computer"] != 1 || devices["Bot"] != 1 {
		t.Fatalf("devices: %+v", report.Devices)
	}
	if deviceKind("Mozilla/5.0 (Linux; Android 13; SM-X700)") != "Tablet" {
		t.Fatal("an Android device without Mobile is a tablet")
	}
}
