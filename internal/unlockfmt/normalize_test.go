package unlockfmt

import (
	"strings"
	"testing"
)

func TestNormalizePrefixesLegacyHeaders(t *testing.T) {
	raw := "IPV4:\n============[ 跨国平台 ]============\nNetflix                   YES\n"
	got := Normalize("ipv4", raw)
	if strings.Contains(got, "IPV4:\n") {
		t.Fatalf("expected standalone IPV4 label to be removed, got %q", got)
	}
	if !strings.Contains(got, "[ IPV4 跨国平台 ]") {
		t.Fatalf("expected legacy header to be prefixed with IP version, got %q", got)
	}
	if !strings.Contains(got, "Netflix") {
		t.Fatalf("expected provider row to remain, got %q", got)
	}
}

func TestNormalizeKeepsVersionedHeaders(t *testing.T) {
	raw := "========[ IPV6 Global ]=========\nClaude                    YES\n"
	got := Normalize("ipv6", raw)
	if strings.Count(got, "IPV6 Global") != 1 {
		t.Fatalf("expected versioned header to remain unchanged, got %q", got)
	}
	if strings.Contains(got, "IPV6 IPV6") {
		t.Fatalf("expected version prefix not to be duplicated, got %q", got)
	}
}

func TestNormalizeRemovesColoredLegacyLabels(t *testing.T) {
	raw := "\x1b[34mIPV6:\x1b[0m\n============[ 跨国平台 ]============\nClaude                    YES\n"
	got := Normalize("ipv6", raw)
	if strings.Contains(got, "IPV6:\n") || strings.Contains(got, "\x1b[34mIPV6") {
		t.Fatalf("expected colored standalone IPV6 label to be removed, got %q", got)
	}
	if !strings.Contains(got, "[ IPV6 跨国平台 ]") {
		t.Fatalf("expected legacy header to be prefixed with IPv6, got %q", got)
	}
}
