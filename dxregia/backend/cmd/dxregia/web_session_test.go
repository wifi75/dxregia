package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestVerifiedWebSessionSurvivesRestart(t *testing.T) {
	root := t.TempDir()
	hash, err := bcrypt.GenerateFromPassword([]byte("example-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".dxregia-sysop"), hash, 0600); err != nil {
		t.Fatal(err)
	}
	key, err := newFactorKey()
	if err != nil {
		t.Fatal(err)
	}
	if err := saveFactor(filepath.Join(root, ".dxregia-totp"), secondFactor{Key: key}); err != nil {
		t.Fatal(err)
	}
	restart := func() *sysopAuth {
		t.Helper()
		auth, err := registerAuth(http.NewServeMux(), root, "")
		if err != nil {
			t.Fatal(err)
		}
		return auth
	}
	post := func(auth *sysopAuth, route, body string, cookie *http.Cookie) *http.Cookie {
		t.Helper()
		request := httptest.NewRequest(http.MethodPost, "/api/auth/"+route, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			request.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		auth.route(route)(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s gave %d: %s", route, response.Code, response.Body)
		}
		return response.Result().Cookies()[0]
	}
	requestWith := func(cookie *http.Cookie) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/session-duration", nil)
		r.AddCookie(cookie)
		return r
	}

	auth := restart()
	pending := post(auth, "login", `{"User":"sysop","Password":"example-password"}`, nil)
	if restart().signedIn(requestWith(pending)) {
		t.Fatal("partially authenticated session survived restart")
	}
	keyBytes, _ := totpEncoding.DecodeString(key)
	code, _ := json.Marshal(factorCode(keyBytes, time.Now().Unix()/30))
	verified := post(auth, "totp", string(code), pending)
	data, err := os.ReadFile(auth.webSessionPath())
	if err != nil || strings.Contains(string(data), verified.Value) {
		t.Fatal("persistent session missing or stores the raw cookie")
	}
	restored := restart()
	if !restored.allowed(requestWith(verified)) || !restored.expires.Equal(auth.expires) {
		t.Fatal("verified session or its original deadline was lost")
	}
	if err := restored.persistWebSession(time.Now().Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	if restart().signedIn(requestWith(verified)) {
		t.Fatal("expired session restored")
	}
	if err := restored.persistWebSession(restored.expires); err != nil {
		t.Fatal(err)
	}
	post(restored, "logout", "{}", verified)
	if restart().signedIn(requestWith(verified)) {
		t.Fatal("revoked session restored")
	}
}

func TestTrustedBrowserSkipsSecondFactor(t *testing.T) {
	root := t.TempDir()
	hash, _ := bcrypt.GenerateFromPassword([]byte("example-password"), bcrypt.MinCost)
	os.WriteFile(filepath.Join(root, ".dxregia-sysop"), hash, 0600)
	key, _ := newFactorKey()
	saveFactor(filepath.Join(root, ".dxregia-totp"), secondFactor{Key: key})
	newAuth := func() *sysopAuth {
		auth, err := registerAuth(http.NewServeMux(), root, "")
		if err != nil {
			t.Fatal(err)
		}
		return auth
	}
	post := func(auth *sysopAuth, route, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/api/auth/"+route, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			request.AddCookie(c)
		}
		response := httptest.NewRecorder()
		auth.route(route)(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s gave %d: %s", route, response.Code, response.Body)
		}
		return response
	}
	named := func(response *httptest.ResponseRecorder, name string) *http.Cookie {
		for _, c := range response.Result().Cookies() {
			if c.Name == name {
				return c
			}
		}
		return nil
	}
	login := `{"User":"sysop","Password":"example-password"}`
	auth := newAuth()
	pending := named(post(auth, "login", login), sysopCookie)
	keyBytes, _ := totpEncoding.DecodeString(key)
	code, _ := json.Marshal(factorCode(keyBytes, time.Now().Unix()/30))
	verified := post(auth, "totp", string(code), pending)
	trust := named(verified, trustCookie)
	if trust == nil || trust.MaxAge != int(trustDuration/time.Second) {
		t.Fatal("trusted device cookie missing or not persistent")
	}
	// Un browser senza cookie fidato deve ancora inserire il codice, anche dopo un riavvio.
	again := newAuth()
	post(again, "login", login)
	if again.factorOK {
		t.Fatal("second factor skipped without trust cookie")
	}
	// Con il cookie fidato il codice non è richiesto, anche dopo un riavvio.
	restarted := newAuth()
	response := post(restarted, "login", login, trust)
	if !restarted.factorOK || strings.Contains(response.Body.String(), "mustVerify") {
		t.Fatal("trusted browser still asked for the second factor")
	}
	// Se la chiave di verifica cambia il dispositivo non è più fidato.
	other, _ := newFactorKey()
	saveFactor(filepath.Join(root, ".dxregia-totp"), secondFactor{Key: other})
	changed := newAuth()
	post(changed, "login", login, trust)
	if changed.factorOK {
		t.Fatal("trust survived a change of the second factor")
	}
}
