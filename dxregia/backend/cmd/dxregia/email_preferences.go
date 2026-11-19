package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
)

type emailPreferences struct {
	Recipient          string
	PortalMessages     bool
	PartnerRequests    bool
	PartnerDown        bool
	PartnerRecovered   bool
	PartnerDownMinutes int
}

func emailPreferencesPath(root string) string {
	return filepath.Join(root, "local_data", "dxregia-email-preferences.json")
}

func readEmailPreferences(root string) (emailPreferences, error) {
	prefs := emailPreferences{PortalMessages: true, PartnerRequests: true, PartnerDownMinutes: 5}
	data, err := os.ReadFile(emailPreferencesPath(root))
	if os.IsNotExist(err) {
		return prefs, nil
	}
	if err == nil {
		err = json.Unmarshal(data, &prefs)
	}
	return prefs, err
}

func (p emailPreferences) valid() bool {
	if p.PartnerDownMinutes < 1 || p.PartnerDownMinutes > 1440 {
		return false
	}
	if p.Recipient == "" {
		return true
	}
	address, err := mail.ParseAddress(p.Recipient)
	return err == nil && address.Address == p.Recipient && len(p.Recipient) <= 254
}

func (p emailPreferences) destination(fallback string) string {
	if p.Recipient != "" {
		return p.Recipient
	}
	return fallback
}

func registerEmailPreferences(mux *http.ServeMux, root string, auth *sysopAuth) {
	if root == "" || auth == nil {
		return
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Method == http.MethodGet {
			prefs, err := readEmailPreferences(root)
			if err != nil {
				http.Error(w, "Impossibile leggere le preferenze email.", http.StatusInternalServerError)
				return
			}
			var node nodeConfig
			if data, err := os.ReadFile(filepath.Join(root, "local", "dxregia-node.json")); err == nil {
				_ = json.Unmarshal(data, &node)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(struct {
				emailPreferences
				SysopEmail string
				SMTPReady  bool
			}{prefs, node.Email, currentMailer(root) != nil})
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return
		}
		var prefs emailPreferences
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&prefs) != nil || decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "Preferenze email non valide.", http.StatusBadRequest)
			return
		}
		prefs.Recipient = strings.TrimSpace(prefs.Recipient)
		if !prefs.valid() {
			http.Error(w, "Controlla il destinatario e scegli da 1 a 1440 minuti.", http.StatusBadRequest)
			return
		}
		if err := os.MkdirAll(filepath.Dir(emailPreferencesPath(root)), 0700); err != nil || writePrivateJSON(emailPreferencesPath(root), prefs) != nil {
			http.Error(w, "Impossibile salvare le preferenze email.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
	mux.HandleFunc("GET /api/email-preferences", handler)
	mux.HandleFunc("PUT /api/email-preferences", handler)
}
