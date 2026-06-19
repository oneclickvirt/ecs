package main

import (
	"testing"

	"github.com/oneclickvirt/ecs/internal/params"
)

func TestApplyEnvironmentDefaultsDisablesMenuWhenNonInteractive(t *testing.T) {
	t.Setenv("noninteractive", "true")
	cfg := params.NewConfig("test")

	applyEnvironmentDefaults(cfg)

	if cfg.MenuMode {
		t.Fatalf("MenuMode should be disabled in non-interactive mode")
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
