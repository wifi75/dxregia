package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Eventi di routing e di accesso del motore, per la pagina Log DXSpider. Si leggono solo frasi note del
// registro di debug e si espongono nominativi, nodi e numeri: mai indirizzi di utenti finali né pacchetti.
// L'etichetta di categoria è "Routing" (spot e pacchetti PC92 scartati) o "Accessi" (accessi respinti).

const mojoLinePrefix = `^([0-9]{9,12})\^(?:RING: [0-9:]{8}\^)?(?:\([a-z*]+\) )?`

var (
	mojoBadSpot            = regexp.MustCompile(mojoLinePrefix + `PCPROT: Bad Spot (\S{1,20}) on ([0-9.]{1,10}) by ([A-Z0-9/-]{1,16})\([0-9A-Fa-f:.]{1,45}\)@([A-Z0-9/-]{1,16}) User \S+ not on node \S+, \S+ via ([A-Z0-9/-]{1,16}) [0-9]+ hops$`)
	mojoNonPublic          = regexp.MustCompile(mojoLinePrefix + `PCPROT: (PC[0-9]{2}) ([0-9A-Fa-f:.]{1,45}) is (?:a localhost address or in RFC1918|not a public Internet address), dropped$`)
	mojoMalformed          = regexp.MustCompile(mojoLinePrefix + `PC92_MALFORMED drop peer=([A-Za-z0-9_.-]{1,40}) origin=([A-Za-z0-9_.-]{1,40}) id=\S* type=(\S{1,8}) reason=(.{1,120}?) packet=`)
	mojoAdmission          = regexp.MustCompile(mojoLinePrefix + `PC92_ADMISSION decision=reject type=([A-Za-z0-9_.-]{1,40}) origin=([A-Za-z0-9_.-]{1,40}) peer=([A-Za-z0-9_.-]{1,40}) reason=([A-Za-z0-9_.-]{1,40})$`)
	mojoRollback           = regexp.MustCompile(mojoLinePrefix + `PC92 A/C/D failure: (.{1,160}?); sql_rollback=(\S{1,20}); graph_rollback=(\S{1,20}); user_cache_rollback=(\S{1,20})`)
	mojoRejected           = regexp.MustCompile(mojoLinePrefix + `PCPROT: PC92([A-Z]) rejected (.{1,120})$`)
	mojoTooManyErrors      = regexp.MustCompile(mojoLinePrefix + `DXChannel ([A-Z0-9/-]{1,16}): too many errors \(([0-9]+) > ([0-9]+)\), disconnecting`)
	mojoTooQuick           = regexp.MustCompile(mojoLinePrefix + `([A-Z0-9/-]{1,16}) on ([0-9A-Fa-f:.]{1,45}) nextlogin .*logging in too quickly, banned for ([0-9]+) seconds`)
	mojoBadList            = regexp.MustCompile(mojoLinePrefix + `([A-Z0-9/-]{1,16}) on ([0-9A-Fa-f:.]{1,45}) is (from a badip [0-9A-Fa-f:.]{1,45}|in badspotter|in badnode), disconnected$`)
	mojoTooManyConnections = regexp.MustCompile(mojoLinePrefix + `([A-Z0-9/-]{1,16}) has too many connections \(([0-9]+) > ([0-9]+)\) disconnected`)
	mojoDuplicate          = regexp.MustCompile(mojoLinePrefix + `([A-Z0-9/-]{1,16}) on ([0-9A-Fa-f:.]{1,45}) - duplicate login from ([0-9A-Fa-f:.]{1,45}) rejected; existing channel preserved`)
)

func shortReason(text string) string {
	var out strings.Builder
	for _, r := range text {
		if r >= 0x20 && r < 0x7f {
			out.WriteRune(r)
		}
	}
	reason := strings.TrimSpace(out.String())
	if len(reason) > 100 {
		reason = reason[:100]
	}
	return reason
}

