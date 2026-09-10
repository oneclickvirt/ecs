package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
)

const suppressedSpeedtestLine = "No match servers"

type nearbySpeedtestRunner func(context.Context, io.Writer, string)

// renderNearbySpeedtest keeps component diagnostics out of the result table
// and guarantees that the fixed nearby slot never silently disappears. The
// N/A row means every real nearby attempt failed; it is not a synthetic speed
// measurement.
func renderNearbySpeedtest(ctx context.Context, writer io.Writer, network string, run nearbySpeedtestRunner) {
	writer = writerOrDiscard(writer)
	var output bytes.Buffer
	if run != nil {
		run(ctx, &output, network)
	}
	_, renderedNearby := writeFilteredSpeedtestOutput(writer, output.String())
	if renderedNearby {
		return
	}
	fmt.Fprint(writer, " ", formatSpeedCell("Speedtest.net"))
	for range 4 {
		fmt.Fprint(writer, formatSpeedCell("N/A"))
	}
	fmt.Fprintln(writer)
}

func renderFilteredSpeedtest(writer io.Writer, run func(io.Writer)) {
	writer = writerOrDiscard(writer)
	var output bytes.Buffer
	if run != nil {
		run(&output)
	}
	writeFilteredSpeedtestOutput(writer, output.String())
}

func renderFilteredSpeedtestWithError(writer io.Writer, run func(io.Writer) error) error {
	writer = writerOrDiscard(writer)
	var output bytes.Buffer
	var err error
	if run != nil {
		err = run(&output)
	}
	writeFilteredSpeedtestOutput(writer, output.String())
	return err
}

// writeFilteredSpeedtestOutput removes only the exact non-result sentinel.
// Other warnings and diagnostic lines remain visible.
func writeFilteredSpeedtestOutput(writer io.Writer, output string) (wrote, renderedNearby bool) {
	writer = writerOrDiscard(writer)
	output = strings.ReplaceAll(output, "\r\n", "\n")
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == suppressedSpeedtestLine {
			continue
		}
		fmt.Fprintln(writer, line)
		wrote = true
		if fields := strings.Fields(trimmed); len(fields) > 0 && fields[0] == "Speedtest.net" {
			renderedNearby = true
		}
	}
	return wrote, renderedNearby
}

func formatSpeedCell(value string) string {
	return fmt.Sprintf("%-16s", value)
}
