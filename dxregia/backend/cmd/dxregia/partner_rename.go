package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Rinomina di un nodo partner: serve a correggere un indicativo sbagliato. Si può fare solo finché il nodo non è
// stato abilitato in Mojo e non è collegato: dopo, l'indicativo vive anche nel motore e si elimina e si ricrea.

func partnerRenameBlocker(root string, entry partner, snapshot presence) string {
	if entry.Provisioning != "" {
		return "Il nodo è già stato abilitato in DXSpider: per cambiare indicativo eliminalo e registralo di nuovo."
	}
	if state := currentConnection(root, entry.Call, snapshot); state == "Attivo" || state == "Collegamento in corso" {
		return "Il nodo è collegato: attendi che la sessione termini."
	}
	files, _ := filepath.Glob(filepath.Join(root, "local_data", "dxregia-partners", entry.Call+".*"))
	for _, file := range files {
		if !strings.HasSuffix(file, ".reconnect.json") {
			return "Il nodo ha già dati di collegamento: per cambiare indicativo eliminalo e registralo di nuovo."
		}
	}
	return ""
}

// renamePartner sposta la scheda e le credenziali salvate da from a to. Restituisce lo stato HTTP e il testo d'errore.
func renamePartner(root, path string, partners map[string]partner, from, to string, snapshot presence, now time.Time) (partner, int, string) {
	d := partnerDetails{Call: to, Host: "x.invalid", Port: 1}
	if !d.validate() {
		return partner{}, http.StatusBadRequest, "Indicativo non valido."
	}
	to = d.Call
	entry, ok := partners[from]
	if !ok || !callsign.MatchString(from) || len(from) > 12 {
		return partner{}, http.StatusNotFound, "Nodo non trovato."
	}
	if to == from {
		return partner{}, http.StatusBadRequest, "Il nuovo indicativo è uguale a quello attuale."
	}
	if _, exists := partners[to]; exists {
		return partner{}, http.StatusConflict, "Indicativo già registrato."
	}
	if reason := partnerRenameBlocker(root, entry, snapshot); reason != "" {
		return partner{}, http.StatusConflict, reason
	}
	oldFile, newFile := reconnectPath(root, from), reconnectPath(root, to)
	moved := false
	if _, err := os.Stat(oldFile); err == nil {
		if _, err := os.Stat(newFile); err == nil {
			return partner{}, http.StatusConflict, "Esistono già dati salvati per il nuovo indicativo."
		}
		if err := os.Rename(oldFile, newFile); err != nil {
			return partner{}, http.StatusInternalServerError, "Impossibile spostare le credenziali del nodo."
		}
		moved = true
	}
	renamed := entry
	renamed.Call = to
	renamed.Notes = append(append([]partnerNote{}, entry.Notes...), partnerNote{Text: "Indicativo cambiato da " + from + " a " + to + ".", At: now.UTC()})
	partners[to] = renamed
	delete(partners, from)
	if err := savePartners(path, partners); err != nil {
		partners[from] = entry
		delete(partners, to)
		if moved {
			_ = os.Rename(newFile, oldFile)
		}
		return partner{}, http.StatusInternalServerError, "Impossibile salvare l'anagrafica."
	}
	return renamed, http.StatusOK, ""
}
