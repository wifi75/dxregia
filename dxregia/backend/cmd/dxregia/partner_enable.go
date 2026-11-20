package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var inboundPasswordPattern = regexp.MustCompile(`^[0-9a-f]{48}$`)

const lockedPartnerStatus = "Nodo bloccato in Mojo"

func queuePartner(root, call string) (string, error) {
	return queuePartnerCredential(root, call, false, false)
}

func queuePartnerCredential(root, call string, resetCCInbound, switchFromIP bool) (string, error) {
	password, err := randomToken()
	if err != nil {
		return "", err
	}
	request := password
	if resetCCInbound {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		request = "reset-cc-inbound:" + string(hash)
		if switchFromIP {
			request = "reset-ip-inbound:" + string(hash)
		}
	}
	if err := queuePartnerProvision(root, call, request); err != nil {
		return "", err
	}
	// Conserva la password in ingresso in un file privato, per poterla reinviare al referente.
	if err := os.WriteFile(inboundPath(root, call), []byte(password), 0600); err != nil {
		log.Printf("Store inbound credential: %v", err)
	}
	return password, nil
}

func inboundPath(root, call string) string {
	return filepath.Join(root, "local_data", "dxregia-partners", call+".inbound")
}

func readInboundPassword(root, call string) string {
	data, err := os.ReadFile(inboundPath(root, call))
	if err != nil || !inboundPasswordPattern.Match(data) {
		return ""
	}
	return string(data)
}

func queueOutgoingPartner(root, call string) error {
	return queueOutgoingPartnerProvision(root, call, "outgoing-only:")
}

func queueOutgoingPartnerProvision(root, call, prefix string) error {
	token, err := randomToken()
	if err != nil {
		return err
	}
	return queuePartnerProvision(root, call, prefix+token)
}

func queuePartnerProvision(root, call, value string) error {
	if !callsign.MatchString(call) || len(call) > 12 {
		return errors.New("invalid partner callsign")
	}
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	target := filepath.Join(dir, call+".enable")
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		return errors.New("partner request already queued or unavailable")
	}
	file, err := os.CreateTemp(dir, ".enable-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString(value)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(dir, call+".result")); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(file.Name(), target); err != nil {
		return err
	}
	return nil
}

func readProvisioningResults(root, path string, partners map[string]partner) {
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	changed := false
	for call, entry := range partners {
		if !callsign.MatchString(call) || len(call) > 12 {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, call+".result"))
		var status string
		if err == nil {
			status = string(data)
		} else if _, err := os.Stat(filepath.Join(dir, call+".enable")); err == nil {
			status = "In attesa di Mojo"
		}
		if status != "" && status != entry.Provisioning {
			entry.Provisioning = status
			partners[call] = entry
			changed = true
		}
	}
	if changed {
		if err := savePartners(path, partners); err != nil {
			log.Printf("Save Mojo partner status: %v", err)
		}
	}
}

// removePartnerFiles cancella i file del nodo e chiede a Mojo di scollegarlo e di cancellarne l'account.
func removePartnerFiles(root, call string) error {
	if !callsign.MatchString(call) || len(call) > 12 {
		return errors.New("invalid partner callsign")
	}
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	files, err := filepath.Glob(filepath.Join(dir, call+".*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, call+".remove"), []byte("remove"), 0600)
}
