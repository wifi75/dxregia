package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func queuePartner(root, call string) (string, error) {
	password, err := randomToken()
	if err != nil {
		return "", err
	}
	if err := queuePartnerProvision(root, call, password); err != nil {
		return "", err
	}
	return password, nil
}

func queueOutgoingPartner(root, call string) error {
	token, err := randomToken()
	if err != nil {
		return err
	}
	return queuePartnerProvision(root, call, "outgoing-only:"+token)
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
