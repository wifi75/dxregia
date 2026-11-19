package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"
)

type storedWebSession struct {
	TokenHash, Credentials string
	Expires                time.Time
}

func sessionHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func (a *sysopAuth) webSessionPath() string {
	return filepath.Join(a.root, ".dxregia-web-session.json")
}

func (a *sysopAuth) sessionCredentials() string {
	return sessionHash(string(a.hash) + ":" + a.factor.Key)
}

func (a *sysopAuth) restoreWebSession() {
	if a.temporary || a.factor.Key == "" || a.factor.ResetPending {
		return
	}
	data, err := os.ReadFile(a.webSessionPath())
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Restore sysop session: %v", err)
		}
		return
	}
	var stored storedWebSession
	if json.Unmarshal(data, &stored) != nil || !time.Now().Before(stored.Expires) || stored.Credentials != a.sessionCredentials() {
		return
	}
	digest, err := hex.DecodeString(stored.TokenHash)
	if err != nil || len(digest) != sha256.Size {
		return
	}
	a.session, a.expires, a.factorOK = stored.TokenHash, stored.Expires, true
}

func (a *sysopAuth) persistWebSession(expires time.Time) error {
	tmp, err := os.CreateTemp(a.root, ".dxregia-session-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	stored := storedWebSession{TokenHash: a.session, Credentials: a.sessionCredentials(), Expires: expires}
	if err := errors.Join(json.NewEncoder(tmp).Encode(stored), tmp.Close()); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), a.webSessionPath())
}

func (a *sysopAuth) removeWebSession() error {
	if err := os.Remove(a.webSessionPath()); !os.IsNotExist(err) {
		return err
	}
	return nil
}
