package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Record storici di utenti collegati contemporaneamente (Telnet e web DXSpider).
//
// Ogni pochi secondi si conta chi è collegato dall'istantanea di presenza scritta dal motore; se il numero supera
// il record salvato, lo sostituisce con la data. I record non scadono e non si azzerano a mezzanotte: contano
// dall'installazione (o dalla prima versione che li raccoglie).

const recordsInterval = 5 * time.Second

type recordEntry struct {
	Count int   // massimo di collegamenti contemporanei
	At    int64 // quando è stato raggiunto (secondi Unix), 0 se mai
}

type recordsData struct {
	Telnet recordEntry
	Web    recordEntry
}

type records struct {
	mu   sync.Mutex
	path string
	data recordsData
}

func loadRecords(path string) *records {
	r := &records{path: path}
	file, err := os.Open(path)
	if err != nil {
		return r
	}
	defer file.Close()
	if json.NewDecoder(io.LimitReader(file, 64<<10)).Decode(&r.data) != nil {
		r.data = recordsData{}
	}
	return r
}

// observe confronta i collegamenti attuali con i record e salva solo se ne batte uno.
func (r *records) observe(telnet, web int, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	changed := false
	if telnet > r.data.Telnet.Count {
		r.data.Telnet, changed = recordEntry{Count: telnet, At: now.Unix()}, true
	}
	if web > r.data.Web.Count {
		r.data.Web, changed = recordEntry{Count: web, At: now.Unix()}, true
	}
	if changed && r.path != "" {
		if err := saveRecords(r.path, r.data); err != nil {
			log.Printf("Save connection records: %v", err)
		}
	}
	return changed
}

func (r *records) snapshot() recordsData {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.data
}

func saveRecords(path string, data recordsData) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".records-")
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

// countSessions conta i collegamenti utente dell'istantanea, con lo stesso criterio dell'elenco in dashboard.
func countSessions(snapshot presence) (telnet, web int) {
	for _, channel := range snapshot.Channels {
		switch channel.Kind {
		case "telnet":
			telnet++
		case "web":
			web++
		}
	}
	return telnet, web
}

func registerRecords(service context.Context, mux *http.ServeMux, root string, auth *sysopAuth) {
	if root == "" || auth == nil {
		return
	}
	presencePath := filepath.Join(root, "local_data", "dxregia-presence.json")
	store := loadRecords(filepath.Join(root, "local_data", "dxregia-records.json"))
	go func() {
		ticker := time.NewTicker(recordsInterval)
		defer ticker.Stop()
		for {
			select {
			case <-service.Done():
				return
			case now := <-ticker.C:
				// Un'istantanea vecchia (motore fermo) torna vuota: non cambia nulla.
				telnet, web := countSessions(readPresence(presencePath))
				store.observe(telnet, web, now)
			}
		}
	}()
	mux.HandleFunc("GET /api/records", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.snapshot())
	})
}
