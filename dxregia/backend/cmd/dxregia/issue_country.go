package main

import (
	"bytes"
	"context"
	"net"
	"os"
	"path/filepath"
	"sort"
	"time"

	maxminddb "github.com/oschwald/maxminddb-golang"
)

type issueCountryRange [34]byte

var issueCountryLanguage = map[string]string{
	"IT": "it", "SM": "it", "VA": "it", "FR": "fr", "MC": "fr", "DE": "de", "AT": "de", "LI": "de",
	"ES": "es", "AR": "es", "BO": "es", "CL": "es", "CO": "es", "CR": "es", "CU": "es",
	"DO": "es", "EC": "es", "GT": "es", "HN": "es", "MX": "es", "NI": "es", "PA": "es",
	"PE": "es", "PR": "es", "PY": "es", "SV": "es", "UY": "es", "VE": "es",
	"PT": "pt", "BR": "pt", "AO": "pt", "CV": "pt", "MZ": "pt",
}

func issueRange(network *net.IPNet, lang string) issueCountryRange {
	var record issueCountryRange
	copy(record[:16], network.IP.To16())
	copy(record[16:32], record[:16])
	offset := 16 - len(network.Mask)
	for i, mask := range network.Mask {
		record[16+offset+i] |= ^mask
	}
	copy(record[32:], lang)
	return record
}

func writeIssueCountryIndex(path, database string) error {
	db, err := maxminddb.Open(database)
	if err != nil {
		return err
	}
	defer db.Close()

	var ranges []issueCountryRange
	networks := db.Networks(maxminddb.SkipAliasedNetworks)
	for networks.Next() {
		var record struct {
			Country struct {
				Code string `maxminddb:"iso_code"`
			} `maxminddb:"country"`
		}
		network, err := networks.Network(&record)
		if err != nil {
			return err
		}
		if lang := issueCountryLanguage[record.Country.Code]; lang != "" {
			ranges = append(ranges, issueRange(network, lang))
		}
	}
	if err := networks.Err(); err != nil {
		return err
	}
	sort.Slice(ranges, func(i, j int) bool { return bytes.Compare(ranges[i][:16], ranges[j][:16]) < 0 })
	data := make([]byte, 0, len(ranges)*len(issueCountryRange{}))
	for _, record := range ranges {
		data = append(data, record[:]...)
	}
	return writeNodeFile(path, data)
}

func refreshIssueCountries(ctx context.Context, root string) {
	if root == "" {
		return
	}
	database := filepath.Join(root, "dbip-country.mmdb")
	index := filepath.Join(root, "local_data", "dxregia-issue-countries.bin")
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for ctx.Err() == nil {
		source, err := os.Stat(database)
		if err == nil {
			if current, err := os.Stat(index); err != nil || current.ModTime().Before(source.ModTime()) {
				_ = writeIssueCountryIndex(index, database)
			}
		} else if os.IsNotExist(err) {
			_ = os.Remove(index)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
