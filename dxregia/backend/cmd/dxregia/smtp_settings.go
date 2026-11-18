package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// smtpSettings è la configurazione SMTP salvata dal pannello. Ha priorità sulle
// variabili d'ambiente DXREGIA_SMTP_*, che restano valide come valore iniziale.
type smtpSettings struct {
	Address  string
	From     string
	User     string
	Password string
	TLS      string
}

func smtpSettingsPath(root string) string {
	return filepath.Join(root, "local_data", "dxregia-smtp.json")
}

func readSMTPSettings(root string) (smtpSettings, bool, error) {
	var settings smtpSettings
	data, err := os.ReadFile(smtpSettingsPath(root))
	if os.IsNotExist(err) {
		return settings, false, nil
	}
	if err != nil {
		return settings, false, err
	}
	return settings, true, json.Unmarshal(data, &settings)
}

func (s smtpSettings) validate() error {
	host, port, err := net.SplitHostPort(s.Address)
	if err != nil || host == "" || port == "" || len(s.Address) > 253 {
		return errors.New("Indirizzo del server non valido. Usa il formato host:porta, ad esempio smtp.example.com:465.")
	}
	sender, err := mail.ParseAddress(s.From)
	if err != nil || sender.Address != s.From || len(s.From) > 254 {
		return errors.New("L’indirizzo del mittente non è valido.")
	}
	if s.TLS != "implicit" && s.TLS != "starttls" {
		return errors.New("Scegli la modalità di cifratura: TLS implicito o STARTTLS.")
	}
	if len(s.User) > 254 || len(s.Password) > 254 || strings.ContainsAny(s.User+s.Password, "\r\n") {
		return errors.New("Utente o password non validi.")
	}
	return nil
}

func (s smtpSettings) mailer() *contactMailer {
	host, _, _ := net.SplitHostPort(s.Address)
	return &contactMailer{address: s.Address, host: host, from: s.From, user: s.User, password: s.Password, startTLS: s.TLS == "starttls"}
}

// currentMailer restituisce il mittente configurato dal pannello oppure, in assenza, quello
// delle variabili d'ambiente. Restituisce nil se l'invio email non è configurato.
func currentMailer(root string) *contactMailer {
	if settings, found, err := readSMTPSettings(root); found {
		if err != nil || settings.validate() != nil {
			return nil
		}
		return settings.mailer()
	}
	mailer, err := loadContactMailer()
	if err != nil {
		return nil
	}
	return mailer
}

func writeSMTPSettings(root string, settings smtpSettings) error {
	dir := filepath.Dir(smtpSettingsPath(root))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".smtp-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0600); err != nil && !os.IsPermission(err) {
		tmp.Close()
		return err
	}
	err = json.NewEncoder(tmp).Encode(settings)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), smtpSettingsPath(root))
}

type smtpView struct {
	Address, From, User, TLS string
	PasswordSet              bool
	Source                   string
}

func smtpViewOf(root string) smtpView {
	if settings, found, err := readSMTPSettings(root); found {
		view := smtpView{Source: "pannello"}
		if err == nil {
			view = smtpView{Address: settings.Address, From: settings.From, User: settings.User, TLS: settings.TLS,
				PasswordSet: settings.Password != "", Source: "pannello"}
		}
		return view
	}
	if mailer, err := loadContactMailer(); err == nil && mailer != nil {
		tls := "implicit"
		if mailer.startTLS {
			tls = "starttls"
		}
		return smtpView{Address: mailer.address, From: mailer.from, User: mailer.user, TLS: tls,
			PasswordSet: mailer.password != "", Source: "ambiente"}
	}
	return smtpView{TLS: "implicit"}
}

func registerSMTP(mux *http.ServeMux, root string, auth *sysopAuth) {
	if auth == nil {
		return
	}
	var lock sync.Mutex
	guard := func(w http.ResponseWriter, r *http.Request) bool {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return false
		}
		return true
	}
	mux.HandleFunc("GET /api/smtp-settings", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r) {
			return
		}
		lock.Lock()
		defer lock.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(smtpViewOf(root))
	})
	mux.HandleFunc("PUT /api/smtp-settings", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r) {
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		var input struct {
			Address, From, User, TLS string
			Password                 *string
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "Controlla i campi del server di posta.", http.StatusBadRequest)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		settings := smtpSettings{Address: strings.TrimSpace(input.Address), From: strings.TrimSpace(input.From),
			User: strings.TrimSpace(input.User), TLS: input.TLS}
		if input.Password != nil {
			settings.Password = *input.Password
		} else if previous, found, err := readSMTPSettings(root); found && err == nil {
			settings.Password = previous.Password
		} else if view := smtpViewOf(root); view.Source == "ambiente" && view.User == settings.User {
			settings.Password = os.Getenv("DXREGIA_SMTP_PASSWORD")
		}
		if err := settings.validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := writeSMTPSettings(root, settings); err != nil {
			http.Error(w, "Impossibile salvare le impostazioni di posta.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(smtpViewOf(root))
	})
	mux.HandleFunc("POST /api/smtp-settings/test", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r) {
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		var input struct{ To string }
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "Indica l’indirizzo a cui inviare la prova.", http.StatusBadRequest)
			return
		}
		to, err := mail.ParseAddress(input.To)
		if err != nil || to.Address != input.To || len(input.To) > 254 {
			http.Error(w, "L’indirizzo del destinatario non è valido.", http.StatusBadRequest)
			return
		}
		lock.Lock()
		mailer := currentMailer(root)
		lock.Unlock()
		if mailer == nil {
			http.Error(w, "Salva prima le impostazioni del server di posta.", http.StatusConflict)
			return
		}
		message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
			mailer.from, to.Address, mime.QEncoding.Encode("utf-8", "DXRegia · prova del server di posta"),
			"Questo è un messaggio di prova. Se lo leggi, il server di posta di DXRegia funziona.\r\n")
		if err := mailer.send(to.Address, message); err != nil {
			http.Error(w, "Invio non riuscito: "+smtpErrorText(err), http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// smtpErrorText riduce l'errore del server di posta a una riga leggibile.
func smtpErrorText(err error) string {
	text := strings.Join(strings.Fields(err.Error()), " ")
	if len(text) > 200 {
		text = text[:200]
	}
	return text
}
