package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type contactInput struct {
	partnerDetails
	Kind    string
	Message string
}

type contactRequest struct {
	contactInput
	ID         string
	ReceivedAt time.Time
	MailStatus string
}

func registerContacts(ctx context.Context, mux *http.ServeMux, root string, auth *sysopAuth) error {
	if auth == nil {
		return nil
	}
	path := filepath.Join(root, ".dxregia-contacts.json")
	config := filepath.Join(root, "local", "dxregia-node.json")
	mailer, err := loadContactMailer()
	if err != nil {
		return err
	}
	var requests []contactRequest
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &requests); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	var lock sync.Mutex
	if mailer != nil {
		go func() {
			timer := time.NewTicker(30 * time.Second)
			defer timer.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					lock.Lock()
					index := slices.IndexFunc(requests, func(item contactRequest) bool { return item.MailStatus == "In coda" })
					var item contactRequest
					if index >= 0 {
						item = requests[index]
					}
					lock.Unlock()
					if index < 0 {
						continue
					}
					status := mailer.notify(config, item)
					lock.Lock()
					requests[index].MailStatus = status
					if err := saveContacts(path, requests); err != nil {
						requests[index].MailStatus = "Esito email da verificare"
					}
					lock.Unlock()
				}
			}
		}()
	}
	mux.HandleFunc("POST /api/contact", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return
		}
		var input contactInput
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF || !input.valid() {
			http.Error(w, "Controlla indicativo, email, messaggio e, per il peering, host e porta.", http.StatusBadRequest)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		now := time.Now().UTC()
		today := 0
		for _, item := range requests {
			if now.Sub(item.ReceivedAt) < 24*time.Hour {
				today++
			}
		}
		if len(requests) >= 1000 || today >= 40 {
			http.Error(w, "Il modulo ha raggiunto il limite di invii. Riprova più tardi.", http.StatusTooManyRequests)
			return
		}
		id, err := randomToken()
		if err != nil {
			http.Error(w, "Richiesta non registrata.", http.StatusInternalServerError)
			return
		}
		item := contactRequest{contactInput: input, ID: id, ReceivedAt: now, MailStatus: "SMTP non configurato"}
		if mailer != nil {
			item.MailStatus = "In coda"
		}
		if err := saveContacts(path, append(requests, item)); err != nil {
			http.Error(w, "Richiesta non registrata. Riprova.", http.StatusInternalServerError)
			return
		}
		requests = append(requests, item)
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("GET /api/contact-requests", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(requests)
	})
	return nil
}

func (c *contactInput) valid() bool {
	c.Message = strings.TrimSpace(c.Message)
	if len(c.Message) > 3000 || c.Kind != "message" && c.Kind != "partner" {
		return false
	}
	if c.Kind == "message" {
		if c.Message == "" {
			return false
		}
		c.Host, c.Port, c.Name = "", 0, ""
		test := c.partnerDetails
		test.Host, test.Port = "localhost", 1
		if !test.validate() || test.Email == "" {
			return false
		}
		c.Call, c.Contact, c.Email = test.Call, test.Contact, test.Email
		return true
	}
	return c.partnerDetails.validate() && c.Email != "" && c.Name != ""
}

func saveContacts(path string, requests []contactRequest) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".contact-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	err = json.NewEncoder(tmp).Encode(requests)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
