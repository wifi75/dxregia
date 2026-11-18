package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var totpEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

type secondFactor struct {
	Key          string
	LastStep     int64
	Recovery     []string
	ResetPending bool
}

func newFactorKey() (string, error) {
	key := make([]byte, 20)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return totpEncoding.EncodeToString(key), nil
}

func factorCode(key []byte, step int64) string {
	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(step))
	mac := hmac.New(sha1.New, key)
	mac.Write(counter[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 15
	number := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", number%1000000)
}

func (f secondFactor) verify(code string, now time.Time) (secondFactor, bool) {
	key, err := totpEncoding.DecodeString(f.Key)
	if err != nil || len(key) != 20 || len(code) != 6 {
		return f, false
	}
	step := now.Unix() / 30
	for _, candidate := range []int64{step, step - 1, step + 1} {
		if candidate > f.LastStep && hmac.Equal([]byte(factorCode(key, candidate)), []byte(code)) {
			f.LastStep = candidate
			return f, true
		}
	}
	return f, false
}

func saveFactor(path string, factor secondFactor) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".dxregia-factor-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	err = json.NewEncoder(tmp).Encode(factor)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func recoveryHash(code string) string {
	digest := sha256.Sum256([]byte(code))
	return hex.EncodeToString(digest[:])
}

func createRecoveryCodes(factor *secondFactor) ([]string, error) {
	codes := make([]string, 8)
	hashes := make([]string, len(codes))
	for i := range codes {
		code, err := randomToken()
		if err != nil {
			return nil, err
		}
		codes[i], hashes[i] = code, recoveryHash(code)
	}
	factor.Recovery = hashes
	factor.ResetPending = false
	return codes, nil
}

func (f secondFactor) allowsRecovery(code string) bool {
	if f.Key == "" || len(code) != 48 {
		return false
	}
	hash := recoveryHash(code)
	for _, expected := range f.Recovery {
		if hmac.Equal([]byte(hash), []byte(expected)) {
			return true
		}
	}
	return false
}
