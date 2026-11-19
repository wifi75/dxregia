package main

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"html"
	"mime"
	"strings"
)

// mailContent è un messaggio pronto: oggetto, versione testo e versione HTML.
type mailContent struct {
	Subject, Text, HTML string
}

// mailRow è una riga etichetta/valore nel riquadro dei dati.
type mailRow struct {
	Label, Value string
	Secret       bool
}

type mailBlocks struct {
	Preheader, Title, Hello, Intro  string
	DataTitle                       string
	Rows                            []mailRow
	StepsTitle                      string
	Steps                           []string
	ScriptTitle, Script, ScriptNote string
	Checks                          []string
	AfterTitle, After               string
	Notice                          string
	ButtonLabel, ButtonURL          string
	Tone                            string // "info" (azzurro), "ok" (verde), "warn" (arancio)
	Footer, Auto, Lang              string
}

const (
	colInk         = "#143a60"
	colText        = "#26384a"
	colMuted       = "#5b6e80"
	colLine        = "#d9e3ec"
	mailProjectURL = "https://github.com/wifi75/dxregia"
	mailPortalURL  = "https://dxregia.iu3wjq.com/"
	mailAuthor     = "Tiziano Cassone · IU3WJQ"
	mailIconID     = "dxregia-icon@iu3wjq.com"
)

//go:embed mail-icon.png
var mailIcon []byte

func toneColors(tone string) (accent, soft string) {
	switch tone {
	case "ok":
		return "#167b50", "#e6f5ee"
	case "warn":
		return "#b8650d", "#fdf0e0"
	}
	return "#2f78b3", "#e7f1fa"
}

func esc(s string) string { return html.EscapeString(s) }

