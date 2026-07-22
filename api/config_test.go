package api

import "testing"

func TestApplyOptionsValidatesConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	ApplyOptions(cfg,
		WithLanguage("EN"),
		WithSpeedTestNum(0),
		WithNt3Location("all"),
		WithUnlockTestIPVersion("IPV6"),
		WithTCPTextFormat("FULL"),
		WithPingSortOrder("NAME"),
		WithPingScope("CHINA"),
		WithTCPSortOrder("LATENCY"),
		nil,
	)

	if cfg.Language != "en" {
		t.Fatalf("Language = %q, want en", cfg.Language)
	}
	if cfg.SpNum != 2 {
		t.Fatalf("SpNum = %d, want default 2", cfg.SpNum)
	}
	if cfg.Nt3Location != "ALL" {
		t.Fatalf("Nt3Location = %q, want ALL", cfg.Nt3Location)
	}
	if cfg.UnlockTestIPVersion != "ipv6" {
		t.Fatalf("UnlockTestIPVersion = %q, want ipv6", cfg.UnlockTestIPVersion)
	}
	if cfg.TCPTextFormat != "full" {
		t.Fatalf("TCPTextFormat = %q, want full", cfg.TCPTextFormat)
	}
	if cfg.PingSortOrder != "name" || cfg.PingScope != "international" || cfg.TCPSortOrder != "latency" {
		t.Fatalf("network ordering options were not normalized: ping=%q scope=%q tcp=%q", cfg.PingSortOrder, cfg.PingScope, cfg.TCPSortOrder)
	}
}

func TestApplyOptionsAllowsNilConfig(t *testing.T) {
	if ApplyOptions(nil, WithLanguage("en")) != nil {
		t.Fatalf("nil config should stay nil")
	}
}
