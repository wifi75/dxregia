package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func registerProfile(mux *http.ServeMux, local string, auth *sysopAuth) {
	path := filepath.Join(local, "dxregia-node.json")
	var lock sync.Mutex
	mux.HandleFunc("PUT /api/node-settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		var config nodeConfig
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&config) != nil || decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "Impostazioni non valide.", http.StatusBadRequest)
			return
		}
		if err := config.validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		data, err := os.ReadFile(path)
		var existing nodeConfig
		if err != nil || json.Unmarshal(data, &existing) != nil {
			http.Error(w, "Prima configurazione richiesta.", http.StatusConflict)
			return
		}
		if existing.Node != config.Node || existing.Sysop != config.Sysop || existing.Port != config.Port {
			http.Error(w, "Indicativi e porta Telnet richiedono una migrazione del nodo.", http.StatusConflict)
			return
		}
		hook, err := os.ReadFile(filepath.Join(local, "Local.pm"))
		if err != nil || !bytes.Equal(hook, presenceHook) {
			http.Error(w, "Serve l’hook Mojo gestito da DXRegia per aggiornare il profilo.", http.StatusConflict)
			return
		}
		if err := saveNodeProfile(path, config); err != nil {
			http.Error(w, "Salvataggio del profilo non riuscito.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	})
}

func saveNodeProfile(path string, config nodeConfig) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".node-profile-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := errors.Join(json.NewEncoder(tmp).Encode(config), tmp.Close()); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
