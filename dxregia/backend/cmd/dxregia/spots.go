package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type publicSpot struct {
	Frequency float64
	Call      string
	At        int64
	Comment   string
	Spotter   string
	DXCC      int
}

func registerSpots(mux *http.ServeMux, root string) {
	mux.HandleFunc("GET /api/spots", func(w http.ResponseWriter, r *http.Request) {
		var spots = make([]publicSpot, 0, 100)
		available := false
		if root != "" {
			path := filepath.Join(root, "local_data", "spots")
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				available = true
				for day := 0; day < 2 && len(spots) < cap(spots); day++ {
					date := time.Now().UTC().AddDate(0, 0, -day)
					file := filepath.Join(path, fmt.Sprintf("%04d", date.Year()), fmt.Sprintf("%03d.dat", date.YearDay()))
					rows, err := recentSpots(file, cap(spots)-len(spots))
					if err != nil && !os.IsNotExist(err) {
						available = false
						break
					}
					spots = append(spots, rows...)
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(map[string]any{"Available": available, "Spots": spots})
	})
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
