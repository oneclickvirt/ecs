package menu

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/oneclickvirt/ecs/internal/params"
	"github.com/oneclickvirt/ecs/utils"
)

func advNameColWidth(advanced []advSetting, lang string) int {
	maxW := 0
	for _, a := range advanced {
		name := a.nameEn
		if lang == "zh" {
			name = a.nameZh
		}
		w := runewidth.StringWidth(name + ":")
		if w > maxW {
			maxW = w
		}
	}
	return maxW + 1 // +1 for guaranteed spacing between name and value
}

// ensureMainCursorVisible adjusts mainScrollOffset so the cursor row is
// within the visible viewport.
func (m *tuiModel) ensureMainCursorVisible() {
	// header ≈ 9 lines (max: blank+title+blank+version+stats+blank+label+blank+blank)
	// footer ≈ 8 lines (blank+panelTitle+panel(3)+blank+help+blank)
	const hdrEst = 9
	const ftrEst = 8
	avail := m.height - hdrEst - ftrEst - 1 // -1 for possible scroll indicator
	if avail < 4 || m.height == 0 {
		return
	}
	// body: mainItems(12) + blank(1) + sectionHdr(1) + quickOpts(2) = 16
	totalBody := len(m.mainItems) + 1 + 1 + m.mainExtraTotal
	maxOff := totalBody - avail
	if maxOff < 0 {
		maxOff = 0
	}
	// cursor → body row: +2 offset for quick-option entries (blank+sectionHdr before them)
	curBodyRow := m.mainCursor
	if m.mainCursor >= len(m.mainItems) {
		curBodyRow = m.mainCursor + 2
	}
	if curBodyRow < m.mainScrollOffset {
		m.mainScrollOffset = curBodyRow
	} else if curBodyRow >= m.mainScrollOffset+avail {
		m.mainScrollOffset = curBodyRow - avail + 1
	}
	if m.mainScrollOffset > maxOff {
		m.mainScrollOffset = maxOff
	}
	if m.mainScrollOffset < 0 {
		m.mainScrollOffset = 0
	}
}

// clampMainScroll clamps mainScrollOffset to valid range after a resize.
func (m *tuiModel) clampMainScroll() {
	const hdrEst = 9
	const ftrEst = 8
	avail := m.height - hdrEst - ftrEst - 1
	if avail < 4 || m.height == 0 {
		m.mainScrollOffset = 0
		return
	}
	totalBody := len(m.mainItems) + 1 + 1 + m.mainExtraTotal
	maxOff := totalBody - avail
	if maxOff < 0 {
		maxOff = 0
	}
	if m.mainScrollOffset > maxOff {
		m.mainScrollOffset = maxOff
	}
	if m.mainScrollOffset < 0 {
		m.mainScrollOffset = 0
	}
}

// customCursorToBodyLine maps a custom-menu cursor index to the body line index.
// Body layout:
//
//	line 0  : toggle section header
//	line 1  : blank
//	lines 2..2+nT-1  : nT toggle items  (cursor 0..nT-1)
//	line 2+nT        : blank
//	line 2+nT+1      : advanced section header
//	line 2+nT+2      : blank
//	lines 2+nT+3..2+nT+3+nA-1 : nA advanced items (cursor nT..nT+nA-1)
//	line 2+nT+3+nA   : blank before confirm
//	line 2+nT+3+nA+1 : confirm button (cursor nT+nA = customTotal-1)
func (m tuiModel) customCursorToBodyLine(cursor int) int {
	nT := len(m.toggles)
	nA := len(m.advanced)
	if cursor < nT {
		return cursor + 2
	}
	if cursor == m.customTotal-1 {
		return nT + nA + 6 // 2+nT+3+nA+1 = nT+nA+6
	}
	// advanced item
	return cursor + 5 // cursor - nT + (2 + nT + 3) = cursor + 5
}

