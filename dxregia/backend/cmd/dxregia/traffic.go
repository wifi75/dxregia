package main

import (
	"encoding/hex"
	"encoding/json"
	"io"
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
	Pages     map[string]uint64
}

type trafficData struct {
	Views uint64
	Days  map[string]trafficDay
}

type traffic struct {
	mu        sync.Mutex
	path      string
	data      trafficData
	active    map[[16]byte]time.Time
	nextSweep time.Time
}

func registerTraffic(mux *http.ServeMux, root string, auth *sysopAuth) (*traffic, error) {
	t := &traffic{data: trafficData{Days: make(map[string]trafficDay)}, active: make(map[[16]byte]time.Time)}
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
	if root != "" {
		mux.HandleFunc("POST /api/visits/heartbeat", t.heartbeat)
	}
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
			t.sweep(time.Now(), true)
			json.NewEncoder(w).Encode(map[string]any{
				"Views": t.data.Views, "Days": t.data.Days, "Active": len(t.active),
			})
		})
	}
	return t, nil
}

func (t *traffic) heartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
		return
	}
	var token string
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128))
	if decoder.Decode(&token) != nil || decoder.Decode(new(any)) != io.EOF || len(token) != 32 {
		http.Error(w, "Identificativo non valido.", http.StatusBadRequest)
		return
	}
	var id [16]byte
	if _, err := hex.Decode(id[:], []byte(token)); err != nil {
		http.Error(w, "Identificativo non valido.", http.StatusBadRequest)
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	t.sweep(now, false)
	if _, exists := t.active[id]; !exists && len(t.active) >= 4096 {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	t.active[id] = now.Add(90 * time.Second)
	w.WriteHeader(http.StatusNoContent)
}

func (t *traffic) sweep(now time.Time, force bool) {
	if !force && now.Before(t.nextSweep) {
		return
	}
	for id, until := range t.active {
		if !now.Before(until) {
			delete(t.active, id)
		}
	}
	t.nextSweep = now.Add(20 * time.Second)
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
	if day.Pages == nil {
		day.Pages = make(map[string]uint64)
		if day.Views > 0 {
			day.Pages["/spots"] = day.Views
		}
	}
	day.Pages[r.URL.Path]++
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
