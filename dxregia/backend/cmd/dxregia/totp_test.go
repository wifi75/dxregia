package main

import (
	"testing"
	"time"
)

func TestFactorDistinguishesReusedAndInvalidCodes(t *testing.T) {
	now := time.Unix(1800000000, 0)
	step := now.Unix() / 30
	key := []byte("12345678901234567890")
	factor := secondFactor{Key: totpEncoding.EncodeToString(key)}
	code := factorCode(key, step)
	verified, accepted, reused := factor.verify(code[:3]+" "+code[3:], now)
	if !accepted || reused || verified.LastStep != step {
		t.Fatal("the current code with a copied separator must be accepted")
	}
	if _, accepted, reused := verified.verify(code, now); accepted || !reused {
		t.Fatal("a replay must be identified and rejected")
	}
	if _, accepted, reused := verified.verify("invalid", now); accepted || reused {
		t.Fatal("an invalid code must not be reported as used")
	}
	if _, accepted, reused := verified.verify(factorCode(key, step+1), now.Add(30*time.Second)); !accepted || reused {
		t.Fatal("a fresh code must still work after a replay")
	}
}
