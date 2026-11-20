package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type rbnConfig struct {
	Enabled     bool
	InboundHash string
}

type rbnSpot struct {
	Call      string
	Skimmer   string
	Mode      string
	Frequency float64
	At        int64
	SNR       int
	Quality   int
}

type rbnSnapshot struct {
	At    int64
	State string
	Spots []rbnSpot
}

func registerRBN(mux *http.ServeMux, root string, auth *sysopAuth) {
	if root == "" {
		return
	}
	configPath := filepath.Join(root, "local", "dxregia-rbn.json")
	snapshotPath := filepath.Join(root, "local_data", "dxregia-rbn.json")
	var updateLock sync.Mutex
	serve := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodPut {
			if !auth.allowed(r) {
				http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
				return
			}
			updateLock.Lock()
			defer updateLock.Unlock()
		}
		config, err := readRBNConfig(configPath)
		if err != nil {
			http.Error(w, "Impossibile leggere la configurazione RBN.", http.StatusInternalServerError)
			return
		}
		if r.Method == http.MethodPut {
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
				return
			}
			var update struct{ Enabled *bool }
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128))
			decoder.DisallowUnknownFields()
			if decoder.Decode(&update) != nil || update.Enabled == nil || decoder.Decode(new(any)) != io.EOF {
				http.Error(w, "Impostazione RBN non valida.", http.StatusBadRequest)
				return
			}
			config.Enabled = *update.Enabled
			if err := saveRBNConfig(configPath, config); err != nil {
				http.Error(w, "Impossibile salvare la configurazione RBN.", http.StatusInternalServerError)
				return
			}
		}
		snapshot := rbnSnapshot{State: "disabled", Spots: []rbnSpot{}}
		if config.Enabled {
			snapshot = readRBNSnapshot(snapshotPath, time.Now())
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Enabled bool
			rbnSnapshot
		}{config.Enabled, snapshot})
	}
	mux.HandleFunc("GET /api/rbn", serve)
	if auth != nil {
		mux.HandleFunc("PUT /api/rbn", serve)
	}
}

func readRBNConfig(path string) (rbnConfig, error) {
	var config rbnConfig
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return config, nil
	}
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func saveRBNConfig(path string, config rbnConfig) error {
	if _, err := bcrypt.Cost([]byte(config.InboundHash)); config.Enabled && err != nil {
		// Only the hash is retained. The random credential is never used to
		// connect to the public feed and is unavailable to inbound clients.
		var secret [32]byte
		if _, err := rand.Read(secret[:]); err != nil {
			return err
		}
		hash, err := bcrypt.GenerateFromPassword(secret[:], bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		config.InboundHash = string(hash)
	}
	return writePrivateJSON(path, config)
}

func readRBNSnapshot(path string, now time.Time) rbnSnapshot {
	waiting := rbnSnapshot{State: "waiting", Spots: []rbnSpot{}}
	file, err := os.Open(path)
	if err != nil {
		return waiting
	}
	defer file.Close()
	var snapshot rbnSnapshot
	decoder := json.NewDecoder(io.LimitReader(file, 64<<10))
	if decoder.Decode(&snapshot) != nil || decoder.Decode(new(any)) != io.EOF {
		return waiting
	}
	if age := now.Sub(time.Unix(snapshot.At, 0)); age < 0 || age > 20*time.Second {
		return waiting
	}
	switch snapshot.State {
	case "connecting", "connected", "conflict", "error":
	default:
		return waiting
	}
	snapshot.Spots = slices.DeleteFunc(snapshot.Spots, func(spot rbnSpot) bool {
		age := now.Sub(time.Unix(spot.At, 0))
		return age > 30*time.Minute || age < -time.Minute || (spot.Mode != "CW" && spot.Mode != "RTTY")
	})
	if snapshot.Spots == nil {
		snapshot.Spots = []rbnSpot{}
	}
	return snapshot
}
