//go:build !ecs_public

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	unlockexecutor "github.com/oneclickvirt/UnlockTests/executor"
	unlockmodel "github.com/oneclickvirt/UnlockTests/model"
	bgptools "github.com/oneclickvirt/backtrace/bgptools"
	basicsmodel "github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/cputest/cpu"
	"github.com/oneclickvirt/disktest/disk"
	"github.com/oneclickvirt/gostun/stuncheck"
	"github.com/oneclickvirt/memorytest/memory"
	nt3model "github.com/oneclickvirt/nt3/model"
	pingprobe "github.com/oneclickvirt/pingtest/pt"
	portemail "github.com/oneclickvirt/portchecker/email"
	privatepst "github.com/oneclickvirt/privatespeedtest/pst"
	securitynetwork "github.com/oneclickvirt/security/network/security"
	speedmodel "github.com/oneclickvirt/speedtest/model"
)

type componentMXResolver struct {
	records map[string][]*net.MX
}

func (resolver componentMXResolver) LookupMX(ctx context.Context, domain string) ([]*net.MX, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return resolver.records[domain], nil
}

type componentMailDialer struct{}

func (componentMailDialer) DialContext(ctx context.Context, _, address string) (net.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.HasSuffix(address, ":465") {
		return nil, errors.New("connection refused")
	}
	client, server := net.Pipe()
	go func() {
		defer server.Close()
		_, _ = server.Write([]byte("220 fixture.example ESMTP ready\r\n"))
	}()
	return client, nil
}

type componentMailListener struct{}

func (componentMailListener) Listen(_, _ string) (net.Listener, error) {
	return componentClosedListener{}, nil
}

type componentClosedListener struct{}

func (componentClosedListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (componentClosedListener) Close() error              { return nil }
func (componentClosedListener) Addr() net.Addr            { return componentFixtureAddr("fixture") }

type componentFixtureAddr string

func (addr componentFixtureAddr) Network() string { return string(addr) }
func (addr componentFixtureAddr) String() string  { return string(addr) }

type componentDNSBLResolver struct{}

func (componentDNSBLResolver) LookupIP(context.Context, string, string) ([]net.IP, error) {
	return nil, &net.DNSError{Err: "fixture not found", IsNotFound: true}
}

func TestLocalComponentAdapterCallsStructuredBasics(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.CpuTestStatus = false
	cfg.MemoryTestStatus = false
	cfg.DiskTestStatus = false
	cfg.TCPProbeStatus = false
	cfg.Nt3Status = false
	components := collectComponentReports(context.Background(), cfg, componentInputs{})
	if len(components) != 1 {
		t.Fatalf("expected one component, got %#v", components)
	}
	if components[0].Name != "basics" || components[0].SchemaVersion != "goecs.system/v1" || len(components[0].Payload) == 0 {
		t.Fatalf("unexpected basics component: %#v", components[0])
	}
}

func TestLocalBuildUsesStructuredComponents(t *testing.T) {
	if !UsesStructuredComponents() {
		t.Fatal("local component build must select bounded structured orchestration")
	}
}

func TestLocalComponentAdapterBuildsProvinceLatencyPayload(t *testing.T) {
	type carrier struct {
		Carrier string `json:"carrier"`
		IPv4    string `json:"ipv4"`
		IPv6    string `json:"ipv6"`
	}
	type province struct {
		Code     string    `json:"code"`
		Name     string    `json:"name"`
		Province int       `json:"province"`
		Short    string    `json:"short"`
		Targets  []carrier `json:"targets"`
	}
	routes := make([]province, 31)
	for index := range routes {
		code := fmt.Sprintf("%c%c", 'A'+rune(index/26), 'A'+rune(index%26))
		prefix := strings.ToLower(code)
		routes[index] = province{Code: code, Name: "Province-" + code, Province: index + 1, Short: code, Targets: []carrier{
			{Carrier: "ct", IPv4: prefix + "-ct.example", IPv6: prefix + "-ct-v6.example"},
			{Carrier: "cu", IPv4: prefix + "-cu.example", IPv6: prefix + "-cu-v6.example"},
			{Carrier: "cm", IPv4: prefix + "-cm.example", IPv6: prefix + "-cm-v6.example"},
		}}
	}
	typedRoutes := make([]nt3model.ProvinceRoute, 0, len(routes))
	for _, route := range routes {
		targets := make([]nt3model.ProvinceCarrierTarget, 0, len(route.Targets))
		for _, target := range route.Targets {
			targets = append(targets, nt3model.ProvinceCarrierTarget{Carrier: target.Carrier, IPv4: target.IPv4, IPv6: target.IPv6})
		}
		typedRoutes = append(typedRoutes, nt3model.ProvinceRoute{Code: route.Code, Name: route.Name, Province: route.Province, Short: route.Short, Targets: targets})
	}
	cfg := NewDefaultConfig()
	cfg.BasicStatus = false
	cfg.CpuTestStatus = false
	cfg.MemoryTestStatus = false
	cfg.DiskTestStatus = false
	cfg.TCPProbeStatus = false
	cfg.Nt3Status = true
	cfg.EmailTestStatus = false
	cfg.Nt3CheckType = "ipv4"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	components := collectComponentReports(ctx, cfg, componentInputs{ProvinceRoutes: typedRoutes, Network: true})
	var provinceComponent *ComponentReport
	for index := range components {
		if components[index].Name == "nt3.province_latency" {
			provinceComponent = &components[index]
			break
		}
	}
	if provinceComponent == nil || provinceComponent.Status != ReportStatusCanceled {
		t.Fatalf("province component missing or not canceled: component=%#v count=%d", provinceComponent, len(components))
	}
	var payload []json.RawMessage
	if err := json.Unmarshal(provinceComponent.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload) != 31*3 {
		t.Fatalf("got %d province targets, want %d", len(payload), 31*3)
	}
}

func TestLocalMailComponentUsesStructuredFixture(t *testing.T) {
	report := collectMailComponent(context.Background(), []portemail.PlatformSpec{{
		Name: "Example", Domain: "example.test", SMTPHost: "smtp.example.test",
	}}, componentMXResolver{records: map[string][]*net.MX{
		"example.test": {{Host: "mx20.example.test.", Pref: 20}, {Host: "mx10.example.test.", Pref: 10}},
	}}, componentMailDialer{}, componentMailListener{})

	if report.Name != "portchecker.email" || report.SchemaVersion != "goecs.portchecker/mail-v1" || report.Status != ReportStatusPartial {
		t.Fatalf("unexpected mail component envelope: %#v", report)
	}
	var payload portemail.MailReport
	if err := json.Unmarshal(report.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Local) != 6 || len(payload.MX) != 2 || payload.MX[0].Preference != 10 || payload.MX[0].Kind != portemail.KindMXSMTP25 {
		t.Fatalf("unexpected structured mail payload: %+v", payload)
	}
}

func TestLocalMailAndSTUNComponentsRespectCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	mail := collectMailComponent(ctx, []portemail.PlatformSpec{{
		Name: "Example", Domain: "example.test", SMTPHost: "smtp.example.test",
	}}, componentMXResolver{}, componentMailDialer{}, componentMailListener{})
	if mail.Status != ReportStatusCanceled {
		t.Fatalf("mail status = %q, want canceled", mail.Status)
	}

	probeCalled := false
	stunConfig := stuncheck.ProbeConfig{Servers: []string{"stun.fixture.test:3478"}, IPVersion: "ipv4"}
	stun := collectSTUNComponent(ctx, stunConfig, func(ctx context.Context, _ stuncheck.ProbeConfig) stuncheck.NATSummary {
		probeCalled = true
		return stuncheck.NATSummary{Status: stuncheck.CapabilityError, Error: ctx.Err().Error()}
	})
	if stun.Status != ReportStatusCanceled || probeCalled {
		t.Fatalf("STUN status = %q, probeCalled = %t; want canceled without probing", stun.Status, probeCalled)
	}
}

