package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

type nodeConfig struct {
	Node           string   `json:"node"`
	Sysop          string   `json:"sysop"`
	Locator        string   `json:"locator"`
	Port           int      `json:"port"`
	Longitude      *float64 `json:"longitude,omitempty"`
	Latitude       *float64 `json:"latitude,omitempty"`
	AdvertisedHost string   `json:"advertisedHost,omitempty"`
	Email          string   `json:"email,omitempty"`
	QTH            string   `json:"qth,omitempty"`
	Name           string   `json:"name,omitempty"`
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
	c.Name = strings.TrimSpace(c.Name)
	c.QTH = strings.TrimSpace(c.QTH)
	c.Email = strings.TrimSpace(c.Email)
	c.AdvertisedHost = strings.TrimSpace(c.AdvertisedHost)
	if len(c.Name) > 120 || len(c.QTH) > 120 || strings.ContainsAny(c.Name+c.QTH, "\r\n\x00") {
		return errors.New("Controlla nome e località del sysop.")
	}
	if c.Email != "" {
		address, err := mail.ParseAddress(c.Email)
		if err != nil || address.Address != c.Email || len(c.Email) > 254 {
			return errors.New("Inserisci un indirizzo email valido.")
		}
	}
	if c.AdvertisedHost != "" && (len(c.AdvertisedHost) > 253 || (net.ParseIP(c.AdvertisedHost) == nil && !partnerHost.MatchString(c.AdvertisedHost))) {
		return errors.New("Inserisci un nome host o un indirizzo IP valido, senza protocollo o porta.")
	}
	if (c.Latitude == nil) != (c.Longitude == nil) ||
		(c.Latitude != nil && (*c.Latitude < -90 || *c.Latitude > 90 || *c.Longitude < -180 || *c.Longitude > 180)) {
		return errors.New("Inserisci entrambe le coordinate: latitudine da -90 a 90 e longitudine da -180 a 180.")
	}
	return nil
}

func registerSetup(service context.Context, children *sync.WaitGroup, mux *http.ServeMux, root, perl string, internalPort, suggestedPort int, fixedPort bool, auth *sysopAuth, setTarget func(int)) error {
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
	registerProfile(mux, local, auth)
	mux.HandleFunc("GET /api/node-settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		data, err := os.ReadFile(filepath.Join(local, "dxregia-node.json"))
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		var config nodeConfig
		if err != nil || json.Unmarshal(data, &config) != nil || config.validate() != nil {
			http.Error(w, "Impostazioni del nodo non disponibili.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	})
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
		if fixedPort && saved.Port != suggestedPort {
			return fmt.Errorf("saved Telnet port %d differs from published port %d", saved.Port, suggestedPort)
		}
		setTarget(saved.Port)
		if perl == "" {
			return nil
		}
		if err := installPresenceHook(local); err != nil {
			return err
		}
		if err := startMojo(service, children, perl, root, *saved); err == nil {
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
		response := map[string]any{"available": !done, "suggestedPort": suggestedPort, "fixedPort": fixedPort}
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
			http.Error(w, "Scegli una porta Telnet diversa dalla porta interna di DXSpider.", http.StatusBadRequest)
			return
		}
		if fixedPort && config.Port != suggestedPort {
			http.Error(w, fmt.Sprintf("Il container pubblica la porta Telnet %d.", suggestedPort), http.StatusBadRequest)
			return
		}
		if saved != nil && !reflect.DeepEqual(*saved, config) {
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
			if err := writeMojoConfig(root, local, string(template), config, internalPort); err != nil {
				log.Printf("Mojo setup: %v", err)
				http.Error(w, "Impossibile salvare la configurazione DXSpider.", http.StatusInternalServerError)
				return
			}
			saved = &config
			setTarget(config.Port)
		}
		if perl != "" {
			if err := startMojo(service, children, perl, root, config); err != nil {
				log.Printf("Mojo startup: %v", err)
				http.Error(w, "Configurazione salvata. Avvio di DXSpider non riuscito. Consulta il registro di avvio e riprova.", http.StatusServiceUnavailable)
				return
			}
		}
		done = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"started": perl != "", "port": config.Port})
	})
	return nil
}

func writeMojoConfig(root, local, template string, config nodeConfig, internalPort int) error {
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
	lat, lon := config.coordinates()
	name, qth := config.Name, config.QTH
	if name == "" {
		name = config.Sysop
	}
	if qth == "" {
		qth = config.Locator
	}
	values := fmt.Sprintf("\n$mycall = %s;\n$myalias = %s;\n$mylocator = %s;\n$myname = %s;\n$myqth = %s;\n$myemail = %s;\n$mylatitude = %.8f;\n$mylongitude = %.8f;\n$clusterport = %d;\n",
		perlLiteral(config.Node), perlLiteral(config.Sysop), perlLiteral(config.Locator), perlLiteral(name),
		perlLiteral(qth), perlLiteral(config.Email), lat, lon, internalPort)
	if err := os.WriteFile(filepath.Join(stage, "DXVars.pm"), []byte(template[:end]+values+template[end:]), 0600); err != nil {
		return err
	}
	listeners := standardMojoListener(config.Port)
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

func perlLiteral(value string) string {
	escape := strings.NewReplacer("\\", "\\\\", "'", "\\'")
	return "'" + escape.Replace(value) + "'"
}

func (c nodeConfig) coordinates() (lat, lon float64) {
	if c.Latitude != nil && c.Longitude != nil {
		return *c.Latitude, *c.Longitude
	}
	grid := c.Locator
	lon = float64(grid[0]-'A')*20 + float64(grid[2]-'0')*2 - 180
	lat = float64(grid[1]-'A')*10 + float64(grid[3]-'0') - 90
	if len(grid) == 6 {
		return lat + (float64(grid[5]-'A')+.5)/24, lon + (float64(grid[4]-'A')+.5)/12
	}
	return lat + .5, lon + 1
}
