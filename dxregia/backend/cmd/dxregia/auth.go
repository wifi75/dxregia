package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sysopCookie = "dxregia_sysop"

type sysopAuth struct {
	mu            sync.Mutex
	file          string
	bootstrapFile string
	hash          []byte
	temporary     bool
	session       string
	expires       time.Time
	nextLogin     time.Time
}

func randomToken() (string, error) {
	key := make([]byte, 24)
	_, err := rand.Read(key)
	return hex.EncodeToString(key), err
}

func registerAuth(mux *http.ServeMux, root, bootstrapFile string) (*sysopAuth, error) {
	if root == "" {
		return nil, nil
	}
	auth := &sysopAuth{file: filepath.Join(root, ".dxregia-sysop"), bootstrapFile: bootstrapFile}
	hash, err := os.ReadFile(auth.file)
	if os.IsNotExist(err) {
		password, err := bootstrapPassword(bootstrapFile)
		if err != nil {
			return nil, err
		}
		hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		auth.temporary = true
		if bootstrapFile == "" {
			log.Printf("Password temporanea sysop: %s", password)
		}
	} else if err != nil {
		return nil, err
	}
	if _, err := bcrypt.Cost(hash); err != nil {
		return nil, errors.New("invalid sysop password file")
	}
	auth.hash = hash
	mux.HandleFunc("GET /api/auth", auth.route("status"))
	mux.HandleFunc("POST /api/auth/login", auth.route("login"))
	mux.HandleFunc("POST /api/auth/password", auth.route("password"))
	mux.HandleFunc("POST /api/auth/logout", auth.route("logout"))
	return auth, nil
}

func bootstrapPassword(file string) (string, error) {
	if file == "" {
		return randomToken()
	}
	existing, err := os.ReadFile(file)
	if err == nil {
		if key, decodeErr := hex.DecodeString(string(existing)); decodeErr != nil || len(key) != 24 {
			return "", errors.New("invalid temporary password file")
		}
		return string(existing), nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	password, err := randomToken()
	if err != nil {
		return "", err
	}
	out, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", err
	}
	_, err = out.WriteString(password)
	closeErr := out.Close()
	return password, errors.Join(err, closeErr)
}

func (a *sysopAuth) signedIn(r *http.Request) bool {
	cookie, err := r.Cookie(sysopCookie)
	return err == nil && a.session != "" && cookie.Value == a.session && time.Now().Before(a.expires)
}

func (a *sysopAuth) allowed(r *http.Request) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.signedIn(r) && !a.temporary
}

func (a *sysopAuth) route(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		a.mu.Lock()
		defer a.mu.Unlock()
		if action != "status" && r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		switch action {
		case "login", "password":
			if action == "password" && (!a.signedIn(r) || !a.temporary) {
				http.Error(w, "Accesso richiesto.", http.StatusUnauthorized)
				return
			}
			var password string
			body := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2048))
			if body.Decode(&password) != nil || body.Decode(new(any)) != io.EOF || len(password) > 72 {
				http.Error(w, "Password non valida.", http.StatusBadRequest)
				return
			}
			if action == "login" {
				if time.Now().Before(a.nextLogin) {
					http.Error(w, "Attendi prima di riprovare.", http.StatusTooManyRequests)
					return
				}
				if bcrypt.CompareHashAndPassword(a.hash, []byte(password)) != nil {
					a.nextLogin = time.Now().Add(time.Second)
					http.Error(w, "Password errata.", http.StatusUnauthorized)
					return
				}
			} else {
				if len(password) < 12 || bcrypt.CompareHashAndPassword(a.hash, []byte(password)) == nil {
					http.Error(w, "Scegli una password diversa di almeno 12 caratteri.", http.StatusBadRequest)
					return
				}
			}
			token, err := randomToken()
			if err == nil && action == "password" {
				err = a.savePassword(password)
			}
			if err != nil {
				log.Printf("Sysop authentication: %v", err)
				http.Error(w, "Impossibile completare l’operazione.", http.StatusInternalServerError)
				return
			}
			a.session, a.expires = token, time.Now().Add(8*time.Hour)
			http.SetCookie(w, &http.Cookie{Name: sysopCookie, Value: token, Path: "/", HttpOnly: true,
				Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode})
		case "logout":
			if a.signedIn(r) {
				a.session = ""
			}
			http.SetCookie(w, &http.Cookie{Name: sysopCookie, Path: "/", MaxAge: -1, HttpOnly: true,
				Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode})
		}
		w.Header().Set("Content-Type", "application/json")
		// The login response precedes the browser's next request with its new cookie.
		signedIn := action == "login" || action == "password" || a.signedIn(r)
		json.NewEncoder(w).Encode(map[string]bool{"signedIn": signedIn, "mustChange": signedIn && a.temporary})
	}
}

func (a *sysopAuth) savePassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(a.file), ".dxregia-auth-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(hash)
	closeErr := tmp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Rename(tmp.Name(), a.file); err != nil {
		return err
	}
	a.hash, a.temporary = hash, false
	if a.bootstrapFile != "" {
		if err := os.Remove(a.bootstrapFile); err != nil && !os.IsNotExist(err) {
			log.Printf("Remove temporary password file: %v", err)
		}
	}
	return nil
}