// ensureCustomCursorVisible adjusts customScrollOffset so the cursor row is visible.
func (m *tuiModel) ensureCustomCursorVisible() {
	const hdrEst = 6
	const ftrEst = 9
	avail := m.height - hdrEst - ftrEst - 1
	if avail < 4 || m.height == 0 {
		return
	}
	nT := len(m.toggles)
	nA := len(m.advanced)
	// total body lines = 2 + nT + 1 + 1 + 1 + nA + 1 + 1 = nT + nA + 7
	totalBody := nT + nA + 7
	maxOff := totalBody - avail
	if maxOff < 0 {
		maxOff = 0
	}
	curBodyLine := m.customCursorToBodyLine(m.customCursor)
	if curBodyLine < m.customScrollOffset {
		m.customScrollOffset = curBodyLine
	} else if curBodyLine >= m.customScrollOffset+avail {
		m.customScrollOffset = curBodyLine - avail + 1
	}
	if m.customScrollOffset > maxOff {
		m.customScrollOffset = maxOff
	}
	if m.customScrollOffset < 0 {
		m.customScrollOffset = 0
	}
}

// clampCustomScroll clamps customScrollOffset to valid range after a resize.
func (m *tuiModel) clampCustomScroll() {
	const hdrEst = 6
	const ftrEst = 9
	avail := m.height - hdrEst - ftrEst - 1
	if avail < 4 || m.height == 0 {
		m.customScrollOffset = 0
		return
	}
	totalBody := len(m.toggles) + len(m.advanced) + 7
	maxOff := totalBody - avail
	if maxOff < 0 {
		maxOff = 0
	}
	if m.customScrollOffset > maxOff {
		m.customScrollOffset = maxOff
	}
	if m.customScrollOffset < 0 {
		m.customScrollOffset = 0
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func RunTuiMenu(preCheck utils.NetCheckResult, config *params.Config) tuiResult {
	var statsTotal, statsDaily int
	var hasStats bool
	var cmpVersion int
	var newVersion string
	if preCheck.Connected {
		var wg sync.WaitGroup
		var stats *utils.StatsResponse
		var statsErr error
		var githubInfo *utils.GitHubRelease
		var githubErr error
		wg.Add(2)
		go func() {
			defer wg.Done()
			stats, statsErr = utils.GetGoescStats()
		}()
		go func() {
			defer wg.Done()
			githubInfo, githubErr = utils.GetLatestEcsRelease()
		}()
		wg.Wait()
		if statsErr == nil {
			statsTotal = stats.Total
			statsDaily = stats.Daily
			hasStats = true
		}
		if githubErr == nil {
			cmpVersion = utils.CompareVersions(config.EcsVersion, githubInfo.TagName)
			newVersion = githubInfo.TagName
		}
	}
	langPreset := config.UserSetFlags["lang"] || config.UserSetFlags["l"]
	m := newTuiModel(preCheck, config, langPreset, statsTotal, statsDaily, hasStats, cmpVersion, newVersion)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running menu: %v\n", err)
		os.Exit(1)
	}
	return finalModel.(tuiModel).result
}

func applyCustomResult(result tuiResult, preCheck utils.NetCheckResult, config *params.Config) {
	for _, t := range result.toggles {
		enabled := t.enabled
		if t.needNet && !preCheck.Connected {
			enabled = false
		}
		switch t.key {
		case "basic":
			config.BasicStatus = enabled
		case "cpu":
			config.CpuTestStatus = enabled
		case "memory":
			config.MemoryTestStatus = enabled
		case "disk":
			config.DiskTestStatus = enabled
		case "ut":
			config.UtTestStatus = enabled
		case "security":
			config.SecurityTestStatus = enabled
		case "email":
			config.EmailTestStatus = enabled
		case "backtrace":
			config.BacktraceStatus = enabled
		case "nt3":
			config.Nt3Status = enabled
		case "speed":
			config.SpeedTestStatus = enabled
		case "ping":
			config.PingTestStatus = enabled
		case "tgdc":
			config.TgdcTestStatus = enabled
		case "web":
			config.WebTestStatus = enabled
		}
	}

	for _, a := range result.advanced {
		switch a.key {
		case "cpum":
			config.CpuTestMethod = a.options[a.current].value
		case "cput":
			config.CpuTestThreadMode = a.options[a.current].value
		case "memorym":
			config.MemoryTestMethod = a.options[a.current].value
		case "diskm":
			config.DiskTestMethod = a.options[a.current].value
		case "diskp":
			config.DiskTestPath = strings.TrimSpace(a.textVal)
		case "diskmc":
			config.DiskMultiCheck = a.boolVal
		case "autodiskm":
			config.AutoChangeDiskMethod = a.boolVal
		case "deep":
			config.DeepMode = a.boolVal
		case "deepdiskpaths":
			config.DeepDiskPaths = strings.TrimSpace(a.textVal)
		case "deepsmartdevices":
			config.DeepSMARTDevices = strings.TrimSpace(a.textVal)
		case "deepburnduration":
			if value := strings.TrimSpace(a.textVal); value == "" {
				config.DeepBurnDuration = 0
			} else if duration, err := time.ParseDuration(value); err == nil {
				config.DeepBurnDuration = duration
			}
		case "deepgpudevice":
			config.DeepGPUDevice = strings.TrimSpace(a.textVal)
		case "timeout":
			if duration, err := time.ParseDuration(strings.TrimSpace(a.textVal)); err == nil {
				config.MaxDuration = duration
			}
		case "hardwarebudget":
			if duration, err := time.ParseDuration(strings.TrimSpace(a.textVal)); err == nil {
				config.HardwareBudget = duration
			}
		case "nt3loc":
			config.Nt3Location = a.options[a.current].value
		case "nt3t":
			config.Nt3CheckType = a.options[a.current].value
		case "spnum":
			if v, err := strconv.Atoi(a.options[a.current].value); err == nil {
				config.SpNum = v
			}
		case "unlockregion":
			config.UnlockTestRegion = a.options[a.current].value
		case "unlockshowip":
			config.UnlockTestShowIP = a.boolVal
		case "unlockipver":
			config.UnlockTestIPVersion = a.options[a.current].value
		case "utinterface":
			config.UnlockTestInterface = strings.TrimSpace(a.textVal)
		case "utdns":
			config.UnlockTestDNSServers = strings.TrimSpace(a.textVal)
		case "uthttpproxy":
			config.UnlockTestHTTPProxy = strings.TrimSpace(a.textVal)
		case "utsocksproxy":
			config.UnlockTestSOCKSProxy = strings.TrimSpace(a.textVal)
		case "utconcurrency":
			if value, err := strconv.Atoi(strings.TrimSpace(a.textVal)); err == nil {
				config.UnlockTestConcurrency = value
			}
		case "log":
			config.EnableLogger = a.boolVal
		case "upload":
			config.EnableUpload = a.boolVal
		case "analysis":
			config.AnalyzeResult = a.boolVal
		case "filepath":
			if strings.TrimSpace(a.textVal) != "" {
				config.FilePath = strings.TrimSpace(a.textVal)
			}
		case "privacy":
			config.PrivacyMode = a.boolVal
		case "tcp":
			config.TCPProbeStatus = a.boolVal
		case "tcpformat":
			config.TCPTextFormat = a.options[a.current].value
		case "pingsort":
			config.PingSortOrder = a.options[a.current].value
		case "pingscope":
			config.PingScope = a.options[a.current].value
		case "tcpsort":
			config.TCPSortOrder = a.options[a.current].value
		case "jsonpath":
			config.JSONPath = strings.TrimSpace(a.textVal)
		case "dataoffline":
			config.DataOffline = a.boolVal
		}
	}

	if !config.BasicStatus && !config.CpuTestStatus && !config.MemoryTestStatus && !config.DiskTestStatus {
		config.OnlyIpInfoCheck = true
	}
}
