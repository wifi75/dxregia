package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const reconnectDelay = 2 * time.Minute

type reconnectPlan struct {
	Host          string
	Port          int
	Password      string
	LastAttempt   time.Time
	AutoProvision bool `json:",omitempty"`
	Disabled      bool `json:",omitempty"`
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

func (p reconnectPlan) eligible(entry partner, state string, snapshot presence) bool {
	if p.Disabled || !partnerOutgoingReady(entry, snapshot) || p.Host != entry.Host || p.Port != entry.Port ||
		!peerPasswordPattern.MatchString(p.Password) {
		return false
	}
	return state == "Disconnesso" || state == "Collegamento non confermato" || state == "Avvio non riuscito" || state == "Account Mojo non abilitato"
}

func (p reconnectPlan) due(entry partner, state string, snapshot presence, now time.Time) bool {
	return p.eligible(entry, state, snapshot) && !now.Before(p.LastAttempt.Add(reconnectDelay))
}

func reconnectPartners(root string, partners map[string]partner, snapshot presence, now time.Time) {
	if !snapshot.PartnerConnect {
		return
	}
	for call, entry := range partners {
		if _, locked := snapshot.LockedPartners[call]; locked {
			continue
		}
		if !entry.partnerDetails.validate() || call != entry.Call || inboundTestUntil(root, call).After(now) {
			continue
		}
		plan, err := readReconnect(root, call)
		if err != nil || plan.Disabled || plan.Host != entry.Host || plan.Port != entry.Port {
			continue
		}
		if plan.AutoProvision && entry.Provisioning == "" && !partnerOutgoingReady(entry, snapshot) {
			if err := queueOutgoingPartner(root, call); err != nil {
				log.Printf("Queue outgoing partner %s: %v", call, err)
			}
			continue
		}
		if !plan.due(entry, currentConnection(root, call, snapshot), snapshot, now) {
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

func partnerOutgoingReady(entry partner, snapshot presence) bool {
	if snapshot.OutgoingReady != nil {
		return snapshot.OutgoingReady[entry.Call]
	}
	return canConnectPartner(entry.Provisioning)
}

func inboundTestPath(root, call string) string {
	return filepath.Join(root, "local_data", "dxregia-partners", call+".inbound-test")
}

func inboundTestUntil(root, call string) time.Time {
	if len(call) > 12 || !callsign.MatchString(call) {
		return time.Time{}
	}
	data, err := os.ReadFile(inboundTestPath(root, call))
	if err != nil {
		return time.Time{}
	}
	seconds, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(seconds, 0)
}

const inboundTestWindow = 5 * time.Minute
const inboundTestInterval = 30 * time.Minute

func inboundAutoPath(root, call string) string {
	return filepath.Join(root, "local_data", "dxregia-partners", call+".inbound-auto.json")
}

func inboundAutoNext(root, call string) (time.Time, error) {
	var next time.Time
	data, err := os.ReadFile(inboundAutoPath(root, call))
	if err == nil {
		err = json.Unmarshal(data, &next)
	}
	return next, err
}

func hasInboundSession(snapshot presence, call string) bool {
	for _, channel := range snapshot.Channels {
		if channel.Call == call && channel.Kind == "node" && !channel.Outbound {
			return true
		}
	}
	return false
}

func runAutomaticInboundTests(root string, partners map[string]partner, snapshot presence, now time.Time) {
	if !snapshot.PartnerConnect {
		return
	}
	for call, entry := range partners {
		if len(call) > 12 || !callsign.MatchString(call) || entry.Provisioning != "Abilitato in Mojo" || hasInboundSession(snapshot, call) {
			continue
		}
		if _, locked := snapshot.LockedPartners[call]; locked {
			continue
		}
		next, err := inboundAutoNext(root, call)
		if err != nil || next.After(now) || inboundTestUntil(root, call).After(now) {
			continue
		}
		if err := writePrivateJSON(inboundTestPath(root, call), now.Add(inboundTestWindow).Unix()); err != nil {
			log.Printf("Open automatic inbound window for %s: %v", call, err)
			continue
		}
		if err := writePrivateJSON(inboundAutoPath(root, call), now.Add(inboundTestInterval)); err != nil {
			log.Printf("Schedule automatic inbound window for %s: %v", call, err)
		}
	}
}
