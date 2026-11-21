package main

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Un browser che ha superato la verifica in due passaggi viene ricordato per dieci giorni:
// al login con password giusta il codice a sei cifre non viene richiesto di nuovo.
// Il cookie resta nel browser anche dopo la chiusura e il riavvio del server.

const (
	trustCookie   = "dxregia_trust"
	trustDuration = 10 * 24 * time.Hour
)

type trustedDevice struct {
	TokenHash, Credentials string
	Expires                time.Time
}

func (a *sysopAuth) trustPath() string {
	return filepath.Join(a.root, ".dxregia-trusted.json")
}

func (a *sysopAuth) loadTrusted() []trustedDevice {
	data, err := os.ReadFile(a.trustPath())
	var devices []trustedDevice
	if err != nil || json.Unmarshal(data, &devices) != nil {
		return nil
	}
	return devices
}

func (a *sysopAuth) saveTrusted(devices []trustedDevice) error {
	tmp, err := os.CreateTemp(a.root, ".dxregia-trust-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	encodeErr := json.NewEncoder(tmp).Encode(devices)
	closeErr := tmp.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(tmp.Name(), a.trustPath())
}

// trustedBrowser dice se la richiesta porta un cookie di dispositivo fidato ancora valido
// per la password e la chiave di verifica attuali.
func (a *sysopAuth) trustedBrowser(r *http.Request) bool {
	cookie, err := r.Cookie(trustCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	hash, credentials, now := sessionHash(cookie.Value), a.sessionCredentials(), time.Now()
	for _, device := range a.loadTrusted() {
		if now.Before(device.Expires) &&
			subtle.ConstantTimeCompare([]byte(device.TokenHash), []byte(hash)) == 1 &&
			subtle.ConstantTimeCompare([]byte(device.Credentials), []byte(credentials)) == 1 {
			return true
		}
	}
	return false
}

// trustBrowser registra un nuovo dispositivo fidato e imposta il cookie nella risposta.
func (a *sysopAuth) trustBrowser(w http.ResponseWriter, secure bool) error {
	token, err := randomToken()
	if err != nil {
		return err
	}
	credentials, now := a.sessionCredentials(), time.Now()
	kept := []trustedDevice{{TokenHash: sessionHash(token), Credentials: credentials, Expires: now.Add(trustDuration)}}
	for _, device := range a.loadTrusted() {
		if now.Before(device.Expires) && device.Credentials == credentials && len(kept) < 20 {
			kept = append(kept, device)
		}
	}
	if err := a.saveTrusted(kept); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: trustCookie, Value: token, Path: "/", HttpOnly: true,
		MaxAge: int(trustDuration / time.Second), Expires: now.Add(trustDuration),
		Secure: secure, SameSite: http.SameSiteStrictMode})
	return nil
}
