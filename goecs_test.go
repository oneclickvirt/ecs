package main

import (
	"testing"

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
