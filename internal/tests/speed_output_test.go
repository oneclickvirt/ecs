package tests

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestWriteFilteredSpeedtestOutputRemovesOnlyNoMatchSentinel(t *testing.T) {
	var output bytes.Buffer
	writeFilteredSpeedtestOutput(&output, "No match servers\n[WARN] useful warning\n Speedtest.net   1 Mbps\n")
	if strings.Contains(output.String(), suppressedSpeedtestLine) {
		t.Fatalf("no-match sentinel leaked into table: %q", output.String())
	}
	for _, want := range []string{"[WARN] useful warning", "Speedtest.net"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("filtered output lost %q: %q", want, output.String())
		}
	}
}

func TestRenderNearbySpeedtestShowsUnavailableRowWhenAllAttemptsFail(t *testing.T) {
	var output bytes.Buffer
	renderNearbySpeedtest(context.Background(), &output, "tcp4", func(context.Context, io.Writer, string) {})
	if !strings.Contains(output.String(), "Speedtest.net") || strings.Count(output.String(), "N/A") != 4 {
		t.Fatalf("missing explicit nearby fallback row: %q", output.String())
	}
}
