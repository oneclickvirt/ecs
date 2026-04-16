package menu

import (
	"fmt"
	"os"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/oneclickvirt/ecs/internal/params"
	"github.com/oneclickvirt/ecs/utils"
)

// --- Styles ---

var (
	tTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	tInfoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	tWarnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	tSelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	tNormStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	tDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	tHelpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	tSectStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	tChkOnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
	tChkOffStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	tBtnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("120")).Bold(true).Padding(0, 2)
	tBtnDimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("238")).Padding(0, 2)
)

// --- Types ---

type menuPhase int

const (
	phaseLang menuPhase = iota
	phaseMain
	phaseCustom
)

type mainMenuItem struct {
	id      string
	zh      string
	en      string
	needNet bool
}

type testToggle struct {
	nameZh  string
	nameEn  string
	enabled bool
	needNet bool
}

type advSetting struct {
	nameZh  string
	nameEn  string
	options []string
	current int
}

type tuiResult struct {
	choice   string
	language string
	quit     bool
	custom   bool
	toggles  []testToggle
	advanced []advSetting
}

type tuiModel struct {
	phase      menuPhase
	config     *params.Config
	preCheck   utils.NetCheckResult
	langPreset bool

	// Language selection
	langCursor int

	// Main menu
	mainCursor int
	mainItems  []mainMenuItem

	// Custom mode
	customCursor int
	toggles      []testToggle
	advanced     []advSetting
	customTotal  int // toggles + advanced + 1 (confirm button)

	// Pre-loaded info
	statsTotal int
	statsDaily int
	hasStats   bool
	cmpVersion int
	newVersion string

	// Result
	result tuiResult

	width  int
	height int
}

// --- Default data ---

func defaultMainItems() []mainMenuItem {
	return []mainMenuItem{
		{id: "1", zh: "融合怪完全体(能测全测)", en: "Full Test (All Available Tests)", needNet: false},
		{id: "2", zh: "极简版(系统信息+CPU+内存+磁盘+测速节点5个)", en: "Minimal Suite (SysInfo+CPU+Mem+Disk+5 Speed Nodes)", needNet: false},
		{id: "3", zh: "精简版(系统信息+CPU+内存+磁盘+解锁+路由+测速节点5个)", en: "Standard Suite (SysInfo+CPU+Mem+Disk+Unlock+Route+5 Speed Nodes)", needNet: false},
		{id: "4", zh: "精简网络版(系统信息+CPU+内存+磁盘+回程+路由+测速节点5个)", en: "Network Suite (SysInfo+CPU+Mem+Disk+Backtrace+Route+5 Speed Nodes)", needNet: false},
		{id: "5", zh: "精简解锁版(系统信息+CPU+内存+磁盘IO+解锁+测速节点5个)", en: "Unlock Suite (SysInfo+CPU+Mem+DiskIO+Unlock+5 Speed Nodes)", needNet: false},
		{id: "6", zh: "网络单项(IP质量+回程+路由+延迟+TGDC+网站+测速节点11个)", en: "Network Only (IPQuality+Backtrace+Route+Latency+TGDC+Web+11 Speed Nodes)", needNet: true},
		{id: "7", zh: "解锁单项(跨国平台解锁)", en: "Unlock Only (International Platform Unlock)", needNet: true},
		{id: "8", zh: "硬件单项(系统信息+CPU+内存+dd磁盘+fio磁盘)", en: "Hardware Only (SysInfo+CPU+Mem+DD Disk+FIO Disk)", needNet: false},
		{id: "9", zh: "IP质量检测(15个数据库+邮件端口检测)", en: "IP Quality (15 Databases + Email Port Test)", needNet: true},
		{id: "10", zh: "三网回程线路+路由+延迟+TGDC+网站延迟", en: "3-Net Backtrace+Route+Latency+TGDC+Websites", needNet: true},
		{id: "custom", zh: ">>> 自定义测试(自由选择测试项和高级设置)", en: ">>> Custom Test (Choose Tests & Advanced Settings)", needNet: false},
		{id: "0", zh: "退出程序", en: "Exit Program", needNet: false},
	}
}

