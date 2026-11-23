package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// L'hook incorporato deve portare il marcatore: senza, le versioni future non si installerebbero sui nodi già in
// funzione e nessun errore lo segnalerebbe (successo tra 0.12.0 e 0.15.2).
func TestEmbeddedHookCarriesTheManagedMarker(t *testing.T) {
	if !bytes.Contains(presenceHook, []byte("\n"+presenceHookMarker)) {
		t.Fatal("presence.pm manca del marcatore " + presenceHookMarker + " sulla seconda riga")
	}
}

func TestInstallPresenceHook(t *testing.T) {
	read := func(path string) []byte {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}

	t.Run("fresh install writes the hook", func(t *testing.T) {
		local := t.TempDir()
		if err := installPresenceHook(local); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(read(filepath.Join(local, "Local.pm")), presenceHook) {
			t.Fatal("the installed hook differs from the embedded one")
		}
	})

	t.Run("an older managed hook is replaced", func(t *testing.T) {
		local := t.TempDir()
		old := []byte("package Local;\n" + presenceHookMarker + ": versione precedente\nsub init {}\n1;\n")
		if err := os.WriteFile(filepath.Join(local, "Local.pm"), old, 0600); err != nil {
			t.Fatal(err)
		}
		if err := installPresenceHook(local); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(read(filepath.Join(local, "Local.pm")), presenceHook) {
			t.Fatal("the older managed hook was not upgraded")
		}
	})

	t.Run("a file the sysop wrote is left alone", func(t *testing.T) {
		local := t.TempDir()
		own := []byte("package Local;\nsub init { print \"mio\"; }\n1;\n")
		if err := os.WriteFile(filepath.Join(local, "Local.pm"), own, 0600); err != nil {
			t.Fatal(err)
		}
		if err := installPresenceHook(local); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(read(filepath.Join(local, "Local.pm")), own) {
			t.Fatal("a custom Local.pm was overwritten")
		}
	})
}
