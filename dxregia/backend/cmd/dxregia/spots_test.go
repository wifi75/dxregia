package main

import (
	"path/filepath"
	"testing"
)

func TestSpotEntitiesFromBundledMojoData(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	entities := spotEntities(root)
	if len(entities) < 300 {
		t.Fatalf("parsed %d entities from the bundled Mojo data", len(entities))
	}
	for code, want := range map[int]string{
		25:  "Madeira Islands",
		85:  "Italy",
		95:  "Japan",
		226: "United States",
		666: "",
	} {
		if got := entities[code]; got != want {
			t.Errorf("DXCC %d = %q, want %q", code, got, want)
		}
	}
}
