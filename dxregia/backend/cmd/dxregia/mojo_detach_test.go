//go:build !windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func engineTestRoot(t *testing.T) (root string, logs *os.File) {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh non disponibile")
	}
	root = filepath.Join(t.TempDir(), "mojo")
	if err := os.MkdirAll(filepath.Join(root, "local_data"), 0700); err != nil {
		t.Fatal(err)
	}
	logs, err := os.Create(filepath.Join(root, "local_data", "dxregia-startup.log"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { logs.Close() })
	return root, logs
}

func writeScript(t *testing.T, root, name, body string) string {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(body), 0700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestEngineRecordsExitCode(t *testing.T) {
	root, logs := engineTestRoot(t)
	script := writeScript(t, root, "fail.sh", "exit 3\n")
	proc, err := startEngineCommand("sh", script, root, logs)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-proc.done:
		if err == nil {
			t.Fatal("exit code 3 reported as success")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("engine did not end")
	}
	if code, _ := os.ReadFile(engineExitFile(root)); strings.TrimSpace(string(code)) != "3" {
		t.Fatalf("exit file: %q", code)
	}
	if readPID(enginePIDFile(root)) <= 1 {
		t.Fatal("pid file missing")
	}
}

func TestEngineAdoptedAfterServiceRestart(t *testing.T) {
	root, logs := engineTestRoot(t)
	script := writeScript(t, root, "long-engine.sh", "sleep 30\n")
	proc, err := startEngineCommand("sh", script, root, logs)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(proc.kill)
	// Un secondo servizio (nuovo processo web) trova il motore già vivo e pronto.
	if _, ok := adoptRunningEngine(root, "long-engine.sh", func() bool { return false }); ok {
		t.Fatal("adopted an engine whose port is not ready")
	}
	if _, ok := adoptRunningEngine(root, "altro-programma", func() bool { return true }); ok {
		t.Fatal("adopted a process that is not the engine")
	}
	adopted, ok := adoptRunningEngine(root, "long-engine.sh", func() bool { return true })
	if !ok || !adopted.adopted || adopted.pid != proc.pid {
		t.Fatalf("not adopted: %+v %v (pid %d)", adopted, ok, proc.pid)
	}
	adopted.kill()
	select {
	case err := <-adopted.done:
		if err == nil {
			t.Fatal("killed engine reported as clean exit")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("adopted engine end not noticed")
	}
}

func TestKeepEngineFileExpires(t *testing.T) {
	root := filepath.Join(t.TempDir(), "mojo")
	if keepEngineOnStop(root) {
		t.Fatal("keep without flag")
	}
	flag := keepEngineFile(root)
	if err := os.WriteFile(flag, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if !keepEngineOnStop(root) {
		t.Fatal("fresh flag ignored")
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(flag, old, old); err != nil {
		t.Fatal(err)
	}
	if keepEngineOnStop(root) {
		t.Fatal("stale flag honored")
	}
}