func TestLocalSTUNComponentBuildsStructuredFixture(t *testing.T) {
	config := stuncheck.ProbeConfig{Servers: []string{"stun.fixture.test:3478"}, IPVersion: "ipv4"}
	report := collectSTUNComponent(context.Background(), config, func(context.Context, stuncheck.ProbeConfig) stuncheck.NATSummary {
		return stuncheck.NATSummary{
			SchemaVersion: "goecs.stun/v1", IPVersion: "ipv4", Status: stuncheck.CapabilityAvailable,
			Successful: 1, MappingConsistency: stuncheck.CapabilityAvailable,
			PortPreservationConsistency: stuncheck.CapabilityAvailable, HairpinConsistency: stuncheck.CapabilityAvailable,
			Results: []stuncheck.NATReport{{
				SchemaVersion: "goecs.stun/v1", IPVersion: "ipv4", Server: "stun.fixture.test:3478",
				Status: stuncheck.CapabilityAvailable, LocalAddress: "192.0.2.10:32000", MappedAddress: "198.51.100.20:32000",
				PortPreservation: stuncheck.CapabilityAvailable, Hairpin: stuncheck.CapabilityAvailable,
			}},
		}
	})
	if report.Name != "gostun.nat" || report.SchemaVersion != "goecs.stun/v1" || report.Status != ReportStatusOK {
		t.Fatalf("unexpected STUN component envelope: %#v", report)
	}
	var payload stuncheck.NATSummary
	if err := json.Unmarshal(report.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Results) != 1 || payload.Results[0].Server != "stun.fixture.test:3478" || payload.PortPreservationConsistency != stuncheck.CapabilityAvailable || payload.HairpinConsistency != stuncheck.CapabilityAvailable {
		t.Fatalf("unexpected structured STUN payload: %+v", payload)
	}
}

