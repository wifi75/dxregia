package main

import (
	"context"
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var managedEnginePID atomic.Int64

// Avvio del servizio e del motore gestito (secondi Unix) e riavvii automatici del motore dopo un arresto anomalo.
var (
	serviceStartedAt     = time.Now().Unix()
	managedEngineStarted atomic.Int64
	mojoAutoRestarts     atomic.Int64
)

type resourcePoint struct {
	At                                           int64
	DiskUsed, DiskTotal, MemoryUsed, MemoryTotal uint64
}

type storageSample struct {
	MemoryFreeBytes, MemoryTotalBytes uint64
	MemoryAvailable                   bool

	At                              time.Time
	LogBytes, FreeBytes, TotalBytes uint64
	GrowthBytes                     int64
	GrowthIntervalSeconds           int64
	DiskAvailable, LogsAvailable    bool
}

var historyFile = regexp.MustCompile(`^[0-9]{20}-[0-9a-f]{48}[.]json$`)
var datedLogFile = regexp.MustCompile(`^[0-9]{4}/[0-9]{2,3}[.]dat$`)

func scanEventStorage(root string, now time.Time, retention int) (uint64, bool) {
	base := filepath.Join(root, "local_data")
	var bytes uint64
	ok := true
	expiration := now.AddDate(0, 0, -retention)
	for _, group := range []string{"log", "debug", "dxregia-partner-mail"} {
		dir := filepath.Join(base, group)
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				ok = false
				return nil
			}
			if !entry.Type().IsRegular() {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				ok = false
				return nil
			}
			relative, _ := filepath.Rel(dir, path)
			relative = filepath.ToSlash(relative)
			eligible := group == "dxregia-partner-mail" && historyFile.MatchString(relative)
			if group == "debug" && datedLogFile.MatchString(relative) {
				date, err := time.Parse("2006/002.dat", relative)
				eligible = err == nil && date.Before(expiration.Truncate(24*time.Hour))
			}
			if group == "log" && datedLogFile.MatchString(relative) {
				date, err := time.Parse("2006/01.dat", relative)
				eligible = err == nil && !date.AddDate(0, 1, 0).After(expiration)
			}
			if retention > 0 && eligible && info.ModTime().Before(expiration) {
				if err := os.Remove(path); err == nil || os.IsNotExist(err) {
					return nil
				} else {
					ok = false
					log.Printf("Event retention: %v", err)
				}
			}
			bytes += uint64(info.Size())
			return nil
		})
		ok = ok && err == nil
	}
	if info, err := os.Stat(filepath.Join(base, "dxregia-startup.log")); err == nil && info.Mode().IsRegular() {
		bytes += uint64(info.Size())
	}
	return bytes, ok
}

func registerNodeHealth(ctx context.Context, mux *http.ServeMux, root string, auth *sysopAuth, spots func() spotSnapshot) {
	if root != "" && auth != nil {
		var mu sync.Mutex
		var storage storageSample
		var resources []resourcePoint
		retention, _ := strconv.Atoi(os.Getenv("DXREGIA_EVENT_RETENTION_DAYS"))
		if retention < 1 || retention > 3650 {
			retention = 0
		}
		go func() {
			storageTick := time.NewTicker(30 * time.Second)
			defer storageTick.Stop()
			lastPrune := time.Time{}
			for {
				if ctx.Err() != nil {
					return
				}
				now := time.Now().UTC()
				days := 0
				if retention > 0 && now.Sub(lastPrune) >= time.Hour {
					days, lastPrune = retention, now
				}
				bytes, ok := scanEventStorage(root, now, days)
				free, total, diskOK := diskSpace(root)
				memFree, memTotal, memOK := serverMemory()
				point := resourcePoint{At: now.Unix()}
				if diskOK && free <= total {
					point.DiskUsed, point.DiskTotal = total-free, total
				}
				if memOK {
					point.MemoryUsed, point.MemoryTotal = memTotal-memFree, memTotal
				}
				mu.Lock()
				next := storageSample{At: now, LogBytes: bytes, FreeBytes: free, TotalBytes: total, DiskAvailable: diskOK, LogsAvailable: ok,
					MemoryFreeBytes: memFree, MemoryTotalBytes: memTotal, MemoryAvailable: memOK}
				if !storage.At.IsZero() && storage.LogsAvailable && ok {
					next.GrowthBytes = int64(bytes) - int64(storage.LogBytes)
					next.GrowthIntervalSeconds = int64(now.Sub(storage.At).Seconds())
				}
				storage = next
				if len(resources) == 120 {
					copy(resources, resources[1:])
					resources = resources[:119]
				}
				resources = append(resources, point)
				mu.Unlock()
				select {
				case <-ctx.Done():
					return
				case <-storageTick.C:
				}
			}
		}()
		mux.HandleFunc("GET /api/node-health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			if !auth.allowed(r) {
				http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
				return
			}
			mu.Lock()
			sample := storage
			history := append([]resourcePoint(nil), resources...)
			mu.Unlock()
			data := struct {
				storageSample
				DXRegiaPID                int
				DXSpiderPID, LatestSpotAt int64
				ServiceStartedAt, EngineStartedAt, EngineRestarts int64
				RetentionDays             int
				Resources                 []resourcePoint
			}{
				storageSample: sample, DXRegiaPID: os.Getpid(), DXSpiderPID: managedEnginePID.Load(),
				ServiceStartedAt: serviceStartedAt, EngineStartedAt: managedEngineStarted.Load(), EngineRestarts: mojoAutoRestarts.Load(),
				LatestSpotAt:  spots().LatestAt,
				RetentionDays: retention,
				Resources:     history,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(data)
		})
	}
	if os.Getenv("DXREGIA_PUBLIC_STATUS") == "on" && root != "" {
		mux.HandleFunc("GET /stato", func(w http.ResponseWriter, r *http.Request) {
			state := readPresence(filepath.Join(root, "local_data", "dxregia-presence.json"))
			latest := spots().LatestAt
			data := struct {
				Online bool
				Spot   string
			}{Online: state.At != 0}
			if latest > 0 {
				data.Spot = time.Unix(latest, 0).UTC().Format(time.RFC3339)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			if !data.Online {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			_ = publicHealthPage.Execute(w, data)
		})
	}
}

var publicHealthPage = template.Must(template.New("status").Parse(`<!doctype html>
<html lang="it"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Stato del cluster · DXRegia</title><style>
body{font:1rem/1.6 system-ui,sans-serif;background:#edf4f8;color:#17384c;margin:0;padding:clamp(16px,5vw,60px)}
main{max-width:48rem;margin:auto;padding:clamp(18px,4vw,36px);background:#fff;border:1px solid #a9c3d0;border-radius:16px}a{color:#176687}
</style></head><body><main><h1>Stato del cluster</h1>
<p>Motore DXSpider · <strong>{{if .Online}}Operativo{{else}}Stato non disponibile{{end}}</strong></p>
{{if .Spot}}<p>Ultimo spot · <time datetime="{{.Spot}}">{{.Spot}}</time></p>{{end}}
<a href="/spots">Torna agli spot</a></main></body></html>`))
