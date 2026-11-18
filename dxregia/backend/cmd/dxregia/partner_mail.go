package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"time"
)

type partnerMail struct {
	Call, Recipient, Password, LocalNode, Host string
	Port                                       int
	VerifyAfter                                time.Time
	Status                                     string
}

func preparePartnerMail(root string, entry partner) (partnerMail, error) {
	var job partnerMail
	data, err := os.ReadFile(filepath.Join(root, "local", "dxregia-node.json"))
	if err != nil {
		return job, err
	}
	var node nodeConfig
	if json.Unmarshal(data, &node) != nil || node.validate() != nil || node.AdvertisedHost == "" || entry.Email == "" {
		return job, fmt.Errorf("missing public endpoint or partner email")
	}
	return partnerMail{Call: entry.Call, Recipient: entry.Email, LocalNode: node.Node,
		Host: node.AdvertisedHost, Port: node.Port, Status: "In attesa dell’abilitazione"}, nil
}

func writePartnerMail(path string, job partnerMail) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".partner-mail-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	err = json.NewEncoder(tmp).Encode(job)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func processPartnerMail(root string, mailer *contactMailer, snapshot presence) {
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	files, err := filepath.Glob(filepath.Join(dir, "*.mail.json"))
	if err != nil {
		return
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		var job partnerMail
		if err != nil || json.Unmarshal(data, &job) != nil {
			continue
		}
		if advancePartnerMail(dir, mailer.sendPartner, snapshot, &job) {
			if err := writePartnerMail(path, job); err != nil {
				return
			}
			return // At most one delivery per worker tick.
		}
	}
}

func advancePartnerMail(dir string, send func(string, string, string) error, snapshot presence, job *partnerMail) bool {
	if job.Password != "" {
		result, err := os.ReadFile(filepath.Join(dir, job.Call+".result"))
		if err != nil {
			return false
		}
		if string(result) != "Abilitato in Mojo" {
			job.Password = ""
			job.Status = "Abilitazione non riuscita. Istruzioni non inviate."
			return true
		}
		text := fmt.Sprintf("Il tuo indicativo %s è stato abilitato sul nodo %s.\r\n\r\nHost: %s\r\nPorta Telnet: %d\r\nProfilo: DXSpider\r\nIndicativo da usare all’accesso: %s\r\nPassword dedicata: %s\r\n\r\nConfigura il tuo nodo e avvia il collegamento. La verifica del link è separata dall’abilitazione.\r\n",
			job.Call, job.LocalNode, job.Host, job.Port, job.Call, job.Password)
		if err := send(job.Recipient, "DXRegia · istruzioni di collegamento", text); err != nil {
			job.Status = "Invio istruzioni non riuscito"
			job.Password = ""
			return true
		}
		job.Password = ""
		job.VerifyAfter = time.Now().Add(5 * time.Minute)
		job.Status = "Istruzioni inviate. Verifica programmata."
		return true
	}
	if job.VerifyAfter.IsZero() || time.Now().Before(job.VerifyAfter) {
		return false
	}

	text := "Mojo non ha confermato un collegamento operativo con il tuo indicativo. Controlla host, porta, password e che il nodo remoto sia stato impostato come DXSpider. Se il problema persiste, contatta il sysop con il messaggio di errore."
	if snapshot.At == 0 {
		text = "Il sysop non ha ancora potuto verificare il collegamento: Mojo non ha fornito dati recenti. Il risultato del tuo collegamento rimane da verificare."
	}
	for _, channel := range snapshot.Channels {
		if channel.Call == job.Call && channel.Kind == "node" && channel.State == "normal" {
			text = "Mojo ha confermato che il collegamento tra i nodi è operativo."
			break
		}
	}
	if err := send(job.Recipient, "DXRegia · esito verifica del collegamento", text+"\r\n"); err != nil {
		job.Status = "Invio esito non riuscito"
	} else {
		job.Status = "Esito inviato"
	}
	job.VerifyAfter = time.Time{}
	return true
}

func (m *contactMailer) sendPartner(to, subject, text string) error {
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		m.from, to, mime.QEncoding.Encode("utf-8", subject), text)
	return m.send(to, message)
}
