package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type inboundDNS struct {
	Host      string
	Addresses []string
	Expires   int64
}

func publicPartnerIP(ip netip.Addr) bool {
	return ip.Zone() == "" && ip.IsGlobalUnicast() && !ip.IsPrivate()
}

func inboundDNSPath(rulePath string) string {
	return strings.TrimSuffix(rulePath, ".inbound-ip") + ".inbound-dns"
}

func resolveInboundDNS(ctx context.Context, host string, lookup func(context.Context, string, string) ([]netip.Addr, error)) (inboundDNS, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	ips, err := lookup(ctx, "ip", host+".")
	if err != nil {
		return inboundDNS{}, err
	}
	resolved := inboundDNS{Host: host, Expires: time.Now().Add(5 * time.Minute).Unix()}
	seen := make(map[netip.Addr]bool)
	for _, ip := range ips {
		ip = ip.Unmap()
		if publicPartnerIP(ip) && !seen[ip] {
			seen[ip] = true
			resolved.Addresses = append(resolved.Addresses, ip.String())
			if len(resolved.Addresses) == 16 {
				break
			}
		}
	}
	if len(resolved.Addresses) == 0 {
		return inboundDNS{}, errors.New("hostname has no public IP")
	}
	return resolved, nil
}

func refreshInboundDNS(ctx context.Context, root string) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			paths, _ := filepath.Glob(filepath.Join(root, "local_data", "dxregia-partners", "*.inbound-ip"))
			for _, path := range paths {
				var host string
				data, err := os.ReadFile(path)
				if err != nil || json.Unmarshal(data, &host) != nil || !validInboundDNSHost(host) {
					continue
				}
				resolved, err := resolveInboundDNS(ctx, host, net.DefaultResolver.LookupNetIP)
				if err != nil {
					continue
				}
				current, err := os.ReadFile(path)
				if err != nil || string(current) != string(data) {
					continue
				}
				if err := writePrivateJSON(inboundDNSPath(path), resolved); err != nil {
					log.Printf("Refresh partner inbound DNS: %v", err)
				}
			}
		}
	}
}

func validInboundDNSHost(host string) bool {
	return len(host) <= 253 && strings.Contains(host, ".") &&
		net.ParseIP(host) == nil && partnerHost.MatchString(host)
}
