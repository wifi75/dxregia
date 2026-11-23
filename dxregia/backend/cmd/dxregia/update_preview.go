package main

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// Anteprima dell'aggiornamento: cosa cambia tra la versione installata e quella disponibile e se serve il riavvio.
// Si ricava dal confronto dei due commit su GitHub; stessa regola di update-native.sh per l'aggiornamento senza riavvio.

const (
	previewMaxChanges = 30
	previewMaxLine    = 160
)

var docsPath = regexp.MustCompile(`^(docs/|[^/]+\.md$|dxregia/deploy/README\.md$)`)

var lightPath =regexp.MustCompile(`^(dxregia/web/|docs/|[^/]+\.md$|dxregia/deploy/README\.md$)`)

type updatePreview struct {
	Changes []string // oggetto dei commit, dal più recente
	Restart bool     // l'aggiornamento riavvia il servizio
	Files   int
	// DocsOnly: cambiano solo documentazione e file di testo del progetto, nessun file del pannello né del servizio.
	DocsOnly bool
}

func parseUpdatePreview(data []byte) (updatePreview, bool) {
	var compare struct {
		Commits []struct {
			Commit struct {
				Message string `json:"message"`
			} `json:"commit"`
		} `json:"commits"`
		Files []struct {
			Filename string `json:"filename"`
		} `json:"files"`
	}
	if json.Unmarshal(data, &compare) != nil || len(compare.Commits) == 0 {
		return updatePreview{}, false
	}
	var preview updatePreview
	for i := len(compare.Commits) - 1; i >= 0 && len(preview.Changes) < previewMaxChanges; i-- {
		line, _, _ := strings.Cut(strings.TrimSpace(compare.Commits[i].Commit.Message), "\n")
		line = strings.TrimSpace(line)
		if len(line) > previewMaxLine {
			line = line[:previewMaxLine]
		}
		if line != "" {
			preview.Changes = append(preview.Changes, line)
		}
	}
	preview.Files = len(compare.Files)
	preview.Restart = len(compare.Files) == 0 || len(compare.Files) >= 300 // 300 = limite dell'API: nel dubbio, completo
	preview.DocsOnly = len(compare.Files) > 0 && len(compare.Files) < 300
	for _, file := range compare.Files {
		if !lightPath.MatchString(file.Filename) {
			preview.Restart = true
		}
		if !docsPath.MatchString(file.Filename) {
			preview.DocsOnly = false
		}
	}
	return preview, true
}

func githubPreview(ctx context.Context, base, head string) (updatePreview, bool) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	data, err := githubGet(ctx, "/compare/"+base+"..."+head, "application/vnd.github+json")
	if err != nil {
		return updatePreview{}, false
	}
	return parseUpdatePreview(data)
}
