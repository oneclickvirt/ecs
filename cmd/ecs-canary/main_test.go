package main

import (
	"testing"
	"time"
)

func TestCanaryConfigDefaultsToDataOnly(t *testing.T) {
	config := canaryConfig(false, false, false, false, "", 0, 30*time.Second)
	if config.BasicStatus || config.CpuTestStatus || config.MemoryTestStatus || config.DiskTestStatus ||
		config.UtTestStatus || config.SecurityTestStatus || config.EmailTestStatus || config.BacktraceStatus ||
		config.Nt3Status || config.PingTestStatus || config.TgdcTestStatus || config.WebTestStatus ||
		config.SpeedTestStatus || config.TCPProbeStatus || canaryRunsHardware(config) {
		t.Fatalf("default canary enabled a probe: %#v", config)
	}
	if !config.PrivacyMode || config.EnableUpload || config.DeepSMARTDevices != "" || config.DeepGPUDevice != "" {
		t.Fatalf("default canary is not private and non-uploading: %#v", config)
	}
}

func TestCanaryConfigAllowsOnlyExplicitSafeDeepTargets(t *testing.T) {
	config := canaryConfig(true, false, false, false, "/tmp", 5*time.Second, 20*time.Second)
	if !config.DeepMode || config.DeepDiskPaths != "/tmp" || config.DeepBurnDuration != 5*time.Second || !canaryRunsHardware(config) {
		t.Fatalf("explicit deep canary was not configured: %#v", config)
	}
	if config.DeepSMARTDevices != "" || config.DeepGPUDevice != "" || config.HardwareBudget != 20*time.Second {
		t.Fatalf("deep canary escaped its safety boundary: %#v", config)
	}
}

func TestCanaryStandardProfileEnablesEveryStandardSection(t *testing.T) {
	config := canaryConfig(false, true, false, false, "", 0, 10*time.Minute)
	if !config.BasicStatus || !config.CpuTestStatus || !config.MemoryTestStatus || !config.DiskTestStatus ||
		!config.UtTestStatus || !config.SecurityTestStatus || !config.EmailTestStatus || !config.BacktraceStatus ||
		!config.Nt3Status || !config.PingTestStatus || !config.TgdcTestStatus || !config.WebTestStatus ||
		!config.SpeedTestStatus || !config.TCPProbeStatus || !canaryRunsHardware(config) {
		t.Fatalf("standard canary omitted a section: %#v", config)
	}
	if config.DeepMode || !config.PrivacyMode || config.EnableUpload || config.HardwareBudget != 2*time.Minute {
		t.Fatalf("standard canary escaped its safety boundary: %#v", config)
	}
}
