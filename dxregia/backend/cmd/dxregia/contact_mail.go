package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"time"
)

type contactMailer struct {
	address, host, from, user, password string
	startTLS                            bool
}

func loadContactMailer() (*contactMailer, error) {
	address := os.Getenv("DXREGIA_SMTP_ADDRESS")
	if address == "" {
		return nil, nil
	}
	host, _, err := net.SplitHostPort(address)
	from := os.Getenv("DXREGIA_SMTP_FROM")
	sender, fromErr := mail.ParseAddress(from)
	if err != nil || host == "" || fromErr != nil || sender.Address != from {
		return nil, errors.New("invalid DXRegia SMTP address or sender")
	}
	mode := os.Getenv("DXREGIA_SMTP_TLS")
	if mode != "" && mode != "implicit" && mode != "starttls" {
		return nil, errors.New("DXREGIA_SMTP_TLS must be implicit or starttls")
	}
	return &contactMailer{
		address: address, host: host, from: from, startTLS: mode == "starttls",
		user: os.Getenv("DXREGIA_SMTP_USER"), password: os.Getenv("DXREGIA_SMTP_PASSWORD"),
	}, nil
}

func (m *contactMailer) notify(config string, request contactRequest) string {
	data, err := os.ReadFile(config)
	var node nodeConfig
	if err != nil || json.Unmarshal(data, &node) != nil || node.Email == "" {
		return "Manca l’email del sysop"
	}
	subject := mime.QEncoding.Encode("utf-8", "DXRegia · nuova richiesta da "+request.Call)
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nReply-To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nTipo: %s\r\nIndicativo: %s\r\nNome: %s\r\nEmail: %s\r\nNodo: %s\r\nHost: %s\r\nPorta: %d\r\n\r\n%s\r\n",
		m.from, node.Email, request.Email, subject,
		request.Kind, request.Call, request.Contact, request.Email, request.Name, request.Host, request.Port, request.Message)
	if err := m.send(node.Email, message); err != nil {
		return "Invio non riuscito"
	}
	return "Notifica inviata al sysop"
}

func (m *contactMailer) send(to, message string) error {
	config := &tls.Config{ServerName: m.host, MinVersion: tls.VersionTLS12}
	dialer := &net.Dialer{Timeout: 8 * time.Second}
	var conn net.Conn
	var err error
	if m.startTLS {
		conn, err = dialer.Dial("tcp", m.address)
	} else {
		conn, err = tls.DialWithDialer(dialer, "tcp", m.address, config)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return err
	}
	defer client.Close()
	if m.startTLS {
		if err := client.StartTLS(config); err != nil {
			return err
		}
	}
	if m.user != "" {
		if err := client.Auth(smtp.PlainAuth("", m.user, m.password, m.host)); err != nil {
			return err
		}
	}
	if err := client.Mail(m.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	body, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := body.Write([]byte(message)); err != nil {
		return err
	}
	return body.Close()
}
