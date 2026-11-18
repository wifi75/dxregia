package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPortalHeartbeat(t *testing.T) {
	mux := http.NewServeMux()
	traffic, err := registerTraffic(mux, t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	post := func(body string, want int) {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/api/visits/heartbeat", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("heartbeat status = %d, want %d", rec.Code, want)
		}
	}
	const token = "0123456789abcdef0123456789abcdef"
	post(`"invalid"`, http.StatusBadRequest)
	post(`"`+token+`"`, http.StatusNoContent)
	post(`"`+token+`"`, http.StatusNoContent)
	if len(traffic.active) != 1 || traffic.data.Views != 0 {
		t.Fatal("repeated heartbeat should track one tab without adding page views")
	}
	traffic.sweep(time.Now().Add(2*time.Minute), true)
	if len(traffic.active) != 0 {
		t.Fatal("inactive tabs should expire")
	}
}
