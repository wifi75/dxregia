package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSecurityHeadersAndHSTSBehindProxy(t *testing.T) {
	handler := secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	request := httptest.NewRequest("GET", "/spots", nil)
	request.RemoteAddr = "127.0.0.1:5000"
	request.Header.Set("X-Forwarded-Proto", "https")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	for _, name := range []string{"Content-Security-Policy", "X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy", "Permissions-Policy", "Strict-Transport-Security"} {
		if recorder.Header().Get(name) == "" {
			t.Errorf("missing header %s", name)
		}
	}
	if !strings.Contains(recorder.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Error("CSP must forbid framing")
	}
	// Un visitatore diretto che dichiara HTTPS non deve ottenere HSTS: l'intestazione si può falsificare.
	request.RemoteAddr = "203.0.113.9:5000"
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Header().Get("Strict-Transport-Security") != "" {
		t.Error("forwarded protocol must be ignored from untrusted clients")
	}
}

func TestClientIPOnlyTrustsLocalProxies(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "10.0.0.2:1"
	request.Header.Set("CF-Connecting-IP", "198.51.100.7")
	if got := clientIP(request); got != "198.51.100.7" {
		t.Fatalf("behind a proxy: %s", got)
	}
	request.RemoteAddr = "203.0.113.9:1"
	if got := clientIP(request); got != "203.0.113.9" {
		t.Fatalf("spoofed header must be ignored: %s", got)
	}
}

func TestCrossSiteWritesAreRefused(t *testing.T) {
	cases := []struct {
		method, site, origin string
		want                 bool
	}{
		{"POST", "same-origin", "", false}, {"POST", "cross-site", "", true}, {"POST", "", "https://evil.example", true},
		{"POST", "", "http://dx.test", false}, {"GET", "cross-site", "", false}, {"POST", "none", "", false},
	}
	for _, c := range cases {
		request := httptest.NewRequest(c.method, "http://dx.test/api/contact", nil)
		if c.site != "" {
			request.Header.Set("Sec-Fetch-Site", c.site)
		}
		if c.origin != "" {
			request.Header.Set("Origin", c.origin)
		}
		if got := crossSite(request); got != c.want {
			t.Errorf("%+v: got %v", c, got)
		}
	}
}

func TestRateLimitRefillsAndIsPerAddress(t *testing.T) {
	stats := &securityStats{buckets: map[string]*bucket{}}
	limit := rule{perMinute: 60, burst: 2}
	now := time.Now()
	if !stats.allow("form", "1.1.1.1", limit, now) || !stats.allow("form", "1.1.1.1", limit, now) {
		t.Fatal("burst must be allowed")
	}
	if stats.allow("form", "1.1.1.1", limit, now) {
		t.Fatal("third request must be refused")
	}
	if !stats.allow("form", "2.2.2.2", limit, now) {
		t.Fatal("another address has its own bucket")
	}
	if !stats.allow("form", "1.1.1.1", limit, now.Add(2*time.Second)) {
		t.Fatal("tokens must refill over time")
	}
}

func TestLoginDelayGrowsAndStops(t *testing.T) {
	if loginDelay(1) != time.Second || loginDelay(3) != 4*time.Second || loginDelay(50) != 30*time.Second {
		t.Fatalf("delays: %v %v %v", loginDelay(1), loginDelay(3), loginDelay(50))
	}
}
