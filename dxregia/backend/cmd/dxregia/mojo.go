package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func startMojo(service context.Context, children *sync.WaitGroup, perl, root string, config nodeConfig) error {
	data := filepath.Join(root, "local_data")
	if err := os.MkdirAll(data, 0700); err != nil {
		return err
	}
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(config.Port))
	if err := checkPortAvailable(config.Port); err != nil {
		return fmt.Errorf("la porta Telnet non è disponibile: %w", err)
	}

	logFile, err := os.OpenFile(filepath.Join(data, "dxregia-startup.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer logFile.Close()
	db := filepath.Join(data, "dxusers.db")
	created := filepath.Join(root, "local", "dxregia-initialized")
	if _, err := os.Stat(created); os.IsNotExist(err) {
		for _, name := range []string{db, filepath.Join(data, "users.v2"), filepath.Join(data, "users.v3"), filepath.Join(data, "users.v3j")} {
			if _, err := os.Lstat(name); !os.IsNotExist(err) {
				return errors.New("banca utenti esistente o inizializzazione precedente incompleta")
			}
		}
		ctx, cancel := context.WithTimeout(service, 12*time.Second)
		err = mojoCommand(ctx, perl, root, "create_sysop.pl", logFile).Run()
		cancel()
		if err != nil {
			return fmt.Errorf("inizializzazione utenti: %w", err)
		}
		if _, err := os.Stat(db); err != nil {
			return fmt.Errorf("banca utenti non creata: %w", err)
		}
		if err := os.WriteFile(created, nil, 0600); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if err := ensureMojoListener(filepath.Join(root, "local"), config.Port); err != nil {
		return fmt.Errorf("configurazione listener Telnet: %w", err)
	}
	cmd := mojoCommand(service, perl, root, "cluster.pl", logFile)
	if runtime.GOOS != "windows" {
		cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
		cmd.WaitDelay = 8 * time.Second
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	pid := int64(cmd.Process.Pid)
	managedEnginePID.Store(pid)
	exited := make(chan error, 1)
	children.Add(1)
	go func() {
		defer children.Done()
		err := cmd.Wait()
		managedEnginePID.CompareAndSwap(pid, 0)
		exited <- err
	}()
	probe, cancel := context.WithTimeout(service, 9*time.Second)
	defer cancel()
	timer := time.NewTicker(250 * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case err := <-exited:
			return fmt.Errorf("Mojo si è chiuso prima di rispondere: %v", err)
		case <-probe.Done():
			_ = cmd.Process.Kill()
			<-exited
			return errors.New("Mojo non ha risposto sulla porta Telnet")
		case <-timer.C:
			if checkListener(probe, address).State == "ready" {
				select {
				case err := <-exited:
					return fmt.Errorf("Mojo si è chiuso durante l'avvio: %v", err)
				default:
					go func() { log.Printf("Mojo terminato: %v", <-exited) }()
					return nil
				}
			}
		}
	}
}

func checkPortAvailable(port int) error {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return err
	}
	return listener.Close()
}

func mojoCommand(ctx context.Context, perl, root, script string, logs io.Writer) *exec.Cmd {
	cmd := exec.CommandContext(ctx, perl, filepath.Join("perl", script))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "DXSPIDER_ROOT="+filepath.ToSlash(root))
	cmd.Stdout = logs
	cmd.Stderr = logs
	return cmd
}

// The public listener uses every IPv4 interface. Mojo's internal port and
// the local health check continue to use loopback.
func standardMojoListener(port int) string {
	return fmt.Sprintf("package main;\nour @listen = ([ '0.0.0.0', %d ]);\n1;\n", port)
}

func ensureMojoListener(local string, port int) error {
	path := filepath.Join(local, "Listeners.pm")
	contents := standardMojoListener(port)
	previous, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if string(previous) == contents {
		return nil
	}
	if err == nil {
		backup := filepath.Join(local, "Listeners.pm.before-dxregia")
		file, err := os.OpenFile(backup, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err == nil {
			_, err = file.Write(previous)
			if closeErr := file.Close(); err == nil {
				err = closeErr
			}
			if err != nil {
				_ = os.Remove(backup)
				return err
			}
		} else if !os.IsExist(err) {
			return err
		}
	}
	file, err := os.CreateTemp(local, ".dxregia-listeners-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString(contents)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
