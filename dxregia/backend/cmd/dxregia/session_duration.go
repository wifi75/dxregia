package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Durata della sessione del pannello: la sceglie il sysop, il valore predefinito è 8 ore.

const defaultSessionHours = 8

var sessionHourChoices = []int{1, 2, 4, 8, 12, 24, 72, 168}

func validSessionHours(hours int) bool {
	for _, choice := range sessionHourChoices {
		if choice == hours {
			return true
		}
	}
	return false
}

func (a *sysopAuth) sessionFile() string { return filepath.Join(a.root, "local", "dxregia-session.json") }

func (a *sysopAuth) sessionHours() int {
	data, err := os.ReadFile(a.sessionFile())
	var stored struct{ Hours int }
	if err != nil || json.Unmarshal(data, &stored) != nil || !validSessionHours(stored.Hours) {
		return defaultSessionHours
	}
	return stored.Hours
}

func (a *sysopAuth) registerSessionDuration(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/session-duration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !a.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		a.mu.Lock()
		hours := a.sessionHours()
		a.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"Hours": hours, "Choices": sessionHourChoices})
	})
	mux.HandleFunc("PUT /api/session-duration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return
		}
		var change struct{ Hours int }
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&change) != nil || decoder.Decode(new(any)) != io.EOF || !validSessionHours(change.Hours) {
			http.Error(w, "Durata non valida.", http.StatusBadRequest)
			return
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if !(a.signedIn(r) && !a.temporary && a.factorOK) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		encoded, _ := json.Marshal(map[string]int{"Hours": change.Hours})
		if err := os.WriteFile(a.sessionFile(), encoded, 0600); err != nil {
			http.Error(w, "Impossibile salvare la durata.", http.StatusInternalServerError)
			return
		}
		a.expires = time.Now().Add(time.Duration(change.Hours) * time.Hour)
		w.WriteHeader(http.StatusNoContent)
	})
}
