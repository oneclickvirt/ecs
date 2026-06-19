package params

import "testing"

func TestParseFlagsNormalizesAndValidatesValues(t *testing.T) {
	cfg := NewConfig("test")

	cfg.ParseFlags([]string{
		"-menu", "false",
		"-l", "EN",
		"-cpum", "GEEKBENCH",
		"-cput", "SINGLE",
		"-memorym", "AUTO",
		"-diskm", "DD",
		"-nt3loc", "all",
		"-nt3t", "BOTH",
		"-utipver", "IPV6",
		"-spnum", "0",
	})

	if cfg.MenuMode {
		t.Fatalf("space-separated bool flag should be parsed")
	}
	if cfg.Language != "en" {
		t.Fatalf("Language = %q, want en", cfg.Language)
	}
	if cfg.CpuTestMethod != "geekbench" {
		t.Fatalf("CpuTestMethod = %q, want geekbench", cfg.CpuTestMethod)
	}
	if cfg.CpuTestThreadMode != "single" {
		t.Fatalf("CpuTestThreadMode = %q, want single", cfg.CpuTestThreadMode)
	}
	if cfg.MemoryTestMethod != "auto" {
		t.Fatalf("MemoryTestMethod = %q, want auto", cfg.MemoryTestMethod)
	}
	if cfg.DiskTestMethod != "dd" {
		t.Fatalf("DiskTestMethod = %q, want dd", cfg.DiskTestMethod)
	}
	if cfg.Nt3Location != "ALL" {
		t.Fatalf("Nt3Location = %q, want ALL", cfg.Nt3Location)
	}
	if cfg.Nt3CheckType != "both" {
		t.Fatalf("Nt3CheckType = %q, want both", cfg.Nt3CheckType)
	}
	if cfg.UnlockTestIPVersion != "ipv6" {
		t.Fatalf("UnlockTestIPVersion = %q, want ipv6", cfg.UnlockTestIPVersion)
	}
	if cfg.SpNum != 2 {
		t.Fatalf("SpNum = %d, want default 2", cfg.SpNum)
	}
}

func TestParseFlagsCanBeCalledRepeatedly(t *testing.T) {
	cfg := NewConfig("test")

	cfg.ParseFlags([]string{"-menu=false", "-l=en"})
	cfg.ParseFlags([]string{"-menu=true", "-l=zh"})

	if !cfg.MenuMode {
		t.Fatalf("MenuMode = false, want true after second parse")
	}
	if cfg.Language != "zh" {
		t.Fatalf("Language = %q, want zh after second parse", cfg.Language)
	}
}
