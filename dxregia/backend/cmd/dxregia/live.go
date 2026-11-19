package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// registerLive espone un flusso di eventi (Server-Sent Events) per la dashboard del sysop: invia un
// segnale appena cambiano la fotografia di Mojo o l'archivio degli spot. Il pannello, al segnale,
// rilegge i dati. Se il flusso cade, il pannello torna a interrogare il server a intervalli.
func registerLive(mux *http.ServeMux, root string, auth *sysopAuth) {
	if auth == nil || root == "" {
		return
	}
	var mu sync.Mutex
	var previous string
	var sequence uint64
	epoch := time.Now().UnixNano()
	currentID := func() string {
		mu.Lock()
		defer mu.Unlock()
		signature := liveSignature(root)
		if signature != previous {
			previous = signature
			sequence++
		}
		return fmt.Sprintf("%d-%d", epoch, sequence)
	}
	mux.HandleFunc("GET /api/live", func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Flusso non supportato.", http.StatusInternalServerError)
			return
		}
		controller := http.NewResponseController(w)
		header := w.Header()
		header.Set("Content-Type", "text/event-stream")
		header.Set("Cache-Control", "no-store")
		header.Set("X-Accel-Buffering", "no")
		send := func(format string, args ...any) bool {
			// Il server limita la durata delle risposte: ogni scrittura rinnova la scadenza.
			_ = controller.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if _, err := fmt.Fprintf(w, format, args...); err != nil {
				return false
			}
			flusher.Flush()
			return true
		}
		last := currentID()
		if r.Header.Get("Last-Event-ID") == last {
			if !send("retry: 3000\n: current\n\n") {
				return
			}
		} else if !send("retry: 3000\nid: %s\nevent: tick\ndata: hello\n\n", last) {
			return
		}
		poll := time.NewTicker(time.Second)
		defer poll.Stop()
		beat := time.NewTicker(15 * time.Second)
		defer beat.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-poll.C:
				if !auth.allowed(r) {
					return
				}
				if current := currentID(); current != last {
					last = current
					if !send("id: %s\nevent: tick\ndata: change\n\n", current) {
						return
					}
				}
			case <-beat.C:
				if !send(": keepalive\n\n") {
					return
				}
			}
		}
	})
}

// liveSignature riassume lo stato osservabile: cambia quando Mojo scrive una nuova fotografia
// o quando arriva uno spot nei file di oggi e di ieri.
func liveSignature(root string) string {
	paths := spotArchivePaths(root, time.Now())
	return fileSignature(filepath.Join(root, "local_data", "dxregia-presence.json"), paths[0], paths[1])
}