func TestLocalMediaComponentPreservesRateLimited(t *testing.T) {
	report := collectMediaComponentWithRunner(context.Background(), unlockexecutor.RunOptions{Selection: "21", IPVersion: "ipv4", Concurrency: 20},
		func(_ context.Context, options unlockexecutor.RunOptions) ([]unlockexecutor.StructuredResult, error) {
			if options.Selection != "21" || options.IPVersion != "ipv4" || options.Concurrency != 20 {
				t.Fatalf("unexpected media options: %+v", options)
			}
			return []unlockexecutor.StructuredResult{
				{Name: "Dola AI", Status: unlockmodel.StatusYes},
				{Name: "Fixture AI", Status: unlockmodel.StatusRateLimited},
			}, nil
		})
	if report.Name != "unlocktests.media" || report.Status != ReportStatusPartial {
		t.Fatalf("unexpected media report: %#v", report)
	}
	var payload mediaComponentPayload
	if err := json.Unmarshal(report.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Selection != "21" || len(payload.Results) != 2 || payload.Results[1].Status != unlockmodel.StatusRateLimited {
		t.Fatalf("unexpected media payload: %+v", payload)
	}
}

func TestLocalMediaComponentPreservesRegistryGroupsAndIPv6Capability(t *testing.T) {
	metadata := []byte(`[{"id":"dola-ai","name":"Dola AI","groups":["ai"],"supports_ipv6":true}]`)
	report := collectMediaComponentWithRegistryRunner(context.Background(), unlockexecutor.RunOptions{Selection: "21", IPVersion: "ipv6"}, metadata,
		func(context.Context, unlockexecutor.RunOptions) ([]unlockexecutor.StructuredResult, error) {
			return []unlockexecutor.StructuredResult{{Name: "Dola AI", Status: unlockmodel.StatusYes}}, nil
		})
	if report.Status != ReportStatusOK {
		t.Fatalf("unexpected media status: %#v", report)
	}
	var payload mediaComponentPayload
	if err := json.Unmarshal(report.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Registry.Matched) != 1 || !payload.Registry.Matched[0].SupportsIPv6 || len(payload.Registry.Matched[0].Groups) != 1 || payload.Registry.Matched[0].Groups[0] != "ai" {
		t.Fatalf("registry metadata was not retained: %+v", payload.Registry)
	}
}

func TestMergeComponentTCPTargetsAddsLocalRegistryWithoutDuplicates(t *testing.T) {
	existing := []TCPTarget{{ID: "existing", Name: "Existing", Host: "existing.test", Port: 443}}
	merged := mergeComponentTCPTargets(existing)
	if len(merged) <= len(existing) {
		t.Fatalf("local TCP registry was not merged: %#v", merged)
	}
	seen := make(map[string]struct{}, len(merged))
	for _, target := range merged {
		key := strings.ToLower(strings.TrimSuffix(target.Host, ".")) + fmt.Sprintf(":%d", target.Port)
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate merged TCP endpoint %q", key)
		}
		seen[key] = struct{}{}
	}
}

func TestRepresentativeICMPTargetsBoundsStandardAndKeepsDeep(t *testing.T) {
	targets := []nt3model.ProvinceLatencyTarget{
		{ProvinceCode: "AA", ProvinceName: "A", Carrier: "ct", IPVersion: "ipv4", Host: "a.test"},
		{ProvinceCode: "BB", ProvinceName: "B", Carrier: "ct", IPVersion: "ipv4", Host: "b.test"},
		{ProvinceCode: "CC", ProvinceName: "C", Carrier: "cu", IPVersion: "ipv4", Host: "c.test"},
	}
	standard := representativeICMPTargets(targets, false)
	deep := representativeICMPTargets(targets, true)
	if len(standard) != 2 || len(deep) != 3 {
		t.Fatalf("unexpected representative target counts: standard=%d deep=%d", len(standard), len(deep))
	}
	if got := pingComponentStatus(context.Background(), []pingprobe.ICMPResult{{Status: "ok"}, {Status: "partial"}}); got != ReportStatusPartial {
		t.Fatalf("ping status = %q", got)
	}
	if got := pingComponentReason([]pingprobe.ICMPResult{{Sent: 3, Received: 2}, {Sent: 3, Received: 1}}, ReportStatusPartial); got != "3/6 ICMP replies received" {
		t.Fatalf("ping reason = %q", got)
	}
	if got := pingTCPComponentReason([]pingprobe.TCPResult{{Attempts: 3, Successful: 2}, {Attempts: 3, Successful: 1}}, ReportStatusPartial); got != "3/6 TCP handshakes succeeded" {
		t.Fatalf("website reason = %q", got)
	}
}