func defaultTestToggles() []testToggle {
	return []testToggle{
		{nameZh: "基础系统信息", nameEn: "Basic System Info", enabled: true, needNet: false},
		{nameZh: "CPU测试", nameEn: "CPU Test", enabled: true, needNet: false},
		{nameZh: "内存测试", nameEn: "Memory Test", enabled: true, needNet: false},
		{nameZh: "磁盘测试", nameEn: "Disk Test", enabled: true, needNet: false},
		{nameZh: "跨国平台解锁", nameEn: "Streaming Unlock", enabled: false, needNet: true},
		{nameZh: "IP质量检测", nameEn: "IP Quality Check", enabled: false, needNet: true},
		{nameZh: "邮件端口检测", nameEn: "Email Port Check", enabled: false, needNet: true},
		{nameZh: "回程路由", nameEn: "Backtrace Route", enabled: false, needNet: true},
		{nameZh: "NT3路由", nameEn: "NT3 Route", enabled: false, needNet: true},
		{nameZh: "测速", nameEn: "Speed Test", enabled: false, needNet: true},
		{nameZh: "Ping测试", nameEn: "Ping Test", enabled: false, needNet: true},
		{nameZh: "TGDC测试", nameEn: "TGDC Test", enabled: false, needNet: true},
		{nameZh: "网站延迟", nameEn: "Website Latency", enabled: false, needNet: true},
	}
}

func defaultAdvSettings(config *params.Config) []advSetting {
	adv := []advSetting{
		{nameZh: "CPU测试方法", nameEn: "CPU Method", options: []string{"sysbench", "geekbench", "winsat"}, current: 0},
		{nameZh: "CPU线程模式", nameEn: "CPU Thread", options: []string{"multi", "single"}, current: 0},
		{nameZh: "内存测试方法", nameEn: "Memory Method", options: []string{"stream", "sysbench", "dd", "winsat", "auto"}, current: 0},
		{nameZh: "磁盘测试方法", nameEn: "Disk Method", options: []string{"fio", "dd", "winsat"}, current: 0},
		{nameZh: "NT3测试位置", nameEn: "NT3 Location", options: []string{"GZ", "SH", "BJ", "CD", "ALL"}, current: 0},
		{nameZh: "NT3测试类型", nameEn: "NT3 Type", options: []string{"ipv4", "ipv6", "both"}, current: 0},
		{nameZh: "测速节点数/运营商", nameEn: "Speed Nodes/ISP", options: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, current: 1},
	}
	// Set current values from config
	setAdvCurrent := func(idx int, val string) {
		for j, opt := range adv[idx].options {
			if opt == val {
				adv[idx].current = j
				return
			}
		}
	}
	setAdvCurrent(0, config.CpuTestMethod)
	setAdvCurrent(1, config.CpuTestThreadMode)
	setAdvCurrent(2, config.MemoryTestMethod)
	setAdvCurrent(3, config.DiskTestMethod)
	setAdvCurrent(4, config.Nt3Location)
	setAdvCurrent(5, config.Nt3CheckType)
	if config.SpNum >= 1 && config.SpNum <= 10 {
		adv[6].current = config.SpNum - 1
	}
	return adv
}

func newTuiModel(preCheck utils.NetCheckResult, config *params.Config, langPreset bool, statsTotal, statsDaily int, hasStats bool, cmpVersion int, newVersion string) tuiModel {
	toggles := defaultTestToggles()
	advanced := defaultAdvSettings(config)
	m := tuiModel{
		config:      config,
		preCheck:    preCheck,
		langPreset:  langPreset,
		mainItems:   defaultMainItems(),
		toggles:     toggles,
		advanced:    advanced,
		customTotal: len(toggles) + len(advanced) + 1,
		statsTotal:  statsTotal,
		statsDaily:  statsDaily,
		hasStats:    hasStats,
		cmpVersion:  cmpVersion,
		newVersion:  newVersion,
		width:       80,
		height:      24,
	}
	if langPreset {
		m.phase = phaseMain
		m.result.language = config.Language
	} else {
		m.phase = phaseLang
	}
	return m
}

// --- Bubbletea Interface ---

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch m.phase {
		case phaseLang:
			return m.updateLang(msg)
		case phaseMain:
			return m.updateMain(msg)
		case phaseCustom:
			return m.updateCustom(msg)
		}
	}
	return m, nil
}

func (m tuiModel) View() string {
	switch m.phase {
	case phaseLang:
		return m.viewLang()
	case phaseMain:
		return m.viewMain()
	case phaseCustom:
		return m.viewCustom()
	}
	return ""
}

// --- Language Selection ---

