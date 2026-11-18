package main

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:embed presence.pm
var presenceHook []byte

type presenceChannel struct {
	Call     string
	Kind     string
	State    string
	Since    int64
	Outbound bool
}

type presence struct {
	Telnet         *telnetOptions
	PartnerConnect bool
	SysopPassword  bool
	At             int64
	Channels       []presenceChannel
}

func registerPresence(mux *http.ServeMux, root string, auth *sysopAuth) {
	if auth == nil {
		return
	}
	path := filepath.Join(root, "local_data", "dxregia-presence.json")
	mux.HandleFunc("GET /api/presence", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(readPresence(path))
	})
}

func readPresence(path string) presence {
	var snapshot presence
	file, err := os.Open(path)
	if err != nil {
		return snapshot
	}
	defer file.Close()
	if err := json.NewDecoder(io.LimitReader(file, 4<<20)).Decode(&snapshot); err != nil {
		return presence{}
	}
	age := time.Since(time.Unix(snapshot.At, 0))
	if age < 0 || age > 15*time.Second {
		return presence{}
	}
	return snapshot
}

func installPresenceHook(local string) error {
	path := filepath.Join(local, "Local.pm")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if os.IsExist(err) {
		current, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(current)
		digest := hex.EncodeToString(sum[:])
		if digest != "539ceccae21bcae0947532e152e99f180b0754aa34faece57c74f0b4a0439e61" &&
			digest != "7205df58566bee8ccce1b013b4646642d268ff16927fe5b691b371e70464a239" &&
			digest != "54df299595d1d8f7bd08a8f2daff868cdc5d1d918802079171d3dc80b2ae4e72" &&
			digest != "6fdf28c9035182dd86ade8fd2c357d7e64fb0079fb8047786084699ef5f5f13f" &&
			digest != "7174f31bd6f577fe03be6bc02f0067ac3887c56d293b12a6785f2d96b98d1742" &&
			digest != "56c2dc038a1b45d9629a69ff0a0fde08cbe6bb334cde185491c6752235fea68e" &&
			digest != "8c0f94bc34d71bc10422aca764a3b21223882e8d5b98dfee0036f45472017ec6" &&
			digest != "ac480a075906701840e342e0e712aa6c0f064d5b65f202b48d6a09a25f2cccd4" &&
			digest != "f655c5fd185d97f83a9d72cc08b62983427b5df718e3352416bc5ca7eb6e603f" &&
			digest != "8b09a687d3058eaeb2292096bf3afbbd024991e7afda5d6b98ad9f9a0286b79c" &&
			digest != "b86cfd5dbaab32dc023192d342c2babf4e3389fbbea4bd0dbddcf5bfccb31327" &&
			digest != "a234b8ce766c47ee29ee6cb9fa1a41f3acae025bcb520fe783f5f43bc613a376" {
			return nil
		}
		return upgradePresenceHook(local, path)
	}
	if err != nil {
		return err
	}
	_, err = f.Write(presenceHook)
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(path)
	}
	return err
}

func upgradePresenceHook(local, path string) error {
	tmp, err := os.CreateTemp(local, ".dxregia-hook-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(presenceHook)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