func TestLegacyConfigDisablesStructuredOwnedStagesOnly(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.BasicStatus = false
	cfg.PingTestStatus = true
	cfg.SpeedTestStatus = true
	legacy := legacyConfigForStructured(cfg)
	if legacy == cfg {
		t.Fatal("local structured build did not create a legacy config copy")
	}
	if legacy.CpuTestStatus || legacy.MemoryTestStatus || legacy.DiskTestStatus || legacy.UtTestStatus || legacy.SecurityTestStatus || legacy.EmailTestStatus || legacy.BacktraceStatus || legacy.Nt3Status {
		t.Fatalf("structured-owned stage remained enabled in legacy copy: %+v", legacy)
	}
	if legacy.PingTestStatus || legacy.SpeedTestStatus || !legacy.OnlyIpInfoCheck {
		t.Fatalf("structured-owned network stages or identity probe are incorrect: %+v", legacy)
	}
	if !cfg.CpuTestStatus || !cfg.SecurityTestStatus || !cfg.Nt3Status {
		t.Fatalf("original config was mutated: %+v", cfg)
	}
}

func TestLocalSpeedComponentProbesAndSelectsAvailableNodes(t *testing.T) {
	speedtestData := []byte(`[
		{"id":"good","host":"good.test:8080","url":"https://good.test/speedtest/upload.php","provider":"fixture","status":"available"},
		{"id":"static-bad","host":"bad.test:8080","status":"unavailable"}
	]`)
	openData := []byte(`[{"id":"dial-bad","host":"open.test","port_from":5201,"status":"available"}]`)
	report := collectSpeedComponentWithDependencies(context.Background(), speedtestData, openData, 2,
		func(_ context.Context, _, address string) (net.Conn, error) {
			if address != "good.test:8080" {
				return nil, errors.New("fixture unavailable")
			}
			client, server := net.Pipe()
			go server.Close()
			return client, nil
		}, func(_ context.Context, server speedmodel.ServerMetadata) speedmodel.ThroughputResult {
			return speedmodel.ThroughputResult{ID: server.ID, Name: server.Name, Status: speedmodel.ThroughputAvailable, DownloadMbps: 100, UploadMbps: 50}
		})
	if report.Name != "speed.registry" || report.Status != ReportStatusOK {
		t.Fatalf("unexpected speed report: %#v", report)
	}
	var payload speedComponentPayload
	if err := json.Unmarshal(report.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Nodes) != 3 || len(payload.Selected) != 1 || payload.Selected[0].ID != "good" || len(payload.Benchmarks) != 1 || payload.Benchmarks[0].DownloadMbps != 100 {
		t.Fatalf("unexpected speed payload: %+v", payload)
	}
	if payload.Nodes[1].Availability != "unavailable" || payload.Nodes[2].Host != "open.test:5201" {
		t.Fatalf("static/dial failures were not retained: %+v", payload.Nodes)
	}
}

func TestLocalSpeedComponentDoesNotTreatTCPAsThroughputSuccess(t *testing.T) {
	data := []byte(`[{"id":"tcp-only","host":"good.test:8080","url":"https://good.test/speedtest/upload.php","status":"available"}]`)
	report := collectSpeedComponentWithDependencies(context.Background(), data, nil, 1,
		func(context.Context, string, string) (net.Conn, error) {
			client, server := net.Pipe()
			go server.Close()
			return client, nil
		}, func(_ context.Context, server speedmodel.ServerMetadata) speedmodel.ThroughputResult {
			return speedmodel.ThroughputResult{ID: server.ID, Status: speedmodel.ThroughputUnavailable, Error: "fixture transfer failed"}
		})
	if report.Status != ReportStatusUnavailable || !strings.Contains(report.Reason, "0/1") {
		t.Fatalf("TCP-only node incorrectly marked speed available: %#v", report)
	}
}

func TestTypedSpeedRegistryKeepsLogicalSourceSelectable(t *testing.T) {
	report := collectSpeedComponentFromRegistryWithDependencies(context.Background(), []speedmodel.ServerMetadata{{
		ID: "typed", Name: "Typed", Host: "typed.test:8080", URL: "https://typed.test/upload",
		Source: "embedded", Availability: speedmodel.ServerCandidate,
	}}, []transferTargetInput{{ID: "transfer", Host: "transfer.test", PortFrom: 5201, PortTo: 5210, Status: "available"}}, 1, func(context.Context, string, string) (net.Conn, error) {
		client, server := net.Pipe()
		go server.Close()
		return client, nil
	}, func(_ context.Context, server speedmodel.ServerMetadata) speedmodel.ThroughputResult {
		return speedmodel.ThroughputResult{ID: server.ID, Status: speedmodel.ThroughputAvailable, DownloadMbps: 100}
	}, nil)
	if report.Status != ReportStatusOK {
		t.Fatalf("typed registry was not selectable: %#v", report)
	}
	var payload speedComponentPayload
	if err := json.Unmarshal(report.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Selected) != 1 || payload.Selected[0].Source != "speedtest" || len(payload.Nodes) != 2 || payload.Nodes[1].Source != "openspeedtest" {
		t.Fatalf("typed source was not normalized: %+v", payload.Selected)
	}
}

