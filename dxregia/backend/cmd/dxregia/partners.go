package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type partnerDetails struct {
	Call, Name, Contact, Email, Host string
	Port                             int
}

type partnerNote struct {
	Text string
	At   time.Time
}

type partner struct {
	partnerDetails
	RegisteredAt     time.Time
	FirstConnectedAt *time.Time
	Notes            []partnerNote
}

var partnerHost = regexp.MustCompile("^[a-zA-Z0-9]([a-zA-Z0-9.-]*[a-zA-Z0-9])?$")

func (d *partnerDetails) validate() bool {
	d.Call = strings.ToUpper(strings.TrimSpace(d.Call))
	d.Name = strings.TrimSpace(d.Name)
	d.Contact = strings.TrimSpace(d.Contact)
	d.Email = strings.TrimSpace(d.Email)
	d.Host = strings.TrimSpace(d.Host)
	if d.Email != "" {
		address, err := mail.ParseAddress(d.Email)
		if err != nil || address.Address != d.Email || len(d.Email) > 254 {
			return false
		}
	}
	return len(d.Call) <= 12 && callsign.MatchString(d.Call) && len(d.Name) <= 120 &&
		len(d.Contact) <= 120 && len(d.Host) <= 253 &&
		(net.ParseIP(d.Host) != nil || partnerHost.MatchString(d.Host)) && d.Port >= 1 && d.Port <= 65535
}

func registerPartners(mux *http.ServeMux, root string, auth *sysopAuth) error {
	if auth == nil {
		return nil
	}
	path := filepath.Join(root, ".dxregia-partners.json")
	partners := make(map[string]partner)
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &partners); err != nil {
			return err
		}
		if partners == nil {
			partners = make(map[string]partner)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	var lock sync.Mutex
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(partners)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		var input struct {
			partnerDetails
			Note string
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF || !input.partnerDetails.validate() || len(input.Note) > 4000 {
			http.Error(w, "Controlla indicativo, host, porta, email e lunghezza dei campi.", http.StatusBadRequest)
			return
		}
		call := input.Call
		old, exists := partners[call]
		if r.Method == http.MethodPost && exists {
			http.Error(w, "Indicativo già registrato.", http.StatusConflict)
			return
		}
		if r.Method == http.MethodPut && (!exists || r.PathValue("call") != call) {
			http.Error(w, "Nodo non trovato.", http.StatusNotFound)
			return
		}
		entry := old
		now := time.Now().UTC()
		if !exists {
			entry.RegisteredAt = now
		}
		entry.partnerDetails = input.partnerDetails
		if note := strings.TrimSpace(input.Note); note != "" {
			entry.Notes = append(entry.Notes, partnerNote{Text: note, At: now})
		}
		partners[call] = entry
		if err := savePartners(path, partners); err != nil {
			if exists {
				partners[call] = old
			} else {
				delete(partners, call)
			}
			http.Error(w, "Impossibile salvare l'anagrafica.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entry)
	}
	mux.HandleFunc("GET /api/partners", handler)
	mux.HandleFunc("POST /api/partners", handler)
	mux.HandleFunc("PUT /api/partners/{call}", handler)
	return nil
}

func savePartners(path string, partners map[string]partner) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".partners-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	err = json.NewEncoder(tmp).Encode(partners)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