func (m tuiModel) updateLang(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.langCursor > 0 {
			m.langCursor--
		}
	case "down", "j":
		if m.langCursor < 1 {
			m.langCursor++
		}
	case "1":
		m.result.language = "zh"
		m.phase = phaseMain
	case "2":
		m.result.language = "en"
		m.phase = phaseMain
	case "enter":
		if m.langCursor == 0 {
			m.result.language = "zh"
		} else {
			m.result.language = "en"
		}
		m.phase = phaseMain
	case "q", "ctrl+c":
		m.result.quit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m tuiModel) viewLang() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(tTitleStyle.Render("  VPS融合怪测试 / VPS Fusion Monster Test"))
	s.WriteString("\n\n")
	s.WriteString(tInfoStyle.Render("  请选择语言 / Please select language:"))
	s.WriteString("\n\n")

	langs := []struct {
		label string
		desc  string
	}{
		{"1. 中文", "Chinese"},
		{"2. English", "英文"},
	}

	for i, l := range langs {
		cursor := "   "
		style := tNormStyle
		if m.langCursor == i {
			cursor = " > "
			style = tSelStyle
		}
		s.WriteString(fmt.Sprintf("%s%s  %s\n", cursor, style.Render(l.label), tDimStyle.Render(l.desc)))
	}

	s.WriteString("\n")
	s.WriteString(tHelpStyle.Render("  ↑/↓ Navigate  Enter Confirm  1/2 Quick-Select  q Quit"))
	s.WriteString("\n")
	return s.String()
}

// --- Main Menu ---

