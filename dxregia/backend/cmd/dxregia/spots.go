package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxSpotTail = 256 * 1024

type publicSpot struct {
	Cursor           string `json:",omitempty"`
	Frequency        float64
	Call             string
	At               int64
	Comment          string
	Spotter          string
	DXCC             int
	SpotterDXCC      int      `json:",omitempty"`
	Continent        string   `json:",omitempty"`
	SpotterContinent string   `json:",omitempty"`
	CQ, ITU          int      `json:",omitempty"`
	Grids            []string `json:",omitempty"`
	Special          bool     `json:",omitempty"`
	Entity           string   `json:",omitempty"`
	LoTW, EQSL       bool     `json:",omitempty"`
}

type spotSnapshot struct {
	Available                    bool
	Spots                        []publicSpot
	Sequence                     uint64
	LatestAt                     int64
	ArchiveUpdatedAt             int64
	LoTWAvailable, EQSLAvailable bool
}

func spotArchivePaths(root string, now time.Time) [2]string {
	var paths [2]string
	for day := range paths {
		date := now.UTC().AddDate(0, 0, -day)
		paths[day] = filepath.Join(root, "local_data", "spots", fmt.Sprintf("%04d", date.Year()), fmt.Sprintf("%03d.dat", date.YearDay()))
	}
	return paths
}

func fileSignature(paths ...string) string {
	var signature strings.Builder
	for _, path := range paths {
		signature.WriteString(path)
		if info, err := os.Stat(path); err == nil {
			fmt.Fprintf(&signature, ":%d:%d;", info.ModTime().UnixNano(), info.Size())
		} else {
			signature.WriteString(":-;")
		}
	}
	return signature.String()
}

type spotFileVersion struct {
	modified, size int64
	directory      bool
}

func spotFileStat(path string) spotFileVersion {
	info, err := os.Stat(path)
	if err != nil {
		return spotFileVersion{}
	}
	return spotFileVersion{info.ModTime().UnixNano(), info.Size(), info.IsDir()}
}

type spotFeed struct {
	snapshot spotSnapshot
	body     []byte
	etag     string
}