// renderMail costruisce la versione HTML (stile inline e tabelle, per compatibilità con i client di posta)
// e la versione testo dello stesso messaggio.
func renderMail(subject string, b mailBlocks) mailContent {
	accent, soft := toneColors(b.Tone)
	var h, t strings.Builder
	lang := normalizeLang(b.Lang)
	if lang == "" {
		lang = "it"
	}
	fmt.Fprintf(&h, `<!DOCTYPE html><html lang="%s"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="color-scheme" content="light"><title>%s</title></head>`, lang, esc(subject))
	fmt.Fprintf(&h, `<body style="margin:0;padding:0;background:#eef3f8;font-family:-apple-system,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:%s;">`, colText)
	fmt.Fprintf(&h, `<div style="display:none;max-height:0;overflow:hidden;opacity:0;color:transparent;">%s</div>`, esc(b.Preheader))
	h.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#eef3f8;"><tr><td align="center" style="padding:24px 12px;">`)
	h.WriteString(`<table role="presentation" width="600" cellpadding="0" cellspacing="0" style="width:100%;max-width:600px;background:#ffffff;border-radius:12px;overflow:hidden;border:1px solid ` + colLine + `;">`)
	// Intestazione
	fmt.Fprintf(&h, `<tr><td style="background:%s;padding:22px 28px;"><table role="presentation" cellpadding="0" cellspacing="0"><tr><td style="padding-right:12px;"><img src="cid:%s" width="40" height="40" alt="" style="display:block;border:0;border-radius:8px;"></td><td><span style="font-size:22px;font-weight:800;letter-spacing:.3px;color:#ffffff;">DX<span style="color:#7fd1ae;">Regia</span></span><br><span style="font-size:12px;color:#b9cfe3;">%s</span></td></tr></table></td></tr>`, colInk, mailIconID, esc(b.Footer))
	fmt.Fprintf(&h, `<tr><td style="height:5px;background:%s;font-size:0;line-height:0;">&nbsp;</td></tr>`, accent)
	// Titolo e introduzione
	fmt.Fprintf(&h, `<tr><td style="padding:30px 28px 6px;"><h1 style="margin:0 0 14px;font-size:24px;line-height:1.25;color:%s;">%s</h1>`, colInk, esc(b.Title))
	if b.Hello != "" {
		fmt.Fprintf(&h, `<p style="margin:0 0 10px;font-size:16px;line-height:1.55;">%s</p>`, esc(b.Hello))
	}
	if b.Intro != "" {
		fmt.Fprintf(&h, `<p style="margin:0;font-size:16px;line-height:1.55;">%s</p>`, esc(b.Intro))
	}
	h.WriteString(`</td></tr>`)
	fmt.Fprintf(&t, "%s\n%s\n\n", b.Title, strings.Repeat("=", min(len(b.Title), 60)))
	if b.Hello != "" {
		t.WriteString(b.Hello + "\n\n")
	}
	if b.Intro != "" {
		t.WriteString(b.Intro + "\n\n")
	}
	if b.ButtonURL != "" {
		fmt.Fprintf(&h, `<tr><td style="padding:22px 28px 6px;"><a href="%s" style="display:inline-block;padding:14px 28px;background:%s;color:#ffffff;font-size:16px;font-weight:700;text-decoration:none;border-radius:8px;">%s</a><p style="margin:12px 0 0;font-size:12.5px;line-height:1.5;color:%s;word-break:break-all;">%s</p></td></tr>`, esc(b.ButtonURL), accent, esc(b.ButtonLabel), colMuted, esc(b.ButtonURL))
		t.WriteString(b.ButtonLabel + ":\n" + b.ButtonURL + "\n\n")
	}
	// Riquadro dati
	if len(b.Rows) > 0 {
		fmt.Fprintf(&h, `<tr><td style="padding:18px 28px 6px;"><table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:%s;border-radius:10px;border:1px solid %s;"><tr><td style="padding:16px 20px 6px;font-size:13px;font-weight:700;letter-spacing:.6px;text-transform:uppercase;color:%s;">%s</td></tr>`, soft, colLine, accent, esc(b.DataTitle))
		t.WriteString(strings.ToUpper(b.DataTitle) + "\n")
		for _, row := range b.Rows {
			value := `<span style="font-size:16px;font-weight:600;color:` + colInk + `;">` + esc(row.Value) + `</span>`
			if row.Secret {
				value = `<span style="display:inline-block;padding:6px 10px;background:#ffffff;border:1px dashed ` + accent + `;border-radius:6px;font-family:Consolas,'Courier New',monospace;font-size:15px;font-weight:700;color:` + colInk + `;word-break:break-all;">` + esc(row.Value) + `</span>`
			}
			fmt.Fprintf(&h, `<tr><td style="padding:8px 20px;border-top:1px solid %s;"><div style="font-size:12px;color:%s;margin-bottom:3px;">%s</div>%s</td></tr>`, colLine, colMuted, esc(row.Label), value)
			fmt.Fprintf(&t, "  %s: %s\n", row.Label, row.Value)
		}
		h.WriteString(`<tr><td style="height:10px;font-size:0;">&nbsp;</td></tr></table></td></tr>`)
		t.WriteString("\n")
	}
	if b.Notice != "" {
		fmt.Fprintf(&h, `<tr><td style="padding:8px 28px 0;"><p style="margin:0;padding:12px 14px;background:#fdf0e0;border-left:4px solid #b8650d;border-radius:4px;font-size:14px;line-height:1.5;color:#6b3d08;">%s</p></td></tr>`, esc(b.Notice))
		t.WriteString("! " + b.Notice + "\n\n")
	}
	// Passi numerati
	if len(b.Steps) > 0 {
		fmt.Fprintf(&h, `<tr><td style="padding:22px 28px 0;"><h2 style="margin:0 0 12px;font-size:18px;color:%s;">%s</h2><table role="presentation" width="100%%" cellpadding="0" cellspacing="0">`, colInk, esc(b.StepsTitle))
		t.WriteString(b.StepsTitle + "\n")
		for i, step := range b.Steps {
			fmt.Fprintf(&h, `<tr><td width="34" valign="top" style="padding:0 0 12px;"><div style="width:26px;height:26px;border-radius:13px;background:%s;color:#ffffff;font-size:14px;font-weight:700;line-height:26px;text-align:center;">%d</div></td><td valign="top" style="padding:2px 0 12px;font-size:15px;line-height:1.5;">%s</td></tr>`, accent, i+1, esc(step))
			fmt.Fprintf(&t, "  %d. %s\n", i+1, step)
		}
		h.WriteString(`</table></td></tr>`)
		t.WriteString("\n")
	}
	// Elenco di controlli
	if len(b.Checks) > 0 {
		h.WriteString(`<tr><td style="padding:14px 28px 0;"><ul style="margin:0;padding-left:20px;font-size:15px;line-height:1.6;">`)
		for _, check := range b.Checks {
			fmt.Fprintf(&h, `<li style="margin-bottom:6px;">%s</li>`, esc(check))
			t.WriteString("  - " + check + "\n")
		}
		h.WriteString(`</ul></td></tr>`)
		t.WriteString("\n")
	}
	// Script d'esempio
	if b.Script != "" {
		fmt.Fprintf(&h, `<tr><td style="padding:12px 28px 0;"><h2 style="margin:0 0 8px;font-size:18px;color:%s;">%s</h2><pre style="margin:0;padding:14px 16px;background:#0f2438;color:#d6e6f5;border-radius:8px;font-family:Consolas,'Courier New',monospace;font-size:13px;line-height:1.6;white-space:pre-wrap;word-break:break-all;">%s</pre><p style="margin:6px 0 0;font-size:13px;color:%s;">%s</p></td></tr>`, colInk, esc(b.ScriptTitle), esc(b.Script), colMuted, esc(b.ScriptNote))
		t.WriteString(b.ScriptTitle + "\n" + b.Script + "\n" + b.ScriptNote + "\n\n")
	}
	// Cosa succede dopo
	if b.After != "" {
		fmt.Fprintf(&h, `<tr><td style="padding:22px 28px 0;"><table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:%s;border-radius:10px;"><tr><td style="padding:16px 20px;"><div style="font-size:15px;font-weight:700;color:%s;margin-bottom:6px;">%s</div><div style="font-size:15px;line-height:1.55;">%s</div></td></tr></table></td></tr>`, soft, accent, esc(b.AfterTitle), esc(b.After))
		t.WriteString(b.AfterTitle + "\n" + b.After + "\n\n")
	}
	// Piè di pagina
	fmt.Fprintf(&h, `<tr><td style="padding:26px 28px 28px;"><p style="margin:0;padding-top:16px;border-top:1px solid %s;font-size:12.5px;line-height:1.55;color:%s;">%s</p><p style="margin:10px 0 0;font-size:13px;line-height:1.6;color:%s;"><a href="%s" style="color:%s;">DXRegia · GitHub</a><br><a href="%s" style="color:%s;">IU3WJQ-7 · DX Cluster</a><br>%s</p></td></tr>`, colLine, colMuted, esc(b.Auto), colMuted, mailProjectURL, colInk, mailPortalURL, colInk, esc(mailAuthor))
	h.WriteString(`</table></td></tr></table></body></html>`)
	t.WriteString("--\n" + b.Auto + "\n" + b.Footer + "\nDXRegia · GitHub: " + mailProjectURL + "\nIU3WJQ-7 · DX Cluster: " + mailPortalURL + "\n" + mailAuthor + "\n")
	return mailContent{Subject: subject, Text: t.String(), HTML: h.String()}
}

