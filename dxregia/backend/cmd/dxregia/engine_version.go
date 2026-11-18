package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// Versione di DXSpider in uso, letta da perl/Version.pm del motore: la mostra il piè di pagina del portale.

var (
	engineVersionLine = regexp.MustCompile(`(?m)^\$version\s*=\s*'([0-9][0-9.]{0,14})'`)
	engineBuildLine   = regexp.MustCompile(`(?m)^\$build\s*=\s*'([0-9]{1,6})'`)
)

type engineVersion struct {
	Name    string
	Version string
	Build   string `json:",omitempty"`
}

func readEngineVersion(root string) engineVersion {
	info := engineVersion{Name: "DXSpider"}
	data, err := os.ReadFile(filepath.Join(root, "perl", "Version.pm"))
	if err != nil {
		return info
	}
	if m := engineVersionLine.FindSubmatch(data); m != nil {
		info.Version = string(m[1])
	}
	if m := engineBuildLine.FindSubmatch(data); m != nil {
		info.Build = string(m[1])
	}
	return info
}

func registerEngineVersion(mux *http.ServeMux, root string) {
	if root == "" {
		return
	}
	info := readEngineVersion(root)
	mux.HandleFunc("GET /api/engine", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		json.NewEncoder(w).Encode(info)
	})
}
