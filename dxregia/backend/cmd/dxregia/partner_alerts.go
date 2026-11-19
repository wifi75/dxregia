package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type partnerOutage struct {
	Since, Recovery, RetryAfter time.Time
	Notified                    bool
}

func processPartnerAlerts(root string, snapshot presence, now time.Time) {
	prefs, err := readEmailPreferences(root)
	if err != nil {
		return
	}
	path := filepath.Join(root, "local_data", "dxregia-partner-outages.json")
	if !prefs.PartnerDown {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Printf("Clear partner outage state: %v", err)
		}
		return
	}
	if !snapshot.PartnerConnect {
		return
	}
	var node nodeConfig
	data, err := os.ReadFile(filepath.Join(root, "local", "dxregia-node.json"))
	if err == nil {
		_ = json.Unmarshal(data, &node)
	}
	destination := prefs.destination(node.Email)
	mailer := currentMailer(root)
	var partners map[string]partner
	data, err = os.ReadFile(filepath.Join(root, ".dxregia-partners.json"))
	if err != nil || json.Unmarshal(data, &partners) != nil {
		return
	}
	outages := make(map[string]partnerOutage)
	if data, err := os.ReadFile(path); err == nil {
		if json.Unmarshal(data, &outages) != nil || outages == nil {
			return
		}
	} else if !os.IsNotExist(err) {
		return
	}
	send := func(content mailContent) error {
		if mailer == nil || destination == "" {
			return fmt.Errorf("SMTP o destinatario non configurato")
		}
		return mailer.sendContent(destination, content)
	}
	if advancePartnerAlerts(root, snapshot, now, prefs, partners, outages, send) {
		if err := writePrivateJSON(path, outages); err != nil {
			log.Printf("Save partner outage state: %v", err)
		}
	}
}

func advancePartnerAlerts(root string, snapshot presence, now time.Time, prefs emailPreferences, partners map[string]partner, outages map[string]partnerOutage, send func(mailContent) error) bool {
	if !snapshot.PartnerConnect || !prefs.PartnerDown {
		return false
	}
	threshold := time.Duration(prefs.PartnerDownMinutes) * time.Minute
	changed := false
	attempted := false
	for call, entry := range partners {
		if call != entry.Call || !entry.partnerDetails.validate() {
			continue
		}
		outage, exists := outages[call]
		state := currentConnection(root, call, snapshot)
		if state == "Attivo" {
			if !exists {
				continue
			}
			if !outage.Notified || !prefs.PartnerRecovered {
				delete(outages, call)
				changed = true
				continue
			}
			if outage.Recovery.IsZero() {
				outage.Recovery, outage.RetryAfter = now, time.Time{}
				outages[call] = outage
				changed = true
			}
			if !attempted && !now.Before(outage.RetryAfter) {
				attempted = true
				if send(partnerAlertMail(entry, outage, "", true, now)) == nil {
					delete(outages, call)
				} else {
					outage.RetryAfter = now.Add(5 * time.Minute)
					outages[call] = outage
				}
				changed = true
			}
			continue
		}
		if !exists {
			// Saved directory entries with no established session or outgoing attempt are not monitored.
			plan, _ := readReconnect(root, call)
			if entry.FirstConnectedAt == nil && (plan.Disabled || plan.LastAttempt.IsZero()) && state == "Disconnesso" {
				continue
			}
		}
		if !exists || !outage.Recovery.IsZero() {
			outage = partnerOutage{Since: now}
			outages[call] = outage
			changed = true
		}
		if !attempted && !outage.Notified && now.Sub(outage.Since) >= threshold && !now.Before(outage.RetryAfter) {
			attempted = true
			reason := partnerOutageReason(root, entry, state, snapshot, outage.Since, now)
			if send(partnerAlertMail(entry, outage, reason, false, now)) == nil {
				outage.Notified = true
			} else {
				outage.RetryAfter = now.Add(5 * time.Minute)
			}
			outages[call] = outage
			changed = true
		}
	}
	for call := range outages {
		if _, exists := partners[call]; !exists {
			delete(outages, call)
			changed = true
		}
	}
	return changed
}

func partnerOutageReason(root string, entry partner, state string, snapshot presence, since, now time.Time) string {
	if _, locked := snapshot.LockedPartners[entry.Call]; locked || entry.Provisioning == lockedPartnerStatus {
		return "Nodo bloccato in Mojo"
	}
	address := net.JoinHostPort(entry.Host, strconv.Itoa(entry.Port))
	utc := now.UTC()
	for day := 0; day < 2; day++ {
		date := utc.AddDate(0, 0, -day)
		path := filepath.Join(root, "local_data", "debug", fmt.Sprintf("%04d", date.Year()), fmt.Sprintf("%03d.dat", date.YearDay()))
		tail := logTail(path)
		parse := mojoDebugParser(tail)
		lines := strings.Split(tail, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			event, ok := parse(strings.TrimSpace(lines[i]))
			if !ok || event.At < since.Unix() || event.Direction != "Uscita" {
				continue
			}
			if (event.Call == entry.Call && (strings.Contains(event.Message, "non ricevuto") || strings.Contains(event.Message, "non completato"))) ||
				(strings.HasPrefix(event.Message, "Connessione TCP non riuscita verso ") && strings.HasSuffix(event.Message, " "+address)) {
				return event.Message
			}
		}
	}
	if state != "Disconnesso" {
		return state + ". Mojo non ha registrato una causa più precisa."
	}
	return "Sessione DXSpider assente. Mojo non ha registrato una causa più precisa."
}

func partnerAlertMail(entry partner, outage partnerOutage, reason string, recovered bool, now time.Time) mailContent {
	end := now
	if recovered {
		end = outage.Recovery
	}
	rows := []mailRow{
		{"Nodo partner", entry.Call, false},
		{"Host e porta", net.JoinHostPort(entry.Host, strconv.Itoa(entry.Port)), false},
		{"Assenza rilevata dalle", outage.Since.UTC().Format("02/01/2006 15:04:05 UTC"), false},
		{"Durata rilevata", fmt.Sprintf("%d min", max(0, int(end.Sub(outage.Since)/time.Minute))), false},
	}
	if recovered {
		rows = append(rows, mailRow{"Ripristinato alle", end.UTC().Format("02/01/2006 15:04:05 UTC"), false})
		return mailForSysop("DXRegia · "+entry.Call+" di nuovo collegato", "Collegamento ripristinato", "Il nodo partner ha di nuovo una sessione DXSpider attiva.", rows, "")
	}
	rows = append(rows, mailRow{"Ultimo esito disponibile", reason, false})
	return mailForSysop("DXRegia · "+entry.Call+" scollegato", "Nodo partner scollegato", "DXRegia non vede una sessione attiva per il partner indicato.", rows, "")
}
