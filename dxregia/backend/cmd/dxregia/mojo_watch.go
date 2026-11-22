package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Sorveglianza del motore: se Mojo termina in modo anomalo mentre il servizio web gira, viene riavviato.
// Un'uscita pulita (codice 0, per esempio il comando shutdown del sysop) o l'arresto del servizio non riavviano.
// Dopo troppi arresti in poco tempo il riavvio automatico si sospende, per non entrare in un ciclo.

const (
	mojoRestartWindow = 10 * time.Minute
	mojoRestartLimit  = 5
)

var mojoRestarts struct {
	sync.Mutex
	times []time.Time
}

// allowMojoRestart registra un riavvio automatico e dice se il limite nella finestra è stato superato.
func allowMojoRestart(now time.Time) bool {
	mojoRestarts.Lock()
	defer mojoRestarts.Unlock()
	kept := mojoRestarts.times[:0]
	for _, at := range mojoRestarts.times {
		if now.Sub(at) < mojoRestartWindow {
			kept = append(kept, at)
		}
	}
	mojoRestarts.times = kept
	if len(kept) >= mojoRestartLimit {
		return false
	}
	mojoRestarts.times = append(kept, now)
	return true
}

// noteMojoWatch aggiunge una riga fissa al registro di avvio, che la pagina Log DXSpider mostra.
func noteMojoWatch(root, event string) {
	file, err := os.OpenFile(filepath.Join(root, "local_data", "dxregia-startup.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer file.Close()
	fmt.Fprintf(file, "%d^DXRegia watchdog %s\n", time.Now().Unix(), event)
}

func watchMojo(service context.Context, children *sync.WaitGroup, perl, root string, config nodeConfig, exited <-chan error) {
	err := <-exited
	log.Printf("Mojo terminato: %v", err)
	if service.Err() != nil || err == nil {
		return
	}
	if !allowMojoRestart(time.Now()) {
		noteMojoWatch(root, "giveup")
		log.Print("Riavvio automatico di Mojo sospeso: troppi arresti ravvicinati")
		return
	}
	noteMojoWatch(root, "restart")
	mojoAutoRestarts.Add(1)
	log.Print("Mojo si è fermato in modo anomalo: riavvio automatico")
	go retryMojo(service, children, perl, root, config)
}
