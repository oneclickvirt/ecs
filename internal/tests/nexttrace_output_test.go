package tests

import (
	"strings"
	"testing"
)

func TestFormatNextTraceOutputSeparatesStopReasonAndNextHeader(t *testing.T) {
	lines := []string{
		"\x1b[33m\x1b[01m广州电信 - ICMP v4 -\x1b[0m",
		"traceroute to 58.60.188.222, 30 hops max, 52 byte packets",
		"1.00 ms AS4134 hop",
		"Trace Stopped: Destination Reached at Hop 3 (ICMP Echo Reply)",
		"\x1b[33m\x1b[01m广州移动 - ICMP v4 -\x1b[0m",
		"traceroute to 120.196.165.24, 30 hops max, 52 byte packets",
	}

	got := formatNextTraceOutput(lines)
	if strings.Contains(got, "Trace Stopped: Destination Reached") {
		t.Fatalf("destination-reached stop line should be filtered: %q", got)
	}
	if strings.Contains(got, "hop广州移动") {
		t.Fatalf("next carrier header was concatenated: %q", got)
	}
	if !strings.Contains(got, "广州移动 - ICMP v4 -\x1b[0mtraceroute to 120.196.165.24") {
		t.Fatalf("header and traceroute body were not joined: %q", got)
	}
}

func TestFormatNextTraceOutputFiltersMaximumHopsReached(t *testing.T) {
	got := formatNextTraceOutput([]string{
		"Trace Stopped: Maximum Hops Reached at Hop 30 (No Destination Response)",
		"广州电信 - ICMP v6 -",
		"traceroute to 240e:e1:aa00:4000::24, 30 hops max, 52 byte packets",
	})
	if strings.Contains(got, "Maximum Hops Reached") {
		t.Fatalf("maximum-hop stop line should be filtered: %q", got)
	}
	if !strings.Contains(got, "广州电信 - ICMP v6 -traceroute to") {
		t.Fatalf("next trace header/body was not preserved: %q", got)
	}
}

func TestFormatNextTraceOutputTerminatesHeaderOnlyAndErrorLines(t *testing.T) {
	got := formatNextTraceOutput([]string{
		"广州电信 - ICMP v4 -",
		"Error: ICMP traceroute unavailable",
	})
	if got != "广州电信 - ICMP v4 -\nError: ICMP traceroute unavailable\n" {
		t.Fatalf("unexpected header/error formatting: %q", got)
	}
}
