package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
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
	Visitors  uint64 `json:",omitempty"`
	Referrers map[string]uint64
	Pages     map[string]uint64
	Languages map[string]uint64 `json:",omitempty"`
	Devices   map[string]uint64 `json:",omitempty"`
	Hours     []uint64          `json:",omitempty"` // 24 fasce orarie UTC
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
	salt      [16]byte
	seen      map[string]map[[16]byte]struct{} // visitatori già contati oggi, solo in memoria
}

func registerTraffic(mux *http.ServeMux, root string, auth *sysopAuth) (*traffic, error) {
	t := &traffic{data: trafficData{Days: make(map[string]trafficDay)}, active: make(map[[16]byte]time.Time),
		seen: make(map[string]map[[16]byte]struct{})}
	if _, err := rand.Read(t.salt[:]); err != nil {
		return nil, err
	}
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
	t.countAudience(&day, r, now, date)
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

// countAudience aggiorna visitatori unici, lingua, dispositivo e orario di una visita. I visitatori unici si
// riconoscono con un'impronta (indirizzo, browser, giorno, sale casuale) tenuta solo in memoria: niente cookie,
// niente indirizzi salvati, e il sale cambia a ogni riavvio.
func (t *traffic) countAudience(day *trafficDay, r *http.Request, now time.Time, date string) {
	if day.Languages == nil {
		day.Languages = make(map[string]uint64)
	}
	if day.Devices == nil {
		day.Devices = make(map[string]uint64)
	}
	if len(day.Hours) != 24 {
		day.Hours = make([]uint64, 24)
	}
	day.Languages[browserLanguage(r.Header.Get("Accept-Language"))]++
	device := deviceKind(r.UserAgent())
	day.Devices[device]++
	day.Hours[now.Hour()]++
	if device == "Bot" {
		return
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	sum := sha256.Sum256([]byte(string(t.salt[:]) + "|" + host + "|" + r.UserAgent() + "|" + date))
	var id [16]byte
	copy(id[:], sum[:16])
	for other := range t.seen {
		if other != date {
			delete(t.seen, other)
		}
	}
	if t.seen[date] == nil {
		t.seen[date] = make(map[[16]byte]struct{})
	}
	if _, known := t.seen[date][id]; !known && len(t.seen[date]) < 100000 {
		t.seen[date][id] = struct{}{}
		day.Visitors++
	}
}

// browserLanguage riduce l'intestazione Accept-Language alla prima lingua, se è tra quelle del portale.
func browserLanguage(header string) string {
	first := strings.TrimSpace(strings.Split(header, ",")[0])
	if code := normalizeLang(first); code != "" {
		return code
	}
	return "altre"
}

func deviceKind(agent string) string {
	lower := strings.ToLower(agent)
	switch {
	case lower == "" || strings.Contains(lower, "bot") || strings.Contains(lower, "spider") || strings.Contains(lower, "crawl") || strings.Contains(lower, "curl/"):
		return "Bot"
	case strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet") || (strings.Contains(lower, "android") && !strings.Contains(lower, "mobi")):
		return "Tablet"
	case strings.Contains(lower, "mobi") || strings.Contains(lower, "android") || strings.Contains(lower, "iphone"):
		return "Telefono"
	}
	return "Computer"
}
