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
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Backup del nodo in funzione, creati dal pannello o a un'ora fissa. L'archivio contiene credenziali e segreti
// del nodo: si scarica e si cancella solo con una sessione completa (password e secondo fattore).

const (
	backupTimeout   = 5 * time.Minute
	backupMinFree   = 256 << 20
	backupRetryWait = 30 * time.Minute
)

var backupName = regexp.MustCompile(`^dxregia-backup-[0-9]{8}-[0-9]{6}\.zip$`)

type backupSchedule struct {
	Enabled bool
	Hour    int // ora UTC, 0-23
	Keep    int // copie da conservare, 1-30
}

func (s backupSchedule) valid() bool {
	return s.Hour >= 0 && s.Hour <= 23 && s.Keep >= 1 && s.Keep <= 30
}

type backupFile struct {
	Name string
	Size int64
	At   int64
}

type backupManager struct {
	root, dir, tool string
	run             func(ctx context.Context, output string) error // sostituibile nei test

	mu         sync.Mutex
	running    bool
	lastError  string
	retryAfter time.Time
}

func newBackupManager(root, tool string) *backupManager {
	b := &backupManager{root: root, dir: filepath.Join(filepath.Dir(filepath.Clean(root)), "backups"), tool: tool}
	b.run = func(ctx context.Context, output string) error {
		return exec.CommandContext(ctx, "python3", "-I", tool, "backup", "--online", root, output).Run()
	}
	return b
}

func (b *backupManager) schedulePath() string {
	return filepath.Join(b.root, "local", "dxregia-backup.json")
}

func (b *backupManager) schedule() backupSchedule {
	s := backupSchedule{Hour: 3, Keep: 7}
	if data, err := os.ReadFile(b.schedulePath()); err == nil {
		var saved backupSchedule
		if json.Unmarshal(data, &saved) == nil && saved.valid() {
			s = saved
		}
	}
	return s
}

func (b *backupManager) list() []backupFile {
	files := []backupFile{}
	entries, _ := os.ReadDir(b.dir)
	for _, entry := range entries {
		if !backupName.MatchString(entry.Name()) || !entry.Type().IsRegular() {
			continue
		}
		if info, err := entry.Info(); err == nil {
			files = append(files, backupFile{entry.Name(), info.Size(), backupTime(entry.Name())})
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name > files[j].Name })
	return files
}

func backupTime(name string) int64 {
	at, err := time.Parse("20060102-150405", strings.TrimSuffix(strings.TrimPrefix(name, "dxregia-backup-"), ".zip"))
	if err != nil {
		return 0
	}
	return at.Unix()
}

// create produce un archivio nuovo e applica la conservazione. Una sola creazione alla volta.
func (b *backupManager) create(ctx context.Context, now time.Time) error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return errors.New("Un backup è già in corso.")
	}
	b.running = true
	b.mu.Unlock()
	err := b.createLocked(ctx, now)
	b.mu.Lock()
	b.running = false
	if err != nil {
		b.lastError = err.Error()
		b.retryAfter = now.Add(backupRetryWait)
	} else {
		b.lastError = ""
	}
	b.mu.Unlock()
	return err
}

func (b *backupManager) createLocked(ctx context.Context, now time.Time) error {
	if err := os.MkdirAll(b.dir, 0700); err != nil {
		return errors.New("Impossibile preparare la cartella dei backup.")
	}
	if free, _, ok := diskSpace(b.dir); ok && free < backupMinFree {
		return errors.New("Spazio su disco insufficiente per un backup.")
	}
	output := filepath.Join(b.dir, "dxregia-backup-"+now.UTC().Format("20060102-150405")+".zip")
	ctx, cancel := context.WithTimeout(ctx, backupTimeout)
	defer cancel()
	if err := b.run(ctx, output); err != nil {
		_ = os.Remove(output)
		log.Printf("Backup: %v", err)
		return errors.New("Backup non riuscito.")
	}
	if err := os.Chmod(output, 0600); err != nil {
		log.Printf("Backup: %v", err)
	}
	b.prune(b.schedule().Keep)
	return nil
}

