package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type trafficDay struct {
	Views     uint64
	Referrers map[string]uint64
}

type trafficData struct {
	Views uint64
	Days  map[string]trafficDay
}

type traffic struct {
	mu   sync.Mutex
	path string
	data trafficData
}

func registerTraffic(mux *http.ServeMux, root string, auth *sysopAuth) (*traffic, error) {
	t := &traffic{data: trafficData{Days: make(map[string]trafficDay)}}
	if root != "" {
		t.path = filepath.Join(root, ".dxregia-traffic.json")
		data, err := os.ReadFile(t.path)
		if err == nil {
			if err := json.Unmarshal(data, &t.data); err != nil {
				return nil, err
			}
			if t.data.Days == nil {
				t.data.Days = make(map[string]trafficDay)
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	mux.HandleFunc("GET /api/visits", func(w http.ResponseWriter, r *http.Request) {
		t.mu.Lock()
		defer t.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(map[string]uint64{"Views": t.data.Views})
	})
	if auth != nil {
		mux.HandleFunc("GET /api/traffic", func(w http.ResponseWriter, r *http.Request) {
			if !auth.allowed(r) {
				http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
				return
			}
			t.mu.Lock()
			defer t.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			json.NewEncoder(w).Encode(t.data)
		})
	}
	return t, nil
}

func (t *traffic) view(r *http.Request) {
	if r.Method != http.MethodGet || t.path == "" {
		return
	}
	referrer := "Diretto o non disponibile"
	if ref, err := url.Parse(r.Referer()); err == nil && (ref.Scheme == "http" || ref.Scheme == "https") {
		host := strings.ToLower(ref.Hostname())
		requestHost := (&url.URL{Host: r.Host}).Hostname()
		if strings.EqualFold(host, requestHost) {
			referrer = "Interno"
		} else if len(host) > 0 && len(host) <= 253 {
			referrer = host
		}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.data.Views++
	now := time.Now().UTC()
	date := now.Format(time.DateOnly)
	day := t.data.Days[date]
	if day.Referrers == nil {
		day.Referrers = make(map[string]uint64)
	}
	if _, known := day.Referrers[referrer]; !known && len(day.Referrers) >= 50 {
		referrer = "Altri domini"
	}
	day.Referrers[referrer]++
	day.Views++
	t.data.Days[date] = day
	cutoff := now.AddDate(0, 0, -89).Format(time.DateOnly)
	for date := range t.data.Days {
		if date < cutoff {
			delete(t.data.Days, date)
		}
	}
	if err := saveTraffic(t.path, t.data); err != nil {
		log.Printf("Save web traffic: %v", err)
	}
}

func saveTraffic(path string, data trafficData) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".traffic-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	err = json.NewEncoder(tmp).Encode(data)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