func registerSpots(mux *http.ServeMux, root string, qsl *qslDirectory) func() spotSnapshot {
	directory := loadDXCC(root)
	registerSpotResearch(mux, root, directory, qsl)
	path := filepath.Join(root, "local_data", "spots")
	var mu sync.Mutex
	var previous [3]spotFileVersion
	var previousQSL qslSnapshot
	var paths [2]string
	var archiveDay time.Time
	var cached spotFeed
	epoch := time.Now().UnixNano()
	feed := func() spotFeed {
		mu.Lock()
		defer mu.Unlock()
		day := time.Now().UTC().Truncate(24 * time.Hour)
		newDay := day != archiveDay
		if newDay {
			paths = spotArchivePaths(root, day)
			archiveDay = day
		}
		var current [3]spotFileVersion
		if root != "" {
			current = [3]spotFileVersion{spotFileStat(path), spotFileStat(paths[0]), spotFileStat(paths[1])}
		}
		membership := qsl.snapshot()
		if !newDay && current == previous && membership == previousQSL {
			return cached
		}
		next := spotSnapshot{Spots: make([]publicSpot, 0, 100), Sequence: cached.snapshot.Sequence + 1,
			LoTWAvailable: membership.lotw != nil, EQSLAvailable: membership.eqsl != nil}
		if current[0].directory {
			next.Available = true
			for _, file := range paths {
				if len(next.Spots) >= 100 {
					break
				}
				// A missing file is normal, but still let the reader report other errors.
				rows, err := recentSpots(file, 100-len(next.Spots))
				if err != nil && !os.IsNotExist(err) {
					next.Available = false
					break
				}
				if len(rows) > 0 {
					if info, err := os.Stat(file); err == nil {
						next.ArchiveUpdatedAt = max(next.ArchiveUpdatedAt, info.ModTime().UnixMilli())
					}
				}
				for i := range rows {
					directory.enrich(&rows[i])
					membership.enrich(&rows[i])
					next.LatestAt = max(next.LatestAt, rows[i].At)
				}
				next.Spots = append(next.Spots, rows...)
			}
		}
		// The same snapshot drives the JSON, ETag and health checks.
		body, err := json.Marshal(next)
		if err == nil {
			body = append(body, '\n')
		}
		previous, previousQSL, cached = current, membership, spotFeed{snapshot: next, body: body, etag: fmt.Sprintf(`"%d-%d"`, epoch, next.Sequence)}
		return cached
	}
	mux.HandleFunc("GET /api/spots", func(w http.ResponseWriter, r *http.Request) {
		current := feed()
		if current.body == nil {
			http.Error(w, "Spot feed unavailable.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("ETag", current.etag)
		w.Header().Set("Cache-Control", "no-cache")
		if r.Header.Get("If-None-Match") == current.etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(current.body)
	})
	return func() spotSnapshot { return feed().snapshot }
}

var prefixEntity = regexp.MustCompile(`^\s*[0-9]+\s*=>\s*bless\(\s*\{\s*name\s*=>\s*'([^']+)',\s*dxcc\s*=>\s*([0-9]+),`)

func spotEntities(root string) map[int]string {
	if root == "" {
		return nil
	}
	f, err := os.Open(spotDataPath(root, "prefix_data.pl"))
	if err != nil {
		return nil
	}
	defer f.Close()
	names := make(map[int]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := prefixEntity.FindStringSubmatch(scanner.Text())
		if fields == nil {
			continue
		}
		code, err := strconv.Atoi(fields[2])
		if err != nil || code == 0 || code == 666 || names[code] != "" {
			continue
		}
		// The canonical entity precedes regional variants in Mojo's data.
		name := fields[1]
		if suffix := strings.LastIndexByte(name, '-'); suffix > 0 {
			name = name[:suffix]
		}
		names[code] = strings.ReplaceAll(name, "-", " ")
	}
	return names
}

func spotDataPath(root, name string) string {
	source := filepath.Join(root, "data", name)
	local := filepath.Join(root, "local_data", name)
	if info, err := os.Stat(local); err == nil {
		if bundled, err := os.Stat(source); err != nil || !bundled.ModTime().After(info.ModTime()) {
			return local
		}
	}
	return source
}

func recentSpots(path string, limit int) ([]publicSpot, error) {
	return readSpotTail(path, limit, maxSpotTail)
}

func readSpotTail(path string, limit int, maxBytes int64) ([]publicSpot, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	start := info.Size() - min(info.Size(), maxBytes)
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, maxBytes))
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(data, []byte{'\n'})
	first := 0
	if start != 0 {
		first = 1
	}
	spots := make([]publicSpot, 0, limit)
	day := filepath.Base(filepath.Dir(path)) + strings.TrimSuffix(filepath.Base(path), ".dat")
	offset := start + int64(len(data)-len(lines[len(lines)-1]))
	for i := len(lines) - 2; i >= first && len(spots) < limit; i-- {
		offset -= int64(len(lines[i]) + 1)
		columns := strings.SplitN(strings.TrimRight(string(lines[i]), "\r"), "^", 11)
		if len(columns) < 6 || len(columns[1]) > 24 || len(columns[4]) > 24 || len(columns[3]) > 1024 {
			continue
		}
		frequency, err := strconv.ParseFloat(columns[0], 64)
		if err != nil || frequency <= 0 || math.IsNaN(frequency) || math.IsInf(frequency, 0) {
			continue
		}
		at, err := strconv.ParseInt(columns[2], 10, 64)
		if err != nil || at <= 0 || at >= 253402300800 {
			continue
		}
		entity, _ := strconv.Atoi(columns[5])
		spot := publicSpot{Cursor: fmt.Sprintf("%s-%020d", day, offset), Frequency: frequency, Call: columns[1],
			At: at, Comment: columns[3], Spotter: columns[4], DXCC: entity}
		if len(columns) > 6 {
			spot.SpotterDXCC, _ = strconv.Atoi(columns[6])
		}
		if len(columns) > 9 {
			spot.ITU, _ = strconv.Atoi(columns[8])
			spot.CQ, _ = strconv.Atoi(columns[9])
		}
		spots = append(spots, spot)
	}
	return spots, nil
}
