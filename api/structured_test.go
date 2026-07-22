package api

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

func TestRunTCPReportsSuccessAndClassification(t *testing.T) {
	targets := []TCPTarget{{ID: "one", Host: "example.test", Port: 443}}
	results := runTCPReports(context.Background(), targets, tcpProbeConfig{
		attempts: 2, timeout: time.Second, concurrency: 1,
		dial: func(context.Context, string, string) (net.Conn, error) {
			local, remote := net.Pipe()
			go remote.Close()
			return local, nil
		},
	})
	if len(results) != 1 || results[0].Successful != 2 || results[0].SuccessRatePercent != 100 || results[0].LossPercent != 0 {
		t.Fatalf("unexpected successful result: %#v", results)
	}

	results = runTCPReports(context.Background(), targets, tcpProbeConfig{
		attempts: 1, timeout: time.Second, concurrency: 1,
		dial: func(context.Context, string, string) (net.Conn, error) {
			return nil, context.DeadlineExceeded
		},
	})
	if results[0].Errors["timeout"] != 1 || results[0].SuccessRatePercent != 0 || results[0].LossPercent != 100 {
		t.Fatalf("unexpected timeout result: %#v", results[0])
	}
	encoded, err := json.Marshal(results[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"success_rate_percent":0`) {
		t.Fatalf("TCP JSON is missing success_rate_percent: %s", encoded)
	}
}

func TestStructuredReportJSONAndPrivacy(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.PrivacyMode = true
	cfg.TCPProbeStatus = false
	started := time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC)
	report := CollectStructuredReport(context.Background(), NetCheckResult{Connected: false, StackType: "None"}, cfg, "sensitive text", started, started.Add(time.Second))
	if report.Text != "" || !report.PrivacyMode || report.SchemaVersion != StructuredReportSchema {
		t.Fatalf("unexpected report: %#v", report)
	}
	encoded, err := report.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) == 0 {
		t.Fatal("empty JSON")
	}
}

func TestStructuredPrivacyRedactsComponentIdentity(t *testing.T) {
	report := &StructuredReport{
		PrivacyMode: true,
		Text:        "public 203.0.113.10 and 2001:db8::10",
		Sections:    []SectionReport{{Name: "security", Reason: "dial 203.0.113.10:443"}},
		Components: []ComponentReport{{
			Name:   "fixture",
			Reason: "lookup [2001:db8::10]:443 failed",
			Payload: json.RawMessage(`{
				"schema_version":"fixture/v1",
				"ip":"203.0.113.10",
				"ip_type":"ipv4",
				"serial_number":"SERIAL-SECRET",
				"device_path":"/dev/disk-secret",
				"target":"/dev/target-secret",
				"mapped_address":"[2001:db8::10]:40000",
				"error":"dial tcp 203.0.113.10:443"
			}`),
		}},
	}
	report.Components = append(report.Components, ComponentReport{
		Name:    "route-fixture",
		Payload: json.RawMessage(`{"target":{"province_name":"Guangdong","carrier":"ct","host":"203.0.113.11"},"status":"ok"}`),
	})
	applyStructuredPrivacy(report)
	encoded, err := report.JSON()
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"203.0.113.10", "2001:db8::10", "SERIAL-SECRET", "/dev/disk-secret", "/dev/target-secret"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("privacy report leaked %q: %s", secret, encoded)
		}
	}
	if !strings.Contains(string(encoded), `"ip_type": "ipv4"`) || !strings.Contains(string(encoded), `"schema_version": "fixture/v1"`) {
		t.Fatalf("privacy filtering removed structural fields: %s", encoded)
	}
	if !strings.Contains(string(encoded), `"province_name": "Guangdong"`) || !strings.Contains(string(encoded), `"carrier": "ct"`) {
		t.Fatalf("privacy filtering removed non-identifying target labels: %s", encoded)
	}
}

func TestDataOfflineForcesEmbeddedSnapshot(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.DataOffline = true
	cfg.TCPProbeStatus = false
	// Data-offline validation must not start live mail/STUN probes in the local
	// component adapter; network capability is covered by dedicated fixtures.
	report := CollectStructuredReport(context.Background(), NetCheckResult{Connected: false, StackType: "None"}, cfg, "", time.Now(), time.Now())
	if report.Data == nil || report.Data.Source != "embedded" || report.Data.Fallback != "embedded" {
		t.Fatalf("unexpected offline data source: %#v", report.Data)
	}
	if len(report.DataFiles) != 8 {
		t.Fatalf("expected all known data files, got %#v", report.DataFiles)
	}
	for index, file := range report.DataFiles {
		if !hasPrivateComponentData() && (file.File == dnsblDataFile || file.File == privateDataFile || file.File == transferDataFile) {
			if file.Status != ReportStatusError || file.Reason == "" {
				t.Fatalf("public-only file %d has unexpected state: %#v", index, file)
			}
			continue
		}
		if file.Status != ReportStatusOK || file.Source != "embedded" || file.Fallback != "embedded" {
			t.Fatalf("file %d has unexpected state: %#v", index, file)
		}
		if index > 0 && report.DataFiles[index-1].File >= file.File {
			t.Fatalf("data files are not stable-sorted: %#v", report.DataFiles)
		}
	}
}

func TestRunAllTestsContextPrivacyOmitsText(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.PrivacyMode = true
	cfg.TCPProbeStatus = false
	cfg.BasicStatus = false
	cfg.CpuTestStatus = false
	cfg.MemoryTestStatus = false
	cfg.DiskTestStatus = false
	cfg.UtTestStatus = false
	cfg.SecurityTestStatus = false
	cfg.EmailTestStatus = false
	cfg.BacktraceStatus = false
	cfg.Nt3Status = false
	cfg.SpeedTestStatus = false
	result := RunAllTestsContext(context.Background(), NetCheckResult{Connected: false, StackType: "None"}, cfg)
	if result.Report == nil || result.Report.Text != "" {
		t.Fatalf("privacy report leaked text: %#v", result.Report)
	}
}

func TestClassifyTCPError(t *testing.T) {
	if got := classifyTCPError(context.Canceled); got != "canceled" {
		t.Fatalf("got %q", got)
	}
	if got := classifyTCPError(errors.New("connection refused")); got != "refused" {
		t.Fatalf("got %q", got)
	}
}

func TestSectionReportsMarksOfflineNetworkUnavailable(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.PingTestStatus = true
	cfg.TCPProbeStatus = true
	sections := sectionReports(cfg, NetCheckResult{Connected: false, StackType: "None"}, structuredExtras{}, ReportStatusOK, "")
	statuses := make(map[string]ReportStatus)
	for _, section := range sections {
		statuses[section.Name] = section.Status
	}
	if statuses["basics"] != ReportStatusPartial || statuses["media"] != ReportStatusUnavailable || statuses["tcp"] != ReportStatusUnavailable || statuses["nat"] != ReportStatusUnavailable {
		t.Fatalf("unexpected statuses: %+v", statuses)
	}
}

func TestSectionReportsDoesNotClaimMissingStructuredComponentsAreOK(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.SecurityTestStatus = true
	sections := sectionReports(cfg, NetCheckResult{Connected: true}, structuredExtras{}, ReportStatusOK, "")
	for _, name := range []string{"media", "security", "backtrace", "speed", "disk"} {
		var found *SectionReport
		for index := range sections {
			if sections[index].Name == name {
				found = &sections[index]
				break
			}
		}
		if found == nil || found.Status != ReportStatusPartial || found.Reason != "structured component unavailable" {
			t.Fatalf("section %q incorrectly claimed success: %#v", name, found)
		}
	}
}

func TestAggregateReportStatusReflectsPartialSections(t *testing.T) {
	sections := []SectionReport{
		{Name: "disabled", Enabled: false, Status: ReportStatusSkipped},
		{Name: "basics", Enabled: true, Status: ReportStatusOK},
		{Name: "media", Enabled: true, Status: ReportStatusPartial},
	}
	if got := aggregateReportStatus(ReportStatusOK, sections); got != ReportStatusPartial {
		t.Fatalf("aggregate status = %q, want partial", got)
	}
	if got := aggregateReportStatus(ReportStatusTimeout, sections); got != ReportStatusTimeout {
		t.Fatalf("timeout was overwritten by section aggregation: %q", got)
	}
}

func TestAggregateComponentSectionIgnoresOptionalSkippedDeepItems(t *testing.T) {
	status, reason := aggregateComponentSectionStatus([]ComponentReport{
		{Name: "cputest", Status: ReportStatusOK},
		{Name: "cputest.burn", Status: ReportStatusSkipped, Reason: "not configured"},
	})
	if status != ReportStatusOK || reason != "" {
		t.Fatalf("optional deep item changed successful section: status=%q reason=%q", status, reason)
	}
	status, _ = aggregateComponentSectionStatus([]ComponentReport{{Name: "cputest.burn", Status: ReportStatusSkipped}})
	if status != ReportStatusSkipped {
		t.Fatalf("all-skipped section status = %q", status)
	}
}

func TestAggregateComponentSectionSuppliesMissingFailureReason(t *testing.T) {
	status, reason := aggregateComponentSectionStatus([]ComponentReport{{Name: "disktest", Status: ReportStatusUnavailable}})
	if status != ReportStatusUnavailable || reason != "unavailable" {
		t.Fatalf("unexpected component fallback reason: status=%q reason=%q", status, reason)
	}
}

func TestTCPSectionStatusUsesSingleTopLevelProbeSet(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.TCPProbeStatus = true
	reports := []TCPReport{{Attempts: 3, Successful: 3}}
	sections := sectionReports(cfg, NetCheckResult{Connected: true}, structuredExtras{tcp: reports}, ReportStatusOK, "")
	for _, section := range sections {
		if section.Name == "tcp" {
			if section.Status != ReportStatusOK || section.Reason != "" {
				t.Fatalf("unexpected successful TCP section: %#v", section)
			}
			return
		}
	}
	t.Fatal("TCP section missing")
}

func TestTCPSectionStatusReportsPartialLoss(t *testing.T) {
	status, reason := tcpSectionStatus([]TCPReport{{Attempts: 3, Successful: 2}, {Attempts: 3, Successful: 1}})
	if status != ReportStatusPartial || !strings.Contains(reason, "3/6") {
		t.Fatalf("unexpected TCP aggregate: status=%q reason=%q", status, reason)
	}
}

func TestRouteSectionAggregatesLatencyAndDeepRouteComponents(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Nt3Status = true
	extras := structuredExtras{components: []ComponentReport{
		{Name: "nt3.province_latency", Status: ReportStatusPartial, Reason: "latency degraded"},
		{Name: "nt3.province_routes", Status: ReportStatusOK},
	}}
	sections := sectionReports(cfg, NetCheckResult{Connected: true}, extras, ReportStatusOK, "")
	for _, section := range sections {
		if section.Name == "routes" {
			if section.Status != ReportStatusPartial || !strings.Contains(section.Reason, "latency degraded") {
				t.Fatalf("route components were not aggregated: %#v", section)
			}
			return
		}
	}
	t.Fatal("routes section missing")
}

func TestStructuredReportComponentEnvelope(t *testing.T) {
	report := &StructuredReport{SchemaVersion: StructuredReportSchema, Status: ReportStatusOK}
	unified := report.WithComponents(ComponentReport{Name: "basics", SchemaVersion: "goecs.system/v1", Status: ReportStatusOK, Payload: json.RawMessage(`{"availability":"available"}`)})
	encoded, err := json.Marshal(unified)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"components"`) || !strings.Contains(string(encoded), `goecs.system/v1`) {
		t.Fatalf("unexpected envelope: %s", encoded)
	}
}

func TestSectionReportsUsesComponentStatus(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.TCPProbeStatus = false
	sections := sectionReports(cfg, NetCheckResult{Connected: true}, structuredExtras{components: []ComponentReport{{
		Name: "cputest", Status: ReportStatusUnavailable, Reason: "temperature unavailable",
	}}}, ReportStatusOK, "")
	for _, section := range sections {
		if section.Name != "cpu" {
			continue
		}
		if section.Status != ReportStatusUnavailable || section.Reason != "temperature unavailable" {
			t.Fatalf("unexpected CPU section: %#v", section)
		}
		return
	}
	t.Fatal("CPU section missing")
}

func TestSectionReportsUsesSuccessfulStructuredComponent(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.BasicStatus = false
	cfg.MemoryTestStatus = false
	cfg.DiskTestStatus = false
	cfg.UtTestStatus = false
	cfg.SecurityTestStatus = false
	cfg.EmailTestStatus = false
	cfg.BacktraceStatus = false
	cfg.Nt3Status = false
	cfg.SpeedTestStatus = false
	cfg.TCPProbeStatus = false
	sections := sectionReports(cfg, NetCheckResult{Connected: true}, structuredExtras{
		components: []ComponentReport{{Name: "cputest", Status: ReportStatusOK}},
	}, ReportStatusOK, "")
	for _, section := range sections {
		if section.Name == "cpu" {
			if section.Status != ReportStatusOK || section.Reason != "" {
				t.Fatalf("unexpected CPU section: %+v", section)
			}
			return
		}
	}
	t.Fatal("CPU section missing")
}

func TestSectionReportsIncludesTelegramAndWebsiteComponents(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.TgdcTestStatus = true
	cfg.WebTestStatus = true
	sections := sectionReports(cfg, NetCheckResult{Connected: true}, structuredExtras{components: []ComponentReport{
		{Name: "ping.telegram", Status: ReportStatusOK},
		{Name: "ping.web_tcp", Status: ReportStatusPartial, Reason: "one website timed out"},
	}}, ReportStatusOK, "")
	statuses := make(map[string]SectionReport, len(sections))
	for _, section := range sections {
		statuses[section.Name] = section
	}
	if statuses["tgdc"].Status != ReportStatusOK {
		t.Fatalf("unexpected Telegram section: %#v", statuses["tgdc"])
	}
	if statuses["web"].Status != ReportStatusPartial || statuses["web"].Reason != "one website timed out" {
		t.Fatalf("unexpected website section: %#v", statuses["web"])
	}
}
