package main

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"
)

// DXSpider asks before replacing an older session of the same callsign: the panel console must answer it.
func TestLoginConsoleConfirmsReplacedSession(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	client.SetDeadline(time.Now().Add(5 * time.Second))
	server.SetDeadline(time.Now().Add(5 * time.Second))
	got := make(chan []string, 1)
	go func() {
		reader := bufio.NewReader(server)
		var lines []string
		read := func() {
			line, _ := reader.ReadString('\n')
			lines = append(lines, strings.TrimSpace(line))
		}
		server.Write([]byte("login: "))
		read()
		server.Write([]byte("Another session for T8TTT is already connected.\r\nDisconnect the existing session and continue? (Y/N): "))
		read()
		server.Write([]byte("T8TTT de T1TTT 22-Nov-2026 0000Z dxspider >"))
		got <- lines
	}()
	if err := loginConsole(client, bufio.NewReader(client), "T8TTT", ""); err != nil {
		t.Fatal(err)
	}
	lines := <-got
	if len(lines) != 2 || lines[0] != "T8TTT" || lines[1] != "Y" {
		t.Fatalf("unexpected exchange %q", lines)
	}
}
