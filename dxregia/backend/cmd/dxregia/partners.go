package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
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
	Lang                             string `json:",omitempty"`
	Person                           string `json:",omitempty"` // nome della persona, solo per l'anagrafica
}

type partnerStatus struct {
	partner
	LastConnectAttempt *time.Time `json:",omitempty"`
	RetryEligibleAt    *time.Time `json:",omitempty"`
	CanResend          bool       `json:",omitempty"`
	Direction          string     `json:",omitempty"`
}

type partnerNote struct {
	Text string
	At   time.Time
}

type partner struct {
	partnerDetails
	RegisteredAt     time.Time
	FirstConnectedAt *time.Time
	Provisioning     string
	Connection       string
	MailStatus       string
	Reconnect        string
	Notes            []partnerNote
}

var partnerHost = regexp.MustCompile("^[a-zA-Z0-9]([a-zA-Z0-9.-]*[a-zA-Z0-9])?$")

func (d *partnerDetails) validate() bool {
	d.Call = strings.ToUpper(strings.TrimSpace(d.Call))
	d.Name = strings.TrimSpace(d.Name)
	d.Contact = strings.TrimSpace(d.Contact)
	d.Email = strings.TrimSpace(d.Email)
	d.Host = strings.TrimSpace(d.Host)
	d.Lang = normalizeLang(d.Lang)
	d.Person = strings.TrimSpace(d.Person)
	if d.Email != "" {
		address, err := mail.ParseAddress(d.Email)
		if err != nil || address.Address != d.Email || len(d.Email) > 254 {
			return false
		}
	}
	return len(d.Call) <= 12 && callsign.MatchString(d.Call) && len(d.Name) <= 120 &&
		len(d.Contact) <= 120 && len(d.Person) <= 120 && len(d.Host) <= 253 &&
		(net.ParseIP(d.Host) != nil || partnerHost.MatchString(d.Host)) && d.Port >= 1 && d.Port <= 65535
}

