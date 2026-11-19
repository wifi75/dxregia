package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Centro di analytics del sysop: pubblico del portale (visite, visitatori, lingue, dispositivi, orari)
// e attività del cluster (spot per giorno, ora, banda, modo, entità, nominativi).

type countItem struct {
	Name  string
	Count uint64
	DXCC  int `json:",omitempty"`
}

type dayPoint struct {
	Date     string
	Views    uint64 `json:",omitempty"`
	Visitors uint64 `json:",omitempty"`
	Spots    uint64 `json:",omitempty"`
}

type portalReport struct {
	AllTime   uint64
	Active    int
	Views     uint64
	Visitors  uint64
	Days      []dayPoint
	Hours     []uint64
	Pages     []countItem
	Referrers []countItem
	Languages []countItem
	Devices   []countItem
}

type spotsReport struct {
	Total                                                 uint64
	Days                                                  []dayPoint
	Hours                                                 []uint64
	Bands                                                 []countItem
	Modes                                                 []countItem
	Entities                                              []countItem
	TopDX                                                 []countItem
	TopSpotters                                           []countItem
	BandHours                                             map[string]map[string][]uint64
	UniqueDX, UniqueSpotters, UniqueEntities, UniqueModes int
}

type analyticsReport struct {
	GeneratedAt time.Time
	Days        int
	Portal      portalReport
}

var analyticsBands = []struct {
	Name      string
	Low, High float64
}{
	{"160m", 1800, 2000}, {"80m", 3500, 4000}, {"60m", 5250, 5450}, {"40m", 7000, 7300}, {"30m", 10100, 10150},
	{"20m", 14000, 14350}, {"17m", 18068, 18168}, {"15m", 21000, 21450}, {"12m", 24890, 24990}, {"10m", 28000, 29700},
	{"6m", 50000, 54000}, {"4m", 70000, 70500}, {"2m", 144000, 148000}, {"70cm", 420000, 450000},
}

var analyticsMode = regexp.MustCompile(`(?i)\b(FT8|FT4|SSB|CW|RTTY|PSK31|USB|LSB)\b`)

func analyticsBand(khz float64) string {
	for _, band := range analyticsBands {
		if khz >= band.Low && khz <= band.High {
			return band.Name
		}
	}
	if khz >= 1000000 {
		return "SHF"
	}
	if khz >= 70000 {
		return "VHF+"
	}
	return "Altre"
}

// Se il commento non indica il modo, lo si deduce dalla frequenza (piano di banda IARU regione 1),
// come fa la pagina pubblica degli spot.
var (
	inferFT8 = []float64{1840, 3573, 5357, 7074, 10136, 14074, 18100, 21074, 24915, 28074, 50313}
	inferFT4 = []float64{3575.5, 7047.5, 10140, 14080, 18104, 21140, 24919, 28180, 50318}
	inferSeg = []struct {
		Low, High float64
		Mode      string
	}{
		{1800, 1838, "CW"}, {1838, 1843, "DIGI"}, {1843, 2000, "SSB"},
		{3500, 3570, "CW"}, {3570, 3600, "DIGI"}, {3600, 3800, "SSB"},
		{5351, 5367, "DIGI"},
		{7000, 7040, "CW"}, {7040, 7060, "DIGI"}, {7060, 7200, "SSB"},
		{10100, 10130, "CW"}, {10130, 10150, "DIGI"},
		{14000, 14070, "CW"}, {14070, 14099, "DIGI"}, {14112, 14350, "SSB"},
		{18068, 18095, "CW"}, {18095, 18109, "DIGI"}, {18111, 18168, "SSB"},
		{21000, 21070, "CW"}, {21070, 21150, "DIGI"}, {21151, 21450, "SSB"},
		{24890, 24915, "CW"}, {24915, 24931, "DIGI"}, {24931, 24990, "SSB"},
		{28000, 28070, "CW"}, {28070, 28190, "DIGI"}, {28225, 29700, "SSB"},
		{50000, 50100, "CW"}, {50100, 50300, "SSB"}, {50300, 50500, "DIGI"},
	}
)

