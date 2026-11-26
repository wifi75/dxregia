package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sysopCookie = "dxregia_sysop"

type sysopAuth struct {
	mu            sync.Mutex
	root          string
	file          string
	bootstrapFile string
	hash          []byte
	temporary     bool
	session       string
	expires       time.Time
	nextLogin     time.Time
	nextFactor    time.Time
	loginFails    int
	factorFails   int
	factorFile    string
	factor        secondFactor
	pendingKey    string
	factorOK      bool
}

// loginName è il nome utente del sysop: l'indicativo scelto nella configurazione guidata.
// Finché la configurazione non esiste si usa «sysop».
func (a *sysopAuth) loginName() string {
	data, err := os.ReadFile(filepath.Join(a.root, "local", "dxregia-node.json"))
	var node nodeConfig
	if err != nil || json.Unmarshal(data, &node) != nil || strings.TrimSpace(node.Sysop) == "" {
		return "sysop"
	}
	return strings.TrimSpace(node.Sysop)
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
	auth := &sysopAuth{root: root, file: filepath.Join(root, ".dxregia-sysop"), bootstrapFile: bootstrapFile}
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
	auth.factorFile = filepath.Join(root, ".dxregia-totp")
	if data, err := os.ReadFile(auth.factorFile); err == nil {
		if err := json.Unmarshal(data, &auth.factor); err != nil {
			return nil, err
		}
		key, err := totpEncoding.DecodeString(auth.factor.Key)
		if auth.factor.Key != "" && (err != nil || len(key) != 20) {
			return nil, errors.New("invalid sysop TOTP key")
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	auth.restoreWebSession()
	auth.registerWebPasswordChange(mux)
	auth.registerSessionDuration(mux)
	auth.registerFactorSettings(mux)
	mux.HandleFunc("GET /api/auth", auth.route("status"))
	mux.HandleFunc("POST /api/auth/login", auth.route("login"))
	mux.HandleFunc("POST /api/auth/password", auth.route("password"))
	mux.HandleFunc("POST /api/auth/totp", auth.route("totp"))
	mux.HandleFunc("POST /api/auth/recovery", auth.route("recovery"))
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
	return err == nil && a.session != "" && time.Now().Before(a.expires) &&
		subtle.ConstantTimeCompare([]byte(sessionHash(cookie.Value)), []byte(a.session)) == 1
}

func (a *sysopAuth) allowed(r *http.Request) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.signedIn(r) && !a.temporary && a.factorOK
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
		var recoveryCodes []string
		switch action {
		case "login", "password":
			if action == "password" && (!a.signedIn(r) || !a.temporary) {
				http.Error(w, "Accesso richiesto.", http.StatusUnauthorized)
				return
			}
			var password, user string
			var raw json.RawMessage
			body := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2048))
			if body.Decode(&raw) != nil || body.Decode(new(any)) != io.EOF {
				http.Error(w, "Password non valida.", http.StatusBadRequest)
				return
			}
			if action == "login" {
				var input struct{ User, Password string }
				if json.Unmarshal(raw, &input) != nil {
					http.Error(w, "Utente o password non validi.", http.StatusBadRequest)
					return
				}
				user, password = input.User, input.Password
			} else if json.Unmarshal(raw, &password) != nil {
				http.Error(w, "Password non valida.", http.StatusBadRequest)
				return
			}
			if len(password) > 72 || len(user) > 64 {
				http.Error(w, "Password non valida.", http.StatusBadRequest)
				return
			}
			if action == "login" {
				if time.Now().Before(a.nextLogin) {
					http.Error(w, "Attendi prima di riprovare.", http.StatusTooManyRequests)
					return
				}
				// La password si verifica sempre, così i tempi non rivelano se l'utente era giusto.
				passwordOK := bcrypt.CompareHashAndPassword(a.hash, []byte(password)) == nil
				userOK := subtle.ConstantTimeCompare([]byte(strings.ToLower(strings.TrimSpace(user))), []byte(strings.ToLower(a.loginName()))) == 1
				if !passwordOK || !userOK {
					a.loginFails++
					a.nextLogin = time.Now().Add(loginDelay(a.loginFails))
					security.record("Accesso errato", clientIPText(r), "/api/auth/login")
					http.Error(w, "Utente o password errati.", http.StatusUnauthorized)
					return
				}
			} else {
				if len(password) < 12 || bcrypt.CompareHashAndPassword(a.hash, []byte(password)) == nil {
					http.Error(w, "Scegli una password diversa di almeno 12 caratteri.", http.StatusBadRequest)
					return
				}
			}
			a.loginFails = 0
			token, err := randomToken()
			if err == nil {
				err = a.removeWebSession()
			}
			if err == nil && action == "password" {
				err = a.savePassword(password)
			}
			if err != nil {
				log.Printf("Sysop authentication: %v", err)
				http.Error(w, "Impossibile completare l’operazione.", http.StatusInternalServerError)
				return
			}
			a.session, a.expires = sessionHash(token), time.Now().Add(20*time.Minute)
			a.factorOK, a.pendingKey = false, ""
			if !a.temporary && (!a.factorEnabled() || (action == "login" && a.factor.Key != "" && a.trustedBrowser(r))) {
				a.factorOK = true
				a.expires = time.Now().Add(time.Duration(a.sessionHours()) * time.Hour)
				if err := a.persistWebSession(a.expires); err != nil {
					log.Printf("Save sysop session: %v", err)
				}
			}
			http.SetCookie(w, &http.Cookie{Name: sysopCookie, Value: token, Path: "/", HttpOnly: true,
				Secure: secureRequest(r), SameSite: http.SameSiteStrictMode})
		case "totp", "recovery":
			if !a.signedIn(r) || a.temporary || a.factorOK || !a.factorEnabled() {
				http.Error(w, "Accesso con password richiesto.", http.StatusUnauthorized)
				return
			}
			if time.Now().Before(a.nextFactor) {
				http.Error(w, "Attendi prima di riprovare.", http.StatusTooManyRequests)
				return
			}
			var code string
			body := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128))
			if body.Decode(&code) != nil || body.Decode(new(any)) != io.EOF {
				http.Error(w, "Codice non valido.", http.StatusBadRequest)
				return
			}
			if action == "recovery" {
				if !a.factor.allowsRecovery(code) {
					a.factorFails++
					a.nextFactor = time.Now().Add(loginDelay(a.factorFails))
					security.record("Codice errato", clientIPText(r), "/api/auth/recovery")
					http.Error(w, "Codice di recupero non valido.", http.StatusUnauthorized)
					return
				}
				key, err := newFactorKey()
				if err == nil {
					err = saveFactor(a.factorFile, secondFactor{ResetPending: true})
				}
				if err != nil {
					log.Printf("Sysop recovery: %v", err)
					http.Error(w, "Impossibile avviare il recupero.", http.StatusInternalServerError)
					return
				}
				a.factor, a.pendingKey = secondFactor{ResetPending: true}, key
				a.expires = time.Now().Add(20 * time.Minute)
				break
			}
			factor := a.factor
			if factor.Key == "" {
				factor.Key = a.pendingKey
			}
			verified, ok, reused := factor.verify(code, time.Now())
			if !ok {
				a.factorFails++
				a.nextFactor = time.Now().Add(loginDelay(a.factorFails))
				security.record("Codice errato", clientIPText(r), "/api/auth/totp")
				message := "Codice non valido. Controlla di usare la voce DXRegia corretta e che telefono e server abbiano l’ora sincronizzata."
				if reused {
					message = "Codice già usato. Attendi che l’app mostri un nuovo codice e riprova."
				}
				http.Error(w, message, http.StatusUnauthorized)
				return
			}
			a.factorFails = 0
			token, err := randomToken()
			if err == nil && len(verified.Recovery) == 0 {
				recoveryCodes, err = createRecoveryCodes(&verified)
			}
			if err == nil {
				err = saveFactor(a.factorFile, verified)
			}
			if err != nil {
				log.Printf("Sysop TOTP: %v", err)
				http.Error(w, "Impossibile completare la verifica.", http.StatusInternalServerError)
				return
			}
			a.factor, a.pendingKey, a.factorOK = verified, "", true
			a.session, a.expires = sessionHash(token), time.Now().Add(time.Duration(a.sessionHours())*time.Hour)
			if err := a.persistWebSession(a.expires); err != nil {
				log.Printf("Save sysop session: %v", err)
			}
			http.SetCookie(w, &http.Cookie{Name: sysopCookie, Value: token, Path: "/", HttpOnly: true,
				Secure: secureRequest(r), SameSite: http.SameSiteStrictMode})
			if err := a.trustBrowser(w, secureRequest(r)); err != nil {
				log.Printf("Save trusted device: %v", err)
			}
		case "logout":
			if a.signedIn(r) {
				if err := a.removeWebSession(); err != nil {
					http.Error(w, "Impossibile terminare la sessione.", http.StatusInternalServerError)
					return
				}
				a.session, a.pendingKey, a.factorOK = "", "", false
			}
			http.SetCookie(w, &http.Cookie{Name: sysopCookie, Path: "/", MaxAge: -1, HttpOnly: true,
				Secure: secureRequest(r), SameSite: http.SameSiteStrictMode})
		}
		w.Header().Set("Content-Type", "application/json")
		// The login response precedes the browser's next request with its new cookie.
		signedIn := action == "login" || action == "password" || action == "totp" || a.signedIn(r)
		result := map[string]any{"signedIn": signedIn, "mustChange": signedIn && a.temporary}
		if len(recoveryCodes) > 0 {
			result["recoveryCodes"] = recoveryCodes
		}
		if signedIn && !a.temporary && !a.factorOK {
			result["mustVerify"] = true
			if a.factor.Key == "" {
				if a.pendingKey == "" {
					key, err := newFactorKey()
					if err != nil {
						http.Error(w, "Impossibile generare la chiave di verifica.", http.StatusInternalServerError)
						return
					}
					a.pendingKey = key
				}
				result["enrollmentKey"] = a.pendingKey
			}
		}
		json.NewEncoder(w).Encode(result)
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
