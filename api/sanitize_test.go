package api

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizePublicTextRemovesPrivateDetails(t *testing.T) {
	input := "load https://private.example/data?token=abc failed at /home/runner/build: Authorization: Bearer header-secret github.com/private-owner/private-repo"
	got := sanitizePublicText(input)
	for _, forbidden := range []string{"private.example", "token=abc", "/home/runner", "header-secret", "private-owner/private-repo"} {
		if contains := len(got) > 0 && stringContains(got, forbidden); contains {
			t.Fatalf("sanitized text still contains %q: %q", forbidden, got)
		}
	}
}

func TestSanitizePublicOutputPreservesProjectLinksAndHidesRegistryDiagnostics(t *testing.T) {
	input := " Go Project: https://github.com/oneclickvirt/ecs\n" +
		" registry source: https://private.example/data?token=abc\n" +
		" 数据源加载失败: git@private.example:owner/repo.git\n"
	got := sanitizePublicOutput(input)
	if !strings.Contains(got, "https://github.com/oneclickvirt/ecs") {
		t.Fatalf("public project link was removed: %q", got)
	}
	for _, forbidden := range []string{"private.example", "token=abc", "owner/repo.git"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("sanitized output still contains %q: %q", forbidden, got)
		}
	}
}

func TestStructuredReportPublicBoundaryRemovesProvenance(t *testing.T) {
	report := &StructuredReport{
		SchemaVersion: StructuredReportSchema,
		Data:          &DataVersion{Source: "https://private.example/manifest"},
		DataFiles:     []DataFileVersion{{File: "private.json", Source: "git@private.example:owner/repo.git"}},
		Sections:      []SectionReport{{Name: "speed", Enabled: true, Status: ReportStatusError, Reason: "fetch https://private.example/list failed"}},
		Components: []ComponentReport{{
			Name: "speed.registry", SchemaVersion: "fixture/v1", Status: ReportStatusError,
			Payload: json.RawMessage(`{
                "source":"speedtest",
                "registry":{"source":"private-loader","fallback":"embedded"},
                "manifest_url":"https://private.example/manifest",
                "private_registry":{"source":"https://private.example/list"},
                "rdap":{"endpoint":"https://private.example/rdap","url":"https://private.example/rdap"},
                "geofeeds":[{"endpoint":"https://private.example/geofeed","status":"available"}],
				"host":"speed-target.example:443",
				"url":"https://speed-target.example/upload?token=node-secret&mode=test",
				"api_key":"payload-secret",
                "error":"download https://private.example/list?key=secret failed",
                "measurement":123
            }`),
		}},
		TCP: []TCPReport{{Target: TCPTarget{Name: "TCP fixture", Host: "tcp-target.example", Port: 443}}},
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"private.example", "owner/repo.git", "private.json", "private_registry", `"data":`, `"data_files":`, `"manifest_url"`, `"endpoint"`, "node-secret", "payload-secret", `"api_key":`, `"registry":{"source"`} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("public report exposed %q: %s", forbidden, encoded)
		}
	}
	if !strings.Contains(string(encoded), `"source": "speedtest"`) {
		t.Fatalf("normal component source was removed: %s", encoded)
	}
	if !strings.Contains(string(encoded), `"measurement": 123`) {
		t.Fatalf("sanitizer removed benchmark data: %s", encoded)
	}
	for _, preserved := range []string{"speed-target.example:443", "https://speed-target.example/upload?token=[redacted]", "mode=test", "tcp-target.example"} {
		if !strings.Contains(string(encoded), preserved) {
			t.Fatalf("sanitizer removed normal test target %q: %s", preserved, encoded)
		}
	}
	if report.TCP[0].Target.Host != "tcp-target.example" {
		t.Fatalf("JSON marshaling mutated TCP target: %#v", report.TCP[0].Target)
	}
}

func TestProgressReasonIsSanitized(t *testing.T) {
	var received ProgressEvent
	ctx := WithProgressObserver(context.Background(), func(event ProgressEvent) { received = event })
	emitProgress(ctx, ProgressEvent{Section: "speed", Reason: "load https://private.example/list?token=secret failed"})
	if strings.Contains(received.Reason, "private.example") || strings.Contains(received.Reason, "token=secret") {
		t.Fatalf("progress reason leaked registry details: %q", received.Reason)
	}
}

func TestSanitizeLogFilePreservesTargetsAndRemovesRegistryDiagnostics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ecs.log")
	input := strings.Join([]string{
		"2026-08-06T00:00:00Z\tinfo\tprobe target https://speed-target.example/upload",
		"2026-08-06T00:00:01Z\terror\tregistry load https://private.example/list?token=secret failed",
		"2026-08-06T00:00:02Z\terror\tRDAP endpoint https://private.example/rdap unavailable",
	}, "\n")
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := SanitizeLogFile(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	if !strings.Contains(output, "https://speed-target.example/upload") {
		t.Fatalf("sanitizer removed normal probe target: %q", output)
	}
	for _, forbidden := range []string{"private.example", "token=secret"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("ecs.log exposed %q: %q", forbidden, output)
		}
	}
}

func TestSanitizeLogFileRestrictsPermissionsWhenContentIsAlreadyClean(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ecs.log")
	if err := os.WriteFile(path, []byte("ordinary benchmark output\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SanitizeLogFile(path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("ecs.log permissions = %o, want 600", got)
	}
}

func stringContains(value, part string) bool {
	for index := 0; index+len(part) <= len(value); index++ {
		if value[index:index+len(part)] == part {
			return true
		}
	}
	return false
}
