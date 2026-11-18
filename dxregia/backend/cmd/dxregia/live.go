package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// registerLive espone un flusso di eventi (Server-Sent Events) per la dashboard del sysop: invia un
// segnale appena cambiano la fotografia di Mojo o l'archivio degli spot. Il pannello, al segnale,
// rilegge i dati. Se il flusso cade, il pannello torna a interrogare il server a intervalli.
func registerLive(mux *http.ServeMux, root string, auth *sysopAuth) {
	if auth == nil || root == "" {
		return
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
		if !send("retry: 3000\nevent: tick\ndata: hello\n\n") {
			return
		}
		last := liveSignature(root)
		poll := time.NewTicker(time.Second)
		defer poll.Stop()
		beat := time.NewTicker(15 * time.Second)
		defer beat.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-poll.C:
				if current := liveSignature(root); current != last {
					last = current
					if !send("event: tick\ndata: change\n\n") {
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
	files := []string{filepath.Join(root, "local_data", "dxregia-presence.json")}
	for day := 0; day < 2; day++ {
		date := time.Now().UTC().AddDate(0, 0, -day)
		files = append(files, filepath.Join(root, "local_data", "spots", fmt.Sprintf("%04d", date.Year()), fmt.Sprintf("%03d.dat", date.YearDay())))
	}
	signature := ""
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			signature += fmt.Sprintf("%d:%d;", info.ModTime().UnixNano(), info.Size())
		} else {
			signature += "-;"
		}
	}
	return signature
}