func (m tuiModel) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "k":
		if m.mainCursor > 0 {
			m.mainCursor--
		}
	case "down", "j":
		if m.mainCursor < len(m.mainItems)-1 {
			m.mainCursor++
		}
	case "home":
		m.mainCursor = 0
	case "end":
		m.mainCursor = len(m.mainItems) - 1
	case "enter":
		item := m.mainItems[m.mainCursor]
		if item.needNet && !m.preCheck.Connected {
			return m, nil // can't select network-dependent items without connection
		}
		if item.id == "custom" {
			m.phase = phaseCustom
			m.customCursor = 0
			return m, nil
		}
		m.result.choice = item.id
		return m, tea.Quit
	case "q", "ctrl+c":
		m.result.quit = true
		return m, tea.Quit
	default:
		// Number quick-select
		for i, item := range m.mainItems {
			if key == item.id {
				if item.needNet && !m.preCheck.Connected {
					return m, nil
				}
				if item.id == "custom" {
					m.phase = phaseCustom
					m.customCursor = 0
					return m, nil
				}
				m.mainCursor = i
				m.result.choice = item.id
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m tuiModel) viewMain() string {
	lang := m.result.language
	var s strings.Builder
	s.WriteString("\n")

	// Title
	if lang == "zh" {
		s.WriteString(tTitleStyle.Render(fmt.Sprintf("  VPS融合怪 %s", m.config.EcsVersion)))
	} else {
		s.WriteString(tTitleStyle.Render(fmt.Sprintf("  VPS Fusion Monster %s", m.config.EcsVersion)))
	}
	s.WriteString("\n")

	// Version warning
	if m.preCheck.Connected && m.cmpVersion == -1 {
		if lang == "zh" {
			s.WriteString(tWarnStyle.Render(fmt.Sprintf("  ! 检测到新版本 %s 如有必要请更新！", m.newVersion)))
		} else {
			s.WriteString(tWarnStyle.Render(fmt.Sprintf("  ! New version %s detected, update if necessary!", m.newVersion)))
		}
		s.WriteString("\n")
	}

	// Stats
	if m.preCheck.Connected && m.hasStats {
		if lang == "zh" {
			s.WriteString(tInfoStyle.Render(fmt.Sprintf("  总使用量: %s | 今日使用: %s",
				utils.FormatGoecsNumber(m.statsTotal), utils.FormatGoecsNumber(m.statsDaily))))
		} else {
			s.WriteString(tInfoStyle.Render(fmt.Sprintf("  Total Usage: %s | Daily Usage: %s",
				utils.FormatGoecsNumber(m.statsTotal), utils.FormatGoecsNumber(m.statsDaily))))
		}
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Section header
	if lang == "zh" {
		s.WriteString(tSectStyle.Render("  请选择测试方案:"))
	} else {
		s.WriteString(tSectStyle.Render("  Select Test Suite:"))
	}
	s.WriteString("\n\n")

	// Menu items
	for i, item := range m.mainItems {
		cursor := "   "
		style := tNormStyle
		if m.mainCursor == i {
			cursor = " > "
			style = tSelStyle
		}

		var label string
		if lang == "zh" {
			label = item.zh
		} else {
			label = item.en
		}

		// Number prefix
		prefix := ""
		switch {
		case item.id == "custom":
			prefix = ""
		case item.id == "0":
			prefix = " 0. "
		default:
			prefix = fmt.Sprintf("%2s. ", item.id)
		}

		// Disabled indicator for network-dependent items when no connection
		suffix := ""
		if item.needNet && !m.preCheck.Connected {
			style = tDimStyle
			if lang == "zh" {
				suffix = " [需要网络]"
			} else {
				suffix = " [No Network]"
			}
		}

		s.WriteString(fmt.Sprintf("%s%s%s\n", cursor, style.Render(prefix+label), tDimStyle.Render(suffix)))
	}

	s.WriteString("\n")
	if lang == "zh" {
		s.WriteString(tHelpStyle.Render("  ↑/↓/j/k 移动  Enter 确认  数字 快速选择  q 退出"))
	} else {
		s.WriteString(tHelpStyle.Render("  Up/Down/j/k Navigate  Enter Confirm  Number Quick-Select  q Quit"))
	}
	s.WriteString("\n")
	return s.String()
}

// --- Custom Mode ---

func (m tuiModel) updateCustom(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "k":
		if m.customCursor > 0 {
			m.customCursor--
		}
	case "down", "j":
		if m.customCursor < m.customTotal-1 {
			m.customCursor++
		}
	case "home":
		m.customCursor = 0
	case "end":
		m.customCursor = m.customTotal - 1
	case " ":
		if m.customCursor < len(m.toggles) {
			// Toggle test item
			t := &m.toggles[m.customCursor]
			if t.needNet && !m.preCheck.Connected {
				break // can't enable without network
			}
			t.enabled = !t.enabled
		} else if m.customCursor < len(m.toggles)+len(m.advanced) {
			// Cycle advanced setting forward
			idx := m.customCursor - len(m.toggles)
			a := &m.advanced[idx]
			a.current = (a.current + 1) % len(a.options)
		}
	case "right", "l":
		if m.customCursor >= len(m.toggles) && m.customCursor < len(m.toggles)+len(m.advanced) {
			idx := m.customCursor - len(m.toggles)
			a := &m.advanced[idx]
			a.current = (a.current + 1) % len(a.options)
		}
	case "left", "h":
		if m.customCursor >= len(m.toggles) && m.customCursor < len(m.toggles)+len(m.advanced) {
			idx := m.customCursor - len(m.toggles)
			a := &m.advanced[idx]
			a.current = (a.current - 1 + len(a.options)) % len(a.options)
		}
	case "a":
		// Toggle all test items
		allEnabled := true
		for _, t := range m.toggles {
			if !t.enabled && (!t.needNet || m.preCheck.Connected) {
				allEnabled = false
				break
			}
		}
		for i := range m.toggles {
			if m.toggles[i].needNet && !m.preCheck.Connected {
				continue
			}
			m.toggles[i].enabled = !allEnabled
		}
	case "enter":
		if m.customCursor == m.customTotal-1 {
			// Confirm button
			m.result.custom = true
			m.result.choice = "custom"
			m.result.toggles = m.toggles
			m.result.advanced = m.advanced
			return m, tea.Quit
		}
		// Toggle current item (same as space)
		if m.customCursor < len(m.toggles) {
			t := &m.toggles[m.customCursor]
			if t.needNet && !m.preCheck.Connected {
				break
			}
			t.enabled = !t.enabled
		} else if m.customCursor < len(m.toggles)+len(m.advanced) {
			idx := m.customCursor - len(m.toggles)
			a := &m.advanced[idx]
			a.current = (a.current + 1) % len(a.options)
		}
	case "esc":
		m.phase = phaseMain
		return m, nil
	case "q", "ctrl+c":
		m.result.quit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m tuiModel) viewCustom() string {
	lang := m.result.language
	var s strings.Builder
	s.WriteString("\n")

	if lang == "zh" {
		s.WriteString(tTitleStyle.Render("  自定义测试配置"))
	} else {
		s.WriteString(tTitleStyle.Render("  Custom Test Configuration"))
	}
	s.WriteString("\n\n")

	// Test toggles
	if lang == "zh" {
		s.WriteString(tSectStyle.Render("  测试项目 (空格切换, a 全选/全不选):"))
	} else {
		s.WriteString(tSectStyle.Render("  Test Items (Space toggle, a Select all/none):"))
	}
	s.WriteString("\n\n")

	for i, t := range m.toggles {
		cursor := "   "
		if m.customCursor == i {
			cursor = " > "
		}

		var check string
		if t.enabled {
			check = tChkOnStyle.Render("[x]")
		} else {
			check = tChkOffStyle.Render("[ ]")
		}

		var name string
		if lang == "zh" {
			name = t.nameZh
		} else {
			name = t.nameEn
		}

		style := tNormStyle
		if m.customCursor == i {
			style = tSelStyle
		}
		if t.needNet && !m.preCheck.Connected {
			style = tDimStyle
		}

		s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, style.Render(name)))
	}

	// Advanced settings
	s.WriteString("\n")
	if lang == "zh" {
		s.WriteString(tSectStyle.Render("  高级设置 (空格/←/→ 切换选项):"))
	} else {
		s.WriteString(tSectStyle.Render("  Advanced Settings (Space/Left/Right cycle):"))
	}
	s.WriteString("\n\n")

	for i, a := range m.advanced {
		idx := len(m.toggles) + i
		cursor := "   "
		if m.customCursor == idx {
			cursor = " > "
		}

		var name string
		if lang == "zh" {
			name = a.nameZh
		} else {
			name = a.nameEn
		}

		style := tNormStyle
		if m.customCursor == idx {
			style = tSelStyle
		}

		// Render current value with arrows
		val := a.options[a.current]
		var optStr string
		if m.customCursor == idx {
			optStr = tSelStyle.Render(fmt.Sprintf("< %s >", val))
		} else {
			optStr = tDimStyle.Render(fmt.Sprintf("< %s >", val))
		}

		s.WriteString(fmt.Sprintf("%s%-22s %s\n", cursor, style.Render(name+":"), optStr))
	}

	// Confirm button
	s.WriteString("\n")
	confirmIdx := m.customTotal - 1
	if m.customCursor == confirmIdx {
		if lang == "zh" {
			s.WriteString(fmt.Sprintf("   %s\n", tBtnStyle.Render(">> 开始测试 <<")))
		} else {
			s.WriteString(fmt.Sprintf("   %s\n", tBtnStyle.Render(">> Start Test <<")))
		}
	} else {
		if lang == "zh" {
			s.WriteString(fmt.Sprintf("   %s\n", tBtnDimStyle.Render(">> 开始测试 <<")))
		} else {
			s.WriteString(fmt.Sprintf("   %s\n", tBtnDimStyle.Render(">> Start Test <<")))
		}
	}

	s.WriteString("\n")
	if lang == "zh" {
		s.WriteString(tHelpStyle.Render("  ↑/↓ 移动  空格/Enter 切换  ←/→ 调整  a 全选  Esc 返回  q 退出"))
	} else {
		s.WriteString(tHelpStyle.Render("  Up/Down Move  Space/Enter Toggle  Left/Right Adjust  a All  Esc Back  q Quit"))
	}
	s.WriteString("\n")
	return s.String()
}