func wrap76(s string) string {
	var out strings.Builder
	for len(s) > 76 {
		out.WriteString(s[:76] + "\r\n")
		s = s[76:]
	}
	return out.String() + s + "\r\n"
}

// buildMessage produce le alternative testo e HTML, con l'icona incorporata nella parte HTML.
func buildMessage(from, to, replyTo string, c mailContent) string {
	boundary := "dxregia-" + fmt.Sprintf("%x", len(c.HTML)*31+len(c.Text))
	var m strings.Builder
	fmt.Fprintf(&m, "From: %s\r\nTo: %s\r\n", from, to)
	if replyTo != "" {
		fmt.Fprintf(&m, "Reply-To: %s\r\n", replyTo)
	}
	fmt.Fprintf(&m, "Subject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", mime.QEncoding.Encode("utf-8", c.Subject), boundary)
	fmt.Fprintf(&m, "--%s\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n%s", boundary, wrap76(base64.StdEncoding.EncodeToString([]byte(c.Text))))
	related := boundary + "-related"
	fmt.Fprintf(&m, "--%s\r\nContent-Type: multipart/related; type=\"text/html\"; boundary=\"%s\"\r\n\r\n", boundary, related)
	fmt.Fprintf(&m, "--%s\r\nContent-Type: text/html; charset=utf-8\r\nContent-Transfer-Encoding: base64\r\n\r\n%s", related, wrap76(base64.StdEncoding.EncodeToString([]byte(c.HTML))))
	fmt.Fprintf(&m, "--%s\r\nContent-Type: image/png; name=\"dxregia.png\"\r\nContent-ID: <%s>\r\nContent-Disposition: inline; filename=\"dxregia.png\"\r\nContent-Transfer-Encoding: base64\r\n\r\n%s", related, mailIconID, wrap76(base64.StdEncoding.EncodeToString(mailIcon)))
	fmt.Fprintf(&m, "--%s--\r\n", related)
	fmt.Fprintf(&m, "--%s--\r\n", boundary)
	return m.String()
}

// ——— Messaggi specifici ———

