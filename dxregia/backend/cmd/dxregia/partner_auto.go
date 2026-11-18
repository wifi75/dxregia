package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Configurazione automatica: quando un partner chiede il collegamento dal modulo Contatti, il nodo
// crea la scheda, abilita l'accesso in ingresso e gli manda le istruzioni. Non tocca mai un nodo già
// registrato e non sostituisce un account Mojo esistente: in quei casi la richiesta resta al sysop.

const autoPartnerNote = "Configurato automaticamente dal modulo Contatti."

type autoPartnerSettings struct {
	Enabled    bool
	DailyLimit int
}

func autoPartnerPath(root string) string {
	return filepath.Join(root, "local_data", "dxregia-autopartner.json")
}

func readAutoPartner(root string) autoPartnerSettings {
	settings := autoPartnerSettings{Enabled: true, DailyLimit: 10}
	if data, err := os.ReadFile(autoPartnerPath(root)); err == nil {
		_ = json.Unmarshal(data, &settings)
	}
	if settings.DailyLimit < 1 || settings.DailyLimit > 100 {
		settings.DailyLimit = 10
	}
	return settings
}

func writeAutoPartner(root string, settings autoPartnerSettings) error {
	dir := filepath.Dir(autoPartnerPath(root))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".autopartner-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	err = json.NewEncoder(tmp).Encode(settings)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), autoPartnerPath(root))
}

// queueAutoPartner mette in coda i dati del nodo: li raccoglie il processo dei partner entro pochi secondi.
func queueAutoPartner(root string, details partnerDetails) error {
	if !details.validate() {
		return os.ErrInvalid
	}
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.Marshal(details)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, details.Call+".auto.json"), data, 0600)
}

// processAutoPartners crea le schede dei partner in coda. Va chiamata con il blocco dei partner già preso.
func processAutoPartners(root, path string, partners map[string]partner, mailReady bool, now time.Time) {
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	files, _ := filepath.Glob(filepath.Join(dir, "*.auto.json"))
	if len(files) == 0 {
		return
	}
	settings := readAutoPartner(root)
	changed := false
	for _, file := range files {
		data, err := os.ReadFile(file)
		_ = os.Remove(file) // ogni richiesta si elabora una sola volta
		var details partnerDetails
		if err != nil || json.Unmarshal(data, &details) != nil || !details.validate() {
			continue
		}
		if !settings.Enabled || !mailReady || details.Email == "" {
			continue
		}
		if _, exists := partners[details.Call]; exists {
			continue
		}
		today := 0
		for _, entry := range partners {
			for _, note := range entry.Notes {
				if note.Text == autoPartnerNote && now.Sub(note.At) < 24*time.Hour {
					today++
				}
			}
		}
		if today >= settings.DailyLimit {
			log.Printf("Auto partner: daily limit reached, %s left for the sysop", details.Call)
			continue
		}
		entry := partner{partnerDetails: details, RegisteredAt: now.UTC(), Provisioning: "In attesa di Mojo",
			Notes: []partnerNote{{Text: autoPartnerNote, At: now.UTC()}}}
		job, err := preparePartnerMail(root, entry)
		if err != nil {
			log.Printf("Auto partner %s: %v", details.Call, err)
			continue
		}
		password, err := queuePartner(root, details.Call)
		if err != nil {
			log.Printf("Auto partner %s: %v", details.Call, err)
			continue
		}
		job.Password = password
		if err := writePartnerMail(filepath.Join(dir, details.Call+".mail.json"), job); err != nil {
			log.Printf("Auto partner %s mail: %v", details.Call, err)
		}
		partners[details.Call] = entry
		changed = true
	}
	if changed {
		if err := savePartners(path, partners); err != nil {
			log.Printf("Save auto partners: %v", err)
		}
	}
}

func registerAutoPartner(mux *http.ServeMux, root string, auth *sysopAuth) {
	if auth == nil {
		return
	}
	mux.HandleFunc("GET /api/auto-partner", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(readAutoPartner(root))
	})
	mux.HandleFunc("PUT /api/auto-partner", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		var input autoPartnerSettings
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF || input.DailyLimit < 1 || input.DailyLimit > 100 {
			http.Error(w, "Il limite giornaliero va da 1 a 100.", http.StatusBadRequest)
			return
		}
		if err := writeAutoPartner(root, input); err != nil {
			http.Error(w, "Impossibile salvare l’impostazione.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(readAutoPartner(root))
	})
}
