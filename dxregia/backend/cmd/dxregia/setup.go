package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type nodeConfig struct {
	Node    string `json:"node"`
	Sysop   string `json:"sysop"`
	Locator string `json:"locator"`
	Port    int    `json:"port"`
}

var callsign = regexp.MustCompile(`^[A-Z0-9]+[0-9][A-Z0-9]+(-([1-9]|1[0-5]))?$`)
var locator = regexp.MustCompile(`^[A-R]{2}[0-9]{2}([A-X]{2})?$`)

func (c *nodeConfig) validate() error {
	c.Node = strings.ToUpper(strings.TrimSpace(c.Node))
	c.Sysop = strings.ToUpper(strings.TrimSpace(c.Sysop))
	c.Locator = strings.ToUpper(strings.TrimSpace(c.Locator))
	if len(c.Node) > 12 || len(c.Sysop) > 12 || !callsign.MatchString(c.Node) || !callsign.MatchString(c.Sysop) || c.Node == c.Sysop {
		return errors.New("Inserisci due nominativi distinti e validi per nodo e sysop.")
	}
	if !locator.MatchString(c.Locator) {
		return errors.New("Il locator deve avere il formato JN55 o JN55AA.")
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("La porta Telnet deve essere compresa tra 1 e 65535.")
	}
	return nil
}

func registerSetup(service context.Context, mux *http.ServeMux, root, perl string, internalPort, suggestedPort int, telnetBind string, auth *sysopAuth, setTarget func(int)) error {
	if root == "" {
		return nil
	}
	if internalPort < 1 || internalPort > 65535 {
		return errors.New("invalid internal port")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	local := filepath.Join(root, "local")
	var template []byte
	var saved *nodeConfig
	if _, err := os.Lstat(local); err == nil {
		state, err := os.ReadFile(filepath.Join(local, "dxregia-node.json"))
		if os.IsNotExist(err) {
			return nil // Existing configuration belongs to the sysop.
		}
		if err != nil {
			return err
		}
		saved = new(nodeConfig)
		if err := json.Unmarshal(state, saved); err != nil {
			return fmt.Errorf("read saved node configuration: %w", err)
		}
		if err := saved.validate(); err != nil {
			return fmt.Errorf("invalid saved node configuration: %w", err)
		}
		setTarget(saved.Port)
		if perl == "" {
			return nil
		}
		if err := installPresenceHook(local); err != nil {
			return err
		}
		if err := startMojo(service, perl, root, *saved); err == nil {
			return nil
		} else {
			log.Printf("Mojo restart: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	} else {
		template, err = os.ReadFile(filepath.Join(root, "perl", "DXVars.pm.issue"))
		if err != nil {
			return err
		}
	}
	var lock sync.Mutex
	done := false
	mux.HandleFunc("GET /api/setup", func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop e cambio password richiesti.", http.StatusUnauthorized)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{"available": !done, "suggestedPort": suggestedPort}
		if !done && saved != nil {
			response["saved"] = *saved
		}
		json.NewEncoder(w).Encode(response)
	})
	mux.HandleFunc("POST /api/setup", func(w http.ResponseWriter, r *http.Request) {
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop e cambio password richiesti.", http.StatusUnauthorized)
			return
		}
		lock.Lock()
		defer lock.Unlock()
		if done {
			http.Error(w, "Configurazione già salvata.", http.StatusConflict)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Richiesta non valida.", http.StatusUnsupportedMediaType)
			return
		}
		var config nodeConfig
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&config) != nil || decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "Configurazione non valida.", http.StatusBadRequest)
			return
		}
		if err := config.validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if saved == nil && config.Port == internalPort {
			http.Error(w, "Scegli una porta Telnet diversa dalla porta interna di Mojo.", http.StatusBadRequest)
			return
		}
		if saved != nil && *saved != config {
			http.Error(w, "Per riprovare usa i dati già salvati.", http.StatusConflict)
			return
		}
		if saved == nil {
			if perl != "" {
				if err := checkPortAvailable(config.Port); err != nil {
					http.Error(w, fmt.Sprintf("La porta Telnet %d è occupata o non disponibile. Scegli un'altra porta.", config.Port), http.StatusConflict)
					return
				}
			}
			if err := writeMojoConfig(root, local, string(template), config, internalPort, telnetBind); err != nil {
				log.Printf("Mojo setup: %v", err)
				http.Error(w, "Impossibile salvare la configurazione Mojo.", http.StatusInternalServerError)
				return
			}
			saved = &config
			setTarget(config.Port)
		}
		if perl != "" {
			if err := startMojo(service, perl, root, config); err != nil {
				log.Printf("Mojo startup: %v", err)
				http.Error(w, "Configurazione salvata. Avvio di Mojo non riuscito. Consulta il registro di avvio e riprova.", http.StatusServiceUnavailable)
				return
			}
		}
		done = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"started": perl != "", "port": config.Port})
	})
	return nil
}

func writeMojoConfig(root, local, template string, config nodeConfig, internalPort int, telnetBind string) error {
	const ending = "\n1;"
	template = strings.ReplaceAll(template, "\r\n", "\n")
	end := strings.LastIndex(template, ending)
	if end < 0 || strings.TrimSpace(template[end:]) != "1;" {
		return errors.New("unexpected DXVars template format")
	}
	stage, err := os.MkdirTemp(root, ".dxregia-setup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	values := fmt.Sprintf("\n$mycall = '%s';\n$myalias = '%s';\n$mylocator = '%s';\n$myname = '%s';\n$myqth = '%s';\n$myemail = '';\n$clusterport = %d;\n", config.Node, config.Sysop, config.Locator, config.Sysop, config.Locator, internalPort)
	if err := os.WriteFile(filepath.Join(stage, "DXVars.pm"), []byte(template[:end]+values+template[end:]), 0600); err != nil {
		return err
	}
	listeners := fmt.Sprintf("package main;\nour @listen = ([ '%s', %d ]);\n1;\n", telnetBind, config.Port)
	if err := os.WriteFile(filepath.Join(stage, "Listeners.pm"), []byte(listeners), 0600); err != nil {
		return err
	}
	if err := installPresenceHook(stage); err != nil {
		return err
	}
	state, err := json.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stage, "dxregia-node.json"), state, 0600); err != nil {
		return err
	}
	if _, err := os.Lstat(local); err == nil || !os.IsNotExist(err) {
		return errors.New("Mojo local directory already exists")
	}
	return os.Rename(stage, local)
}
