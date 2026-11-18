package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
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
	kind := "Messaggio al sysop"
	if request.Kind == "partner" {
		kind = "Richiesta di collegamento tra nodi"
	}
	rows := []mailRow{{"Tipo", kind, false}, {"Indicativo", request.Call, false}, {"Referente", request.Contact, false},
		{"Email", request.Email, false}, {"Paese", request.Country, false}, {"Lingua del portale", strings.ToUpper(normalizeLang(request.Lang)), false}}
	if request.Kind == "partner" {
		rows = append(rows, mailRow{"Nodo", request.Name, false}, mailRow{"Host e porta", fmt.Sprintf("%s:%d", request.Host, request.Port), false})
	}
	content := mailForSysop("DXRegia · nuova richiesta da "+request.Call, "Nuova richiesta da "+request.Call,
		"Hai ricevuto una richiesta dal portale. Puoi rispondere direttamente a questa email.", rows, request.Message)
	if err := m.send(node.Email, buildMessage(m.from, node.Email, request.Email, content)); err != nil {
		return "Invio non riuscito"
	}
	// Ricevuta per chi ha scritto, nella sua lingua. Un errore qui non invalida la notifica al sysop.
	ack := mailRequestReceived(request.Lang, node.Node, request)
	_ = m.send(request.Email, buildMessage(m.from, request.Email, node.Email, ack))
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
