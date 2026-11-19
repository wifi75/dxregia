package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/netip"
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
	LastConnectAttempt   *time.Time `json:",omitempty"`
	RetryEligibleAt      *time.Time `json:",omitempty"`
	PublicMailAcceptedAt *time.Time `json:",omitempty"`
	InboundTestUntil     *time.Time `json:",omitempty"`
	InboundAutoNext      *time.Time `json:",omitempty"`
	MailRetryAt          *time.Time `json:",omitempty"`
	IPMailStatus         string     `json:",omitempty"`
	OutgoingReady        bool
	CanResetCCInbound    bool
	CanSetInboundIP      bool   `json:",omitempty"`
	InboundIP            string `json:",omitempty"`
	CanResend            bool   `json:",omitempty"`
	Locked               bool   `json:",omitempty"`
	CanUnlock            bool   `json:",omitempty"`
	Direction            string `json:",omitempty"`
}

type partnerNote struct {
	Text string
	At   time.Time
}

type partner struct {
	partnerDetails
	RegisteredAt     time.Time
	FirstConnectedAt *time.Time
	FirstInboundAt   *time.Time
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
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					m := mailer()
					snapshot := readPresence(filepath.Join(root, "local_data", "dxregia-presence.json"))
					if m != nil {
						processPartnerMail(root, m, snapshot)
					}
					processPartnerAlerts(root, snapshot, time.Now().UTC())
				}
			}
		}()
	}
	var lock sync.Mutex
	readProvisioningResults(root, path, partners)
	go refreshInboundDNS(ctx, root)
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
				now := time.Now()
				runAutomaticInboundTests(root, partners, snapshot, now)
				reconnectPartners(root, partners, snapshot, now)
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
				status := partnerStatus{partner: entry, OutgoingReady: partnerOutgoingReady(entry, snapshot)}
				if until := inboundTestUntil(root, call); until.After(time.Now()) {
					status.InboundTestUntil = &until
				}
				status.CanSetInboundIP = snapshot.CCInboundReset[call]
				status.InboundIP = partnerInboundIP(root, call)
				status.CanResetCCInbound = snapshot.CCInboundReset[call] && entry.Provisioning != "In attesa di Mojo" && entry.Provisioning != "Abilitato in Mojo"
				status.CanUnlock, status.Locked = snapshot.LockedPartners[call]
				if snapshot.LockedPartners == nil && entry.Provisioning == lockedPartnerStatus {
					status.Locked, status.CanUnlock = true, true
				}
				status.CanUnlock = status.CanUnlock && entry.Provisioning != "In attesa di Mojo"
				for _, channel := range snapshot.Channels {
					if channel.Call == call && channel.Kind == "node" {
						status.Direction = "Ingresso"
						if channel.Outbound {
							status.Direction = "Uscita"
						}
					}
				}
				if callsign.MatchString(call) && len(call) <= 12 {
					if next, err := inboundAutoNext(root, call); err == nil {
						status.InboundAutoNext = &next
					}
					if data, err := os.ReadFile(publicMailAcceptedPath(root, call)); err == nil {
						var accepted time.Time
						if json.Unmarshal(data, &accepted) == nil {
							status.PublicMailAcceptedAt = &accepted
						}
					}
					if plan, err := readReconnect(root, call); err == nil && !plan.Disabled {
						if !plan.LastAttempt.IsZero() {
							status.LastConnectAttempt = &plan.LastAttempt
						}
						status.Reconnect = "Abilitata"
						if plan.Host != entry.Host || plan.Port != entry.Port {
							status.Reconnect = "Sospesa: indirizzo modificato"
						} else if plan.eligible(entry, entry.Connection, snapshot) {
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
					if !job.RetryAt.IsZero() {
						status.MailRetryAt = &job.RetryAt
					}
				}
				if status.InboundIP != "" {
					var ipJob partnerMail
					dir := filepath.Join(root, "local_data", "dxregia-partners")
					if data, err := os.ReadFile(partnerIPMailPath(dir, call)); err == nil &&
						json.Unmarshal(data, &ipJob) == nil && ipJob.InboundSource == status.InboundIP {
						status.IPMailStatus = ipJob.Status
					}
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
		var options struct{ SendMail, OutgoingOnly, Unlock, ResetCCInbound, SwitchFromIP bool }
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
		snapshot := readPresence(filepath.Join(root, "local_data", "dxregia-presence.json"))
		if options.SwitchFromIP && (!options.ResetCCInbound || !options.SendMail || partnerInboundIP(root, call) == "") {
			http.Error(w, "Per passare dall’IP alla password serve una nuova password da inviare al partner.", http.StatusConflict)
			return
		}
		if options.ResetCCInbound && (options.OutgoingOnly || options.Unlock || !snapshot.CCInboundReset[call]) {
			http.Error(w, "Questo account CC Cluster non può essere aggiornato dal pannello.", http.StatusConflict)
			return
		}
		unlockable, locked := snapshot.LockedPartners[call]
		if snapshot.LockedPartners == nil && entry.Provisioning == lockedPartnerStatus {
			unlockable, locked = true, true
		}
		if options.Unlock && (!options.OutgoingOnly || options.SendMail || !unlockable) {
			http.Error(w, "Lo sblocco richiede un nodo partner bloccato e una richiesta di preparazione in uscita.", http.StatusConflict)
			return
		}
		if locked && !options.Unlock {
			http.Error(w, lockedPartnerStatus, http.StatusConflict)
			return
		}
		if entry.Provisioning == "In attesa di Mojo" || (!options.Unlock && !options.SwitchFromIP && entry.Provisioning == "Abilitato in Mojo") {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if options.OutgoingOnly {
			if !options.Unlock && entry.Provisioning == "Pronto in uscita" {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			prefix := "outgoing-only:"
			if options.Unlock {
				prefix = "unlock-outgoing:"
			}
			if err := queueOutgoingPartnerProvision(root, call, prefix); err != nil {
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
		password, err := queuePartnerCredential(root, call, options.ResetCCInbound)
		if err != nil {
			log.Printf("Queue partner: %v", err)
			http.Error(w, "Impossibile inviare la richiesta a Mojo.", http.StatusServiceUnavailable)
			return
		}
		mailStatus := ""
		if options.SendMail {
			job.Password = password
			job.SwitchFromIP = options.SwitchFromIP
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
	mux.HandleFunc("POST /api/partners/{call}/send-node-details", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		call := r.PathValue("call")
		lock.Lock()
		entry, ok := partners[call]
		lock.Unlock()
		if !ok || !callsign.MatchString(call) || len(call) > 12 {
			http.Error(w, "Nodo non trovato.", http.StatusNotFound)
			return
		}
		smtp := mailer()
		if smtp == nil {
			http.Error(w, "Configura SMTP prima di inviare i dati del nodo.", http.StatusConflict)
			return
		}
		job, err := preparePartnerMail(root, entry)
		if err != nil {
			http.Error(w, "Per l’email servono l’indirizzo del referente e l’host pubblico del nodo.", http.StatusBadRequest)
			return
		}
		if err := archivePartnerDelivery(root, call, "node-details", job.Recipient, "", mailNodeDetails(job), smtp.sendContent); err != nil {
			http.Error(w, "Invio dei dati del nodo non confermato. Controlla lo storico email e la configurazione SMTP.", http.StatusBadGateway)
			return
		}
		if err := writePrivateJSON(publicMailAcceptedPath(root, call), time.Now().UTC()); err != nil {
			log.Printf("Save public partner mail confirmation: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
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
	inboundIP := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		call := r.PathValue("call")
		lock.Lock()
		locked := true
		defer func() {
			if locked {
				lock.Unlock()
			}
		}()
		entry, ok := partners[call]
		if !ok || !validPartnerIPCall(call) {
			http.NotFound(w, r)
			return
		}
		path := partnerInboundIPPath(root, call)
		sendMail := r.Method == http.MethodPut && r.URL.Query().Get("send-mail") == "true"
		var job partnerMail
		var smtp *contactMailer
		var source string
		var err error
		if r.Method == http.MethodDelete {
			err = os.Remove(path)
			if os.IsNotExist(err) {
				err = nil
			}
		} else {
			var address string
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
			if decoder.Decode(&address) != nil || decoder.Decode(new(any)) != io.EOF {
				http.Error(w, "Indirizzo IP non valido.", http.StatusBadRequest)
				return
			}
			address = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(address), "."))
			ip, parseErr := netip.ParseAddr(address)
			host := parseErr != nil && validInboundDNSHost(address)
			if !host && (parseErr != nil || !publicPartnerIP(ip.Unmap())) {
				http.Error(w, "Inserisci un IP pubblico o un nome host del partner.", http.StatusBadRequest)
				return
			}
			snapshot := readPresence(filepath.Join(root, "local_data", "dxregia-presence.json"))
			if !snapshot.CCInboundReset[call] {
				http.Error(w, "Serve un account CC Cluster locale e sbloccato.", http.StatusConflict)
				return
			}
			var pending partnerMail
			if data, readErr := os.ReadFile(filepath.Join(filepath.Dir(path), call+".mail.json")); readErr == nil {
				_ = json.Unmarshal(data, &pending)
			}
			if entry.Provisioning == "In attesa di Mojo" || pending.SwitchFromIP && pending.Password != "" {
				http.Error(w, "Attendi l’esito della configurazione in corso prima di cambiare modalità.", http.StatusConflict)
				return
			}
			if sendMail {
				smtp = mailer()
				if smtp == nil {
					http.Error(w, "Configura SMTP prima di inviare le istruzioni al partner.", http.StatusConflict)
					return
				}
				job, err = preparePartnerMail(root, entry)
				if err != nil {
					http.Error(w, "Per l’email servono l’indirizzo del referente e l’host pubblico del nodo.", http.StatusBadRequest)
					return
				}
			}
			if err = os.MkdirAll(filepath.Dir(path), 0700); err == nil {
				if host {
					resolved, resolveErr := resolveInboundDNS(r.Context(), address, net.DefaultResolver.LookupNetIP)
					if resolveErr != nil {
						http.Error(w, "Il nome host non risolve a un indirizzo IP pubblico.", http.StatusBadRequest)
						return
					}
					err = writePrivateJSON(inboundDNSPath(path), resolved)
				} else {
					address = ip.Unmap().String()
				}
				if err == nil {
					source = address
					err = writePrivateJSON(path, source)
				}
			}
		}
		if err != nil {
			http.Error(w, "Impossibile aggiornare l’accesso per IP.", http.StatusServiceUnavailable)
			return
		}
		lock.Unlock()
		locked = false
		if sendMail {
			content := mailIPInstructions(job)
			if err := archivePartnerDelivery(root, call, "ip-instructions", job.Recipient, "", content, smtp.sendContent); err != nil {
				http.Error(w, "Accesso per IP salvato. L’invio dell’email non è confermato: controlla SMTP e riprova.", http.StatusBadGateway)
				return
			}
			lock.Lock()
			_, exists := partners[call]
			if exists && partnerInboundIP(root, call) == source {
				job.InboundSource = source
				job.WatchUntil = time.Now().Add(7 * 24 * time.Hour)
				job.Status = "Istruzioni senza password inviate. In attesa del collegamento."
				jobPath := partnerIPMailPath(filepath.Dir(path), call)
				err = writePartnerMail(jobPath, job)
			}
			lock.Unlock()
			if err != nil {
				http.Error(w, "Email accettata da SMTP. Impossibile programmare la conferma del collegamento.", http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}
	mux.HandleFunc("PUT /api/partners/{call}/inbound-ip", inboundIP)
	mux.HandleFunc("DELETE /api/partners/{call}/inbound-ip", inboundIP)
	inboundTest := func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		call := r.PathValue("call")
		lock.Lock()
		defer lock.Unlock()
		entry, ok := partners[call]
		if !ok || !callsign.MatchString(call) || len(call) > 12 {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			if err := os.Remove(inboundAutoPath(root, call)); err != nil && !os.IsNotExist(err) {
				http.Error(w, "Impossibile disattivare la prova automatica.", http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		snapshot := readPresence(filepath.Join(root, "local_data", "dxregia-presence.json"))
		if _, locked := snapshot.LockedPartners[call]; !snapshot.PartnerConnect || locked || !partnerInboundEnabled(root, entry, snapshot) {
			http.Error(w, "Per la prova serve un nodo attivo con accesso in ingresso abilitato.", http.StatusConflict)
			return
		}
		if r.URL.Query().Get("automatic") == "true" {
			if err := os.MkdirAll(filepath.Dir(inboundAutoPath(root, call)), 0700); err != nil {
				http.Error(w, "Impossibile attivare la prova automatica.", http.StatusServiceUnavailable)
				return
			}
			if _, err := inboundAutoNext(root, call); err != nil {
				if err := writePrivateJSON(inboundAutoPath(root, call), time.Now()); err != nil {
					http.Error(w, "Impossibile attivare la prova automatica.", http.StatusServiceUnavailable)
					return
				}
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if hasInboundSession(snapshot, call) {
			http.Error(w, "Il partner è già collegato in ingresso.", http.StatusConflict)
			return
		}
		if inboundTestUntil(root, call).After(time.Now()) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		until := time.Now().Add(inboundTestWindow)
		if err := writePrivateJSON(inboundTestPath(root, call), until.Unix()); err != nil {
			http.Error(w, "Impossibile avviare la prova.", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
	mux.HandleFunc("POST /api/partners/{call}/inbound-test", inboundTest)
	mux.HandleFunc("DELETE /api/partners/{call}/inbound-test", inboundTest)
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
		if _, locked := snapshot.LockedPartners[call]; locked {
			http.Error(w, lockedPartnerStatus, http.StatusConflict)
			return
		}
		if inboundTestUntil(root, call).After(time.Now()) {
			http.Error(w, "Il collegamento in uscita è sospeso durante la prova in ingresso.", http.StatusConflict)
			return
		}
		if !partnerOutgoingReady(entry, snapshot) {
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
	setReconnect := func(w http.ResponseWriter, r *http.Request) {
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
		entry, ok := partners[call]
		if !ok || !entry.partnerDetails.validate() || entry.Call != call {
			http.Error(w, "Nodo non trovato o configurazione non valida.", http.StatusNotFound)
			return
		}
		plan, err := readReconnect(root, call)
		if r.Method == http.MethodPut {
			var password string
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256))
			if decoder.Decode(&password) != nil || decoder.Decode(new(any)) != io.EOF || !peerPasswordPattern.MatchString(password) {
				http.Error(w, "Per il collegamento usa una password alfanumerica di massimo 72 caratteri.", http.StatusBadRequest)
				return
			}
			if err == nil || os.IsNotExist(err) {
				plan = reconnectPlan{Host: entry.Host, Port: entry.Port, Password: password, LastAttempt: plan.LastAttempt, AutoProvision: plan.AutoProvision}
				err = saveReconnect(root, call, plan)
			}
		} else if err == nil {
			plan.Disabled = true
			err = saveReconnect(root, call, plan)
		} else if os.IsNotExist(err) {
			err = nil
		}
		if err != nil {
			http.Error(w, "Impossibile salvare la riconnessione automatica.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
	mux.HandleFunc("PUT /api/partners/{call}/reconnect", setReconnect)
	mux.HandleFunc("DELETE /api/partners/{call}/reconnect", setReconnect)
	return nil
}

func recordFirstConnections(path string, partners map[string]partner, snapshot presence) {
	type before struct{ first, inbound *time.Time }
	changed := map[string]before{}
	for _, channel := range snapshot.Channels {
		if channel.Kind != "node" || channel.State != "normal" || channel.Since <= 0 || channel.Since > snapshot.At {
			continue
		}
		entry, ok := partners[channel.Call]
		if !ok {
			continue
		}
		since := time.Unix(channel.Since, 0).UTC()
		old := before{entry.FirstConnectedAt, entry.FirstInboundAt}
		touched := false
		if entry.FirstConnectedAt == nil {
			entry.FirstConnectedAt, touched = &since, true
		}
		if !channel.Outbound && entry.FirstInboundAt == nil {
			entry.FirstInboundAt, touched = &since, true
		}
		if touched {
			partners[channel.Call] = entry
			if _, seen := changed[channel.Call]; !seen {
				changed[channel.Call] = old
			}
		}
	}
	if len(changed) == 0 {
		return
	}
	if err := savePartners(path, partners); err != nil {
		log.Printf("Save first partner connections: %v", err)
		for call, old := range changed {
			entry := partners[call]
			entry.FirstConnectedAt, entry.FirstInboundAt = old.first, old.inbound
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

func publicMailAcceptedPath(root, call string) string {
	return filepath.Join(root, "local_data", "dxregia-partners", call+".public-mail-accepted.json")
}