func inferMode(khz float64) string {
	for _, value := range inferFT8 {
		if khz >= value-1.2 && khz <= value+1.2 {
			return "FT8"
		}
	}
	for _, value := range inferFT4 {
		if khz >= value-1.2 && khz <= value+1.2 {
			return "FT4"
		}
	}
	for _, segment := range inferSeg {
		if khz >= segment.Low && khz < segment.High {
			return segment.Mode
		}
	}
	return ""
}

func topItems(counts map[string]uint64, limit int) []countItem {
	items := make([]countItem, 0, len(counts))
	for name, count := range counts {
		items = append(items, countItem{Name: name, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return items[i].Name < items[j].Name
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

// analyzeSpots legge l'archivio degli spot degli ultimi giorni. Una riga ha: frequenza^nominativo^ora^commento^spotter^entità.
func analyzeSpots(root string, days int, now time.Time, names, continents map[int]string) spotsReport {
	report := spotsReport{Hours: make([]uint64, 24), BandHours: make(map[string]map[string][]uint64)}
	since := now.Add(-time.Duration(days) * 24 * time.Hour).Unix()
	if days == 0 {
		since = now.UTC().Truncate(24 * time.Hour).Unix()
	}
	perDay := map[string]uint64{}
	bands, modes, dx, spotters := map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]uint64{}
	entityCounts := map[int]uint64{}
	dxEntities, spotterEntities := make(map[string]int), make(map[string]int)
	addHour := func(continent, band string, hour int) {
		if report.BandHours[continent] == nil {
			report.BandHours[continent] = make(map[string][]uint64)
		}
		if report.BandHours[continent][band] == nil {
			report.BandHours[continent][band] = make([]uint64, 24)
		}
		report.BandHours[continent][band][hour]++
	}
	for back := 0; back <= days; back++ {
		date := now.UTC().AddDate(0, 0, -back)
		file := filepath.Join(root, "local_data", "spots", fmt.Sprintf("%04d", date.Year()), fmt.Sprintf("%03d.dat", date.YearDay()))
		handle, err := os.Open(file)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(handle)
		scanner.Buffer(make([]byte, 64*1024), 1<<20)
		for scanner.Scan() {
			columns := strings.SplitN(scanner.Text(), "^", 8)
			if len(columns) < 6 || len(columns[1]) > 24 || len(columns[4]) > 24 {
				continue
			}
			at, err := strconv.ParseInt(columns[2], 10, 64)
			if err != nil || at < since || at > now.Unix()+60 {
				continue
			}
			khz, err := strconv.ParseFloat(columns[0], 64)
			if err != nil || khz <= 0 {
				continue
			}
			when := time.Unix(at, 0).UTC()
			report.Total++
			perDay[when.Format(time.DateOnly)]++
			report.Hours[when.Hour()]++
			band := analyticsBand(khz)
			bands[band]++
			addHour("all", band, when.Hour())
			mode := strings.ToUpper(analyticsMode.FindString(columns[3]))
			if mode == "" {
				mode = inferMode(khz)
			}
			if mode == "USB" || mode == "LSB" {
				mode = "SSB"
			}
			if mode != "" {
				modes[mode]++
			}
			dx[columns[1]]++
			spotters[columns[4]]++
			if entity, err := strconv.Atoi(columns[5]); err == nil && entity > 0 && entity != 666 {
				entityCounts[entity]++
				dxEntities[columns[1]] = entity
				if continent := continents[entity]; continent != "" {
					addHour(continent, band, when.Hour())
				}
			}
			if len(columns) > 6 {
				if entity, err := strconv.Atoi(columns[6]); err == nil && entity > 0 && entity != 666 {
					spotterEntities[columns[4]] = entity
				}
			}
		}
		handle.Close()
	}
	for back := days - 1; back >= 0; back-- {
		date := now.UTC().AddDate(0, 0, -back).Format(time.DateOnly)
		report.Days = append(report.Days, dayPoint{Date: date, Spots: perDay[date]})
	}
	order := map[string]int{}
	for i, band := range analyticsBands {
		order[band.Name] = i
	}
	for name, count := range bands {
		report.Bands = append(report.Bands, countItem{Name: name, Count: count})
	}
	sort.Slice(report.Bands, func(i, j int) bool {
		a, aok := order[report.Bands[i].Name]
		b, bok := order[report.Bands[j].Name]
		if aok != bok {
			return aok
		}
		if aok {
			return a < b
		}
		return report.Bands[i].Name < report.Bands[j].Name
	})
	report.Modes = topItems(modes, 8)
	report.TopDX = topItems(dx, 10)
	report.TopSpotters = topItems(spotters, 10)
	for i := range report.TopDX {
		report.TopDX[i].DXCC = dxEntities[report.TopDX[i].Name]
	}
	for i := range report.TopSpotters {
		report.TopSpotters[i].DXCC = spotterEntities[report.TopSpotters[i].Name]
	}
	report.UniqueDX, report.UniqueSpotters, report.UniqueEntities, report.UniqueModes = len(dx), len(spotters), len(entityCounts), len(modes)
	for code, count := range entityCounts {
		name := names[code]
		if name == "" {
			name = "DXCC " + strconv.Itoa(code)
		}
		report.Entities = append(report.Entities, countItem{Name: name, Count: count, DXCC: code})
	}
	sort.Slice(report.Entities, func(i, j int) bool {
		if report.Entities[i].Count != report.Entities[j].Count {
			return report.Entities[i].Count > report.Entities[j].Count
		}
		return report.Entities[i].Name < report.Entities[j].Name
	})
	return report
}

func (t *traffic) report(days int, now time.Time) portalReport {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sweep(now, true)
	report := portalReport{AllTime: t.data.Views, Active: len(t.active), Hours: make([]uint64, 24)}
	pages, referrers, languages, devices := map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]uint64{}
	for back := days - 1; back >= 0; back-- {
		date := now.UTC().AddDate(0, 0, -back).Format(time.DateOnly)
		day := t.data.Days[date]
		report.Days = append(report.Days, dayPoint{Date: date, Views: day.Views, Visitors: day.Visitors})
		report.Views += day.Views
		report.Visitors += day.Visitors
		for name, count := range day.Pages {
			pages[name] += count
		}
		for name, count := range day.Referrers {
			referrers[name] += count
		}
		for name, count := range day.Languages {
			languages[name] += count
		}
		for name, count := range day.Devices {
			devices[name] += count
		}
		for hour, count := range day.Hours {
			if hour < 24 {
				report.Hours[hour] += count
			}
		}
	}
	report.Pages = topItems(pages, 8)
	report.Referrers = topItems(referrers, 8)
	report.Languages = topItems(languages, 8)
	report.Devices = topItems(devices, 6)
	return report
}

func registerAnalytics(mux *http.ServeMux, root string, auth *sysopAuth, t *traffic) {
	if auth == nil || root == "" || t == nil {
		return
	}
	var lock sync.Mutex
	cache := map[int]analyticsReport{}
	mux.HandleFunc("GET /api/analytics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		days, err := strconv.Atoi(r.URL.Query().Get("days"))
		if err != nil || (days != 1 && days != 7 && days != 30) {
			days = 7
		}
		now := time.Now()
		lock.Lock()
		report, ok := cache[days]
		lock.Unlock()
		if !ok || now.Sub(report.GeneratedAt) > 15*time.Second {
			report = analyticsReport{GeneratedAt: now, Days: days, Portal: t.report(days, now)}
			lock.Lock()
			cache[days] = report
			lock.Unlock()
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	})
}