func registerPartners(ctx context.Context, mux *http.ServeMux, root string, auth *sysopAuth) error {
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
	mailer := func() *contactMailer { return currentMailer(root) }
	{
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					m := mailer()
					if m == nil {
						continue
					}
					processPartnerMail(root, m, readPresence(filepath.Join(root, "local_data", "dxregia-presence.json")))
				}
			}
		}()
	}
	var lock sync.Mutex
	readProvisioningResults(root, path, partners)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				snapshot := readPresence(filepath.Join(root, "local_data", "dxregia-presence.json"))
				lock.Lock()
				recordFirstConnections(path, partners, snapshot)
				processAutoPartners(root, path, partners, mailer() != nil, time.Now())
				readProvisioningResults(root, path, partners)
				reconnectPartners(root, partners, snapshot, time.Now())
				lock.Unlock()
			}
		}
	}()

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
			snapshot := readPresence(filepath.Join(root, "local_data", "dxregia-presence.json"))
			view := make(map[string]partnerStatus, len(partners))
			for call, entry := range partners {
				entry.Connection = currentConnection(root, call, snapshot)
				status := partnerStatus{partner: entry}
				for _, channel := range snapshot.Channels {
					if channel.Call == call && channel.Kind == "node" {
						status.Direction = "Ingresso"
						if channel.Outbound {
							status.Direction = "Uscita"
						}
					}
				}
				if callsign.MatchString(call) && len(call) <= 12 {
					if plan, err := readReconnect(root, call); err == nil && !plan.Disabled {
						if !plan.LastAttempt.IsZero() {
							status.LastConnectAttempt = &plan.LastAttempt
						}
						status.Reconnect = "Abilitata"
						if plan.Host != entry.Host || plan.Port != entry.Port {
							status.Reconnect = "Sospesa: indirizzo modificato"
						} else if plan.eligible(entry, entry.Connection) {
							earliest := plan.LastAttempt.Add(reconnectDelay)
							status.RetryEligibleAt = &earliest
						}
					}
				}
				status.CanResend = callsign.MatchString(call) && len(call) <= 12 && entry.Provisioning == "Abilitato in Mojo" &&
					entry.Email != "" && readInboundPassword(root, call) != ""
				var job partnerMail
				if data, err := os.ReadFile(filepath.Join(root, "local_data", "dxregia-partners", call+".mail.json")); err == nil && json.Unmarshal(data, &job) == nil {
					status.MailStatus = job.Status
				}
				view[call] = status
			}
			json.NewEncoder(w).Encode(view)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		var input struct {
			partnerDetails
			Note           string
			ConnectNow     bool
			RemotePassword *string
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF || !input.partnerDetails.validate() || len(input.Note) > 4000 {
			http.Error(w, "Controlla indicativo, host, porta, email e lunghezza dei campi.", http.StatusBadRequest)
			return
		}
		if input.RemotePassword != nil && !peerPasswordPattern.MatchString(*input.RemotePassword) {
			http.Error(w, "La password del nodo remoto può contenere al massimo 72 lettere o cifre.", http.StatusBadRequest)
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
		auto := input.ConnectNow && !exists
		saveCredential := auto || input.RemotePassword != nil
		previousPlan, previousErr := readReconnect(root, call)
		if saveCredential {
			if previousErr != nil && !os.IsNotExist(previousErr) {
				http.Error(w, "Impossibile leggere le credenziali precedenti.", http.StatusInternalServerError)
				return
			}
			plan := reconnectPlan{Host: entry.Host, Port: entry.Port, AutoProvision: auto, Disabled: !auto}
			if previousErr == nil && previousPlan.Host == entry.Host && previousPlan.Port == entry.Port {
				plan = previousPlan
				if auto {
					plan.AutoProvision, plan.Disabled = true, false
				}
			}
			if input.RemotePassword != nil {
				plan.Password = *input.RemotePassword
			}
			if err := saveReconnect(root, call, plan); err != nil {
				http.Error(w, "Impossibile salvare le credenziali del nodo.", http.StatusInternalServerError)
				return
			}
		}
		partners[call] = entry
		if err := savePartners(path, partners); err != nil {
			if saveCredential {
				if previousErr == nil {
					_ = saveReconnect(root, call, previousPlan)
				} else {
					_ = os.Remove(reconnectPath(root, call))
				}
			}
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
	mux.HandleFunc("GET /api/partners/{call}/credential", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		call := r.PathValue("call")
		lock.Lock()
		defer lock.Unlock()
		entry, ok := partners[call]
		if !ok || !callsign.MatchString(call) {
			http.NotFound(w, r)
			return
		}
		password := ""
		if plan, err := readReconnect(root, call); err == nil && plan.Host == entry.Host && plan.Port == entry.Port {
			password = plan.Password
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"password": password})
	})
	mux.HandleFunc("POST /api/partners/{call}/enable", func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		var options struct{ SendMail, OutgoingOnly bool }
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&options); err != io.EOF && (err != nil || decoder.Decode(new(any)) != io.EOF) {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return
		}
		call := r.PathValue("call")
		lock.Lock()
		defer lock.Unlock()
		entry, ok := partners[call]
		if !ok {
			http.Error(w, "Nodo non trovato.", http.StatusNotFound)
			return
		}
		if entry.Provisioning == "In attesa di Mojo" || entry.Provisioning == "Abilitato in Mojo" {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if options.OutgoingOnly {
			if entry.Provisioning == "Pronto in uscita" {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			if err := queueOutgoingPartner(root, call); err != nil {
				http.Error(w, "Impossibile preparare il collegamento in uscita.", http.StatusServiceUnavailable)
				return
			}
			entry.Provisioning = "In attesa di Mojo"
			partners[call] = entry
			w.WriteHeader(http.StatusAccepted)
			return
		}
		var job partnerMail
		if options.SendMail {
			if mailer() == nil {
				http.Error(w, "Configura SMTP prima di inviare le istruzioni al partner.", http.StatusConflict)
				return
			}
			var err error
			job, err = preparePartnerMail(root, entry)
			if err != nil {
				http.Error(w, "Per l’email servono l’indirizzo del referente e l’host pubblico del nodo.", http.StatusBadRequest)
				return
			}
		}
		password, err := queuePartner(root, call)
		if err != nil {
			log.Printf("Queue partner: %v", err)
			http.Error(w, "Impossibile inviare la richiesta a Mojo.", http.StatusServiceUnavailable)
			return
		}
		mailStatus := ""
		if options.SendMail {
			job.Password = password
			if err := writePartnerMail(filepath.Join(root, "local_data", "dxregia-partners", call+".mail.json"), job); err != nil {
				log.Printf("Queue partner instructions: %v", err)
				mailStatus = "Istruzioni non accodate. Puoi inviarle manualmente."
			} else {
				mailStatus = job.Status
			}
		}
		entry.Provisioning = "In attesa di Mojo"
		partners[call] = entry
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"Password": password, "MailStatus": mailStatus})
	})
	mux.HandleFunc("DELETE /api/partners/{call}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		call := r.PathValue("call")
		lock.Lock()
		defer lock.Unlock()
		entry, ok := partners[call]
		if !ok || !callsign.MatchString(call) || len(call) > 12 {
			http.Error(w, "Nodo non trovato.", http.StatusNotFound)
			return
		}
		delete(partners, call)
		if err := savePartners(path, partners); err != nil {
			partners[call] = entry
			http.Error(w, "Impossibile eliminare il nodo.", http.StatusInternalServerError)
			return
		}
		if err := removePartnerFiles(root, call); err != nil {
			log.Printf("Remove partner %s: %v", call, err)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/partners/{call}/inbound-password", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		call := r.PathValue("call")
		lock.Lock()
		defer lock.Unlock()
		entry, ok := partners[call]
		if !ok || !callsign.MatchString(call) || len(call) > 12 || entry.Provisioning != "Abilitato in Mojo" {
			http.NotFound(w, r)
			return
		}
		password := readInboundPassword(root, call)
		if password == "" {
			http.Error(w, "Questa password non è stata conservata: è stata generata prima di questa versione.", http.StatusConflict)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"password": password})
	})
	mux.HandleFunc("POST /api/partners/{call}/resend", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		call := r.PathValue("call")
		lock.Lock()
		defer lock.Unlock()
		entry, ok := partners[call]
		if !ok || !callsign.MatchString(call) || len(call) > 12 {
			http.Error(w, "Nodo non trovato.", http.StatusNotFound)
			return
		}
		if entry.Provisioning != "Abilitato in Mojo" {
			http.Error(w, "L’accesso in ingresso non è abilitato per questo nodo.", http.StatusConflict)
			return
		}
		password := readInboundPassword(root, call)
		if password == "" {
			http.Error(w, "La password di questo nodo non è stata conservata. Non posso reinviarla.", http.StatusConflict)
			return
		}
		if mailer() == nil {
			http.Error(w, "Configura SMTP prima di inviare le istruzioni al partner.", http.StatusConflict)
			return
		}
		job, err := preparePartnerMail(root, entry)
		if err != nil {
			http.Error(w, "Per l’email servono l’indirizzo del referente e l’host pubblico del nodo.", http.StatusBadRequest)
			return
		}
		job.Password = password
		if err := writePartnerMail(filepath.Join(root, "local_data", "dxregia-partners", call+".mail.json"), job); err != nil {
			log.Printf("Queue partner instructions: %v", err)
			http.Error(w, "Istruzioni non accodate. Riprova.", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"MailStatus": job.Status})
	})
	mux.HandleFunc("POST /api/partners/{call}/connect", func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		var password string
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256))
		if decoder.Decode(&password) != nil || decoder.Decode(new(any)) != io.EOF || !peerPasswordPattern.MatchString(password) {
			http.Error(w, "Per il collegamento usa una password alfanumerica di massimo 72 caratteri.", http.StatusBadRequest)
			return
		}
		call := r.PathValue("call")
		lock.Lock()
		defer lock.Unlock()
		entry, ok := partners[call]
		if !ok || !entry.partnerDetails.validate() || entry.Call != call {
			http.Error(w, "Nodo non trovato o configurazione non valida.", http.StatusNotFound)
			return
		}
		snapshot := readPresence(filepath.Join(root, "local_data", "dxregia-presence.json"))
		if !snapshot.PartnerConnect {
			http.Error(w, "Mojo deve essere attivo con l’hook DXRegia aggiornato.", http.StatusServiceUnavailable)
			return
		}
		if !canConnectPartner(entry.Provisioning) {
			http.Error(w, "Abilita prima l’account partner in Mojo.", http.StatusConflict)
			return
		}
		if r.URL.Query().Get("retry") == "true" {
			plan := reconnectPlan{Host: entry.Host, Port: entry.Port, Password: password, LastAttempt: time.Now()}
			if err := saveReconnect(root, call, plan); err != nil {
				http.Error(w, "Impossibile salvare la riconnessione automatica.", http.StatusInternalServerError)
				return
			}
		}
		switch currentConnection(root, call, snapshot) {
		case "Attivo", "In coda", "Collegamento in corso":
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if err := queueConnection(root, entry, password); err != nil {
			log.Printf("Queue partner connection: %v", err)
			http.Error(w, "Impossibile accodare il collegamento.", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("DELETE /api/partners/{call}/reconnect", func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		call := r.PathValue("call")
		if !callsign.MatchString(call) || len(call) > 12 {
			http.Error(w, "Indicativo non valido.", http.StatusBadRequest)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		plan, err := readReconnect(root, call)
		if err == nil {
			plan.Disabled = true
			err = saveReconnect(root, call, plan)
		} else if os.IsNotExist(err) {
			err = nil
		}
		if err != nil {
			http.Error(w, "Impossibile disattivare la riconnessione.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return nil
}

func recordFirstConnections(path string, partners map[string]partner, snapshot presence) {
	changed := make([]string, 0)
	for _, channel := range snapshot.Channels {
		if channel.Kind != "node" || channel.State != "normal" || channel.Since <= 0 || channel.Since > snapshot.At {
			continue
		}
		entry, ok := partners[channel.Call]
		if !ok || entry.FirstConnectedAt != nil {
			continue
		}
		since := time.Unix(channel.Since, 0).UTC()
		entry.FirstConnectedAt = &since
		partners[channel.Call] = entry
		changed = append(changed, channel.Call)
	}
	if len(changed) == 0 {
		return
	}
	if err := savePartners(path, partners); err != nil {
		log.Printf("Save first partner connections: %v", err)
		for _, call := range changed {
			entry := partners[call]
			entry.FirstConnectedAt = nil
			partners[call] = entry
		}
	}
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
