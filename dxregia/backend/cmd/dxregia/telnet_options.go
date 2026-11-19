package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type telnetOptions struct {
	PasswordRequired bool
	SlashCallsigns   bool
	ReplaceSession   bool
	Registration     int
	ConcurrentUsers  int
	ConcurrentNodes  int
	MultipleSessions int
	ReconnectSeconds int
	MaxSSID          int
}

func (o telnetOptions) valid() bool {
	return o.Registration >= 0 && o.Registration <= 2 &&
		o.ConcurrentUsers >= 0 && o.ConcurrentUsers <= 1000 &&
		o.ConcurrentNodes >= 0 && o.ConcurrentNodes <= 1000 &&
		o.MultipleSessions >= 0 && o.MultipleSessions <= 15 &&
		o.ReconnectSeconds >= 0 && o.ReconnectSeconds <= 86400 &&
		o.MaxSSID >= 0 && o.MaxSSID <= 15
}

func registerTelnetOptions(mux *http.ServeMux, auth *sysopAuth, root string) {
	if root == "" || auth == nil {
		return
	}
	var lock sync.Mutex
	path := filepath.Join(root, "local", "dxregia-telnet.json")
	presencePath := filepath.Join(root, "local_data", "dxregia-presence.json")
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		snapshot := readPresence(presencePath)
		if snapshot.Telnet == nil {
			http.Error(w, "Opzioni non disponibili. Verifica che DXSpider sia operativo e aggiornato.", http.StatusServiceUnavailable)
			return
		}
		desired := *snapshot.Telnet
		if data, err := os.ReadFile(path); err == nil {
			if json.Unmarshal(data, &desired) != nil || !desired.valid() {
				http.Error(w, "Configurazione Telnet non valida.", http.StatusInternalServerError)
				return
			}
		} else if !os.IsNotExist(err) {
			http.Error(w, "Impossibile leggere la configurazione Telnet.", http.StatusInternalServerError)
			return
		}
		if r.Method == http.MethodPut {
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
				return
			}
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
			decoder.DisallowUnknownFields()
			if decoder.Decode(&desired) != nil || decoder.Decode(new(any)) != io.EOF || !desired.valid() {
				http.Error(w, "Controlla i valori delle opzioni Telnet.", http.StatusBadRequest)
				return
			}
			if err := saveTelnetOptions(path, desired); err != nil {
				http.Error(w, "Impossibile salvare le opzioni Telnet.", http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Options telnetOptions
			Applied bool
		}{desired, desired == *snapshot.Telnet})
	}
	mux.HandleFunc("GET /api/telnet-options", handler)
	mux.HandleFunc("PUT /api/telnet-options", handler)
}

func saveTelnetOptions(path string, options telnetOptions) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".dxregia-telnet-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = json.NewEncoder(file).Encode(options)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
