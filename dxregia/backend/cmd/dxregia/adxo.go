package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Spedizioni DX annunciate: il feed pubblico ADXO di NG3K (https://www.ng3k.com/Misc/adxo.html).
// Il nodo lo scarica ogni 6 ore e lo mostra accanto ai nominativi corrispondenti, citando la fonte.
// Si disattiva con DXREGIA_ADXO=off.

const (
	adxoFeed     = "https://www.ng3k.com/adxo.xml"
	adxoSource   = "https://www.ng3k.com/Misc/adxo.html"
	adxoInterval = 6 * time.Hour
	adxoMaxBytes = 1 << 20
)

type adxoEntry struct {
	Call   string
	Entity string
	Dates  string
	QSL    string
	Info   string
	Link   string
}

type adxoData struct {
	Updated time.Time
	Source  string
	Entries []adxoEntry
}

var adxoTitle = regexp.MustCompile(`^(.*?): (.*?) -- (.*?)(?: -- QSL via: (.*))?$`)

// parseADXO legge il feed RSS e ne ricava le voci. Le voci che non seguono il formato atteso si scartano.
func parseADXO(data []byte) ([]adxoEntry, error) {
	var feed struct {
		Items []struct {
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Link        string `xml:"link"`
		} `xml:"channel>item"`
	}
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}
	entries := make([]adxoEntry, 0, len(feed.Items))
	for _, item := range feed.Items {
		match := adxoTitle.FindStringSubmatch(strings.TrimSpace(item.Title))
		if match == nil || len(entries) >= 600 {
			continue
		}
		call := strings.ToUpper(strings.TrimSpace(match[3]))
		if call == "" || len(call) > 40 {
			continue
		}
		info := strings.Join(strings.Fields(strings.ReplaceAll(item.Description, "--", "·")), " ")
		if len(info) > 700 {
			info = info[:700] + "…"
		}
		link := strings.TrimSpace(item.Link)
		if !strings.HasPrefix(link, "https://") && !strings.HasPrefix(link, "http://") {
			link = adxoSource
		}
		entries = append(entries, adxoEntry{Call: call, Entity: strings.TrimSpace(match[1]), Dates: strings.TrimSpace(match[2]),
			QSL: strings.TrimSpace(match[4]), Info: info, Link: link})
	}
	return entries, nil
}

func registerADXO(ctx context.Context, mux *http.ServeMux, root string) {
	if root == "" || os.Getenv("DXREGIA_ADXO") == "off" {
		return
	}
	path := filepath.Join(root, "local_data", "dxregia-adxo.json")
	var lock sync.Mutex
	current := adxoData{Source: adxoSource}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &current)
		current.Source = adxoSource
	}
	refresh := func() {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, adxoFeed, nil)
		if err != nil {
			return
		}
		request.Header.Set("User-Agent", "DXRegia (+https://github.com/wifi75/dxregia)")
		client := &http.Client{Timeout: 20 * time.Second}
		response, err := client.Do(request)
		if err != nil {
			log.Printf("ADXO: %v", err)
			return
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			log.Printf("ADXO: status %d", response.StatusCode)
			return
		}
		body, err := io.ReadAll(io.LimitReader(response.Body, adxoMaxBytes))
		if err != nil {
			return
		}
		entries, err := parseADXO(body)
		if err != nil || len(entries) == 0 {
			log.Printf("ADXO: feed not understood (%v)", err)
			return
		}
		fresh := adxoData{Updated: time.Now().UTC(), Source: adxoSource, Entries: entries}
		lock.Lock()
		current = fresh
		lock.Unlock()
		if encoded, err := json.Marshal(fresh); err == nil {
			_ = os.WriteFile(path, encoded, 0600)
		}
	}
	go func() {
		lock.Lock()
		stale := time.Since(current.Updated) > adxoInterval
		lock.Unlock()
		if stale {
			refresh()
		}
		ticker := time.NewTicker(adxoInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				refresh()
			}
		}
	}()
	mux.HandleFunc("GET /api/adxo", func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		snapshot := current
		lock.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=900")
		json.NewEncoder(w).Encode(snapshot)
	})
}
