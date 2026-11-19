package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type clusterBulletin struct {
	At                                              int64
	Call, Text                                      string
	SFI, A, K                                       string `json:",omitempty"`
	ExpectedK, Sunspots, Solar, Geomagnetic, Aurora string `json:",omitempty"`
}

func parseBulletin(kind, line string) (clusterBulletin, bool) {
	var row clusterBulletin
	var timestamp string
	switch kind {
	case "wwv":
		fields := strings.Split(line, "^")
		if len(fields) < 7 {
			return row, false
		}
		row.Call, timestamp, row.SFI, row.A, row.K, row.Text = fields[0], fields[1], fields[2], fields[3], fields[4], fields[5]
		if len(fields) >= 8 {
			row.Sunspots = fields[7]
		}
	case "wcy":
		fields := strings.Split(line, "^")
		if len(fields) < 11 {
			return row, false
		}
		timestamp, row.SFI, row.A, row.K = fields[0], fields[1], fields[2], fields[3]
		row.ExpectedK, row.Sunspots, row.Solar, row.Geomagnetic, row.Aurora, row.Call = fields[4], fields[5], fields[6], fields[7], fields[8], fields[9]
	default:
		return row, false
	}
	row.At, _ = strconv.ParseInt(timestamp, 10, 64)
	row.Call = strings.ToUpper(strings.TrimSpace(row.Call))
	row.Text = strings.TrimSpace(row.Text)
	return row, row.At > 0 && row.At < 1<<32 && researchCall.MatchString(row.Call)
}

func readBulletins(path, kind string) ([]clusterBulletin, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	const maxBytes = 256 * 1024
	start := max(int64(0), info.Size()-maxBytes)
	data, err := io.ReadAll(io.NewSectionReader(file, start, maxBytes))
	if err != nil {
		return nil, err
	}
	if start > 0 {
		_, data, _ = bytes.Cut(data, []byte{'\n'})
	}
	if end := bytes.LastIndexByte(data, '\n'); end >= 0 {
		data = data[:end+1]
	} else {
		return nil, nil
	}
	rows := make([]clusterBulletin, 0)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if row, ok := parseBulletin(kind, scanner.Text()); ok {
			rows = append(rows, row)
		}
	}
	return rows, scanner.Err()
}

func registerBulletins(mux *http.ServeMux, root string) {
	type snapshot struct {
		at   time.Time
		rows []clusterBulletin
	}
	var mu sync.Mutex
	cache := map[string]snapshot{}
	mux.HandleFunc("GET /api/cluster-bulletins/{kind}", func(w http.ResponseWriter, r *http.Request) {
		kind := r.PathValue("kind")
		if kind != "wwv" && kind != "wcy" {
			http.NotFound(w, r)
			return
		}
		mu.Lock()
		current := cache[kind]
		if time.Since(current.at) >= 15*time.Second {
			now := time.Now().UTC()
			rows := make([]clusterBulletin, 0)
			if root != "" {
				month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
				for offset := range 3 {
					if len(rows) >= 50 {
						break
					}
					date := month.AddDate(0, -offset, 0)
					path := filepath.Join(root, "local_data", kind, fmt.Sprintf("%04d", date.Year()), fmt.Sprintf("%02d.dat", date.Month()))
					found, err := readBulletins(path, kind)
					if err != nil && !os.IsNotExist(err) {
						mu.Unlock()
						http.Error(w, "Bulletin archive unavailable.", http.StatusServiceUnavailable)
						return
					}
					rows = append(rows, found...)
				}
			}
			sort.SliceStable(rows, func(i, j int) bool { return rows[i].At > rows[j].At })
			seen := make(map[clusterBulletin]bool)
			kept := rows[:0]
			for _, row := range rows {
				if !seen[row] {
					seen[row] = true
					kept = append(kept, row)
				}
				if len(kept) == 50 {
					break
				}
			}
			current = snapshot{now, kept}
			cache[kind] = current
		}
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=15")
		_ = json.NewEncoder(w).Encode(current.rows)
	})
}
