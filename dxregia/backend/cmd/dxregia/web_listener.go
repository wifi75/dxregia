package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type webListener struct {
	mu       sync.Mutex
	server   *http.Server
	listener net.Listener
	host     string
	port     int
	path     string
}

func (w *webListener) bind(port int) (net.Listener, error) {
	return net.Listen("tcp", net.JoinHostPort(w.host, strconv.Itoa(port)))
}

func (w *webListener) serve(listener net.Listener) {
	log.Printf("DXRegia: http://%s", listener.Addr())
	if err := w.server.Serve(listener); err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("HTTP listener: %v", err)
	}
}

func (w *webListener) register(mux *http.ServeMux, auth *sysopAuth) {
	if auth == nil {
		return
	}
	handler := func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(request) {
			http.Error(response, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		w.mu.Lock()
		defer w.mu.Unlock()
		if request.Method == http.MethodPut {
			if w.path == "" {
				http.Error(response, "La porta web viene gestita dal servizio di installazione.", http.StatusConflict)
				return
			}
			if request.Header.Get("Content-Type") != "application/json" {
				http.Error(response, "Richiesta non valida.", http.StatusUnsupportedMediaType)
				return
			}
			var port int
			decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 64))
			if decoder.Decode(&port) != nil || decoder.Decode(new(any)) != io.EOF || port < 1024 || port > 65535 {
				http.Error(response, "Scegli una porta tra 1024 e 65535.", http.StatusBadRequest)
				return
			}
			if port != w.port {
				listener, err := w.bind(port)
				if err != nil {
					http.Error(response, "Impossibile aprire la porta richiesta. Controlla che sia libera.", http.StatusConflict)
					return
				}
				if err := saveWebPort(w.path, port); err != nil {
					_ = listener.Close()
					log.Printf("Save HTTP port: %v", err)
					http.Error(response, "Impossibile salvare la porta web.", http.StatusInternalServerError)
					return
				}
				old := w.listener
				w.listener, w.port = listener, port
				go w.serve(listener)
				_ = old.Close()
			}
		}
		response.Header().Set("Content-Type", "application/json")
		json.NewEncoder(response).Encode(map[string]any{"port": w.port, "editable": w.path != ""})
	}
	mux.HandleFunc("GET /api/web-settings", handler)
	mux.HandleFunc("PUT /api/web-settings", handler)
}

func loadWebPort(path string, fallback int) (int, error) {
	if path == "" {
		return fallback, nil
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return fallback, nil
	}
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || port < 1024 || port > 65535 {
		return 0, errors.New("invalid saved HTTP port")
	}
	return port, nil
}

func saveWebPort(path string, port int) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".web-port-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, err = io.WriteString(file, strconv.Itoa(port)+"\n")
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
