package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func registerTelnetPassword(mux *http.ServeMux, auth *sysopAuth, root string) {
	if auth == nil || root == "" {
		return
	}
	var lock sync.Mutex
	dir := filepath.Join(root, "local_data")
	pending := filepath.Join(dir, ".dxregia-telnet-password")
	result := filepath.Join(dir, "dxregia-telnet-password.result")
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		status := "idle"
		if _, err := os.Stat(pending); err == nil {
			status = "pending"
		} else if data, err := os.ReadFile(result); err == nil {
			switch string(data) {
			case "applied", "failed":
				status = string(data)
			}
		}
		if r.Method == http.MethodPut {
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
				return
			}
			var change struct{ Password, Code string }
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256))
			if decoder.Decode(&change) != nil || decoder.Decode(new(any)) != io.EOF || !validTelnetPassword(change.Password) {
				http.Error(w, "Usa da 12 a 72 caratteri ASCII stampabili, senza spazi.", http.StatusBadRequest)
				return
			}
			if status == "pending" {
				http.Error(w, "Un cambio di password è già in corso.", http.StatusConflict)
				return
			}
			snapshot := readPresence(filepath.Join(dir, "dxregia-presence.json"))
			if !snapshot.SysopPassword {
				http.Error(w, "Il nodo Mojo deve essere attivo con l’integrazione DXRegia aggiornata.", http.StatusServiceUnavailable)
				return
			}
			auth.mu.Lock()
			defer auth.mu.Unlock()
			if !auth.confirmPasswordChange(w, r, change.Code) {
				return
			}
			if err := os.Remove(result); err != nil && !os.IsNotExist(err) {
				http.Error(w, "Impossibile preparare il cambio password.", http.StatusInternalServerError)
				return
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(change.Password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Impossibile proteggere la password.", http.StatusInternalServerError)
				return
			}
			file, err := os.CreateTemp(dir, ".telnet-password-")
			if err != nil {
				http.Error(w, "Impossibile preparare il cambio password.", http.StatusInternalServerError)
				return
			}
			defer os.Remove(file.Name())
			_, err = file.Write(hash)
			if closeErr := file.Close(); err == nil {
				err = closeErr
			}
			if err == nil {
				err = os.Rename(file.Name(), pending)
			}
			if err != nil {
				http.Error(w, "Impossibile accodare il cambio password.", http.StatusInternalServerError)
				return
			}
			status = "pending"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": status})
	}
	mux.HandleFunc("GET /api/telnet-password", handler)
	mux.HandleFunc("PUT /api/telnet-password", handler)
}

func validTelnetPassword(password string) bool {
	if len(password) < 12 || len(password) > 72 {
		return false
	}
	for i := 0; i < len(password); i++ {
		if password[i] < '!' || password[i] > '~' {
			return false
		}
	}
	return true
}

func (a *sysopAuth) registerWebPasswordChange(mux *http.ServeMux) {
	mux.HandleFunc("PUT /api/web-password", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return
		}
		var change struct{ Password, Code string }
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&change) != nil || decoder.Decode(new(any)) != io.EOF ||
			len(change.Password) < 12 || len(change.Password) > 72 {
			http.Error(w, "Usa una password tra 12 e 72 byte.", http.StatusBadRequest)
			return
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if !a.confirmPasswordChange(w, r, change.Code) {
			return
		}
		if err := a.removeWebSession(); err != nil {
			http.Error(w, "Impossibile revocare la sessione corrente.", http.StatusInternalServerError)
			return
		}
		if err := a.savePassword(change.Password); err != nil {
			http.Error(w, "Impossibile salvare la password web.", http.StatusInternalServerError)
			return
		}
		a.session, a.pendingKey, a.factorOK = "", "", false
		http.SetCookie(w, &http.Cookie{Name: sysopCookie, Path: "/", MaxAge: -1, HttpOnly: true,
			Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode})
		w.WriteHeader(http.StatusNoContent)
	})
}

// The caller holds a.mu until the authorized change has been applied.
func (a *sysopAuth) confirmPasswordChange(w http.ResponseWriter, r *http.Request, code string) bool {
	if !a.signedIn(r) || a.temporary || !a.factorOK {
		http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
		return false
	}
	now := time.Now()
	if now.Before(a.nextFactor) {
		http.Error(w, "Attendi prima di riprovare.", http.StatusTooManyRequests)
		return false
	}
	verified, accepted, reused := a.factor.verify(code, now)
	if !accepted {
		a.nextFactor = now.Add(time.Second)
		message := "Codice 2FA non valido."
		if reused {
			message = "Codice 2FA già usato. Attendi il prossimo codice."
		}
		http.Error(w, message, http.StatusUnauthorized)
		return false
	}
	if err := saveFactor(a.factorFile, verified); err != nil {
		http.Error(w, "Impossibile confermare il 2FA.", http.StatusInternalServerError)
		return false
	}
	a.factor = verified
	return true
}
