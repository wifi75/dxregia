package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Aggiornamento di DXRegia dal pannello di amministrazione.
//
// Il servizio web non gira come root e viene riavviato dall'aggiornamento stesso, quindi non può aggiornarsi
// da solo. Il pannello scrive soltanto un file di richiesta, senza contenuto scelto dall'utente: una unità
// systemd (dxregia-update.path) lo vede e avvia dxregia-update.service, che esegue lo script ufficiale
// update-native.sh con i permessi necessari. Lo stato si legge dal registro dello script e da systemd.

const (
	updateRepo      = "wifi75/dxregia"
	updateCheckAge  = 10 * time.Minute
	updateCheckGap  = 20 * time.Second
	updateLogLines  = 80
	updateLogMaxCol = 300
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

type updateInfo struct {
	Installed    bool // le unità systemd dell'aggiornamento sono presenti
	State        string
	Current      string
	Available    string
	CurrentTag   string `json:",omitempty"` // nome della versione, se il commit ha un tag
	AvailableTag string `json:",omitempty"`
	UpToDate     bool
	CheckedAt    int64
	CheckError   string   `json:",omitempty"`
	Log          []string `json:",omitempty"`
}

type updater struct {
	revisionFile, pathUnit, requestFile, logFile string
	remote                                       func(context.Context) (head string, tags map[string]string, err error)
	running                                      func() bool

	mu       sync.Mutex
	head     string
	tags     map[string]string
	checked  time.Time
	checkErr string
}

func newUpdater(stateDir string) *updater {
	return &updater{
		revisionFile: "/opt/dxregia/revision",
		pathUnit:     "/etc/systemd/system/dxregia-update.path",
		requestFile:  filepath.Join(stateDir, "update-request"),
		logFile:      filepath.Join(stateDir, "update.log"),
		remote:       githubHead,
		running:      systemdUpdateRunning,
	}
}

func githubGet(ctx context.Context, path, accept string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+updateRepo+path, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", accept)
	request.Header.Set("User-Agent", "dxregia-update-check")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("GitHub ha risposto " + response.Status)
	}
	return io.ReadAll(io.LimitReader(response.Body, 1<<20))
}

// githubHead restituisce l'ultimo commit del ramo main, quello che installa update-native.sh,
// e i tag pubblicati (commit completo -> nome), per dare un nome di versione ai commit.
func githubHead(ctx context.Context) (string, map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	data, err := githubGet(ctx, "/commits/main", "application/vnd.github.sha")
	if err != nil {
		return "", nil, err
	}
	head := strings.TrimSpace(string(data))
	if len(head) != 40 || strings.Trim(head, "0123456789abcdef") != "" {
		return "", nil, errors.New("risposta di GitHub non valida")
	}
	tags := map[string]string{}
	if data, err := githubGet(ctx, "/tags?per_page=50", "application/vnd.github+json"); err == nil {
		var list []struct {
			Name   string `json:"name"`
			Commit struct {
				Sha string `json:"sha"`
			} `json:"commit"`
		}
		if json.Unmarshal(data, &list) == nil {
			for _, tag := range list {
				if len(tag.Name) <= 40 && len(tag.Commit.Sha) == 40 {
					if _, seen := tags[tag.Commit.Sha]; !seen {
						tags[tag.Commit.Sha] = tag.Name
					}
				}
			}
		}
	}
	return head, tags, nil
}

func systemdUpdateRunning() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "systemctl", "show", "dxregia-update.service", "-p", "ActiveState", "--value").Output()
	if err != nil {
		return false
	}
	state := strings.TrimSpace(string(out))
	return state == "active" || state == "activating"
}

func (u *updater) refresh(ctx context.Context, force bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	age := time.Since(u.checked)
	if age < updateCheckGap || (!force && age < updateCheckAge && u.head != "") {
		return
	}
	head, tags, err := u.remote(ctx)
	u.checked = time.Now()
	if err != nil {
		u.checkErr = "Impossibile controllare gli aggiornamenti su GitHub."
		return
	}
	u.head, u.tags, u.checkErr = head, tags, ""
}

