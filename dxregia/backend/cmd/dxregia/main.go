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
	"strings"
	"strconv"
	"sync"
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
	target := flag.String("telnet", "127.0.0.1:7300", "DXSpider host:port")
	web := flag.String("web", "../web/dist", "built Svelte directory")
	bootstrapFile := flag.String("bootstrap-password-file", "", "private file for the temporary sysop password")
	setupRoot := flag.String("setup-root", "", "fresh Mojo checkout for local setup")
	setupPerl := flag.String("setup-perl", "", "Perl executable for starting Mojo during setup")
	setupFixedPort := flag.Bool("setup-fixed-telnet-port", false, "require the Telnet port published by the container")
	setupInternal := flag.Int("setup-internal-port", 27754, "Mojo internal loopback port")
	backupTool := flag.String("backup-tool", "", "path to the installed backup utility")
	webPortFile := flag.String("web-port-file", "", "persistent HTTP port settings for native installations")
	flag.Parse()

	if net.ParseIP(*httpBind) == nil {
		log.Fatal("Invalid HTTP bind IP")
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
	var children sync.WaitGroup
	mux := http.NewServeMux()
	auth, err := registerAuth(mux, *setupRoot, *bootstrapFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := registerPartners(service, mux, *setupRoot, auth); err != nil {
		log.Fatal(err)
	}
	registerBackupVerification(mux, auth, *backupTool)
	registerBackups(service, mux, auth, *setupRoot, *backupTool)
	registerPresence(mux, *setupRoot, auth)
	registerRecords(service, mux, *setupRoot, auth)
	registerSMTP(mux, *setupRoot, auth)
	registerNodeFiles(mux, *setupRoot, auth)
	go refreshIssueCountries(service, *setupRoot)
	registerNodeLists(mux, *setupRoot, auth)
	registerConnections(mux, *setupRoot, auth)
	registerPartnerMailHistory(mux, *setupRoot, auth)
	registerEmailPreferences(mux, *setupRoot, auth)
	registerUsers(mux, *setupRoot, auth)
	registerAutoPartner(mux, *setupRoot, auth)
	registerLive(mux, *setupRoot, auth)
	registerMojoLog(mux, auth, *setupRoot)
	registerTelnetPassword(mux, auth, *setupRoot)
	registerTelnetOptions(mux, auth, *setupRoot)
	spots := registerSpots(mux, *setupRoot, loadQSLDirectory(service, *setupRoot))
	registerBulletins(mux, *setupRoot)
	registerSolarConditions(service, mux, *setupRoot)
	registerRBN(mux, *setupRoot, auth)
	registerNodeHealth(service, mux, *setupRoot, auth, spots)
	registerADXO(service, mux, *setupRoot)
	registerEngineVersion(mux, *setupRoot)
	registerUpdate(mux, *setupRoot, auth)
	if err := registerContacts(service, mux, *setupRoot, auth); err != nil {
		log.Fatal(err)
	}
	traffic, err := registerTraffic(mux, *setupRoot, auth)
	if err != nil {
		log.Fatal(err)
	}
	registerAnalytics(mux, *setupRoot, auth, traffic)
	registerClusterStats(mux, *setupRoot)
	var activeTarget atomic.Value
	activeTarget.Store(*target)
	if err := registerSetup(service, &children, mux, *setupRoot, *setupPerl, *setupInternal, number, *setupFixedPort, auth, func(port int) {
		activeTarget.Store(net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	}); err != nil {
		log.Fatal(err)
	}
	registerTelnetConsole(mux, auth, func() string { return activeTarget.Load().(string) })
	mux.HandleFunc("GET /api/listener", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if err := json.NewEncoder(w).Encode(checkListener(r.Context(), activeTarget.Load().(string))); err != nil {
			log.Print(err)
		}
	})
	for _, route := range []string{"GET /registrazione", "GET /registrazione/conferma"} {
		mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFile(w, r, filepath.Join(*web, "index.html"))
		})
	}
	mux.HandleFunc("GET /contatti", func(w http.ResponseWriter, r *http.Request) {
		traffic.view(r)
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(*web, "index.html"))
	})
	mux.HandleFunc("GET /spots", func(w http.ResponseWriter, r *http.Request) {
		traffic.view(r)
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(*web, "index.html"))
	})
	mux.HandleFunc("GET /statistiche", func(w http.ResponseWriter, r *http.Request) {
		traffic.view(r)
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(*web, "index.html"))
	})
	// Il pannello sysop vive su /admin e non è collegato da nessuna pagina pubblica.
	adminPage := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, filepath.Join(*web, "index.html"))
	}
	mux.HandleFunc("GET /admin", adminPage)
	mux.HandleFunc("GET /admin/", adminPage)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/spots", http.StatusFound)
	})
	files := http.FileServer(http.Dir(*web))
	// I file in /assets hanno il codice nel nome e non cambiano mai; il resto va ricontrollato ad ogni apertura.
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
	httpPort, err := loadWebPort(*webPortFile, *port)
	if err != nil {
		log.Fatalf("Load HTTP port: %v", err)
	}
	server := &http.Server{Addr: net.JoinHostPort(*httpBind, strconv.Itoa(httpPort)), Handler: mux,
		ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 25 * time.Second}
	webListener := &webListener{server: server, host: *httpBind, port: httpPort, path: *webPortFile}
	webListener.listener, err = webListener.bind(httpPort)
	if err != nil {
		log.Fatalf("HTTP listen: %v", err)
	}
	webListener.register(mux, auth)
	go webListener.serve(webListener.listener)
	<-service.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
	children.Wait()
}