func parseMojoRouting(line string) (mojoLogEvent, bool) {
	at := func(text string) int64 {
		value, _ := strconv.ParseInt(text, 10, 64)
		return value
	}
	if m := mojoBadSpot.FindStringSubmatch(line); m != nil {
		return mojoLogEvent{At: at(m[1]), Kind: "Routing", Call: m[6],
			Message: fmt.Sprintf("Spot di %s su %s kHz scartato: %s non risulta collegato al nodo %s", m[2], m[3], m[4], m[5])}, true
	}
	if m := mojoNonPublic.FindStringSubmatch(line); m != nil {
		return mojoLogEvent{At: at(m[1]), Kind: "Routing", Host: m[3],
			Message: "Spot " + m[2] + " scartato: l’indirizzo di origine non è un indirizzo pubblico"}, true
	}
	if m := mojoMalformed.FindStringSubmatch(line); m != nil {
		return mojoLogEvent{At: at(m[1]), Kind: "Routing", Call: m[2],
			Message: fmt.Sprintf("Pacchetto PC92 %s da %s scartato: %s", m[4], m[3], shortReason(m[5]))}, true
	}
	if m := mojoAdmission.FindStringSubmatch(line); m != nil {
		return mojoLogEvent{At: at(m[1]), Kind: "Routing", Call: m[4],
			Message: fmt.Sprintf("Pacchetto PC92 %s di %s non ammesso: %s", m[2], m[3], m[5])}, true
	}
	if m := mojoRollback.FindStringSubmatch(line); m != nil {
		return mojoLogEvent{At: at(m[1]), Kind: "Routing",
			Message: fmt.Sprintf("Modifica delle rotte annullata (SQL %s, rotte %s, utenti %s): %s", m[3], m[4], m[5], shortReason(m[2]))}, true
	}
	if m := mojoRejected.FindStringSubmatch(line); m != nil {
		return mojoLogEvent{At: at(m[1]), Kind: "Routing", Message: "PC92" + m[2] + " rifiutato: " + shortReason(m[3])}, true
	}
	if m := mojoTooManyErrors.FindStringSubmatch(line); m != nil {
		return mojoLogEvent{At: at(m[1]), Kind: "Accessi", Call: m[2],
			Message: fmt.Sprintf("Collegamento chiuso per troppi errori (%s su un massimo di %s)", m[3], m[4])}, true
	}
	if m := mojoTooQuick.FindStringSubmatch(line); m != nil {
		return mojoLogEvent{At: at(m[1]), Kind: "Accessi", Call: m[2], Host: m[3], Direction: "Ingresso",
			Message: "Accesso troppo ravvicinato: bloccato per " + m[4] + " secondi"}, true
	}
	if m := mojoBadList.FindStringSubmatch(line); m != nil {
		reason := "indirizzo in badip"
		switch {
		case m[4] == "in badspotter":
			reason = "nominativo in badspotter"
		case m[4] == "in badnode":
			reason = "nominativo in badnode"
		}
		return mojoLogEvent{At: at(m[1]), Kind: "Accessi", Call: m[2], Host: m[3], Direction: "Ingresso",
			Message: "Accesso respinto: " + reason}, true
	}
	if m := mojoTooManyConnections.FindStringSubmatch(line); m != nil {
		return mojoLogEvent{At: at(m[1]), Kind: "Accessi", Call: m[2],
			Message: fmt.Sprintf("Accesso respinto: troppe connessioni (%s su un massimo di %s)", m[3], m[4])}, true
	}
	if m := mojoDuplicate.FindStringSubmatch(line); m != nil {
		return mojoLogEvent{At: at(m[1]), Kind: "Accessi", Call: m[2], Host: m[4], Direction: "Ingresso",
			Message: "Secondo accesso respinto: la sessione esistente è stata conservata"}, true
	}
	return mojoLogEvent{}, false
}
