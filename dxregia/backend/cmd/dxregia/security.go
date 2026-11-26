package main

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Protezione del sito: intestazioni di sicurezza, controllo delle richieste da altri siti, limite di richieste per
// indirizzo e registro (solo in memoria) di ciò che è stato respinto, letto dal pannello Sicurezza.

const contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: https://prop.kc2g.com; font-src 'self' data:; connect-src 'self'; worker-src 'self'; " +
	"manifest-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"

// trustedProxy dice se la richiesta arriva da un proxy locale o di rete privata: solo allora si leggono gli
// indirizzi e il protocollo dichiarati nelle intestazioni, altrimenti chiunque potrebbe falsificarli.
func trustedProxy(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast())
}

// clientIP è l'indirizzo del visitatore, anche dietro un proxy fidato (HAProxy, Cloudflare, nginx). Il secondo valore
// è falso quando la richiesta viene da un proxy che non dichiara l'indirizzo del visitatore: in quel caso l'unico
// indirizzo noto è quello del proxy, uguale per tutti, e non si può usare per limitare le richieste.
func clientIP(r *http.Request) (string, bool) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if trustedProxy(r) {
		for _, name := range []string{"CF-Connecting-IP", "X-Real-IP"} {
			if ip := net.ParseIP(strings.TrimSpace(r.Header.Get(name))); ip != nil {
				return ip.String(), true
			}
		}
		parts := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		if ip := net.ParseIP(strings.TrimSpace(parts[len(parts)-1])); ip != nil {
			return ip.String(), true
		}
		return host, false
	}
	return host, true
}

// secureRequest dice se il visitatore usa HTTPS, direttamente o tramite un proxy fidato che termina TLS.
func secureRequest(r *http.Request) bool {
	return r.TLS != nil || trustedProxy(r) && strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

type bucket struct {
	tokens float64
	at     time.Time
}

type rule struct {
	perMinute float64
	burst     float64
}

type securityEvent struct {
	At   int64
	Kind string
	IP   string
	Path string
}

type securityStats struct {
	mu          sync.Mutex
	buckets     map[string]*bucket
	cleaned     time.Time
	Started     int64
	RateLimited uint64
	CrossSite   uint64
	LoginFails  uint64
	events      []securityEvent
}

var security = &securityStats{Started: time.Now().Unix(), buckets: map[string]*bucket{}}

func (s *securityStats) record(kind, ip, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch kind {
	case "Troppe richieste":
		s.RateLimited++
	case "Richiesta da altro sito":
		s.CrossSite++
	case "Accesso errato":
		s.LoginFails++
	}
	s.events = append(s.events, securityEvent{At: time.Now().Unix(), Kind: kind, IP: ip, Path: path})
	if len(s.events) > 30 {
		s.events = s.events[len(s.events)-30:]
	}
}

// allow consuma un gettone del secchio di (classe, indirizzo); false se è vuoto.
func (s *securityStats) allow(class, ip string, r rule, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if now.Sub(s.cleaned) > 5*time.Minute {
		for key, b := range s.buckets {
			if now.Sub(b.at) > 10*time.Minute {
				delete(s.buckets, key)
			}
		}
		s.cleaned = now
	}
	key := class + "|" + ip
	b := s.buckets[key]
	if b == nil {
		if len(s.buckets) > 50000 { // difesa da chi usa moltissimi indirizzi: si smette di ricordarli
			return true
		}
		b = &bucket{tokens: r.burst, at: now}
		s.buckets[key] = b
	}
	b.tokens = min(r.burst, b.tokens+now.Sub(b.at).Seconds()*r.perMinute/60)
	b.at = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (s *securityStats) snapshot() (stats struct {
	Started, RateLimited, CrossSite, LoginFails uint64
}, events []securityEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stats.Started, stats.RateLimited, stats.CrossSite, stats.LoginFails = uint64(s.Started), s.RateLimited, s.CrossSite, s.LoginFails
	for i := len(s.events) - 1; i >= 0; i-- {
		events = append(events, s.events[i])
	}
	return stats, events
}

// limitFor sceglie il limite in base al tipo di richiesta. Il limite per le pagine statiche non esiste.
func limitFor(r *http.Request) (string, rule, bool) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/") {
		return "", rule{}, false
	}
	switch {
	case strings.HasPrefix(path, "/api/auth/") && r.Method == http.MethodPost:
		return "auth", rule{perMinute: 10, burst: 10}, true
	case r.Method == http.MethodPost && (path == "/api/contact" || strings.HasPrefix(path, "/api/register")):
		return "form", rule{perMinute: 5, burst: 5}, true
	case r.Method == http.MethodPost && strings.HasPrefix(path, "/api/visits"):
		return "visit", rule{perMinute: 60, burst: 20}, true
	case r.Method != http.MethodGet && r.Method != http.MethodHead:
		return "write", rule{perMinute: 120, burst: 60}, true
	}
	return "read", rule{perMinute: 900, burst: 300}, true
}

// crossSite riconosce una richiesta che modifica dati e arriva da un altro sito.
func crossSite(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return false
	}
	switch r.Header.Get("Sec-Fetch-Site") {
	case "same-origin", "none":
		return false
	case "cross-site", "same-site":
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return true
	}
	return !strings.EqualFold(parsed.Host, r.Host) && !strings.EqualFold(parsed.Host, r.Header.Get("X-Forwarded-Host"))
}

func secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=(), interest-cohort=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		if secureRequest(r) {
			h.Set("Strict-Transport-Security", "max-age=31536000")
		}
		ip, known := clientIP(r)
		if crossSite(r) {
			security.record("Richiesta da altro sito", ip, r.URL.Path)
			http.Error(w, "Richiesta non consentita.", http.StatusForbidden)
			return
		}
		if class, limit, ok := limitFor(r); ok && known && !security.allow(class, ip, limit, time.Now()) {
			security.record("Troppe richieste", ip, r.URL.Path)
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Troppe richieste: riprova tra poco.", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loginDelay è l'attesa dopo un errore: cresce con i tentativi sbagliati di fila, fino a mezzo minuto.
func loginDelay(failures int) time.Duration {
	if failures <= 1 {
		return time.Second
	}
	return min(time.Duration(1<<min(failures-1, 5))*time.Second, 30*time.Second)
}

func registerSecurity(mux *http.ServeMux, auth *sysopAuth) {
	if auth == nil {
		return
	}
	mux.HandleFunc("GET /api/security", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		stats, events := security.snapshot()
		auth.mu.Lock()
		twoFactor, hours := auth.factorEnabled(), auth.sessionHours()
		auth.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"Https": secureRequest(r), "TwoFactor": twoFactor, "SessionHours": hours,
			"Started": stats.Started, "RateLimited": stats.RateLimited, "CrossSite": stats.CrossSite, "LoginFails": stats.LoginFails,
			"Events": events,
		})
	})
}

func clientIPText(r *http.Request) string { ip, _ := clientIP(r); return ip }
