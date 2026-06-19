package menu

import (
	"testing"

	"github.com/oneclickvirt/ecs/internal/params"
	"github.com/oneclickvirt/ecs/utils"
)

func findToggle(t *testing.T, toggles []testToggle, key string) testToggle {
	t.Helper()
	for _, toggle := range toggles {
		if toggle.key == key {
			return toggle
		}
	}
	t.Fatalf("toggle %q not found", key)
	return testToggle{}
}

func findAdvanced(t *testing.T, advanced []advSetting, key string) advSetting {
	t.Helper()
	for _, setting := range advanced {
		if setting.key == key {
			return setting
		}
	}
	t.Fatalf("advanced setting %q not found", key)
	return advSetting{}
}

func TestApplyMenuResultQuickOptionsSurviveSavedParamsRestore(t *testing.T) {
	cfg := params.NewConfig("test")
	cfg.AnalyzeResult = true
	cfg.UserSetFlags["analysis"] = true
	saved := cfg.SaveUserSetParams()

	applyMenuResult(utils.NetCheckResult{}, cfg, tuiResult{
		choice:      "2",
		language:    "zh",
		mainAnalyze: false,
		mainUpload:  false,
	}, saved)

	if cfg.AnalyzeResult {
		t.Fatalf("TUI quick analysis toggle should override restored saved analysis flag")
	}
	if cfg.EnableUpload {
		t.Fatalf("TUI quick upload toggle should be applied")
	}
}

func TestApplyMenuResultRestoresExplicitTestFlagForPreset(t *testing.T) {
	cfg := params.NewConfig("test")
	cfg.CpuTestStatus = false
	cfg.UserSetFlags["cpu"] = true
	saved := cfg.SaveUserSetParams()

	applyMenuResult(utils.NetCheckResult{}, cfg, tuiResult{
		choice:      "2",
		language:    "zh",
		mainAnalyze: cfg.AnalyzeResult,
		mainUpload:  cfg.EnableUpload,
	}, saved)

	if cfg.CpuTestStatus {
		t.Fatalf("explicit CLI cpu=false should override the selected preset")
	}
	if !cfg.BasicStatus || !cfg.MemoryTestStatus || !cfg.DiskTestStatus {
		t.Fatalf("minimal preset should still enable basic, memory and disk tests")
	}
}

func TestApplyMenuResultCustomUsesTuiResultAfterSavedParams(t *testing.T) {
	cfg := params.NewConfig("test")
	cfg.CpuTestStatus = false
	cfg.UserSetFlags["cpu"] = true
	saved := cfg.SaveUserSetParams()

	toggles := defaultTestToggles()
	for i := range toggles {
		toggles[i].enabled = toggles[i].key == "cpu"
	}
	applyMenuResult(utils.NetCheckResult{}, cfg, tuiResult{
		custom:   true,
		choice:   "custom",
		language: "zh",
		toggles:  toggles,
		advanced: defaultAdvSettings(cfg),
	}, saved)

	if !cfg.CpuTestStatus {
		t.Fatalf("advanced custom TUI result should be able to override a saved cpu flag")
	}
	if cfg.BasicStatus || cfg.MemoryTestStatus || cfg.DiskTestStatus {
		t.Fatalf("advanced custom toggles should be applied exactly")
	}
}

func TestDefaultTogglesRespectExplicitConfigFlags(t *testing.T) {
	cfg := params.NewConfig("test")
	cfg.UserSetFlags["web"] = true
	cfg.WebTestStatus = false

	toggles := defaultTestToggles()
	applyExplicitConfigToToggles(toggles, cfg)

	if findToggle(t, toggles, "web").enabled {
		t.Fatalf("explicit web=false should be reflected in custom-menu toggles")
	}
	if !findToggle(t, defaultTestToggles(), "web").enabled {
		t.Fatalf("default custom-menu web toggle should remain enabled without an explicit flag")
	}
}

func TestMainQuickOptionsSyncToCustomAdvanced(t *testing.T) {
	cfg := params.NewConfig("test")
	model := newTuiModel(utils.NetCheckResult{}, cfg, true, 0, 0, false, 0, "")
	model.mainAnalyze = true
	model.mainUpload = false

	model.syncMainQuickOptionsToAdvanced()

	if !findAdvanced(t, model.advanced, "analysis").boolVal {
		t.Fatalf("analysis advanced setting should follow main quick option")
	}
	if findAdvanced(t, model.advanced, "upload").boolVal {
		t.Fatalf("upload advanced setting should follow main quick option")
	}
}
