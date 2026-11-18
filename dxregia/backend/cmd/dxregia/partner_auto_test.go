package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func autoRoot(t *testing.T) string {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "local"), 0700); err != nil {
		t.Fatal(err)
	}
	node := `{"node":"IU3WJQ-7","sysop":"IU3WJQ","locator":"JN55TE","port":8000,"advertisedHost":"cluster.example.org","email":"sysop@example.org"}`
	if err := os.WriteFile(filepath.Join(root, "local", "dxregia-node.json"), []byte(node), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

func autoDetails(call string) partnerDetails {
	return partnerDetails{Call: call, Name: "node", Contact: "Anna", Email: "anna@example.org", Host: "node.example.org", Port: 7300, Lang: "fr"}
}

func TestAutoPartnerCreatesCardEnablesAccessAndQueuesMail(t *testing.T) {
	root := autoRoot(t)
	path := filepath.Join(root, ".dxregia-partners.json")
	partners := map[string]partner{}
	if err := queueAutoPartner(root, autoDetails("F4ABC-2")); err != nil {
		t.Fatal(err)
	}
	processAutoPartners(root, path, partners, true, time.Now())
	entry, ok := partners["F4ABC-2"]
	if !ok || entry.Provisioning != "In attesa di Mojo" || entry.Lang != "fr" || len(entry.Notes) != 1 {
		t.Fatalf("partner not registered: %+v", entry)
	}
	dir := filepath.Join(root, "local_data", "dxregia-partners")
	for _, name := range []string{"F4ABC-2.enable", "F4ABC-2.mail.json", "F4ABC-2.inbound"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("missing %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "F4ABC-2.auto.json")); err == nil {
		t.Error("the request must be consumed")
	}
}

func TestAutoPartnerLeavesExistingDisabledAndLimitedRequestsToTheSysop(t *testing.T) {
	root := autoRoot(t)
	path := filepath.Join(root, ".dxregia-partners.json")
	partners := map[string]partner{"F4ABC-2": {partnerDetails: autoDetails("F4ABC-2")}}
	_ = queueAutoPartner(root, autoDetails("F4ABC-2"))
	processAutoPartners(root, path, partners, true, time.Now())
	if len(partners) != 1 || len(partners["F4ABC-2"].Notes) != 0 {
		t.Fatal("an existing node must never be touched")
	}

	_ = queueAutoPartner(root, autoDetails("G1AAA"))
	processAutoPartners(root, path, partners, false, time.Now())
	if _, ok := partners["G1AAA"]; ok {
		t.Fatal("without working mail there is nobody to send the password to")
	}

	if err := writeAutoPartner(root, autoPartnerSettings{Enabled: false, DailyLimit: 5}); err != nil {
		t.Fatal(err)
	}
	_ = queueAutoPartner(root, autoDetails("G1AAA"))
	processAutoPartners(root, path, partners, true, time.Now())
	if _, ok := partners["G1AAA"]; ok {
		t.Fatal("disabled setting ignored")
	}

	if err := writeAutoPartner(root, autoPartnerSettings{Enabled: true, DailyLimit: 1}); err != nil {
		t.Fatal(err)
	}
	_ = queueAutoPartner(root, autoDetails("G1AAA"))
	_ = queueAutoPartner(root, autoDetails("G2BBB"))
	processAutoPartners(root, path, partners, true, time.Now())
	auto := 0
	for _, entry := range partners {
		if len(entry.Notes) == 1 && entry.Notes[0].Text == autoPartnerNote {
			auto++
		}
	}
	if auto != 1 {
		t.Fatalf("daily limit not enforced: %d automatic partners", auto)
	}
}
