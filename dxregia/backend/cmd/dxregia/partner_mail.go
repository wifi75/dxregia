package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/textproto"
	"os"
	"path/filepath"
	"time"
)

type partnerMail struct {
	Call, Recipient, Password, LocalNode, Host string
	Port                                       int
	VerifyAfter                                time.Time
	Status                                     string
	Lang, Contact                              string
	PendingSent                                bool
	QueuedAt, RetryAt                          time.Time
	Attempts                                   int
	WatchUntil                                 time.Time
}

func preparePartnerMail(root string, entry partner) (partnerMail, error) {
	var job partnerMail
	data, err := os.ReadFile(filepath.Join(root, "local", "dxregia-node.json"))
	if err != nil {
		return job, err
	}
	var node nodeConfig
	if json.Unmarshal(data, &node) != nil || node.validate() != nil || node.AdvertisedHost == "" || entry.Email == "" {
		return job, fmt.Errorf("missing public endpoint or partner email")
	}
	return partnerMail{Call: entry.Call, Recipient: entry.Email, LocalNode: node.Node,
		Host: node.AdvertisedHost, Port: node.Port, Lang: entry.Lang, Contact: entry.Contact, QueuedAt: time.Now(), Status: "In attesa dell’abilitazione"}, nil
}

func writePartnerMail(path string, job partnerMail) error {
	return writePrivateJSON(path, job)
}

func writePrivateJSON(path string, value any) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".dxregia-json-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	err = json.NewEncoder(tmp).Encode(value)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func processPartnerMail(root string, mailer *contactMailer, snapshot presence) {
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	files, err := filepath.Glob(filepath.Join(dir, "*.mail.json"))
	if err != nil {
		return
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		var job partnerMail
		if err != nil || json.Unmarshal(data, &job) != nil {
			continue
		}
		send := func(kind, to string, content mailContent) error {
			return archivePartnerDelivery(root, job.Call, kind, to, job.Password, content, mailer.sendContent)
		}
		if advancePartnerMail(dir, send, snapshot, &job) {
			if err := writePartnerMail(path, job); err != nil {
				return
			}
			return // At most one delivery per worker tick.
		}
	}
}

var partnerMailBackoff = [...]time.Duration{time.Minute, 3 * time.Minute, 10 * time.Minute}

func retryPartnerMail(job *partnerMail, err error, failed string, now time.Time) bool {
	var status *textproto.Error
	job.Attempts++
	if job.Attempts <= len(partnerMailBackoff) && (!errors.As(err, &status) || status.Code < 500) {
		job.RetryAt = now.Add(partnerMailBackoff[job.Attempts-1])
		job.Status = "Invio non riuscito. Nuovo tentativo automatico programmato."
		return true
	}
	job.Attempts, job.RetryAt = 0, time.Time{}
	job.Status = failed
	return false
}

func advancePartnerMail(dir string, send func(string, string, mailContent) error, snapshot presence, job *partnerMail) bool {
	now := time.Now()
	if now.Before(job.RetryAt) {
		return false
	}
	if job.Password != "" {
		result, err := os.ReadFile(filepath.Join(dir, job.Call+".result"))
		if err != nil {
			if !job.QueuedAt.IsZero() && now.Sub(job.QueuedAt) >= 2*time.Minute {
				job.Password = ""
				job.Status = "Nessuna risposta all’abilitazione. Istruzioni non inviate."
				return true
			}
			return false
		}
		if string(result) != "Abilitato in Mojo" {
			job.Password = ""
			job.Status = "Abilitazione non riuscita. Istruzioni non inviate."
			return true
		}
		if err := send("instructions", job.Recipient, mailInstructions(job.Lang, *job, job.Contact)); err != nil {
			if !retryPartnerMail(job, err, "Invio istruzioni non riuscito", now) {
				job.Password = ""
			}
			return true
		}
		job.Password = ""
		job.Attempts, job.RetryAt = 0, time.Time{}
		job.PendingSent = false
		job.VerifyAfter = time.Now().Add(5 * time.Minute)
		job.WatchUntil = time.Now().Add(7 * 24 * time.Hour)
		job.Status = "Istruzioni inviate. In attesa del collegamento."
		return true
	}
	if job.WatchUntil.IsZero() {
		return false
	}
	connected := false
	for _, channel := range snapshot.Channels {
		if channel.Call == job.Call && channel.Kind == "node" && channel.State == "normal" {
			connected = true
			break
		}
	}
	if connected {
		if err := send("confirmed", job.Recipient, mailLinkOK(job.Lang, *job, job.Contact)); err != nil {
			if retryPartnerMail(job, err, "Invio conferma non riuscito", now) {
				return true
			}
		} else {
			job.Attempts, job.RetryAt = 0, time.Time{}
			job.Status = "Conferma inviata: collegamento operativo"
		}
		job.VerifyAfter, job.WatchUntil = time.Time{}, time.Time{}
		return true
	}
	if snapshot.At == 0 {
		return false
	}
	if time.Now().After(job.WatchUntil) {
		job.VerifyAfter, job.WatchUntil = time.Time{}, time.Time{}
		return true
	}
	if !job.PendingSent && !job.VerifyAfter.IsZero() && time.Now().After(job.VerifyAfter) {
		if err := send("pending", job.Recipient, mailLinkPending(job.Lang, *job, job.Contact)); err != nil {
			if retryPartnerMail(job, err, "Invio avviso non riuscito", now) {
				return true
			}
		} else {
			job.Attempts, job.RetryAt = 0, time.Time{}
			job.Status = "Collegamento non ancora attivo: avviso inviato"
		}
		job.PendingSent = true
		return true
	}
	return false
}

func (m *contactMailer) sendContent(to string, content mailContent) error {
	return m.send(to, buildMessage(m.from, to, "", content))
}
