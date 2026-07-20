package params

import (
	"testing"
	"time"
)

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

func TestNT3DefaultsToBothAndAllowsExplicitSingleStack(t *testing.T) {
	if got := NewConfig("test").Nt3CheckType; got != "both" {
		t.Fatalf("default Nt3CheckType = %q, want both", got)
	}
	for _, value := range []string{"ipv4", "ipv6"} {
		cfg := NewConfig("test")
		cfg.ParseFlags([]string{"-nt3-type=" + value})
		if cfg.Nt3CheckType != value {
			t.Fatalf("explicit Nt3CheckType = %q, want %q", cfg.Nt3CheckType, value)
		}
	}
	invalid := NewConfig("test")
	invalid.Nt3CheckType = "invalid"
	invalid.ValidateParams()
	if invalid.Nt3CheckType != "both" {
		t.Fatalf("invalid Nt3CheckType fallback = %q, want both", invalid.Nt3CheckType)
	}
}

func TestValidateParamsCapsStandardHardwareBudget(t *testing.T) {
	cfg := NewConfig("test")
	cfg.MaxDuration = 15 * time.Minute
	cfg.HardwareBudget = 5 * time.Minute
	cfg.ValidateParams()
	if cfg.HardwareBudget != 2*time.Minute {
		t.Fatalf("HardwareBudget = %s, want 2m standard cap", cfg.HardwareBudget)
	}

	cfg.MaxDuration = 30 * time.Second
	cfg.HardwareBudget = 2 * time.Minute
	cfg.ValidateParams()
	if cfg.HardwareBudget != 30*time.Second {
		t.Fatalf("HardwareBudget = %s, want MaxDuration cap", cfg.HardwareBudget)
	}
}

func TestValidateParamsAllowsExplicitDeepHardwareBudget(t *testing.T) {
	cfg := NewConfig("test")
	cfg.DeepMode = true
	cfg.MaxDuration = 10 * time.Minute
	cfg.HardwareBudget = 5 * time.Minute
	cfg.ValidateParams()
	if cfg.HardwareBudget != 5*time.Minute {
		t.Fatalf("deep HardwareBudget = %s, want 5m", cfg.HardwareBudget)
	}
	cfg.HardwareBudget = 12 * time.Minute
	cfg.ValidateParams()
	if cfg.HardwareBudget != 10*time.Minute {
		t.Fatalf("deep HardwareBudget = %s, want MaxDuration cap", cfg.HardwareBudget)
	}
}

func TestValidateParamsDisablesExplicitDeepTargetsOutsideDeepMode(t *testing.T) {
	cfg := NewConfig("test")
	cfg.DeepDiskPaths = "/mnt/a,/mnt/b"
	cfg.DeepSMARTDevices = "/dev/sda"
	cfg.DeepBurnDuration = time.Minute
	cfg.DeepGPUDevice = "0"
	cfg.ValidateParams()
	if cfg.DeepDiskPaths != "" || cfg.DeepSMARTDevices != "" || cfg.DeepBurnDuration != 0 || cfg.DeepGPUDevice != "" {
		t.Fatalf("deep-only targets survived standard validation: %+v", cfg)
	}
}

func TestValidateParamsCapsExplicitDeepBurnToHardwareBudget(t *testing.T) {
	cfg := NewConfig("test")
	cfg.DeepMode = true
	cfg.HardwareBudget = 3 * time.Minute
	cfg.DeepBurnDuration = 5 * time.Minute
	cfg.ValidateParams()
	if cfg.DeepBurnDuration != 3*time.Minute {
		t.Fatalf("DeepBurnDuration = %s, want 3m", cfg.DeepBurnDuration)
	}
}

func TestParseFlagsAcceptsExplicitDeepTargets(t *testing.T) {
	cfg := NewConfig("test")
	cfg.ParseFlags([]string{
		"-deep", "-hardware-budget=4m", "-deep-disk-paths=/mnt/a,/mnt/b",
		"-deep-smart-devices=/dev/sda", "-deep-burn-duration=30s", "-deep-gpu-device=0",
	})
	if !cfg.DeepMode || cfg.DeepDiskPaths != "/mnt/a,/mnt/b" || cfg.DeepSMARTDevices != "/dev/sda" || cfg.DeepBurnDuration != 30*time.Second || cfg.DeepGPUDevice != "0" {
		t.Fatalf("explicit deep flags not retained: %+v", cfg)
	}
}
