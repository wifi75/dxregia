package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Riepilogo giornaliero per il sysop: stato dei partner e fatti delle ultime 24 ore (routing, accessi respinti,
// riavvii automatici del motore). Si invia una volta al giorno dopo le 06:00 UTC, solo se il sysop lo ha scelto.

const summaryHourUTC = 6

type summaryState struct {
	Date       string // ultimo giorno inviato, AAAA-MM-GG UTC
	RetryAfter time.Time
}

type summaryCounts struct {
	Routing, Accessi, Riavvii int
}

func summaryDue(state summaryState, now time.Time) bool {
	now = now.UTC()
	return now.Hour() >= summaryHourUTC && state.Date != now.Format("2006-01-02") && !now.Before(state.RetryAfter)
}

// countSummaryEvents conta gli eventi delle ultime 24 ore nel registro di debug e in quello di avvio.
func countSummaryEvents(root string, now time.Time) summaryCounts {
	var counts summaryCounts
	since := now.Add(-24 * time.Hour).Unix()
	base := filepath.Join(root, "local_data")
	for day := 0; day < 2; day++ {
		date := now.UTC().AddDate(0, 0, -day)
		tail := logTail(filepath.Join(base, "debug", fmt.Sprintf("%04d", date.Year()), fmt.Sprintf("%03d.dat", date.YearDay())))
		for _, line := range strings.Split(tail, "\n") {
			if event, ok := parseMojoRouting(strings.TrimSpace(line)); ok && event.At >= since {
				switch event.Kind {
				case "Routing":
					counts.Routing++
				case "Accessi":
					counts.Accessi++
				}
			}
		}
	}
	for _, line := range strings.Split(logTail(filepath.Join(base, "dxregia-startup.log")), "\n") {
		if event, ok := parseMojoStartup(strings.TrimSpace(line)); ok && event.At >= since && strings.Contains(event.Message, "riavvio automatico") &&
			!strings.Contains(event.Message, "sospeso") {
			counts.Riavvii++
		}
	}
	return counts
}

func summaryMail(partners []string, down []string, counts summaryCounts, now time.Time) mailContent {
	downText := "nessuno"
	if len(down) > 0 {
		downText = strings.Join(down, ", ")
	}
	rows := []mailRow{
		{"Nodi partner attivi", fmt.Sprintf("%d su %d", len(partners)-len(down), len(partners)), false},
		{"Non attivi", downText, false},
		{"Spot e pacchetti PC92 scartati (24 ore)", fmt.Sprint(counts.Routing), false},
		{"Accessi respinti (24 ore)", fmt.Sprint(counts.Accessi), false},
		{"Riavvii automatici del motore (24 ore)", fmt.Sprint(counts.Riavvii), false},
	}
	return mailForSysop("DXRegia · riepilogo del "+now.UTC().Format("02/01/2006"), "Riepilogo giornaliero",
		"Lo stato del nodo alle "+now.UTC().Format("15:04")+" UTC e i fatti delle ultime 24 ore. I dettagli sono in Diagnostica > Log DXSpider.", rows, "")
}

func processDailySummary(root string, snapshot presence, now time.Time) {
	prefs, err := readEmailPreferences(root)
	if err != nil || !prefs.DailySummary || !snapshot.PartnerConnect {
		return
	}
	path := filepath.Join(root, "local_data", "dxregia-summary.json")
	var state summaryState
	if data, err := os.ReadFile(path); err == nil {
		if json.Unmarshal(data, &state) != nil {
			return
		}
	} else if !os.IsNotExist(err) {
		return
	}
	if !summaryDue(state, now) {
		return
	}
	mailer := currentMailer(root)
	var node nodeConfig
	if data, err := os.ReadFile(filepath.Join(root, "local", "dxregia-node.json")); err == nil {
		_ = json.Unmarshal(data, &node)
	}
	destination := prefs.destination(node.Email)
	if mailer == nil || destination == "" {
		return
	}
	var saved map[string]partner
	data, err := os.ReadFile(filepath.Join(root, ".dxregia-partners.json"))
	if err != nil || json.Unmarshal(data, &saved) != nil {
		return
	}
	var all, down []string
	for call, entry := range saved {
		if call != entry.Call || !entry.partnerDetails.validate() {
			continue
		}
		all = append(all, call)
		if currentConnection(root, call, snapshot) != "Attivo" {
			down = append(down, call)
		}
	}
	sort.Strings(all)
	sort.Strings(down)
	if mailer.sendContent(destination, summaryMail(all, down, countSummaryEvents(root, now), now)) != nil {
		state.RetryAfter = now.Add(30 * time.Minute)
	} else {
		state = summaryState{Date: now.UTC().Format("2006-01-02")}
	}
	if err := writePrivateJSON(path, state); err != nil {
		log.Printf("Save daily summary state: %v", err)
	}
}