func (u *updater) installedRevision() string {
	data, err := os.ReadFile(u.revisionFile)
	if err != nil {
		return ""
	}
	revision := strings.TrimSpace(string(data))
	if len(revision) != 40 || strings.Trim(revision, "0123456789abcdef") != "" {
		return ""
	}
	return revision
}

func shortRevision(revision string) string {
	if len(revision) > 12 {
		return revision[:12]
	}
	return revision
}

// readLog restituisce le ultime righe del registro dello script, senza i codici colore.
func (u *updater) readLog() []string {
	data, err := os.ReadFile(u.logFile)
	if err != nil {
		return nil
	}
	if len(data) > 64<<10 {
		data = data[len(data)-64<<10:]
	}
	var lines []string
	for _, line := range strings.Split(ansiEscape.ReplaceAllString(string(data), ""), "\n") {
		line = strings.TrimRight(line, " \r\t")
		if line == "" {
			continue
		}
		if len(line) > updateLogMaxCol {
			line = line[:updateLogMaxCol]
		}
		lines = append(lines, line)
	}
	if len(lines) > updateLogLines {
		lines = lines[len(lines)-updateLogLines:]
	}
	return lines
}

func updateState(running, queued bool, log []string) string {
	switch {
	case running:
		return "running"
	case queued:
		return "queued"
	}
	for i := len(log) - 1; i >= 0; i-- {
		line := log[i]
		switch {
		case strings.Contains(line, "Ripristino della versione precedente"), strings.Contains(line, "Errore:"):
			return "failed"
		case strings.Contains(line, "DXRegia aggiornato a"), strings.Contains(line, "già aggiornato"):
			return "success"
		}
	}
	return "idle"
}

func (u *updater) info() updateInfo {
	u.mu.Lock()
	head, tags, checked, checkErr := u.head, u.tags, u.checked, u.checkErr
	u.mu.Unlock()
	_, unitErr := os.Stat(u.pathUnit)
	_, requestErr := os.Stat(u.requestFile)
	log := u.readLog()
	current := u.installedRevision()
	info := updateInfo{
		Installed:    unitErr == nil,
		State:        updateState(u.running(), requestErr == nil, log),
		Current:      shortRevision(current),
		Available:    shortRevision(head),
		CurrentTag:   tags[current],
		AvailableTag: tags[head],
		UpToDate:     current != "" && head != "" && current == head,
		CheckError:   checkErr,
		Log:          log,
	}
	if !checked.IsZero() {
		info.CheckedAt = checked.Unix()
	}
	return info
}

func registerUpdate(mux *http.ServeMux, root string, auth *sysopAuth) {
	if root == "" || auth == nil {
		return
	}
	u := newUpdater(filepath.Dir(filepath.Clean(root)))
	u.register(mux, auth)
}

func (u *updater) register(mux *http.ServeMux, auth *sysopAuth) {
	mux.HandleFunc("GET /api/update", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		u.refresh(r.Context(), r.URL.Query().Get("check") == "1")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u.info())
	})
	mux.HandleFunc("POST /api/update/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return
		}
		// Il corpo deve essere un oggetto JSON vuoto: nessun dato dell'utente arriva allo script.
		var request struct{}
		body := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64))
		body.DisallowUnknownFields()
		if body.Decode(&request) != nil || body.Decode(new(any)) != io.EOF {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return
		}
		auth.mu.Lock()
		allowed := auth.signedIn(r) && !auth.temporary && auth.factorOK
		auth.mu.Unlock()
		if !allowed {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		u.refresh(r.Context(), true)
		info := u.info()
		switch {
		case !info.Installed:
			http.Error(w, "L’aggiornamento dal pannello non è ancora installato su questo server.", http.StatusConflict)
		case info.State == "running" || info.State == "queued":
			http.Error(w, "Un aggiornamento è già in corso.", http.StatusConflict)
		case info.Available == "":
			http.Error(w, "Impossibile controllare gli aggiornamenti su GitHub.", http.StatusBadGateway)
		case info.UpToDate:
			http.Error(w, "DXRegia è già aggiornato.", http.StatusConflict)
		default:
			// La richiesta non porta dati: la unità systemd la rimuove all'avvio dello script.
			if err := os.WriteFile(u.requestFile, nil, 0600); err != nil {
				http.Error(w, "Impossibile avviare l’aggiornamento.", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusAccepted)
		}
	})
}
