package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Registrazione degli utenti dal portale. Chi vuole inviare spot sceglie nominativo, email e password;
// riceve un'email con un link, e solo cliccandolo il nominativo viene registrato in DXSpider con quella
// password (salvata come hash bcrypt). Il collegamento Telnet chiede poi la password.

const (
	registrationTTL      = 24 * time.Hour
	registrationsPerDay  = 40
	registrationsPerCall = 3
	registrationMax      = 5000
)

var (
	userCall     = regexp.MustCompile(`^[A-Z0-9]+[0-9][A-Z0-9]+$`)
	userPassword = regexp.MustCompile(`^[\x21-\x7e]{8,72}$`)
)

type registerSettings struct {
	Enabled   bool
	PublicURL string
}

func registerSettingsPath(root string) string {
	return filepath.Join(root, "local", "dxregia-register.json")
}

func readRegisterSettings(root string) registerSettings {
	var settings registerSettings
	if data, err := os.ReadFile(registerSettingsPath(root)); err == nil {
		_ = json.Unmarshal(data, &settings)
	}
	if _, ok := cleanPublicURL(settings.PublicURL); !ok {
		settings.Enabled = false
	}
	return settings
}

// cleanPublicURL accetta solo l'indirizzo del sito (schema e host): è la base dei link nelle email.
func cleanPublicURL(raw string) (string, bool) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.User != nil {
		return "", false
	}
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && (parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1")) {
		return "", false
	}
	return parsed.Scheme + "://" + parsed.Host, true
}

type registration struct {
	Call        string
	Email       string
	Lang        string
	Hash        string `json:",omitempty"`
	TokenHash   string `json:",omitempty"`
	RequestedAt time.Time
	ExpiresAt   time.Time
	ConfirmedAt *time.Time `json:",omitempty"`
	Status      string
}

type registrationView struct {
	Call        string
	Email       string
	Lang        string
	Status      string
	RequestedAt time.Time
	ConfirmedAt *time.Time `json:",omitempty"`
}

const (
	regPending    = "Da confermare"
	regQueued     = "In attesa di Mojo"
	regDone       = "Registrato"
	regExpired    = "Scaduta"
	regRefused    = "Rifiutata"
	regRevokeWait = "Revoca in corso"
)

func tokenDigest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func saveRegistrations(path string, items map[string]registration) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".registrations-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	err = json.NewEncoder(tmp).Encode(items)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func registrationQueueDir(root string) string {
	return filepath.Join(root, "local_data", "dxregia-users")
}

func nodeCallOf(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "local", "dxregia-node.json"))
	var node nodeConfig
	if err != nil || json.Unmarshal(data, &node) != nil || node.Node == "" {
		return "DXRegia"
	}
	return node.Node
}

