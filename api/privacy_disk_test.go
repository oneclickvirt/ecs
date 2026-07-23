package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStructuredPrivacyRedactsLegacyDiskPathsButKeepsMetrics(t *testing.T) {
	report := &StructuredReport{Components: []ComponentReport{{
		Name:    "disktest",
		Payload: json.RawMessage(`{"method":"dd","legacy_output":"Test Path    Block Size    Direct Write(IOPS)\n/root        4k            10 MB/s(10)\n/dev/sda     1M            20 MB/s(20)\n"}`),
	}}}
	applyStructuredPrivacy(report)
	encoded := string(report.Components[0].Payload)
	for _, forbidden := range []string{"/root", "/dev/sda"} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("privacy payload retained disk path %q: %s", forbidden, encoded)
		}
	}
	for _, want := range []string{"Test Path", "[redacted-path]", "10 MB/s(10)", "20 MB/s(20)"} {
		if !strings.Contains(encoded, want) {
			t.Fatalf("privacy payload lost %q: %s", want, encoded)
		}
	}
}

func TestRedactLegacyDiskOutputKeepsRelativeDeviceLabels(t *testing.T) {
	const input = "Test Path  Block\nsda        4k\n"
	if got := redactLegacyDiskOutput(input); got != input {
		t.Fatalf("relative device label changed: %q", got)
	}
}
