package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPartnerMailHistoryRedactsCredentialsAndRequiresSysop(t *testing.T) {
	root := t.TempDir()
	job := partnerMail{Call: "Q3QQQ-2", Recipient: "partner@example.org", Password: "private-link-password", LocalNode: "IU3WJQ-7", Host: "node.example.org", Port: 8000}
	content := mailInstructions("en", job, "Nick")
	if err := archivePartnerDelivery(root, job.Call, "instructions", job.Recipient, job.Password, content, func(to string, sent mailContent) error {
		if to != job.Recipient || sent != content {
			t.Fatal("archiving changed the outgoing message")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("SMTP error with sensitive detail")
	if err := archivePartnerDelivery(root, job.Call, "pending", job.Recipient, "", mailLinkPending("en", job, "Nick"), func(string, mailContent) error {
		return failure
	}); !errors.Is(err, failure) {
		t.Fatalf("expected send failure, got %v", err)
	}
	dir := partnerMailHistoryDir(root)
	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 2 {
		t.Fatalf("expected two archived messages, got %d, err %v", len(files), err)
	}
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil || strings.Contains(string(data), job.Password) || strings.Contains(string(data), failure.Error()) || strings.Contains(string(data), `"HTML"`) {
			t.Fatal("unsafe content in history file")
		}
	}
	mux := http.NewServeMux()
	auth := &sysopAuth{session: sessionHash("session"), expires: time.Now().Add(time.Hour)}
	registerPartnerMailHistory(mux, root, auth)
	req := httptest.NewRequest(http.MethodGet, "/api/partners/mail-history?q=Q3QQQ-2", nil)
	req.AddCookie(&http.Cookie{Name: sysopCookie, Value: "session"})
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusUnauthorized {
		t.Fatal("history exposed before second factor verification")
	}
	auth.factorOK = true
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	var page struct{ Items []partnerMailRecord }
	if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil || len(page.Items) != 2 {
		t.Fatalf("unexpected history response: %v, status %d", err, response.Code)
	}
	if page.Items[0].Status != "failed" || page.Items[1].Status != "accepted" || strings.Count(page.Items[1].Text, "••••••••") < 2 {
		t.Fatal("missing delivery statuses or masked credentials")
	}
}
