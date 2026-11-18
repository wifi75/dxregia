package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var peerPasswordPattern = regexp.MustCompile("^[A-Za-z0-9]{0,72}$")

func queueConnection(root string, entry partner, password string) error {
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	target := filepath.Join(dir, entry.Call+".connect")
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		return errors.New("connection already queued or unavailable")
	}
	tmp, err := os.CreateTemp(dir, ".outgoing-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, err = fmt.Fprintf(tmp, "%s\n%d\n%s", entry.Host, entry.Port, password)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(dir, entry.Call+".connect-result")); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(tmp.Name(), target)
}

func currentConnection(root, call string, snapshot presence) string {
	if !snapshot.PartnerConnect {
		return "Stato Mojo non disponibile"
	}
	for _, channel := range snapshot.Channels {
		if channel.Call == call && channel.Kind == "node" {
			if channel.State == "normal" {
				return "Attivo"
			}
			return "Collegamento in corso"
		}
	}
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	if _, err := os.Stat(filepath.Join(dir, call+".connect")); err == nil {
		return "In coda"
	}
	path := filepath.Join(dir, call+".connect-result")
	data, err := os.ReadFile(path)
	if err != nil {
		return "Disconnesso"
	}
	if string(data) == "Tentativo avviato" {
		if info, err := os.Stat(path); err == nil && time.Since(info.ModTime()) < 25*time.Second {
			return "Collegamento in corso"
		}
		return "Collegamento non confermato"
	}
	if string(data) == "Sessione gia presente" {
		return "Disconnesso"
	}
	return string(data)
}
