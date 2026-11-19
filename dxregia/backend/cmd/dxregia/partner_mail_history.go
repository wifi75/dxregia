package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type partnerMailRecord struct {
	ID, Call, Recipient, Kind, Subject, Text, Status string
	At                                               time.Time
}

const partnerMailPageSize = 50

func partnerMailHistoryDir(root string) string {
	return filepath.Join(root, "local_data", "dxregia-partner-mail")
}

func archivePartnerDelivery(root, call, kind, to, password string, content mailContent, send func(string, mailContent) error) error {
	at := time.Now().UTC()
	token, err := randomToken()
	if err != nil {
		return err
	}
	dir := partnerMailHistoryDir(root)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	redact := func(value string) string {
		if password == "" {
			return value
		}
		return strings.ReplaceAll(value, password, "••••••••")
	}
	record := partnerMailRecord{
		ID: fmt.Sprintf("%020d-%s", at.UnixNano(), token), At: at,
		Call: redact(call), Kind: kind, Recipient: redact(to), Subject: redact(content.Subject), Text: redact(content.Text),
		Status: "unknown",
	}
	path := filepath.Join(dir, record.ID+".json")
	if err := writePrivateJSON(path, record); err != nil {
		return err
	}
	deliveryErr := send(to, content)
	record.Status = "accepted"
	if deliveryErr != nil {
		record.Status = "failed"
	}
	if err := writePrivateJSON(path, record); err != nil {
		log.Printf("Save partner mail history status: %v", err)
	}
	return deliveryErr
}

func registerPartnerMailHistory(mux *http.ServeMux, root string, auth *sysopAuth) {
	if root == "" || auth == nil {
		return
	}
	mux.HandleFunc("GET /api/partners/mail-history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		dir := partnerMailHistoryDir(root)
		files, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			http.Error(w, "Impossibile leggere lo storico email.", http.StatusInternalServerError)
			return
		}
		query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
		before := r.URL.Query().Get("before")
		items := make([]partnerMailRecord, 0, partnerMailPageSize)
		var next string
		for i := len(files) - 1; i >= 0; i-- {
			file := files[i]
			id, ok := strings.CutSuffix(file.Name(), ".json")
			if !ok || !file.Type().IsRegular() || (before != "" && id >= before) {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, file.Name()))
			var record partnerMailRecord
			if err != nil || json.Unmarshal(data, &record) != nil || record.ID != id {
				continue
			}
			if !strings.Contains(strings.ToLower(record.Call+" "+record.Recipient+" "+record.Subject), query) {
				continue
			}
			if len(items) == partnerMailPageSize {
				next = items[len(items)-1].ID
				break
			}
			items = append(items, record)
		}
		w.Header().Set("Content-Type", "application/json")
		type latestStatus struct{ Call, Recipient, Status string }
		latest := make([]latestStatus, 0)
		if before == "" {
			paths, _ := filepath.Glob(filepath.Join(root, "local_data", "dxregia-partners", "*.mail.json"))
			for _, path := range paths {
				var job partnerMail
				data, err := os.ReadFile(path)
				if err == nil && json.Unmarshal(data, &job) == nil && job.Status != "" &&
					strings.Contains(strings.ToLower(job.Call+" "+job.Recipient+" "+job.Status), query) {
					latest = append(latest, latestStatus{job.Call, job.Recipient, job.Status})
				}
			}
		}
		json.NewEncoder(w).Encode(map[string]any{"Items": items, "Next": next, "Latest": latest})
	})
}
