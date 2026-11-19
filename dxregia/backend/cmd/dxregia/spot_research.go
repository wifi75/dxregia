package main

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type dxccEntry struct {
	Code          int
	Name, Prefix  string
	Continent     string
	CQ, ITU       int
	ReferenceGrid string
}

type dxccDirectory struct {
	entries    map[int]dxccEntry
	exactCalls map[string]bool
}

var prefixZones = regexp.MustCompile("itu => ([0-9]+), cq => ([0-9]+),.*qra => '([^']*)'")
var reportedGrid = regexp.MustCompile("(?i)^[A-R]{2}[0-9]{2}(?:[A-X]{2}(?:[0-9]{2})?)?$")
var spotCursor = regexp.MustCompile("^[0-9]{7}-[0-9]{20}$")
var researchCall = regexp.MustCompile("^[A-Z0-9][A-Z0-9/-]{0,23}$")

func loadDXCC(root string) dxccDirectory {
	directory := dxccDirectory{entries: make(map[int]dxccEntry), exactCalls: make(map[string]bool)}
	if root == "" {
		return directory
	}
	names, continents := spotEntities(root), spotContinents(root)
	for code, name := range names {
		directory.entries[code] = dxccEntry{Code: code, Name: name, Continent: continents[code]}
	}
	file, err := os.Open(spotDataPath(root, "prefix_data.pl"))
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			fields := prefixEntity.FindStringSubmatch(line)
			if fields == nil {
				continue
			}
			code, _ := strconv.Atoi(fields[2])
			entry, exists := directory.entries[code]
			suffix := strings.LastIndexByte(fields[1], '-')
			if !exists || suffix < 0 || entry.Prefix != "" {
				continue
			}
			entry.Prefix = fields[1][suffix+1:]
			if zone := prefixZones.FindStringSubmatch(line); zone != nil {
				entry.CQ, _ = strconv.Atoi(zone[2])
				entry.ITU, _ = strconv.Atoi(zone[1])
				entry.ReferenceGrid = zone[3]
			}
			directory.entries[code] = entry
		}
		file.Close()
	}
	if file, err := os.Open(spotDataPath(root, "cty.dat")); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || (line[0] != ' ' && line[0] != '\t') {
				continue
			}
			for _, alias := range strings.FieldsFunc(line, func(r rune) bool { return r == ',' || r == ';' }) {
				if call, exact := strings.CutPrefix(strings.TrimSpace(alias), "="); exact {
					if end := strings.IndexAny(call, "([<{~"); end >= 0 {
						call = call[:end]
					}
					if researchCall.MatchString(call) {
						directory.exactCalls[call] = true
					}
				}
			}
		}
		file.Close()
	}
	return directory
}

func (d dxccDirectory) enrich(spot *publicSpot) {
	entity := d.entries[spot.DXCC]
	spot.Entity, spot.Continent = entity.Name, entity.Continent
	spot.SpotterContinent = d.entries[spot.SpotterDXCC].Continent
	spot.Special = d.exactCalls[strings.ToUpper(spot.Call)]
	for _, field := range strings.FieldsFunc(spot.Comment, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
	}) {
		if reportedGrid.MatchString(field) {
			spot.Grids = append(spot.Grids, strings.ToUpper(field))
		}
	}
}

const researchArchiveDays = 7

// A cursor is a day and byte offset in a newline-delimited DXSpider archive.
func resumeArchivedSpots(root, cursor string, directory dxccDirectory, now time.Time) ([]publicSpot, bool) {
	const limit = 1000
	rows := make([]publicSpot, 0)
	if root == "" {
		return rows, false
	}
	gap := false
	for day := range researchArchiveDays {
		path := spotArchivePaths(root, now.AddDate(0, 0, -day))[0]
		stored, err := recentSpots(path, limit+1)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			gap = true
			continue
		}
		for _, spot := range stored {
			if spot.Cursor == cursor {
				return rows, !gap
			}
			if spot.Cursor < cursor || len(rows) == limit {
				return rows, false
			}
			directory.enrich(&spot)
			rows = append(rows, spot)
		}
		// A bounded tail may omit earlier records from a day after the cursor.
		if info, err := os.Stat(path); err != nil || info.Size() > maxSpotTail {
			gap = true
		}
	}
	return rows, false
}

func registerSpotResearch(mux *http.ServeMux, root string, directory dxccDirectory, qsl *qslDirectory) {
	entities := make([]dxccEntry, 0, len(directory.entries))
	for _, entry := range directory.entries {
		entities = append(entities, entry)
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].Name < entities[j].Name })
	mux.HandleFunc("GET /api/dxcc", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=900")
		_ = json.NewEncoder(w).Encode(entities)
	})

	mux.HandleFunc("GET /api/spot-resume", func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("after")
		if !spotCursor.MatchString(cursor) {
			http.Error(w, "Invalid spot cursor.", http.StatusBadRequest)
			return
		}
		rows, complete := resumeArchivedSpots(root, cursor, directory, time.Now().UTC())
		membership := qsl.snapshot()
		for i := range rows {
			membership.enrich(&rows[i])
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(struct {
			Spots    []publicSpot
			Complete bool
		}{rows, complete})
	})
	const (
		resultLimit     = 200
		historyDayRows  = 8000
		historyDayBytes = 1 << 20
	)
	var mu sync.Mutex
	var sampledAt time.Time
	var archive []publicSpot
	mux.HandleFunc("GET /api/spot-history", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		call := strings.ToUpper(strings.TrimSpace(q.Get("call")))
		entity := q.Get("dxcc")
		code, _ := strconv.Atoi(entity)
		if (call == "") == (entity == "") || (call != "" && !researchCall.MatchString(call)) ||
			(entity != "" && directory.entries[code].Code == 0) {
			http.Error(w, "Choose one callsign or DXCC entity.", http.StatusBadRequest)
			return
		}
		now := time.Now().UTC()
		mu.Lock()
		if now.Sub(sampledAt) >= 30*time.Second {
			next := make([]publicSpot, 0)
			if root != "" {
				for day := range researchArchiveDays {
					rows, _ := readSpotTail(spotArchivePaths(root, now.AddDate(0, 0, -day))[0], historyDayRows, historyDayBytes)
					next = append(next, rows...)
				}
			}
			for i := range next {
				directory.enrich(&next[i])
			}
			sort.SliceStable(next, func(i, j int) bool { return next[i].At > next[j].At })
			archive, sampledAt = next, now
		}
		rows := archive
		mu.Unlock()

		membership := qsl.snapshot()
		results := make([]publicSpot, 0)
		bands := make(map[string]int)
		for _, spot := range rows {
			if (call != "" && strings.EqualFold(spot.Call, call)) || (code != 0 && spot.DXCC == code) {
				if len(results) < resultLimit {
					membership.enrich(&spot)
					results = append(results, spot)
				}
				if code != 0 {
					bands[analyticsBand(spot.Frequency)]++
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=15")
		_ = json.NewEncoder(w).Encode(struct {
			Spots []publicSpot
			Bands map[string]int
		}{results, bands})
	})
}
