package main

import (
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:embed presence.pm
var presenceHook []byte

type presenceChannel struct {
	Call  string
	Kind  string
	State string
	Since int64
}

type presence struct {
	At       int64
	Channels []presenceChannel
}

func registerPresence(mux *http.ServeMux, root string, auth *sysopAuth) {
	if auth == nil {
		return
	}
	path := filepath.Join(root, "local_data", "dxregia-presence.json")
	mux.HandleFunc("GET /api/presence", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(readPresence(path))
	})
}

func readPresence(path string) presence {
	var snapshot presence
	file, err := os.Open(path)
	if err != nil {
		return snapshot
	}
	defer file.Close()
	if err := json.NewDecoder(io.LimitReader(file, 4<<20)).Decode(&snapshot); err != nil {
		return presence{}
	}
	age := time.Since(time.Unix(snapshot.At, 0))
	if age < 0 || age > 15*time.Second {
		return presence{}
	}
	return snapshot
}

func installPresenceHook(local string) error {
	path := filepath.Join(local, "Local.pm")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = f.Write(presenceHook)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(path)
	}
	return err
}