// --- Public Interface ---

// RunTuiMenu runs the interactive TUI menu and returns the result
func RunTuiMenu(preCheck utils.NetCheckResult, config *params.Config) tuiResult {
	// Pre-load stats
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

// applyCustomResult applies custom mode selections to config
func applyCustomResult(result tuiResult, preCheck utils.NetCheckResult, config *params.Config) {
	// Apply test toggles (order must match defaultTestToggles)
	for i, t := range result.toggles {
		enabled := t.enabled
		if t.needNet && !preCheck.Connected {
			enabled = false
		}
		switch i {
		case 0:
			config.BasicStatus = enabled
		case 1:
			config.CpuTestStatus = enabled
		case 2:
			config.MemoryTestStatus = enabled
		case 3:
			config.DiskTestStatus = enabled
		case 4:
			config.UtTestStatus = enabled
		case 5:
			config.SecurityTestStatus = enabled
		case 6:
			config.EmailTestStatus = enabled
		case 7:
			config.BacktraceStatus = enabled
		case 8:
			config.Nt3Status = enabled
		case 9:
			config.SpeedTestStatus = enabled
		case 10:
			config.PingTestStatus = enabled
		case 11:
			config.TgdcTestStatus = enabled
		case 12:
			config.WebTestStatus = enabled
		}
	}

	// Apply advanced settings (order must match defaultAdvSettings)
	if len(result.advanced) >= 7 {
		config.CpuTestMethod = result.advanced[0].options[result.advanced[0].current]
		config.CpuTestThreadMode = result.advanced[1].options[result.advanced[1].current]
		config.MemoryTestMethod = result.advanced[2].options[result.advanced[2].current]
		config.DiskTestMethod = result.advanced[3].options[result.advanced[3].current]
		config.Nt3Location = result.advanced[4].options[result.advanced[4].current]
		config.Nt3CheckType = result.advanced[5].options[result.advanced[5].current]
		spIdx := result.advanced[6].current
		config.SpNum = spIdx + 1
	}

	// Set OnlyIpInfoCheck if no hardware tests are enabled
	if !config.BasicStatus && !config.CpuTestStatus && !config.MemoryTestStatus && !config.DiskTestStatus {
		config.OnlyIpInfoCheck = true
	}

	config.AutoChangeDiskMethod = true
}