func TestLocalSpeedComponentRunsPrivateHTTPThroughput(t *testing.T) {
	report := collectSpeedComponentWithAllDependencies(context.Background(), nil, nil, 1, nil, nil,
		func(context.Context, int) (any, int, []privateSpeedBenchmark) {
			return privatepst.RegistryReport{
				SchemaVersion: "privatespeedtest.registry/v1", Availability: privatepst.ServerAvailable,
				Selected: []privatepst.RegistryNode{{ID: "private", Name: "Private", Availability: privatepst.ServerAvailable}},
			}, 1, []privateSpeedBenchmark{{ID: "private", Name: "Private", Source: "privatespeedtest", Status: "available", DownloadMbps: 200, UploadMbps: 100}}
		})
	if report.Status != ReportStatusOK {
		t.Fatalf("private throughput did not satisfy speed component: %#v", report)
	}
	var payload speedComponentPayload
	if err := json.Unmarshal(report.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.PrivateRegistry == nil || len(payload.PrivateBenchmarks) != 1 || payload.PrivateBenchmarks[0].DownloadMbps != 200 {
		t.Fatalf("private throughput evidence missing: %+v", payload)
	}
}

func TestLocalSecurityComponentUsesProviderAndDNSBLFixtures(t *testing.T) {
	zones := []byte(`[{"zone":"clean.fixture.test","ipv4":true,"ipv6":false}]`)
	report := collectSecurityComponentWithDeps(context.Background(), "198.51.100.10", "", zones,
		func(ipVersion string) []securitynetwork.ProviderProbe {
			return []securitynetwork.ProviderProbe{{
				Name: "fixture", IPVersions: []string{ipVersion},
				FetchDetailed: func(context.Context, string) securitynetwork.ProviderFetchResult {
					return securitynetwork.ProviderFetchResult{Score: &basicsmodel.SecurityScore{}}
				},
			}}
		}, componentDNSBLResolver{})
	if report.Name != "security.evidence" || report.Status != ReportStatusOK {
		t.Fatalf("unexpected security report: %#v", report)
	}
	var payload securityComponentPayload
	if err := json.Unmarshal(report.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Addresses) != 1 || len(payload.Addresses[0].Providers) != 1 || payload.Addresses[0].Providers[0].Status != securitynetwork.ProviderAvailable {
		t.Fatalf("unexpected provider payload: %+v", payload)
	}
	if payload.Addresses[0].DNSBL == nil || payload.Addresses[0].DNSBL.Counts[securitynetwork.DNSBLClean] != 1 {
		t.Fatalf("unexpected DNSBL payload: %+v", payload.Addresses[0].DNSBL)
	}
}

func TestLocalBacktraceComponentUsesStructuredRunner(t *testing.T) {
	report := collectBacktraceComponentWithRunner(context.Background(), "192.0.2.1", "2001:db8::1",
		func(_ context.Context, ip string, config bgptools.IPBGPReportConfig) (*bgptools.IPBGPReport, error) {
			if !config.EnableWHOISFallback || !config.FetchGeofeed || config.WHOISTimeout <= 0 || config.ResolveASN == nil {
				t.Fatalf("unexpected backtrace config: %+v", config)
			}
			return &bgptools.IPBGPReport{IP: ip, Status: bgptools.ReportAvailable}, nil
		})
	if report.Name != "backtrace.ip_bgp" || report.Status != ReportStatusOK {
		t.Fatalf("unexpected backtrace report: %#v", report)
	}
	var payload struct {
		Reports []*bgptools.IPBGPReport `json:"reports"`
	}
	if err := json.Unmarshal(report.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Reports) != 2 || payload.Reports[0].IP != "192.0.2.1" || payload.Reports[1].IP != "2001:db8::1" {
		t.Fatalf("unexpected backtrace payload: %+v", payload)
	}
}

func TestLocalBacktraceComponentResolvesASNAndRelationshipsOffline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rdap/192.0.2.1":
			_, _ = io.WriteString(w, `{"handle":"NET-ARIN-TEST","cidr0_cidrs":[{"v4prefix":"192.0.2.0","length":24}],"events":[{"eventAction":"registration","eventDate":"2020-01-02T03:04:05Z"}]}`)
		case "/origin":
			if got := r.URL.Query().Get("resource"); got != "192.0.2.1" {
				t.Errorf("origin resource = %q", got)
			}
			_, _ = io.WriteString(w, `{"status":"ok","data":{"asns":[64500]}}`)
		case "/relationships":
			if got := r.URL.Query().Get("resource"); got != "64500" {
				t.Errorf("relationship resource = %q", got)
			}
			_, _ = io.WriteString(w, `{"data":{"neighbours":[{"asn":64501,"relationship":"upstream"},{"asn":64502,"relationship":"peer"}]}}`)
		case "/peering":
			if got := r.URL.Query().Get("asn"); got != "64500" {
				t.Errorf("peering ASN = %q", got)
			}
			_, _ = io.WriteString(w, `{"data":[{"ix_id":7,"name":"Fixture IX"}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	report := collectBacktraceComponentWithDependencies(context.Background(), "192.0.2.1", "", nil, bgptools.QueryIPBGPReport,
		func() bgptools.IPBGPReportConfig {
			return bgptools.IPBGPReportConfig{
				Timeout: time.Second, RDAPClient: server.Client(), RDAPBaseURL: server.URL + "/rdap",
				ResolveASN: func(ctx context.Context, ip string) (string, error) {
					return bgptools.ResolveOriginASNWithConfig(ctx, ip, bgptools.OriginASNConfig{
						Client: server.Client(), BaseURL: server.URL + "/origin", Timeout: time.Second,
					})
				},
				Relationships: bgptools.RelationshipConfig{
					Client: server.Client(), RIPEstatURL: server.URL + "/relationships",
					PeeringDBURL: server.URL + "/peering", Timeout: time.Second,
				},
			}
		})
	if report.Status != ReportStatusOK {
		t.Fatalf("backtrace component status = %q, reason = %q", report.Status, report.Reason)
	}
	var payload struct {
		Reports []*bgptools.IPBGPReport `json:"reports"`
	}
	if err := json.Unmarshal(report.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Reports) != 1 || payload.Reports[0].ASN != "64500" || payload.Reports[0].Relationships == nil {
		t.Fatalf("unexpected resolved report: %+v", payload.Reports)
	}
	relationships := payload.Reports[0].Relationships
	if len(relationships.Upstreams) != 1 || len(relationships.Peers) != 1 || len(relationships.IXPs) != 1 {
		t.Fatalf("relationships were not executed: %+v", relationships)
	}
}

func TestHardwareStageRunsEachBenchmarkOnceWithOneDeadline(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.BasicStatus = false
	cfg.CpuTestStatus = true
	cfg.MemoryTestStatus = true
	cfg.DiskTestStatus = true
	cfg.HardwareBudget = time.Second
	cfg.DiskTestPath = t.TempDir()

	type observed struct {
		name     string
		deadline time.Time
	}
	observedRuns := make([]observed, 0, 3)
	runners := hardwareComponentRunners{
		CPU: func(ctx context.Context, _ cpu.StructuredConfig) cpu.StructuredResult {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Fatal("CPU context has no shared deadline")
			}
			observedRuns = append(observedRuns, observed{"cpu", deadline})
			return cpu.StructuredResult{SchemaVersion: "goecs.cpu/v1", Status: "ok"}
		},
		Memory: func(ctx context.Context, _ memory.BenchmarkConfig) (memory.BenchmarkResult, error) {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Fatal("memory context has no shared deadline")
			}
			observedRuns = append(observedRuns, observed{"memory", deadline})
			return memory.BenchmarkResult{SchemaVersion: "goecs.memory/v1", Status: memory.BenchmarkOK}, nil
		},
		Disk: func(ctx context.Context, _ disk.MatrixConfig) disk.MatrixResult {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Fatal("disk context has no shared deadline")
			}
			observedRuns = append(observedRuns, observed{"disk", deadline})
			return disk.MatrixResult{SchemaVersion: "goecs.disk/v1", Status: "ok"}
		},
	}

	reports := collectHardwareComponentReports(context.Background(), cfg, runners)
	if len(reports) != 3 || len(observedRuns) != 3 {
		t.Fatalf("reports=%d runs=%d, want three each: %#v", len(reports), len(observedRuns), reports)
	}
	for index := range observedRuns {
		if observedRuns[index].name != []string{"cpu", "memory", "disk"}[index] {
			t.Fatalf("unexpected execution order: %#v", observedRuns)
		}
		if !observedRuns[index].deadline.Equal(observedRuns[0].deadline) {
			t.Fatalf("benchmark %q received a separate budget: %#v", observedRuns[index].name, observedRuns)
		}
	}
}

func TestHardwareStageKeepsStandardProfiles(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.BasicStatus = false
	cfg.CpuTestStatus = true
	cfg.MemoryTestStatus = true
	cfg.DiskTestStatus = true
	cfg.HardwareBudget = 2 * time.Minute
	cfg.DiskTestPath = t.TempDir()

	var cpuConfig cpu.StructuredConfig
	var memoryConfig memory.BenchmarkConfig
	var diskConfig disk.MatrixConfig
	deepDiskCalls := 0
	reports := collectHardwareComponentReports(context.Background(), cfg, hardwareComponentRunners{
		CPU: func(_ context.Context, config cpu.StructuredConfig) cpu.StructuredResult {
			cpuConfig = config
			return cpu.StructuredResult{SchemaVersion: "goecs.cpu/v1", Status: "ok"}
		},
		Memory: func(_ context.Context, config memory.BenchmarkConfig) (memory.BenchmarkResult, error) {
			memoryConfig = config
			return memory.BenchmarkResult{SchemaVersion: "goecs.memory/v1", Status: memory.BenchmarkOK}, nil
		},
		Disk: func(_ context.Context, config disk.MatrixConfig) disk.MatrixResult {
			diskConfig = config
			return disk.MatrixResult{SchemaVersion: "goecs.disk/v1", Status: "ok"}
		},
		DeepDisk: func(context.Context, disk.MatrixConfig) disk.MatrixResult {
			deepDiskCalls++
			return disk.MatrixResult{SchemaVersion: "goecs.disk/v1", Status: "ok"}
		},
	})

	defaultMemory := memory.DefaultBenchmarkConfig()
	if len(reports) != 3 || deepDiskCalls != 0 {
		t.Fatalf("reports=%d deep disk calls=%d, want 3 and 0", len(reports), deepDiskCalls)
	}
	if cpuConfig.Duration != 5*time.Second || cpuConfig.MaxPrime != 10000 {
		t.Fatalf("unexpected standard CPU config: %+v", cpuConfig)
	}
	if memoryConfig.WorkingSetBytes != defaultMemory.WorkingSetBytes || memoryConfig.Iterations != defaultMemory.Iterations {
		t.Fatalf("unexpected standard memory config: %+v", memoryConfig)
	}
	if diskConfig.SizeBytes != 16<<20 || diskConfig.Runtime != time.Second || diskConfig.MaxDuration != 45*time.Second {
		t.Fatalf("unexpected standard disk config: %+v", diskConfig)
	}
}

func TestHardwareStageSelectsDeepProfiles(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.BasicStatus = false
	cfg.CpuTestStatus = true
	cfg.MemoryTestStatus = true
	cfg.DiskTestStatus = true
	cfg.DeepMode = true
	cfg.HardwareBudget = 5 * time.Minute
	cfg.DiskTestPath = t.TempDir()

	var cpuConfig cpu.StructuredConfig
	var memoryConfig memory.BenchmarkConfig
	var deepDiskConfig disk.MatrixConfig
	standardDiskCalls := 0
	reports := collectHardwareComponentReports(context.Background(), cfg, hardwareComponentRunners{
		CPU: func(_ context.Context, config cpu.StructuredConfig) cpu.StructuredResult {
			cpuConfig = config
			return cpu.StructuredResult{SchemaVersion: "goecs.cpu/v1", Status: "ok"}
		},
		Memory: func(_ context.Context, config memory.BenchmarkConfig) (memory.BenchmarkResult, error) {
			memoryConfig = config
			return memory.BenchmarkResult{SchemaVersion: "goecs.memory/v1", Status: memory.BenchmarkOK}, nil
		},
		Disk: func(context.Context, disk.MatrixConfig) disk.MatrixResult {
			standardDiskCalls++
			return disk.MatrixResult{SchemaVersion: "goecs.disk/v1", Status: "ok"}
		},
		DeepDisk: func(_ context.Context, config disk.MatrixConfig) disk.MatrixResult {
			deepDiskConfig = config
			return disk.MatrixResult{SchemaVersion: "goecs.disk/v1", Status: "ok"}
		},
	})

	if len(reports) != 7 || standardDiskCalls != 0 {
		t.Fatalf("reports=%d standard disk calls=%d, want 7 and 0", len(reports), standardDiskCalls)
	}
	for _, report := range reports[3:] {
		if report.Status != ReportStatusSkipped || report.Reason == "" {
			t.Fatalf("unconfigured optional deep operation was not explicit: %+v", report)
		}
	}
	if cpuConfig.Duration != 20*time.Second || cpuConfig.MaxPrime != 50000 {
		t.Fatalf("unexpected deep CPU config: %+v", cpuConfig)
	}
	if memoryConfig.WorkingSetBytes != 256<<20 || memoryConfig.Iterations != 8 {
		t.Fatalf("unexpected deep memory config: %+v", memoryConfig)
	}
	if deepDiskConfig.SizeBytes != 256<<20 || deepDiskConfig.Runtime != 2*time.Second || deepDiskConfig.MaxDuration != 3*time.Minute {
		t.Fatalf("unexpected deep disk config: %+v", deepDiskConfig)
	}
}

func TestDeepDeviceOperationsRequireExplicitTargets(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.DeepMode = true
	reports := collectExplicitDeepHardwareReports(context.Background(), cfg)
	if len(reports) != 4 {
		t.Fatalf("got %d deep placeholders", len(reports))
	}
	for _, report := range reports {
		if report.Status != ReportStatusSkipped || report.Reason == "" {
			t.Fatalf("implicit deep operation was not skipped: %+v", report)
		}
	}
	if got := splitExplicitTargets(" /mnt/a,/mnt/a, /mnt/b "); len(got) != 2 || got[0] != "/mnt/a" || got[1] != "/mnt/b" {
		t.Fatalf("unexpected explicit targets: %#v", got)
	}
}

func TestExplicitDeepHardwareCanRunWithoutStandardHardwareFlags(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.DeepMode = true
	cfg.CpuTestStatus = false
	cfg.MemoryTestStatus = false
	cfg.DiskTestStatus = false
	cfg.DeepBurnDuration = 5 * time.Millisecond
	cfg.HardwareBudget = time.Second
	reports := collectHardwareComponentReports(context.Background(), cfg, hardwareComponentRunners{})
	if len(reports) != 4 {
		t.Fatalf("explicit deep-only run produced %d reports: %+v", len(reports), reports)
	}
	if reports[2].Name != "cputest.burn" || reports[2].Status != ReportStatusOK {
		t.Fatalf("explicit burn did not run: %+v", reports[2])
	}
}

func TestHardwareStageStopsBeforeStartingNextBenchmarkOnCancel(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.BasicStatus = false
	cfg.CpuTestStatus = true
	cfg.MemoryTestStatus = true
	cfg.DiskTestStatus = true
	cfg.HardwareBudget = time.Second
	cfg.DiskTestPath = t.TempDir()
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	memoryCalls, diskCalls := 0, 0
	runners := hardwareComponentRunners{
		CPU: func(ctx context.Context, _ cpu.StructuredConfig) cpu.StructuredResult {
			cancel()
			<-ctx.Done()
			return cpu.StructuredResult{SchemaVersion: "goecs.cpu/v1", Status: "canceled", Error: ctx.Err().Error()}
		},
		Memory: func(context.Context, memory.BenchmarkConfig) (memory.BenchmarkResult, error) {
			memoryCalls++
			return memory.BenchmarkResult{SchemaVersion: "goecs.memory/v1", Status: memory.BenchmarkOK}, nil
		},
		Disk: func(context.Context, disk.MatrixConfig) disk.MatrixResult {
			diskCalls++
			return disk.MatrixResult{SchemaVersion: "goecs.disk/v1", Status: "ok"}
		},
	}

	reports := collectHardwareComponentReports(parent, cfg, runners)
	if memoryCalls != 0 || diskCalls != 0 {
		t.Fatalf("later benchmarks started after cancellation: memory=%d disk=%d", memoryCalls, diskCalls)
	}
	if len(reports) != 3 || reports[0].Status != ReportStatusCanceled || reports[1].Status != ReportStatusCanceled || reports[2].Status != ReportStatusCanceled {
		t.Fatalf("cancellation was not propagated to all enabled components: %#v", reports)
	}
}

func TestHardwareStageUsesOneBudgetAndStopsAtDeadline(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.BasicStatus = false
	cfg.CpuTestStatus = true
	cfg.MemoryTestStatus = true
	cfg.DiskTestStatus = true
	cfg.HardwareBudget = 20 * time.Millisecond
	cfg.DiskTestPath = t.TempDir()
	memoryCalls, diskCalls := 0, 0
	started := time.Now()
	reports := collectHardwareComponentReports(context.Background(), cfg, hardwareComponentRunners{
		CPU: func(ctx context.Context, _ cpu.StructuredConfig) cpu.StructuredResult {
			<-ctx.Done()
			return cpu.StructuredResult{SchemaVersion: "goecs.cpu/v1", Status: "canceled", Error: ctx.Err().Error()}
		},
		Memory: func(context.Context, memory.BenchmarkConfig) (memory.BenchmarkResult, error) {
			memoryCalls++
			return memory.BenchmarkResult{}, nil
		},
		Disk: func(context.Context, disk.MatrixConfig) disk.MatrixResult {
			diskCalls++
			return disk.MatrixResult{}
		},
	})
	if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
		t.Fatalf("hardware stage did not return promptly after deadline: %s", elapsed)
	}
	if memoryCalls != 0 || diskCalls != 0 {
		t.Fatalf("later benchmarks started after shared deadline: memory=%d disk=%d", memoryCalls, diskCalls)
	}
	if len(reports) != 3 {
		t.Fatalf("got %d component reports, want 3: %#v", len(reports), reports)
	}
	for _, report := range reports {
		if report.Status != ReportStatusTimeout || report.Reason == "" {
			t.Fatalf("component %q did not retain shared deadline result: %#v", report.Name, report)
		}
	}
}

func TestStructuredPostLegacyCollectionSkipsHardware(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.BasicStatus = false
	cfg.CpuTestStatus = true
	cfg.MemoryTestStatus = true
	cfg.DiskTestStatus = true
	reports := collectHardwareComponentReports(skipStructuredHardware(context.Background()), cfg, hardwareComponentRunners{
		CPU: func(context.Context, cpu.StructuredConfig) cpu.StructuredResult {
			t.Fatal("CPU benchmark repeated")
			return cpu.StructuredResult{}
		},
		Memory: func(context.Context, memory.BenchmarkConfig) (memory.BenchmarkResult, error) {
			t.Fatal("memory benchmark repeated")
			return memory.BenchmarkResult{}, nil
		},
		Disk: func(context.Context, disk.MatrixConfig) disk.MatrixResult {
			t.Fatal("disk benchmark repeated")
			return disk.MatrixResult{}
		},
	})
	if len(reports) != 0 {
		t.Fatalf("post-legacy hardware should be omitted, got %#v", reports)
	}
}
