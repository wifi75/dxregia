package main

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func spotContinents(root string) map[int]string {
	cty, err := os.Open(spotDataPath(root, "cty.dat"))
	if err != nil {
		return nil
	}
	defer cty.Close()
	byPrefix := make(map[string]string)
	scanner := bufio.NewScanner(cty)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == ' ' || line[0] == '\t' {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 9 {
			continue
		}
		continent := strings.TrimSpace(fields[3])
		switch continent {
		case "AF", "AN", "AS", "EU", "NA", "OC", "SA":
			byPrefix[strings.TrimLeft(strings.TrimSpace(fields[7]), "*")] = continent
		}
	}
	prefixes, err := os.Open(spotDataPath(root, "prefix_data.pl"))
	if err != nil {
		return nil
	}
	defer prefixes.Close()
	byEntity := make(map[int]string)
	scanner = bufio.NewScanner(prefixes)
	for scanner.Scan() {
		fields := prefixEntity.FindStringSubmatch(scanner.Text())
		if fields == nil {
			continue
		}
		code, err := strconv.Atoi(fields[2])
		if err != nil || byEntity[code] != "" {
			continue
		}
		if suffix := strings.LastIndexByte(fields[1], '-'); suffix >= 0 {
			if continent := byPrefix[fields[1][suffix+1:]]; continent != "" {
				byEntity[code] = continent
			}
		}
	}
	return byEntity
}

func registerClusterStats(mux *http.ServeMux, root string) {
	if root == "" {
		return
	}
	names, continents := spotEntities(root), spotContinents(root)
	var lock sync.Mutex
	var cache struct {
		GeneratedAt time.Time
		Date        string
		Spots       spotsReport
	}
	mux.HandleFunc("GET /api/cluster-stats", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		date := now.Format(time.DateOnly)
		lock.Lock()
		if cache.Date != date || now.Sub(cache.GeneratedAt) >= 30*time.Second {
			cache.Spots = analyzeSpots(root, 0, now, names, continents)
			cache.Date, cache.GeneratedAt = date, now
		}
		report := cache
		lock.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=15")
		json.NewEncoder(w).Encode(report)
	})
}
