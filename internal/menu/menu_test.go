package menu

import (
	"testing"
	"time"

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

func TestFullPresetEnablesEnhancedChecks(t *testing.T) {
	cfg := params.NewConfig("test")
	applyMenuResult(utils.NetCheckResult{Connected: true, StackType: "IPv4"}, cfg, tuiResult{
		choice:     "1",
		language:   "zh",
		mainUpload: cfg.EnableUpload,
	}, nil)

	if !cfg.DiskMultiCheck || !cfg.DeepMode || cfg.DeepBurnDuration != 20*time.Second || !cfg.TCPProbeStatus || cfg.TCPTextFormat != "compact" || !cfg.UnlockTestShowIP || !cfg.PingTestStatus {
		t.Fatalf("full preset did not enable enhanced checks: disk_multi=%t deep=%t burn=%s tcp=%t tcp_format=%s show_ip=%t ping=%t", cfg.DiskMultiCheck, cfg.DeepMode, cfg.DeepBurnDuration, cfg.TCPProbeStatus, cfg.TCPTextFormat, cfg.UnlockTestShowIP, cfg.PingTestStatus)
	}
}

func TestFullPresetRespectsExplicitEnhancedFlagOverrides(t *testing.T) {
	cfg := params.NewConfig("test")
	cfg.ParseFlags([]string{"-deep=false", "-diskmc=false", "-tcp=false", "-tcp-format=full", "-utshowip=false", "-deep-burn-duration=0s"})
	saved := cfg.SaveUserSetParams()
	applyMenuResult(utils.NetCheckResult{Connected: true, StackType: "IPv4"}, cfg, tuiResult{
		choice:     "1",
		language:   "zh",
		mainUpload: cfg.EnableUpload,
	}, saved)

	if cfg.DiskMultiCheck || cfg.DeepMode || cfg.DeepBurnDuration != 0 || cfg.TCPProbeStatus || cfg.TCPTextFormat != "full" || cfg.UnlockTestShowIP {
		t.Fatalf("explicit enhanced flag overrides were ignored: disk_multi=%t deep=%t burn=%s tcp=%t tcp_format=%s show_ip=%t", cfg.DiskMultiCheck, cfg.DeepMode, cfg.DeepBurnDuration, cfg.TCPProbeStatus, cfg.TCPTextFormat, cfg.UnlockTestShowIP)
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

func TestCustomAdvancedCarriesStructuredRuntimeParameters(t *testing.T) {
	cfg := params.NewConfig("test")
	advanced := defaultAdvSettings(cfg)
	for index := range advanced {
		switch advanced[index].key {
		case "deep":
			advanced[index].boolVal = true
		case "deepdiskpaths":
			advanced[index].textVal = "/mnt/data"
		case "deepburnduration":
			advanced[index].textVal = "30s"
		case "timeout":
			advanced[index].textVal = "10m"
		case "hardwarebudget":
			advanced[index].textVal = "3m"
		case "utinterface":
			advanced[index].textVal = "eth0"
		case "utdns":
			advanced[index].textVal = "1.1.1.1"
		case "utconcurrency":
			advanced[index].textVal = "12"
		case "dataoffline":
			advanced[index].boolVal = true
		case "privacy":
			advanced[index].boolVal = true
		case "tcpformat":
			advanced[index].current = optionIndexByValue(advanced[index].options, "full")
		case "pingsort":
			advanced[index].current = optionIndexByValue(advanced[index].options, "name")
		case "pingscope":
			advanced[index].current = optionIndexByValue(advanced[index].options, "international")
		case "tcpsort":
			advanced[index].current = optionIndexByValue(advanced[index].options, "latency")
		}
	}
	applyCustomResult(tuiResult{toggles: defaultTestToggles(), advanced: advanced}, utils.NetCheckResult{Connected: true}, cfg)
	cfg.ValidateParams()
	if !cfg.DeepMode || cfg.DeepDiskPaths != "/mnt/data" || cfg.DeepBurnDuration != 30*time.Second {
		t.Fatalf("deep parameters were not applied: %#v", cfg)
	}
	if cfg.MaxDuration != 10*time.Minute || cfg.HardwareBudget != 3*time.Minute {
		t.Fatalf("budgets were not applied: max=%s hardware=%s", cfg.MaxDuration, cfg.HardwareBudget)
	}
	if cfg.UnlockTestInterface != "eth0" || cfg.UnlockTestDNSServers != "1.1.1.1" || cfg.UnlockTestConcurrency != 12 {
		t.Fatalf("unlock network parameters were not applied: %#v", cfg)
	}
	if !cfg.DataOffline || !cfg.PrivacyMode || cfg.EnableUpload {
		t.Fatalf("data/privacy parameters were not applied: %#v", cfg)
	}
	if cfg.TCPTextFormat != "full" {
		t.Fatalf("TCP text format was not applied: %q", cfg.TCPTextFormat)
	}
	if cfg.PingSortOrder != "name" || cfg.PingScope != "international" || cfg.TCPSortOrder != "latency" {
		t.Fatalf("network ordering settings were not applied: ping=%q scope=%q tcp=%q", cfg.PingSortOrder, cfg.PingScope, cfg.TCPSortOrder)
	}
}

func TestEnglishCustomPingScopeCannotSelectMainlandChina(t *testing.T) {
	cfg := params.NewConfig("test")
	cfg.Language = "en"
	advanced := defaultAdvSettings(cfg)
	for index := range advanced {
		if advanced[index].key == "pingscope" {
			advanced[index].current = optionIndexByValue(advanced[index].options, "china")
		}
	}
	applyCustomResult(tuiResult{toggles: defaultTestToggles(), advanced: advanced}, utils.NetCheckResult{Connected: true}, cfg)
	cfg.ValidateParams()
	if cfg.PingScope != "international" {
		t.Fatalf("English custom Ping scope = %q, want international", cfg.PingScope)
	}
}

func TestCustomAdvancedIncludesAIOnlyUnlockRegion(t *testing.T) {
	setting := findAdvanced(t, defaultAdvSettings(params.NewConfig("test")), "unlockregion")
	if optionIndexByValue(setting.options, "21") == 0 {
		t.Fatalf("AI-only unlock region is missing: %#v", setting.options)
	}
}
