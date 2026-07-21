//go:build !ecs_public

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	securitynetwork "github.com/oneclickvirt/security/network/security"
)

type securityAddressPayload struct {
	IP        string                             `json:"ip"`
	IPType    string                             `json:"ip_type"`
	Providers []securitynetwork.ProviderEvidence `json:"providers,omitempty"`
	DNSBL     *securitynetwork.DNSBLReport       `json:"dnsbl,omitempty"`
}

type securityComponentPayload struct {
	SchemaVersion string                   `json:"schema_version"`
	Addresses     []securityAddressPayload `json:"addresses"`
	Error         string                   `json:"error,omitempty"`
}

type securityProviderFactory func(string) []securitynetwork.ProviderProbe

func loadSecurityComponentData(ctx context.Context, offline bool) componentDataResult {
	var zones []securitynetwork.DNSBLZone
	var source securitynetwork.DNSBLDataSource
	var err error
	if offline {
		zones, source, err = securitynetwork.EmbeddedDNSBLZoneRegistrySnapshot()
	} else {
		zones, source, err = securitynetwork.LoadDNSBLZoneRegistry(ctx, nil)
	}
	if err != nil {
		return failedComponentData(ctx, dnsblDataFile, err)
	}
	file := timeMetadataFile(dnsblDataFile, source.Schema, source.GeneratedAt, source.Source, source.Fallback, source.Count)
	inputs := make([]dnsblZoneInput, 0, len(zones))
	for _, zone := range zones {
		inputs = append(inputs, dnsblZoneInput{Zone: zone.Zone, IPv4: zone.IPv4, IPv6: zone.IPv6})
	}
	return componentDataResult{file: file, apply: func(component *componentInputs) { component.DNSBLZones = inputs }}
}

func collectSecurityComponent(ctx context.Context, ipv4, ipv6 string, zones []dnsblZoneInput) ComponentReport {
	converted := make([]securitynetwork.DNSBLZone, 0, len(zones))
	for _, zone := range zones {
		converted = append(converted, securitynetwork.DNSBLZone{Zone: zone.Zone, IPv4: zone.IPv4, IPv6: zone.IPv6})
	}
	return collectSecurityComponentWithTypedDeps(ctx, ipv4, ipv6, converted, securitynetwork.DefaultProviderProbes, nil)
}

func collectSecurityComponentWithDeps(ctx context.Context, ipv4, ipv6 string, zonesData []byte, providerFactory securityProviderFactory, resolver securitynetwork.DNSBLResolver) ComponentReport {
	var zones []securitynetwork.DNSBLZone
	if len(zonesData) > 0 {
		if err := json.Unmarshal(zonesData, &zones); err != nil {
			return componentPayload("security.evidence", "goecs.security/v1", ReportStatusError, time.Now(), nil, fmt.Errorf("decode DNSBL zones: %w", err))
		}
	}
	return collectSecurityComponentWithTypedDeps(ctx, ipv4, ipv6, zones, providerFactory, resolver)
}

func collectSecurityComponentWithTypedDeps(ctx context.Context, ipv4, ipv6 string, zones []securitynetwork.DNSBLZone, providerFactory securityProviderFactory, resolver securitynetwork.DNSBLResolver) ComponentReport {
	started := time.Now()
	if ctx == nil {
		ctx = context.Background()
	}
	payload := securityComponentPayload{SchemaVersion: "goecs.security/v1"}
	if providerFactory == nil {
		providerFactory = securitynetwork.DefaultProviderProbes
	}
	for _, ip := range []string{strings.TrimSpace(ipv4), strings.TrimSpace(ipv6)} {
		parsed := net.ParseIP(ip)
		if parsed == nil {
			continue
		}
		address := securityAddressPayload{IP: ip, IPType: "ipv6"}
		if parsed.To4() != nil {
			address.IPType = "ipv4"
		}
		var wg sync.WaitGroup
		wg.Add(2)
		var providers []securitynetwork.ProviderEvidence
		var dnsbl *securitynetwork.DNSBLReport
		go func() {
			defer wg.Done()
			providers = securitynetwork.CollectProviderEvidence(ctx, ip, providerFactory(address.IPType), securitynetwork.EvidenceOptions{Timeout: 5 * time.Second, Concurrency: 8})
		}()
		go func() {
			defer wg.Done()
			if len(zones) > 0 {
				report := securitynetwork.CheckDNSBLZones(ctx, ip, zones, resolver)
				dnsbl = &report
			}
		}()
		wg.Wait()
		address.Providers = providers
		address.DNSBL = dnsbl
		payload.Addresses = append(payload.Addresses, address)
	}
	status := securityComponentStatus(ctx, payload)
	report := componentPayload("security.evidence", payload.SchemaVersion, status, started, payload, nil)
	if report.Status != ReportStatusOK {
		report.Reason = securityComponentReason(payload)
	}
	return report
}

func securityComponentStatus(ctx context.Context, payload securityComponentPayload) ReportStatus {
	if ctx != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ReportStatusTimeout
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return ReportStatusCanceled
		}
	}
	if len(payload.Addresses) == 0 {
		return ReportStatusUnavailable
	}
	valid, degraded := 0, 0
	for _, address := range payload.Addresses {
		for _, evidence := range address.Providers {
			switch evidence.Status {
			case securitynetwork.ProviderAvailable:
				valid++
			case securitynetwork.ProviderMissingFields:
				valid++
				degraded++
			default:
				degraded++
			}
		}
		if address.DNSBL != nil {
			for state, count := range address.DNSBL.Counts {
				if count == 0 {
					continue
				}
				switch state {
				case securitynetwork.DNSBLClean, securitynetwork.DNSBLListed, securitynetwork.DNSBLMarked:
					valid += count
				default:
					degraded += count
				}
			}
		}
	}
	if valid == 0 {
		return ReportStatusUnavailable
	}
	if degraded > 0 {
		return ReportStatusPartial
	}
	return ReportStatusOK
}

func securityComponentReason(payload securityComponentPayload) string {
	if len(payload.Addresses) == 0 {
		return "no valid public IP address"
	}
	return "one or more security sources unavailable or degraded"
}
