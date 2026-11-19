package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

const propagationWindow = 30 * time.Minute

type propagationRoute struct {
	From, To, Band string
	Count          int
}

type propagationReport struct {
	At      time.Time
	Minutes int
	Routes  []propagationRoute
}

func propagationRoutes(root string, now time.Time, directory dxccDirectory) ([]propagationRoute, error) {
	type routeKey struct{ from, to, band string }
	counts := map[routeKey]int{}
	since := now.Add(-propagationWindow)
	for day, path := range spotArchivePaths(root, now) {
		if root == "" || (day != 0 && now.UTC().YearDay() == since.UTC().YearDay()) {
			break
		}
		spots, err := readSpotTail(path, 8000, 1<<20)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, spot := range spots {
			if spot.At < since.Unix() || spot.At > now.Unix() {
				continue
			}
			from, to := directory.entries[spot.SpotterDXCC].Continent, directory.entries[spot.DXCC].Continent
			band := analyticsBand(spot.Frequency)
			if from == "" || to == "" || band == "Altre" {
				continue
			}
			counts[routeKey{from, to, band}]++
		}
	}
	routes := make([]propagationRoute, 0, len(counts))
	for key, count := range counts {
		routes = append(routes, propagationRoute{key.from, key.to, key.band, count})
	}
	sort.Slice(routes, func(i, j int) bool {
		a, b := routes[i], routes[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		return a.Band < b.Band
	})
	return routes, nil
}

func registerPropagation(mux *http.ServeMux, root string, directory dxccDirectory) {
	var lock sync.Mutex
	var report propagationReport
	mux.HandleFunc("GET /api/propagation", func(w http.ResponseWriter, _ *http.Request) {
		lock.Lock()
		now := time.Now().UTC()
		if now.Sub(report.At) >= 30*time.Second {
			routes, err := propagationRoutes(root, now, directory)
			if err != nil {
				lock.Unlock()
				http.Error(w, "Spot archive unavailable.", http.StatusServiceUnavailable)
				return
			}
			report = propagationReport{now, int(propagationWindow.Minutes()), routes}
		}
		current := report
		lock.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=15")
		_ = json.NewEncoder(w).Encode(current)
	})
}