func (b *backupManager) prune(keep int) {
	files := b.list()
	for i := keep; i < len(files); i++ {
		if err := os.Remove(filepath.Join(b.dir, files[i].Name)); err != nil {
			log.Printf("Backup: %v", err)
		}
	}
}

// due dice se l'ora pianificata è arrivata e oggi (UTC) non c'è ancora un backup.
func (b *backupManager) due(now time.Time) bool {
	s := b.schedule()
	now = now.UTC()
	if !s.Enabled || now.Hour() < s.Hour {
		return false
	}
	b.mu.Lock()
	wait := b.running || now.Before(b.retryAfter)
	b.mu.Unlock()
	if wait {
		return false
	}
	files := b.list()
	return len(files) == 0 || time.Unix(files[0].At, 0).UTC().Format("2006-01-02") != now.Format("2006-01-02")
}

func (b *backupManager) watch(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if b.due(time.Now()) {
				if err := b.create(ctx, time.Now()); err != nil {
					log.Printf("Backup pianificato: %v", err)
				}
			}
		}
	}
}

func registerBackups(ctx context.Context, mux *http.ServeMux, auth *sysopAuth, root, tool string) *backupManager {
	if root == "" || auth == nil || tool == "" {
		return nil
	}
	b := newBackupManager(root, tool)
	go b.watch(ctx)
	b.routes(mux, auth)
	return b
}

func (b *backupManager) routes(mux *http.ServeMux, auth *sysopAuth) {
	strong := func(w http.ResponseWriter, r *http.Request) bool {
		w.Header().Set("Cache-Control", "no-store")
		auth.mu.Lock()
		ok := auth.signedIn(r) && !auth.temporary && auth.factorOK
		auth.mu.Unlock()
		if !ok {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
		}
		return ok
	}
	decode := func(w http.ResponseWriter, r *http.Request, into any) bool {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Formato non valido.", http.StatusUnsupportedMediaType)
			return false
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if decoder.Decode(into) != nil || decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return false
		}
		return true
	}
	mux.HandleFunc("GET /api/backups", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !auth.allowed(r) {
			http.Error(w, "Accesso sysop richiesto.", http.StatusUnauthorized)
			return
		}
		b.mu.Lock()
		running, lastError := b.running, b.lastError
		b.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"Schedule": b.schedule(), "Backups": b.list(), "Running": running, "LastError": lastError})
	})
	mux.HandleFunc("PUT /api/backups/schedule", func(w http.ResponseWriter, r *http.Request) {
		if !strong(w, r) {
			return
		}
		var s backupSchedule
		if !decode(w, r, &s) {
			return
		}
		if !s.valid() {
			http.Error(w, "Ora o numero di copie non valido.", http.StatusBadRequest)
			return
		}
		data, _ := json.Marshal(s)
		if err := writeNodeFile(b.schedulePath(), data); err != nil {
			http.Error(w, "Impossibile salvare la pianificazione.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/backups/run", func(w http.ResponseWriter, r *http.Request) {
		if !strong(w, r) {
			return
		}
		var empty struct{}
		if !decode(w, r, &empty) {
			return
		}
		b.mu.Lock()
		busy := b.running
		b.mu.Unlock()
		if busy {
			http.Error(w, "Un backup è già in corso.", http.StatusConflict)
			return
		}
		go func() {
			if err := b.create(context.Background(), time.Now()); err != nil {
				log.Printf("Backup: %v", err)
			}
		}()
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("GET /api/backups/files/{name}", func(w http.ResponseWriter, r *http.Request) {
		if !strong(w, r) {
			return
		}
		name := r.PathValue("name")
		if !backupName.MatchString(name) {
			http.NotFound(w, r)
			return
		}
		file, err := os.Open(filepath.Join(b.dir, name))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
		http.ServeContent(w, r, name, info.ModTime(), file)
	})
	mux.HandleFunc("DELETE /api/backups/files/{name}", func(w http.ResponseWriter, r *http.Request) {
		if !strong(w, r) {
			return
		}
		name := r.PathValue("name")
		if !backupName.MatchString(name) {
			http.NotFound(w, r)
			return
		}
		if err := os.Remove(filepath.Join(b.dir, name)); err != nil {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
