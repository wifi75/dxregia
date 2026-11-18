package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

type listenerStatus struct {
	Address   string    `json:"address"`
	State     string    `json:"state"`
	CheckedAt time.Time `json:"checkedAt"`
}

func checkListener(ctx context.Context, address string) listenerStatus {
	status := listenerStatus{Address: address, State: "unreachable"}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	deadline, _ := ctx.Deadline()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err == nil {
		defer conn.Close()
		if conn.SetReadDeadline(deadline) == nil {
			status.State = "unexpected"
			scanner := bufio.NewScanner(io.LimitReader(conn, 65536))
			scanner.Split(func(data []byte, _ bool) (int, []byte, error) {
				if bytes.Contains(data, []byte("login: ")) {
					return len(data), data, nil
				}
				return 0, nil, nil
			})
			if scanner.Scan() {
				status.State = "ready"
			}
		}
	}
	status.CheckedAt = time.Now().UTC()
	return status
}

func main() {
	port := flag.Int("port", 8080, "HTTP port")
	httpBind := flag.String("http-bind", "127.0.0.1", "HTTP bind address")
	telnetBind := flag.String("setup-telnet-bind", "127.0.0.1", "IPv4 address for the generated Mojo listener")
	target := flag.String("telnet", "127.0.0.1:7300", "DXSpider host:port")
	web := flag.String("web", "../web/dist", "built Svelte directory")
	bootstrapFile := flag.String("bootstrap-password-file", "", "private file for the temporary sysop password")
	setupRoot := flag.String("setup-root", "", "fresh Mojo checkout for local setup")
	setupPerl := flag.String("setup-perl", "", "Perl executable for starting Mojo during setup")
	setupInternal := flag.Int("setup-internal-port", 27754, "Mojo internal loopback port")
	flag.Parse()

	if net.ParseIP(*httpBind) == nil || net.ParseIP(*telnetBind).To4() == nil {
		log.Fatal("Invalid HTTP or Telnet bind IP")
	}
	if *port < 1 || *port > 65535 {
		log.Fatal("HTTP port must be between 1 and 65535")
	}
	host, targetPort, err := net.SplitHostPort(*target)
	number, portErr := strconv.Atoi(targetPort)
	if err != nil || host == "" || portErr != nil || number < 1 || number > 65535 {
		log.Fatal("Telnet endpoint must be host:port")
	}
	if _, err := os.Stat(filepath.Join(*web, "index.html")); err != nil {
		log.Fatalf("Build the web frontend first: %v", err)
	}

	service, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	mux := http.NewServeMux()
	auth, err := registerAuth(mux, *setupRoot, *bootstrapFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := registerPartners(mux, *setupRoot, auth); err != nil {
		log.Fatal(err)
	}
	var activeTarget atomic.Value
	activeTarget.Store(*target)
	if err := registerSetup(service, mux, *setupRoot, *setupPerl, *setupInternal, number, *telnetBind, auth, func(port int) {
		activeTarget.Store(net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	}); err != nil {
		log.Fatal(err)
	}
	mux.HandleFunc("GET /api/listener", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if err := json.NewEncoder(w).Encode(checkListener(r.Context(), activeTarget.Load().(string))); err != nil {
			log.Print(err)
		}
	})
	mux.Handle("GET /", http.FileServer(http.Dir(*web)))
	address := net.JoinHostPort(*httpBind, strconv.Itoa(*port))
	log.Printf("DXRegia: http://%s", address)
	server := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 25 * time.Second}
	go func() {
		<-service.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Print(err)
	}
}