func registerUsers(mux *http.ServeMux, root string, auth *sysopAuth) {
	if root == "" || auth == nil {
		return
	}
	path := filepath.Join(root, ".dxregia-registrations.json")
	items := map[string]registration{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &items)
	}
	var lock sync.Mutex
	save := func() {
		if err := saveRegistrations(path, items); err != nil {
			log.Printf("Save registrations: %v", err)
		}
	}
	decode := func(w http.ResponseWriter, r *http.Request, into any) bool {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return false
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if decoder.Decode(into) != nil || decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return false
		}
		return true
	}
	reply := func(w http.ResponseWriter, value any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(value)
	}

	// Stato pubblico: serve al portale per mostrare o nascondere il pulsante «Registrati».
	mux.HandleFunc("GET /api/register/status", func(w http.ResponseWriter, r *http.Request) {
		settings := readRegisterSettings(root)
		reply(w, map[string]bool{"Enabled": settings.Enabled && currentMailer(root) != nil})
	})

	mux.HandleFunc("POST /api/register", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Call, Email, Password, Lang, Website string
		}
		if !decode(w, r, &input) {
			return
		}
		settings := readRegisterSettings(root)
		mailer := currentMailer(root)
		if !settings.Enabled || mailer == nil {
			http.Error(w, "La registrazione non è disponibile.", http.StatusServiceUnavailable)
			return
		}
		if input.Website != "" { // campo trappola per i programmi automatici
			w.WriteHeader(http.StatusAccepted)
			return
		}
		call := strings.ToUpper(strings.TrimSpace(input.Call))
		email := strings.TrimSpace(input.Email)
		address, err := mail.ParseAddress(email)
		if !userCall.MatchString(call) || len(call) > 10 || err != nil || address.Address != email || len(email) > 254 || !userPassword.MatchString(input.Password) {
			http.Error(w, "Controlla indicativo, email e password.", http.StatusBadRequest)
			return
		}
		lang := normalizeLang(input.Lang)
		if lang == "" {
			lang = "en"
		}
		token, err := randomToken()
		if err != nil {
			http.Error(w, "Richiesta non registrata.", http.StatusInternalServerError)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
		if err != nil {
			http.Error(w, "Richiesta non registrata.", http.StatusInternalServerError)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		now := time.Now().UTC()
		recent, sameCall := 0, 0
		for _, item := range items {
			if now.Sub(item.RequestedAt) < 24*time.Hour {
				recent++
				if item.Call == call || strings.EqualFold(item.Email, email) {
					sameCall++
				}
			}
		}
		if existing, ok := items[call]; ok && existing.Status == regDone {
			http.Error(w, "Questo indicativo è già registrato.", http.StatusConflict)
			return
		}
		if len(items) >= registrationMax || recent >= registrationsPerDay || sameCall >= registrationsPerCall {
			http.Error(w, "Troppe richieste. Riprova più tardi.", http.StatusTooManyRequests)
			return
		}
		item := registration{Call: call, Email: email, Lang: lang, Hash: string(hash), TokenHash: tokenDigest(token),
			RequestedAt: now, ExpiresAt: now.Add(registrationTTL), Status: regPending}
		content := mailRegistrationConfirm(lang, call, nodeCallOf(root), settings.PublicURL+"/registrazione/conferma?token="+token)
		if err := mailer.send(email, buildMessage(mailer.from, email, "", content)); err != nil {
			log.Printf("Registration mail to %s: %v", email, err)
			http.Error(w, "Non riesco a inviare l'email di conferma. Riprova più tardi.", http.StatusBadGateway)
			return
		}
		items[call] = item
		save()
		w.WriteHeader(http.StatusAccepted)
	})

	mux.HandleFunc("POST /api/register/confirm", func(w http.ResponseWriter, r *http.Request) {
		var input struct{ Token string }
		if !decode(w, r, &input) {
			return
		}
		if len(input.Token) != 48 {
			http.Error(w, "Link non valido.", http.StatusBadRequest)
			return
		}
		digest := tokenDigest(input.Token)
		lock.Lock()
		var found *registration
		for call, item := range items {
			if item.TokenHash != "" && subtle.ConstantTimeCompare([]byte(item.TokenHash), []byte(digest)) == 1 {
				snapshot := item
				found = &snapshot
				_ = call
				break
			}
		}
		if found == nil {
			lock.Unlock()
			http.Error(w, "Link non valido o già usato.", http.StatusNotFound)
			return
		}
		if time.Now().After(found.ExpiresAt) {
			found.Status, found.TokenHash, found.Hash = regExpired, "", ""
			items[found.Call] = *found
			save()
			lock.Unlock()
			http.Error(w, "Il link è scaduto. Ripeti la registrazione.", http.StatusGone)
			return
		}
		dir := registrationQueueDir(root)
		if err := os.MkdirAll(dir, 0700); err == nil {
			_ = os.Remove(filepath.Join(dir, found.Call+".result"))
			err = os.WriteFile(filepath.Join(dir, found.Call+".register"), []byte(found.Hash), 0600)
			if err != nil {
				lock.Unlock()
				http.Error(w, "Registrazione non completata. Riprova.", http.StatusInternalServerError)
				return
			}
		} else {
			lock.Unlock()
			http.Error(w, "Registrazione non completata. Riprova.", http.StatusInternalServerError)
			return
		}
		now := time.Now().UTC()
		found.Status, found.TokenHash, found.ConfirmedAt = regQueued, "", &now
		items[found.Call] = *found
		save()
		lock.Unlock()
		// Mojo elabora la coda ogni pochi secondi: si attende l'esito per mostrarlo subito.
		result := ""
		for attempt := 0; attempt < 14 && result == ""; attempt++ {
			time.Sleep(time.Second)
			if data, err := os.ReadFile(filepath.Join(dir, found.Call+".result")); err == nil {
				result = strings.TrimSpace(string(data))
			}
		}
		lock.Lock()
		current := items[found.Call]
		switch result {
		case "Registrato":
			current.Status, current.Hash = regDone, ""
		case "":
		default:
			current.Status, current.Hash = regRefused, ""
		}
		items[found.Call] = current
		save()
		lock.Unlock()
		reply(w, map[string]string{"Call": found.Call, "Result": result})
	})

	// Area sysop: impostazioni, elenco e revoca.
	mux.HandleFunc("GET /api/register/settings", func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		settings := readRegisterSettings(root)
		raw := settings
		if data, err := os.ReadFile(registerSettingsPath(root)); err == nil {
			_ = json.Unmarshal(data, &raw)
		}
		reply(w, map[string]any{"Enabled": raw.Enabled, "PublicURL": raw.PublicURL, "SMTP": currentMailer(root) != nil})
	})
	mux.HandleFunc("PUT /api/register/settings", func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		var input registerSettings
		if !decode(w, r, &input) {
			return
		}
		clean, ok := cleanPublicURL(input.PublicURL)
		if !ok {
			http.Error(w, "Indirizzo del sito non valido: usa https://esempio.it senza percorso.", http.StatusBadRequest)
			return
		}
		input.PublicURL = clean
		encoded, _ := json.Marshal(input)
		if err := writeNodeFile(registerSettingsPath(root), encoded); err != nil {
			http.Error(w, "Impossibile salvare le impostazioni.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/registrations", func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		lock.Lock()
		views := make([]registrationView, 0, len(items))
		for _, item := range items {
			if item.Status == regPending && time.Now().After(item.ExpiresAt) {
				item.Status = regExpired
			}
			views = append(views, registrationView{Call: item.Call, Email: item.Email, Lang: item.Lang, Status: item.Status,
				RequestedAt: item.RequestedAt, ConfirmedAt: item.ConfirmedAt})
		}
		lock.Unlock()
		sort.Slice(views, func(i, j int) bool { return views[i].RequestedAt.After(views[j].RequestedAt) })
		reply(w, views)
	})
	mux.HandleFunc("DELETE /api/registrations/{call}", func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		call := strings.ToUpper(r.PathValue("call"))
		if !userCall.MatchString(call) {
			http.Error(w, "Indicativo non valido.", http.StatusBadRequest)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		item, ok := items[call]
		if !ok {
			http.Error(w, "Registrazione non trovata.", http.StatusNotFound)
			return
		}
		if item.Status == regDone || item.Status == regQueued {
			dir := registrationQueueDir(root)
			if err := os.MkdirAll(dir, 0700); err != nil || os.WriteFile(filepath.Join(dir, call+".revoke"), []byte("1"), 0600) != nil {
				http.Error(w, "Revoca non riuscita.", http.StatusInternalServerError)
				return
			}
		}
		delete(items, call)
		save()
		w.WriteHeader(http.StatusNoContent)
	})
}
