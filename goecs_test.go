package main

import (
	"testing"
	"time"

	"github.com/oneclickvirt/ecs/internal/params"
)

func TestApplyEnvironmentDefaultsPreservesMenuWhenNonInteractive(t *testing.T) {
	// noninteractive env var should NOT disable the menu.
	// Menu should only be disabled via explicit CLI flag -menu=false.
	// The noninteractive env var is used by goecs.sh install script
	// and should not leak to affect goecs runtime behavior.
	t.Setenv("noninteractive", "true")
	cfg := params.NewConfig("test")

	applyEnvironmentDefaults(cfg)

	if !cfg.MenuMode {
		t.Fatalf("MenuMode should remain enabled even in non-interactive mode; menu is controlled by -menu flag only")
	}
}

func TestApplyEnvironmentDefaultsKeepsExplicitMenuFlag(t *testing.T) {
	t.Setenv("noninteractive", "true")
	cfg := params.NewConfig("test")
	cfg.UserSetFlags["menu"] = true
	cfg.MenuMode = true

	applyEnvironmentDefaults(cfg)

	if !cfg.MenuMode {
		t.Fatalf("explicit menu flag should be preserved")
	}
}

func TestApplyEnvironmentDefaultsRespectsExplicitMenuFalse(t *testing.T) {
	// Explicit -menu=false should still work
	cfg := params.NewConfig("test")
	cfg.UserSetFlags["menu"] = true
	cfg.MenuMode = false

	applyEnvironmentDefaults(cfg)

	if cfg.MenuMode {
		t.Fatalf("explicit -menu=false should be respected")
	}
}

func TestStructuredCLIRequiresExplicitJSONOutput(t *testing.T) {
	cfg := params.NewConfig("test")
	if shouldRunStructuredCLI(cfg) {
		t.Fatal("structured adapters must not replace the default streaming text runner")
	}
	for _, path := range []string{"-", "report.json"} {
		cfg.JSONPath = path
		if !shouldRunStructuredCLI(cfg) {
			t.Fatalf("JSON path %q did not select structured CLI mode", path)
		}
	}
}

func TestLegacyDeadlineKeepsOneCleanupWindow(t *testing.T) {
	for _, test := range []struct {
		maximum, soft time.Duration
	}{
		{maximum: 15 * time.Second, soft: 12 * time.Second},
		{maximum: 15 * time.Minute, soft: 14*time.Minute + 30*time.Second},
	} {
		soft, hard := legacyDeadlineWindows(test.maximum)
		if soft != test.soft || hard != test.maximum {
			t.Fatalf("legacyDeadlineWindows(%s) = %s, %s", test.maximum, soft, hard)
		}
	}
}

func TestUploadDisabledDoesNotDisableSecuritySection(t *testing.T) {
	previous := configs
	defer func() { configs = previous }()
	configs = params.NewConfig("test")
	configs.Language = "zh"
	configs.EnableUpload = false
	configs.SecurityTestStatus = true
	handleLanguageSpecificSettings()
	if !configs.SecurityTestStatus {
		t.Fatal("security section was disabled when only result upload was disabled")
	}
}
