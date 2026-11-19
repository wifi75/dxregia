package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (a *sysopAuth) factorEnabled() bool {
	return a.factor.Key != "" || a.factor.ResetPending
}

func (a *sysopAuth) registerFactorSettings(mux *http.ServeMux) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		a.mu.Lock()
		defer a.mu.Unlock()
		if !a.signedIn(r) || a.temporary || !a.factorOK {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"Enabled": a.factorEnabled()})
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return
		}
		var change struct {
			Enabled        *bool
			Password, Code string
		}
		body := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		body.DisallowUnknownFields()
		if body.Decode(&change) != nil || body.Decode(new(any)) != io.EOF || change.Enabled == nil || len(change.Password) > 72 {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return
		}
		if time.Now().Before(a.nextFactor) {
			http.Error(w, "Attendi prima di riprovare.", http.StatusTooManyRequests)
			return
		}
		if bcrypt.CompareHashAndPassword(a.hash, []byte(change.Password)) != nil {
			a.nextFactor = time.Now().Add(time.Second)
			http.Error(w, "Password web attuale non valida.", http.StatusForbidden)
			return
		}
		if *change.Enabled == a.factorEnabled() {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if a.factorEnabled() && !a.confirmPasswordChange(w, r, change.Code, change.Password) {
			return
		}
		next := secondFactor{ResetPending: *change.Enabled}
		if err := a.removeWebSession(); err != nil || saveFactor(a.factorFile, next) != nil {
			http.Error(w, "Impossibile salvare la verifica in due passaggi.", http.StatusInternalServerError)
			return
		}
		a.factor, a.pendingKey, a.factorOK = next, "", false
		if *change.Enabled {
			a.expires = time.Now().Add(20 * time.Minute)
		} else {
			a.session = ""
			http.SetCookie(w, &http.Cookie{Name: sysopCookie, Path: "/", MaxAge: -1, HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode})
		}
		w.WriteHeader(http.StatusNoContent)
	}
	mux.HandleFunc("GET /api/factor-settings", handler)
	mux.HandleFunc("PUT /api/factor-settings", handler)
}
