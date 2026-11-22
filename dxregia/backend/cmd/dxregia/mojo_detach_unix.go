//go:build !windows

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Su Linux il motore gira in un gruppo di processi proprio, avviato da una piccola shell che registra il numero
// del processo e il codice di uscita. Così il servizio web può essere riavviato (aggiornamento del solo pannello)
// senza fermare Telnet e i collegamenti: il nuovo servizio trova il motore ancora vivo e lo adotta.
//
// Il motore resta in vita all'arresto del servizio solo se lo script di aggiornamento ha lasciato il file
// «keep-engine» (recente) nella cartella di stato. In ogni altro caso l'arresto del servizio ferma il motore.

const (
	keepEngineMaxAge = 10 * time.Minute
	engineScript     = `rm -f "$3" "$4"; "$1" "$2" & p=$!; echo $p > "$3"; wait $p; c=$?; echo $c > "$4"; exit $c`
)

func enginePIDFile(root string) string  { return filepath.Join(root, "local_data", "dxregia-engine.pid") }
func engineExitFile(root string) string { return filepath.Join(root, "local_data", "dxregia-engine.exit") }

func keepEngineFile(root string) string {
	return filepath.Join(filepath.Dir(filepath.Clean(root)), "keep-engine")
}

func keepEngineOnStop(root string) bool {
	info, err := os.Stat(keepEngineFile(root))
	return err == nil && time.Since(info.ModTime()) < keepEngineMaxAge
}

func processAlive(pid int) bool {
	return pid > 1 && syscall.Kill(pid, 0) == nil
}

// isEngineProcess evita di adottare un numero di processo riassegnato a un altro programma.
func isEngineProcess(pid int, marker string) bool {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil {
		return processAlive(pid) && marker == ""
	}
	return strings.Contains(string(data), marker)
}

func readPID(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return pid
}

// startEngineProcess avvia perl <script> dentro la shell di sorveglianza, in un gruppo di processi proprio.
func startEngineProcess(_ context.Context, perl, root string, logs *os.File) (*engineProc, error) {
	return startEngineCommand(perl, filepath.Join("perl", "cluster.pl"), root, logs)
}

func startEngineCommand(program, script, root string, logs *os.File) (*engineProc, error) {
	cmd := exec.Command("sh", "-c", engineScript, "dxregia-engine", program, script, enginePIDFile(root), engineExitFile(root))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "DXSPIDER_ROOT="+filepath.ToSlash(root))
	cmd.Stdout, cmd.Stderr = logs, logs
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	group := cmd.Process.Pid
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	pid := int64(group)
	for i := 0; i < 20; i++ {
		if p := readPID(enginePIDFile(root)); p > 1 {
			pid = int64(p)
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return &engineProc{
		pid:  pid,
		done: done,
		// Solo perl riceve SIGTERM: la shell attende la sua uscita e registra il codice, così l'arresto è ordinato.
		term: func() { _ = syscall.Kill(int(pid), syscall.SIGTERM) },
		kill: func() { _ = syscall.Kill(-group, syscall.SIGKILL) },
	}, nil
}

// adoptRunningEngine riprende un motore ancora attivo di un servizio precedente. Restituisce il processo adottato.
func adoptRunningEngine(root, marker string, ready func() bool) (*engineProc, bool) {
	pid := readPID(enginePIDFile(root))
	if pid <= 1 || !isEngineProcess(pid, marker) || !ready() {
		return nil, false
	}
	done := make(chan error, 1)
	var once sync.Once
	go func() {
		for processAlive(pid) {
			time.Sleep(2 * time.Second)
		}
		once.Do(func() {
			time.Sleep(300 * time.Millisecond) // la shell scrive il codice di uscita un istante dopo
			if code, err := os.ReadFile(engineExitFile(root)); err == nil && strings.TrimSpace(string(code)) == "0" {
				done <- nil
			} else {
				done <- errors.New("il motore adottato si è fermato in modo anomalo")
			}
		})
	}()
	return &engineProc{
		pid: int64(pid), done: done, adopted: true,
		term: func() { _ = syscall.Kill(pid, syscall.SIGTERM) },
		kill: func() { _ = syscall.Kill(pid, syscall.SIGKILL) },
	}, true
}
