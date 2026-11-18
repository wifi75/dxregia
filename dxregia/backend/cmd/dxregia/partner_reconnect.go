package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

const reconnectDelay = 2 * time.Minute

type reconnectPlan struct {
	Host          string
	Port          int
	Password      string
	LastAttempt   time.Time
	AutoProvision bool `json:",omitempty"`
}

func reconnectPath(root, call string) string {
	return filepath.Join(root, "local_data", "dxregia-partners", call+".reconnect.json")
}

func readReconnect(root, call string) (reconnectPlan, error) {
	var plan reconnectPlan
	data, err := os.ReadFile(reconnectPath(root, call))
	if err == nil {
		err = json.Unmarshal(data, &plan)
	}
	return plan, err
}

func saveReconnect(root, call string, plan reconnectPlan) error {
	path := reconnectPath(root, call)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".reconnect-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	err = json.NewEncoder(tmp).Encode(plan)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func (p reconnectPlan) eligible(entry partner, state string) bool {
	if !canConnectPartner(entry.Provisioning) || p.Host != entry.Host || p.Port != entry.Port ||
		!peerPasswordPattern.MatchString(p.Password) {
		return false
	}
	return state == "Disconnesso" || state == "Collegamento non confermato" || state == "Avvio non riuscito"
}

func (p reconnectPlan) due(entry partner, state string, now time.Time) bool {
	return p.eligible(entry, state) && !now.Before(p.LastAttempt.Add(reconnectDelay))
}

func reconnectPartners(root string, partners map[string]partner, snapshot presence, now time.Time) {
	if !snapshot.PartnerConnect {
		return
	}
	for call, entry := range partners {
		if !entry.partnerDetails.validate() || call != entry.Call {
			continue
		}
		plan, err := readReconnect(root, call)
		if err != nil || plan.Host != entry.Host || plan.Port != entry.Port {
			continue
		}
		if plan.AutoProvision && entry.Provisioning == "" {
			if err := queueOutgoingPartner(root, call); err != nil {
				log.Printf("Queue outgoing partner %s: %v", call, err)
			}
			continue
		}
		if !plan.due(entry, currentConnection(root, call, snapshot), now) {
			continue
		}
		plan.LastAttempt = now
		if err := saveReconnect(root, call, plan); err != nil {
			log.Printf("Save reconnect schedule for %s: %v", call, err)
			continue
		}
		if err := queueConnection(root, entry, plan.Password); err != nil {
			log.Printf("Queue reconnect for %s: %v", call, err)
		}
	}
}

func canConnectPartner(status string) bool {
	return status == "Abilitato in Mojo" || status == "Pronto in uscita"
}
