package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestSysopLoginNeedsUserAndPassword(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "local"), 0700); err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("right-password-1"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	auth := &sysopAuth{root: root, hash: hash}
	if auth.loginName() != "sysop" {
		t.Fatalf("default login name: %q", auth.loginName())
	}
	if err := os.WriteFile(filepath.Join(root, "local", "dxregia-node.json"), []byte(`{"node":"IU3WJQ-7","sysop":"IU3WJQ"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if auth.loginName() != "IU3WJQ" {
		t.Fatalf("login name from the node config: %q", auth.loginName())
	}
	post := func(body string) int {
		auth.nextLogin = auth.nextLogin.AddDate(-1, 0, 0)
		request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		auth.route("login")(response, request)
		if response.Code == http.StatusOK {
			if strings.Contains(response.Body.String(), "mustVerify") {
				t.Fatal("2FA was required with no second factor configured")
			}
			request.AddCookie(response.Result().Cookies()[0])
			if !auth.allowed(request) {
				t.Fatal("password-only login did not grant access with 2FA disabled")
			}
		}
		return response.Code
	}
	for name, body := range map[string]string{
		"wrong user":     `{"User":"nobody","Password":"right-password-1"}`,
		"wrong password": `{"User":"IU3WJQ","Password":"wrong-password-1"}`,
		"password only":  `"right-password-1"`,
		"missing user":   `{"Password":"right-password-1"}`,
	} {
		if code := post(body); code == http.StatusOK {
			t.Errorf("%s: login accepted", name)
		}
	}
	if code := post(`{"User":"iu3wjq","Password":"right-password-1"}`); code != http.StatusOK {
		t.Fatalf("valid login (case-insensitive user) rejected: %d", code)
	}
}
