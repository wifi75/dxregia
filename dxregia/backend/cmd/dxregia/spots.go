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

type publicSpot struct {
	Frequency float64
	Call      string
	At        int64
	Comment   string
	Spotter   string
	DXCC      int
	Entity    string `json:",omitempty"`
}

type spotSnapshot struct {
	Available        bool
	Spots            []publicSpot
	Sequence         uint64
	LatestAt         int64
	ArchiveUpdatedAt int64
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

func registerSpots(mux *http.ServeMux, root string) func() spotSnapshot {
	entities := spotEntities(root)
	var mu sync.Mutex
	var signature string
	var cached spotSnapshot
	epoch := time.Now().UnixNano()
	snapshot := func() spotSnapshot {
		mu.Lock()
		defer mu.Unlock()
		paths := spotArchivePaths(root, time.Now())
		path := filepath.Join(root, "local_data", "spots")
		nextSignature := fileSignature(path, paths[0], paths[1])
		if nextSignature == signature {
			return cached
		}
		next := spotSnapshot{Spots: make([]publicSpot, 0, 100), Sequence: cached.Sequence + 1}
		if info, err := os.Stat(path); root != "" && err == nil && info.IsDir() {
			next.Available = true
			for _, file := range paths {
				if len(next.Spots) >= 100 {
					break
				}
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
					rows[i].Entity = entities[rows[i].DXCC]
					next.LatestAt = max(next.LatestAt, rows[i].At)
				}
				next.Spots = append(next.Spots, rows...)
			}
		}
		signature, cached = nextSignature, next
		return cached
	}
	mux.HandleFunc("GET /api/spots", func(w http.ResponseWriter, r *http.Request) {
		spots := snapshot()
		etag := fmt.Sprintf("\"%d-%d\"", epoch, spots.Sequence)
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "no-cache")
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(spots)
	})
	return snapshot
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
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	const maxTail = 256 * 1024
	start := info.Size() - min(info.Size(), maxTail)
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, maxTail))
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(data, []byte{'\n'})
	first := 0
	if start != 0 {
		first = 1
	}
	spots := make([]publicSpot, 0, limit)
	for i := len(lines) - 2; i >= first && len(spots) < limit; i-- {
		columns := strings.SplitN(strings.TrimRight(string(lines[i]), "\r"), "^", 7)
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
		spots = append(spots, publicSpot{Frequency: frequency, Call: columns[1],
			At: at, Comment: columns[3], Spotter: columns[4], DXCC: entity})
	}
	return spots, nil
}