func mailValues(call, node, name, host string, port int) map[string]string {
	if name == "" {
		name = call
	}
	return map[string]string{"call": call, "node": node, "name": name, "host": host, "port": fmt.Sprint(port)}
}

func mailInstructions(lang string, job partnerMail, contact string) mailContent {
	s := textsFor(lang)
	v := mailValues(job.Call, job.LocalNode, contact, job.Host, job.Port)
	steps := make([]string, 0, len(s.Steps))
	for _, step := range s.Steps {
		steps = append(steps, fill(step, v))
	}
	script := fmt.Sprintf("timeout 15\nconnect telnet %s %d\n'login' '%s'\n'password:' '%s'\nclient %s telnet", job.Host, job.Port, job.Call, job.Password, job.LocalNode)
	subject := fill(s.InstrSubject, v)
	return renderMail(subject, mailBlocks{
		Preheader: fill(s.InstrTitle, v), Title: fill(s.InstrTitle, v), Hello: fill(s.Hello, v), Intro: fill(s.InstrIntro, v),
		DataTitle: s.DataTitle, Tone: "info", Lang: lang,
		Rows: []mailRow{{s.LNode, job.LocalNode, false}, {s.LHost, job.Host, false}, {s.LPort, fmt.Sprint(job.Port), false},
			{s.LCall, job.Call, false}, {s.LPass, job.Password, true}, {s.LProfile, s.VProf, false}},
		Notice: s.SecurityNote, StepsTitle: s.StepsTitle, Steps: steps,
		ScriptTitle: s.ScriptTitle, Script: script, ScriptNote: s.ScriptNote,
		AfterTitle: s.AfterTitle, After: fill(s.AfterText, v), Footer: s.Footer, Auto: fill(s.AutoNote, v),
	})
}

func mailLinkOK(lang string, job partnerMail, contact string) mailContent {
	s := textsFor(lang)
	v := mailValues(job.Call, job.LocalNode, contact, job.Host, job.Port)
	return renderMail(fill(s.OKSubject, v), mailBlocks{
		Preheader: fill(s.OKTitle, v), Title: s.OKTitle, Hello: fill(s.Hello, v), Intro: fill(s.OKText, v),
		Tone: "ok", Lang: lang, AfterTitle: "✓", After: fill(s.OKNext, v), Footer: s.Footer, Auto: fill(s.AutoNote, v),
	})
}

func mailLinkPending(lang string, job partnerMail, contact string) mailContent {
	s := textsFor(lang)
	v := mailValues(job.Call, job.LocalNode, contact, job.Host, job.Port)
	checks := make([]string, 0, len(s.NoChecks))
	for _, check := range s.NoChecks {
		checks = append(checks, fill(check, v))
	}
	return renderMail(fill(s.NoSubject, v), mailBlocks{
		Preheader: fill(s.NoTitle, v), Title: s.NoTitle, Hello: fill(s.Hello, v), Intro: fill(s.NoText, v),
		Tone: "warn", Lang: lang, Checks: checks, Footer: s.Footer, Auto: fill(s.AutoNote, v),
	})
}

func mailRequestReceived(lang, node string, request contactRequest) mailContent {
	s := textsFor(lang)
	v := mailValues(request.Call, node, request.Contact, request.Host, request.Port)
	rows := []mailRow{{"Call", request.Call, false}}
	if request.Name != "" {
		rows = append(rows, mailRow{"Node", request.Name, false})
	}
	if request.Host != "" {
		rows = append(rows, mailRow{"Host", fmt.Sprintf("%s:%d", request.Host, request.Port), false})
	}
	return renderMail(fill(s.RecvSubject, v), mailBlocks{
		Preheader: fill(s.RecvTitle, v), Title: s.RecvTitle, Hello: fill(s.Hello, v), Intro: fill(s.RecvText, v),
		DataTitle: s.YourData, Rows: rows, Tone: "info", Lang: lang,
		AfterTitle: "→", After: fill(s.RecvNext, v), Footer: s.Footer, Auto: fill(s.AutoNote, v),
	})
}

// mailForSysop genera le email destinate al sysop, sempre in italiano.
func mailForSysop(subject, title, intro string, rows []mailRow, note string) mailContent {
	s := textsFor("it")
	return renderMail(subject, mailBlocks{
		Preheader: title, Title: title, Intro: intro, DataTitle: "Dettagli", Rows: rows, After: note,
		AfterTitle: "Messaggio", Tone: "info", Lang: "it", Footer: s.Footer,
		Auto: "Notifica automatica di DXRegia per il sysop.",
	})
}
