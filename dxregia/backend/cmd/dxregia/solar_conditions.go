package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

const hamQSLFeed = "https://www.hamqsl.com/solarxml.php"
const hamQSLInterval = time.Hour
const maxHamQSLBytes = 64 << 10

type hfCondition struct {
	Band, Day, Night string
}

type vhfCondition struct {
	Name, Location, Condition string
}

type solarConditions struct {
	UpdatedAt                                                     time.Time
	SFI, A, K, Sunspots, AuroraLatitude                           string
	SolarWind, MagneticField, GeomagneticField, SignalNoise, FOF2 string
	HF                                                            []hfCondition
	VHF                                                           []vhfCondition
}

func parseHamQSL(body []byte) (solarConditions, error) {
	var feed struct {
		XMLName xml.Name `xml:"solar"`
		Data    struct {
			Updated          string `xml:"updated"`
			SFI              string `xml:"solarflux"`
			A                string `xml:"aindex"`
			K                string `xml:"kindex"`
			Sunspots         string `xml:"sunspots"`
			AuroraLatitude   string `xml:"latdegree"`
			SolarWind        string `xml:"solarwind"`
			MagneticField    string `xml:"magneticfield"`
			GeomagneticField string `xml:"geomagfield"`
			SignalNoise      string `xml:"signalnoise"`
			FOF2             string `xml:"fof2"`
			HF               []struct {
				Band  string `xml:"name,attr"`
				Time  string `xml:"time,attr"`
				Value string `xml:",chardata"`
			} `xml:"calculatedconditions>band"`
			VHF []struct {
				Name     string `xml:"name,attr"`
				Location string `xml:"location,attr"`
				Value    string `xml:",chardata"`
			} `xml:"calculatedvhfconditions>phenomenon"`
		} `xml:"solardata"`
	}
	if err := xml.Unmarshal(body, &feed); err != nil {
		return solarConditions{}, err
	}
	data := feed.Data
	at, err := time.Parse("2 Jan 2006 1504 MST", strings.TrimSpace(data.Updated))
	if err != nil {
		return solarConditions{}, fmt.Errorf("invalid HamQSL timestamp: %w", err)
	}
	clean := strings.TrimSpace
	out := solarConditions{
		UpdatedAt: at.UTC(), SFI: clean(data.SFI), A: clean(data.A), K: clean(data.K),
		Sunspots: clean(data.Sunspots), AuroraLatitude: clean(data.AuroraLatitude), SolarWind: clean(data.SolarWind),
		MagneticField: clean(data.MagneticField), GeomagneticField: clean(data.GeomagneticField),
		SignalNoise: clean(data.SignalNoise), FOF2: clean(data.FOF2),
		HF: []hfCondition{}, VHF: []vhfCondition{},
	}
	for _, band := range []string{"80m-40m", "30m-20m", "17m-15m", "12m-10m"} {
		row := hfCondition{Band: band}
		for _, source := range data.HF {
			if source.Band != band {
				continue
			}
			switch source.Time {
			case "day":
				row.Day = clean(source.Value)
			case "night":
				row.Night = clean(source.Value)
			}
		}
		if row.Day != "" || row.Night != "" {
			out.HF = append(out.HF, row)
		}
	}
	for _, row := range data.VHF {
		if len(out.VHF) >= 25 {
			break
		}
		out.VHF = append(out.VHF, vhfCondition{clean(row.Name), clean(row.Location), clean(row.Value)})
	}
	return out, nil
}

func registerSolarConditions(ctx context.Context, mux *http.ServeMux, root string) {
	var current atomic.Pointer[solarConditions]
	mux.HandleFunc("GET /api/solar-conditions", func(w http.ResponseWriter, _ *http.Request) {
		if latest := current.Load(); latest != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "public, max-age=300")
			_ = json.NewEncoder(w).Encode(latest)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		http.Error(w, "Solar data unavailable.", http.StatusServiceUnavailable)
	})
	if root == "" || os.Getenv("DXREGIA_HAMQSL") == "off" {
		return
	}
	cache := filepath.Join(root, "local_data", "dxregia-hamqsl.xml")
	due := time.Now()
	if info, err := os.Stat(cache); err == nil && info.Size() <= maxHamQSLBytes {
		if cached, err := os.ReadFile(cache); err == nil {
			if report, err := parseHamQSL(cached); err == nil {
				current.Store(&report)
				due = info.ModTime().Add(hamQSLInterval)
			}
		}
	}
	go func() {
		client := &http.Client{Timeout: 15 * time.Second}
		for {
			timer := time.NewTimer(time.Until(due))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			due = time.Now().Add(hamQSLInterval)
			body, err := readHamQSL(ctx, client)
			if err == nil {
				var report solarConditions
				report, err = parseHamQSL(body)
				if err == nil {
					current.Store(&report)
					err = writeNodeFile(cache, body)
				}
			}
			if err != nil && ctx.Err() == nil {
				log.Printf("HamQSL: %v", err)
			}
		}
	}()
}

func readHamQSL(ctx context.Context, client *http.Client) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, hamQSLFeed, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "DXRegia (+https://github.com/wifi75/dxregia)")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxHamQSLBytes+1))
	if len(body) > maxHamQSLBytes {
		return nil, fmt.Errorf("solar feed exceeds size limit")
	}
	return body, err
}
