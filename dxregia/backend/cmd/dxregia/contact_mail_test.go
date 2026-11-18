package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestStartTLSRequiresUpgradeBeforeAuthentication(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	server := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			server <- err
			return
		}
		defer conn.Close()
		server <- rejectStartTLS(conn)
	}()

	mailer := contactMailer{
		address:  listener.Addr().String(),
		host:     "localhost",
		from:     "sysop@example.org",
		user:     "sysop",
		password: "test-password",
		startTLS: true,
	}
	if err := mailer.send("peer@example.org", "test message"); err == nil {
		t.Fatal("sending unexpectedly succeeded after STARTTLS was rejected")
	}
	if err := <-server; err != nil {
		t.Fatal(err)
	}
}

func rejectStartTLS(conn net.Conn) error {
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return err
	}
	reader := bufio.NewReader(conn)
	if _, err := fmt.Fprint(conn, "220 localhost ESMTP\r\n"); err != nil {
		return err
	}
	line, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, "EHLO ") {
		return fmt.Errorf("expected SMTP greeting, got %q (%v)", line, err)
	}
	if _, err := fmt.Fprint(conn, "250-localhost\r\n250-AUTH PLAIN\r\n250 STARTTLS\r\n"); err != nil {
		return err
	}
	line, err = reader.ReadString('\n')
	if err != nil || line != "STARTTLS\r\n" {
		return fmt.Errorf("expected STARTTLS immediately after greeting, got %q (%v)", line, err)
	}
	if _, err := fmt.Fprint(conn, "454 TLS unavailable\r\n"); err != nil {
		return err
	}
	if line, err = reader.ReadString('\n'); line != "" || err != io.EOF {
		return fmt.Errorf("unexpected SMTP command after TLS rejection: %q", line)
	}
	return nil
}
