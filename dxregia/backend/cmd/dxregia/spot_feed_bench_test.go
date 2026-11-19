package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type spotBenchmarkWriter struct {
	headers http.Header
	status  int
}

func (w *spotBenchmarkWriter) Header() http.Header    { return w.headers }
func (w *spotBenchmarkWriter) WriteHeader(status int) { w.status = status }
func (w *spotBenchmarkWriter) Write(data []byte) (int, error) {
	w.status = http.StatusOK
	return len(data), nil
}

func BenchmarkSpotFeed(b *testing.B) {
	root := b.TempDir()
	now := time.Now().UTC()
	archive := spotArchivePaths(root, now)[0]
	if err := os.MkdirAll(filepath.Dir(archive), 0700); err != nil {
		b.Fatal(err)
	}
	var rows strings.Builder
	for i := range 100 {
		fmt.Fprintf(&rows, "14074^JA1ABC^%d^FT8 PM95AB^IU3WJQ^95^85^NODE^45^25\n", now.Unix()+int64(i))
	}
	if err := os.WriteFile(archive, []byte(rows.String()), 0600); err != nil {
		b.Fatal(err)
	}
	mux := http.NewServeMux()
	registerSpots(mux, root)
	var response = spotBenchmarkWriter{headers: make(http.Header)}
	initial := httptest.NewRequest(http.MethodGet, "/api/spots", nil)
	mux.ServeHTTP(&response, initial)
	if response.status != http.StatusOK {
		b.Fatalf("spot feed returned %d", response.status)
	}
	tag := response.headers.Get("ETag")

	for _, cached := range []bool{false, true} {
		name, expected := "FullResponse", http.StatusOK
		if cached {
			name, expected = "NotModified", http.StatusNotModified
		}
		b.Run(name, func(b *testing.B) {
			req := httptest.NewRequest(http.MethodGet, "/api/spots", nil)
			if cached {
				req.Header.Set("If-None-Match", tag)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				clear(response.headers)
				response.status = 0
				mux.ServeHTTP(&response, req)
				if response.status != expected {
					b.Fatalf("got %d, want %d", response.status, expected)
				}
			}
		})
	}
}
