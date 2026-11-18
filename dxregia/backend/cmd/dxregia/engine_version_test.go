package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEngineVersion(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "perl"), 0o755); err != nil {
		t.Fatal(err)
	}
	text := "package main;\n$version = '1.57';\n$build = '46';\n$gitversion = '855513b[i]';\n1;\n"
	if err := os.WriteFile(filepath.Join(root, "perl", "Version.pm"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	info := readEngineVersion(root)
	if info.Version != "1.57" || info.Build != "46" {
		t.Fatalf("got %+v", info)
	}
	if empty := readEngineVersion(t.TempDir()); empty.Version != "" {
		t.Fatalf("missing file should give no version: %+v", empty)
	}
}
