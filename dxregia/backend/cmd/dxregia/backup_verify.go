package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func registerBackupVerification(mux *http.ServeMux, auth *sysopAuth, tool string) {
	if auth == nil || tool == "" {
		return
	}
	// Limit simultaneous uploads and decompression by the backup tool.
	busy := make(chan struct{}, 1)
	mux.HandleFunc("POST /api/backups/verify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/zip" {
			http.Error(w, "Seleziona un archivio ZIP.", http.StatusUnsupportedMediaType)
			return
		}
		select {
		case busy <- struct{}{}:
			defer func() { <-busy }()
		default:
			http.Error(w, "Una verifica è già in corso.", http.StatusConflict)
			return
		}
		file, err := os.CreateTemp("", "dxregia-verify-*.zip")
		if err != nil {
			http.Error(w, "Impossibile preparare la verifica.", http.StatusInternalServerError)
			return
		}
		defer os.Remove(file.Name())
		const timeout = 2 * time.Minute
		controller := http.NewResponseController(w)
		_ = controller.SetReadDeadline(time.Now().Add(timeout))
		_ = controller.SetWriteDeadline(time.Now().Add(2 * timeout))
		_, err = io.Copy(file, http.MaxBytesReader(w, r.Body, 512<<20))
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			var limit *http.MaxBytesError
			if errors.As(err, &limit) {
				http.Error(w, "L’archivio supera 512 MiB.", http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "Caricamento non riuscito.", http.StatusBadRequest)
			}
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		err = exec.CommandContext(ctx, "python3", "-I", tool, "verify", file.Name()).Run()
		if err != nil {
			var exit *exec.ExitError
			if ctx.Err() == nil && errors.As(err, &exit) && exit.ExitCode() == 1 {
				http.Error(w, "Archivio non valido o danneggiato.", http.StatusUnprocessableEntity)
			} else {
				http.Error(w, "Verifica non disponibile o scaduta.", http.StatusServiceUnavailable)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"verified": true})
	})
}
