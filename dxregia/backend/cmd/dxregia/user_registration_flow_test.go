package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Percorso completo: richiesta dal modulo, email con il link, conferma, Mojo (simulato) che crea l'utente.
func TestRegistrationFlowEndToEnd(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "local"), 0o755); err != nil {
		t.Fatal(err)
	}
	settings, _ := json.Marshal(registerSettings{Enabled: true, PublicURL: "https://dxregia.example.org"})
	if err := os.WriteFile(registerSettingsPath(root), settings, 0o600); err != nil {
		t.Fatal(err)
	}

	var sentTo, sentMessage string
	previous := registrationMailer
	registrationMailer = func(string) *registrationSender {
		return &registrationSender{from: "dx@example.org", send: func(to, message string) error {
			sentTo, sentMessage = to, message
			return nil
		}}
	}
	defer func() { registrationMailer = previous }()

	mux := http.NewServeMux()
	registerUsers(mux, root, &sysopAuth{})
	server := httptest.NewServer(mux)
	defer server.Close()

	post := func(path string, body any) (*http.Response, []byte) {
		encoded, _ := json.Marshal(body)
		response, err := http.Post(server.URL+path, "application/json", bytes.NewReader(encoded))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		data, _ := io.ReadAll(response.Body)
		return response, data
	}

	// Il pulsante del portale compare solo quando la funzione è attiva e la posta è pronta.
	status, err := http.Get(server.URL + "/api/register/status")
	if err != nil {
		t.Fatal(err)
	}
	var enabled map[string]bool
	_ = json.NewDecoder(status.Body).Decode(&enabled)
	status.Body.Close()
	if !enabled["Enabled"] {
		t.Fatal("registration should be reported as open")
	}

	// Dati non validi.
	for _, bad := range []map[string]string{
		{"Call": "IU3WJQ-7", "Email": "a@example.org", "Password": "Sup3rS3gret0!"},
		{"Call": "IU3WJQ", "Email": "not-an-email", "Password": "Sup3rS3gret0!"},
		{"Call": "IU3WJQ", "Email": "a@example.org", "Password": "short"},
	} {
		if response, _ := post("/api/register", bad); response.StatusCode != http.StatusBadRequest {
			t.Fatalf("bad input %v → %d", bad, response.StatusCode)
		}
	}

	// Richiesta valida: nessun utente finché non si conferma.
	password := "Sup3rS3gret0!"
	if response, _ := post("/api/register", map[string]string{"Call": "iu3wjq", "Email": "ham@example.org", "Password": password, "Lang": "de-AT"}); response.StatusCode != http.StatusAccepted {
		t.Fatalf("register → %d", response.StatusCode)
	}
	if sentTo != "ham@example.org" {
		t.Fatalf("email sent to %q", sentTo)
	}
	if _, err := os.Stat(filepath.Join(registrationQueueDir(root), "IU3WJQ.register")); err == nil {
		t.Fatal("nothing must be queued before the confirmation")
	}
	text := decodedMessage(t, sentMessage)
	link := regexp.MustCompile(`https://dxregia\.example\.org/registrazione/conferma\?token=([0-9a-f]{48})`).FindStringSubmatch(text)
	if link == nil {
		t.Fatalf("confirmation link not found in the email:\n%s", text)
	}
	if !strings.Contains(text, "IU3WJQ") || strings.Contains(text, password) {
		t.Fatal("the email must name the callsign and never contain the password")
	}

	// Mojo simulato: legge la coda, controlla la password e scrive l'esito.
	done := make(chan string, 1)
	go func() {
		queued := filepath.Join(registrationQueueDir(root), "IU3WJQ.register")
		for i := 0; i < 100; i++ {
			if hash, err := os.ReadFile(queued); err == nil {
				result := "Richiesta non valida"
				if bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil {
					result = "Registrato"
				}
				_ = os.WriteFile(filepath.Join(registrationQueueDir(root), "IU3WJQ.result"), []byte(result), 0o600)
				_ = os.Remove(queued)
				done <- result
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
		done <- "timeout"
	}()

	response, data := post("/api/register/confirm", map[string]string{"Token": link[1]})
	if response.StatusCode != http.StatusOK || !strings.Contains(string(data), `"Result":"Registrato"`) {
		t.Fatalf("confirm → %d %s", response.StatusCode, data)
	}
	if result := <-done; result != "Registrato" {
		t.Fatalf("simulated Mojo saw %q: the queued hash must match the chosen password", result)
	}

	// Il link vale una sola volta, un link inventato non vale, e l'indicativo non si registra due volte.
	if response, _ := post("/api/register/confirm", map[string]string{"Token": link[1]}); response.StatusCode != http.StatusNotFound {
		t.Fatalf("reused token → %d", response.StatusCode)
	}
	if response, _ := post("/api/register/confirm", map[string]string{"Token": strings.Repeat("a", 48)}); response.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown token → %d", response.StatusCode)
	}
	if response, _ := post("/api/register", map[string]string{"Call": "IU3WJQ", "Email": "other@example.org", "Password": password}); response.StatusCode != http.StatusConflict {
		t.Fatalf("second registration → %d", response.StatusCode)
	}
	stored, _ := os.ReadFile(filepath.Join(root, ".dxregia-registrations.json"))
	if strings.Contains(string(stored), password) || strings.Contains(string(stored), link[1]) {
		t.Fatal("neither the password nor the token may be stored in clear")
	}
}

func TestRegistrationDisabledOrWithoutMail(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "local"), 0o755); err != nil {
		t.Fatal(err)
	}
	previous := registrationMailer
	registrationMailer = func(string) *registrationSender { return nil }
	defer func() { registrationMailer = previous }()
	mux := http.NewServeMux()
	registerUsers(mux, root, &sysopAuth{})
	server := httptest.NewServer(mux)
	defer server.Close()
	body := `{"Call":"IU3WJQ","Email":"a@example.org","Password":"Sup3rS3gret0!"}`
	response, err := http.Post(server.URL+"/api/register", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("closed registration → %d", response.StatusCode)
	}
}

// decodedMessage restituisce il testo di tutte le parti (testo e HTML) dell'email, già decodificate.
func decodedMessage(t *testing.T, raw string) string {
	t.Helper()
	message, err := mail.ReadMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	_, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	return decodedTextParts(t, multipart.NewReader(message.Body, params["boundary"]))
}

func decodedTextParts(t *testing.T, reader *multipart.Reader) string {
	t.Helper()
	var out strings.Builder
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		mediaType, params, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(mediaType, "multipart/") {
			out.WriteString(decodedTextParts(t, multipart.NewReader(part, params["boundary"])))
		} else if strings.HasPrefix(mediaType, "text/") {
			if _, err := io.Copy(&out, base64.NewDecoder(base64.StdEncoding, part)); err != nil {
				t.Fatal(err)
			}
			out.WriteByte('\n')
		}
	}
	return out.String()
}
