package menu

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	textinput "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/oneclickvirt/ecs/internal/params"
	"github.com/oneclickvirt/ecs/utils"
)

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
	tPanelStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("238")).Padding(0, 1)
)

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
	descZh  string
	descEn  string
	needNet bool
}

type testToggle struct {
	key     string
	nameZh  string
	nameEn  string
	descZh  string
	descEn  string
	enabled bool
	needNet bool
}

type advOption struct {
	value   string
	labelZh string
	labelEn string
	descZh  string
	descEn  string
}

type advSetting struct {
	key     string
	nameZh  string
	nameEn  string
	descZh  string
	descEn  string
	kind    string // option | bool | text
	options []advOption
	current int
	boolVal bool
	textVal string
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

	langCursor int
	mainCursor int
	mainItems  []mainMenuItem

	customCursor int
	toggles      []testToggle
	advanced     []advSetting
	customTotal  int

	editingText bool
	editingIdx  int
	textInput   textinput.Model

	statsTotal int
	statsDaily int
	hasStats   bool
	cmpVersion int
	newVersion string

	result tuiResult
	width  int
	height int
}

func defaultMainItems() []mainMenuItem {
	return []mainMenuItem{
		{id: "1", zh: "融合怪完全体(能测全测)", en: "Full Test (All Available Tests)", descZh: "系统信息、CPU、内存、磁盘、解锁、IP质量、邮件端口、回程、NT3、测速、TGDC、网站延迟。", descEn: "Runs all available modules: system, compute, memory, disk, unlock, security, routing and speed.", needNet: false},
		{id: "2", zh: "极简版", en: "Minimal Suite", descZh: "系统信息+CPU+内存+磁盘+测速节点×5，不含解锁/网络/路由测试。", descEn: "System info + CPU + memory + disk + 5 speed nodes. No unlock/network/routing tests.", needNet: false},
		{id: "3", zh: "精简版", en: "Standard Suite", descZh: "系统信息+CPU+内存+磁盘+跨国平台解锁+三网回程路由+测速节点×5。", descEn: "System info + CPU + memory + disk + streaming unlock + 3-network routing + 5 speed nodes.", needNet: false},
		{id: "4", zh: "精简网络版", en: "Network Suite", descZh: "系统信息+CPU+内存+磁盘+上游及三网回程路由+测速节点×5。", descEn: "System info + CPU + memory + disk + upstream/3-network backtrace routing + 5 speed nodes.", needNet: false},
		{id: "5", zh: "精简解锁版", en: "Unlock Suite", descZh: "系统信息+CPU+内存+磁盘IO+跨国平台解锁+测速节点×5。", descEn: "System info + CPU + memory + disk IO + streaming unlock + 5 speed nodes.", needNet: false},
		{id: "6", zh: "网络单项", en: "Network Only", descZh: "仅网络维度：IP质量、回程、NT3、延迟、TGDC、网站和测速。", descEn: "Network-only profile: IP quality, route, latency, TGDC, websites, speed.", needNet: true},
		{id: "7", zh: "解锁单项", en: "Unlock Only", descZh: "仅进行跨国平台解锁与流媒体可用性检测。", descEn: "Unlock-only profile for cross-border media/service availability.", needNet: true},
		{id: "8", zh: "硬件单项", en: "Hardware Only", descZh: "系统信息、CPU、内存、dd/fio 磁盘测试。", descEn: "Hardware-only profile with system, CPU, memory and disk tests.", needNet: false},
		{id: "9", zh: "IP质量检测", en: "IP Quality", descZh: "15个数据库IP质量检测+邮件端口连通性检测。", descEn: "IP quality check across 15 databases + email port connectivity test.", needNet: true},
		{id: "10", zh: "三网回程线路", en: "3-Network Route", descZh: "三网回程、NT3路由、延迟、TGDC、网站延迟专项。", descEn: "3-network backtrace + NT3 route + latency/TGDC/website checks.", needNet: true},
		{id: "custom", zh: ">>> 高级自定义(全参数模式)", en: ">>> Advanced Custom (Full Parameters)", descZh: "按参数逐项配置，支持测试项、方法、路径、上传和结果分析。", descEn: "Configure per-parameter with test toggles, methods, paths, upload and analysis.", needNet: false},
		{id: "0", zh: "退出程序", en: "Exit Program", descZh: "退出当前程序。", descEn: "Exit program.", needNet: false},
	}
}

func defaultTestToggles() []testToggle {
	return []testToggle{
		{key: "basic", nameZh: "基础系统信息", nameEn: "Basic System Info", descZh: "操作系统、CPU型号、内核、虚拟化等基础信息。", descEn: "OS, CPU model, kernel, virtualization and base environment info.", enabled: true, needNet: false},
		{key: "cpu", nameZh: "CPU测试", nameEn: "CPU Test", descZh: "按所选方法执行 CPU 计算性能测试。", descEn: "Run CPU compute benchmarks using selected method.", enabled: true, needNet: false},
		{key: "memory", nameZh: "内存测试", nameEn: "Memory Test", descZh: "按所选方法测试内存吞吐和访问性能。", descEn: "Run memory throughput and access benchmarks by selected method.", enabled: true, needNet: false},
		{key: "disk", nameZh: "磁盘测试", nameEn: "Disk Test", descZh: "按所选方法执行磁盘读写性能测试。", descEn: "Run disk read/write benchmark using selected method/path.", enabled: true, needNet: false},
		{key: "ut", nameZh: "跨国平台解锁", nameEn: "Streaming Unlock", descZh: "检测多类海外流媒体与服务可用性。", descEn: "Check availability of cross-border streaming/services.", enabled: false, needNet: true},
		{key: "security", nameZh: "IP质量检测", nameEn: "IP Quality Check", descZh: "多库 IP 信誉、风险和质量信息检测。", descEn: "IP reputation/risk/quality checks across multiple datasets.", enabled: false, needNet: true},
		{key: "email", nameZh: "邮件端口检测", nameEn: "Email Port Check", descZh: "检查常见邮件相关端口连通能力。", descEn: "Check common mail-related port connectivity.", enabled: false, needNet: true},
		{key: "backtrace", nameZh: "回程路由", nameEn: "Backtrace Route", descZh: "检测上游及三网回程路径。", descEn: "Inspect upstream and 3-network return routes.", enabled: false, needNet: true},
		{key: "nt3", nameZh: "NT3路由", nameEn: "NT3 Route", descZh: "按指定地区与协议执行详细路由追踪。", descEn: "Run detailed route trace by selected location/protocol.", enabled: false, needNet: true},
		{key: "speed", nameZh: "测速", nameEn: "Speed Test", descZh: "测试下载/上传带宽与延迟。", descEn: "Measure download/upload bandwidth and latency.", enabled: false, needNet: true},
		{key: "ping", nameZh: "Ping测试", nameEn: "Ping Test", descZh: "全国/多地区延迟质量测试。", descEn: "Latency quality checks across multiple regions.", enabled: false, needNet: true},
		{key: "tgdc", nameZh: "Telegram DC测试", nameEn: "Telegram DC Test", descZh: "检测各 Telegram 数据中心节点延迟。", descEn: "Measure latency to each Telegram data center node.", enabled: false, needNet: true},
		{key: "web", nameZh: "网站延迟", nameEn: "Website Latency", descZh: "检测常见网站访问延迟。", descEn: "Check latency to commonly used websites.", enabled: false, needNet: true},
	}
}

func option(value, zh, en, descZh, descEn string) advOption {
	return advOption{value: value, labelZh: zh, labelEn: en, descZh: descZh, descEn: descEn}
}

func defaultAdvSettings(config *params.Config) []advSetting {
	adv := []advSetting{
		{
			key: "cpum", nameZh: "CPU测试方法", nameEn: "CPU Method", kind: "option",
			descZh: "选择 CPU 基准测试工具（sysbench/geekbench/winsat）。",
			descEn: "Choose CPU benchmark tool (sysbench/geekbench/winsat).",
			options: []advOption{
				option("sysbench", "Sysbench", "Sysbench", "通用 CPU 基准测试工具。", "General-purpose CPU benchmark tool."),
				option("geekbench", "Geekbench", "Geekbench", "综合场景 CPU 基准测试工具。", "Synthetic benchmark simulating real-world application workloads."),
				option("winsat", "WinSAT", "WinSAT", "Windows 环境下的 CPU 基准测试工具。", "CPU benchmark tool for Windows environments."),
			},
		},
		{
			key: "cput", nameZh: "CPU线程模式", nameEn: "CPU Thread Mode", kind: "option",
			descZh: "单线程: 测试单核最高运算速度; 多线程: 测试全核并发吞吐。",
			descEn: "Single-thread: peak single-core speed; Multi-thread: full-core parallel throughput.",
			options: []advOption{
				option("multi", "多线程", "Multi-thread", "测试所有核心并发运算吞吐。", "Measure parallel compute throughput across all cores."),
				option("single", "单线程", "Single-thread", "测试单核最高运算速度。", "Measure peak single-core compute speed."),
			},
		},
		{
			key: "memorym", nameZh: "内存测试方法", nameEn: "Memory Method", kind: "option",
			descZh: "选择内存基准测试工具。",
			descEn: "Choose memory benchmark tool.",
			options: []advOption{
				option("stream", "STREAM", "STREAM", "专项内存带宽基准测试工具（STREAM）。", "Memory bandwidth benchmark tool (STREAM)."),
				option("sysbench", "Sysbench", "Sysbench", "通用内存基准测试工具。", "General-purpose memory benchmark tool."),
				option("dd", "dd", "dd", "使用 dd 命令测量内存顺序读写。", "Measure memory sequential R/W using dd command."),
				option("winsat", "WinSAT", "WinSAT", "Windows 环境内存基准测试工具。", "Memory benchmark tool for Windows environments."),
				option("auto", "自动", "Auto", "按优先级自动选择可用测试工具。", "Automatically select the preferred available tool."),
			},
		},
		{
			key: "diskm", nameZh: "磁盘测试方法", nameEn: "Disk Method", kind: "option",
			descZh: "选择磁盘基准测试工具。",
			descEn: "Choose disk benchmark tool.",
			options: []advOption{
				option("fio", "FIO", "FIO", "多队列深度顺序/随机 I/O 全面基准测试。", "Comprehensive sequential/random I/O benchmark with multiple queue depths."),
				option("dd", "dd", "dd", "使用 dd 命令进行顺序读写基准测试。", "Sequential read/write benchmark using dd command."),
				option("winsat", "WinSAT", "WinSAT", "Windows 环境磁盘基准测试工具。", "Disk benchmark tool for Windows environments."),
			},
		},
		{
			key: "diskp", nameZh: "磁盘测试路径", nameEn: "Disk Test Path", kind: "text",
			descZh:  "自定义磁盘测试目录。留空表示默认路径。",
			descEn:  "Custom disk test directory. Empty means default path.",
			textVal: config.DiskTestPath,
		},
		{
			key: "diskmc", nameZh: "多磁盘检测", nameEn: "Multi-Disk Check", kind: "bool",
			descZh:  "启用后检测并测试所有已挂载磁盘路径。",
			descEn:  "When enabled, detect and benchmark all mounted disk paths.",
			boolVal: config.DiskMultiCheck,
		},
		{
			key: "autodiskm", nameZh: "磁盘方法失败自动切换", nameEn: "Auto Switch Disk Method", kind: "bool",
			descZh:  "所选磁盘测试方法失败时自动切换为其他可用方法。",
			descEn:  "Automatically try another available disk method if the selected method fails.",
			boolVal: config.AutoChangeDiskMethod,
		},
		{
			key: "nt3loc", nameZh: "NT3测试地区", nameEn: "NT3 Location", kind: "option",
			descZh: "选择路由追踪地区。显示中文全称，内部仍使用标准参数值。",
			descEn: "Choose route trace region. Full names are shown while preserving standard values.",
			options: []advOption{
				option("GZ", "广州", "Guangzhou", "从广州节点进行追踪。", "Trace from Guangzhou node."),
				option("SH", "上海", "Shanghai", "从上海节点进行追踪。", "Trace from Shanghai node."),
				option("BJ", "北京", "Beijing", "从北京节点进行追踪。", "Trace from Beijing node."),
				option("CD", "成都", "Chengdu", "从成都节点进行追踪。", "Trace from Chengdu node."),
				option("ALL", "全部地区", "All Regions", "依次测试全部地区节点。", "Run route traces from all supported regions."),
			},
		},
		{
			key: "nt3t", nameZh: "NT3协议类型", nameEn: "NT3 Protocol", kind: "option",
			descZh: "指定 NT3 路由检测协议栈。",
			descEn: "Select protocol stack used by NT3 route checks.",
			options: []advOption{
				option("ipv4", "仅 IPv4", "IPv4 Only", "仅测试 IPv4 路由路径。", "Test IPv4 routing only."),
				option("ipv6", "仅 IPv6", "IPv6 Only", "仅测试 IPv6 路由路径。", "Test IPv6 routing only."),
				option("both", "IPv4 + IPv6", "IPv4 + IPv6", "同时测试 IPv4 与 IPv6。", "Test both IPv4 and IPv6."),
			},
		},
		{
			key: "spnum", nameZh: "测速节点数/运营商", nameEn: "Speed Nodes per ISP", kind: "option",
			descZh: "每个运营商参与测速的节点数量。",
			descEn: "Number of speed test nodes per ISP.",
			options: []advOption{
				option("1", "1 个", "1 node", "每运营商1节点，耗时最短，覆盖面最小。", "1 node per ISP, shortest runtime, least coverage."),
				option("2", "2 个", "2 nodes", "每运营商2节点（默认值）。", "2 nodes per ISP (default)."),
				option("3", "3 个", "3 nodes", "每运营商3节点，覆盖面扩大，耗时增加。", "3 nodes per ISP, wider coverage, longer runtime."),
				option("4", "4 个", "4 nodes", "每运营商4节点。", "4 nodes per ISP."),
				option("5", "5 个", "5 nodes", "每运营商5节点，覆盖面宽，耗时较高。", "5 nodes per ISP, wide coverage, higher runtime."),
				option("6", "6 个", "6 nodes", "每运营商6节点。", "6 nodes per ISP."),
				option("7", "7 个", "7 nodes", "每运营商7节点。", "7 nodes per ISP."),
				option("8", "8 个", "8 nodes", "每运营商8节点。", "8 nodes per ISP."),
				option("9", "9 个", "9 nodes", "每运营商9节点。", "9 nodes per ISP."),
				option("10", "10 个", "10 nodes", "每运营商10节点，覆盖面最宽，耗时最高。", "10 nodes per ISP, widest coverage, longest runtime."),
			},
		},
		{
			key: "log", nameZh: "调试日志", nameEn: "Debug Logger", kind: "bool",
			descZh:  "启用后输出更多调试日志，便于排障。",
			descEn:  "Enable verbose logs for troubleshooting.",
			boolVal: config.EnableLogger,
		},
		{
			key: "upload", nameZh: "上传并生成分享链接", nameEn: "Upload & Share Link", kind: "bool",
			descZh:  "启用后上传测试结果并生成可分享链接。",
			descEn:  "Upload final result and generate a shareable link.",
			boolVal: config.EnableUpload,
		},
		{
			key: "analysis", nameZh: "测试后结果总结分析", nameEn: "Post-Test Summary Analysis", kind: "bool",
			descZh:  "测试结束后输出简明总结（含CPU排名、带宽和延迟数据）。",
			descEn:  "Output a concise summary after tests (CPU rank, bandwidth, latency scores).",
			boolVal: config.AnalyzeResult,
		},
		{
			key: "filepath", nameZh: "结果文件名", nameEn: "Result File Name", kind: "text",
			descZh:  "上传前本地结果文件名。",
			descEn:  "Local result filename used before upload.",
			textVal: config.FilePath,
		},
		{
			key: "width", nameZh: "输出宽度", nameEn: "Output Width", kind: "option",
			descZh: "控制终端输出排版宽度。",
			descEn: "Controls console output formatting width.",
			options: []advOption{
				option("72", "72 列", "72 cols", "紧凑显示。", "Compact layout."),
				option("82", "82 列", "82 cols", "默认宽度。", "Default width."),
				option("100", "100 列", "100 cols", "更宽显示。", "Wider layout."),
				option("120", "120 列", "120 cols", "宽屏显示。", "Wide-screen layout."),
			},
		},
	}

	for i := range adv {
		switch adv[i].key {
		case "cpum":
			adv[i].current = optionIndexByValue(adv[i].options, config.CpuTestMethod)
		case "cput":
			adv[i].current = optionIndexByValue(adv[i].options, config.CpuTestThreadMode)
		case "memorym":
			adv[i].current = optionIndexByValue(adv[i].options, config.MemoryTestMethod)
		case "diskm":
			adv[i].current = optionIndexByValue(adv[i].options, config.DiskTestMethod)
		case "nt3loc":
			adv[i].current = optionIndexByValue(adv[i].options, config.Nt3Location)
		case "nt3t":
			adv[i].current = optionIndexByValue(adv[i].options, config.Nt3CheckType)
		case "spnum":
			adv[i].current = optionIndexByValue(adv[i].options, strconv.Itoa(config.SpNum))
		case "width":
			adv[i].current = optionIndexByValue(adv[i].options, strconv.Itoa(config.Width))
		}
	}

	return adv
}

func optionIndexByValue(options []advOption, value string) int {
	for i, opt := range options {
		if opt.value == value {
			return i
		}
	}
	return 0
}

func newTuiModel(preCheck utils.NetCheckResult, config *params.Config, langPreset bool, statsTotal, statsDaily int, hasStats bool, cmpVersion int, newVersion string) tuiModel {
	toggles := defaultTestToggles()
	advanced := defaultAdvSettings(config)
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = ""
	ti.CharLimit = 255
	ti.Width = 45
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
		width:       config.Width,
		height:      24,
		textInput:   ti,
	}
	if langPreset {
		m.phase = phaseMain
		m.result.language = config.Language
	} else {
		m.phase = phaseLang
	}
	return m
}

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
	langs := []string{"1. 中文", "2. English"}
	for i, l := range langs {
		cursor := "   "
		style := tNormStyle
		if m.langCursor == i {
			cursor = " > "
			style = tSelStyle
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(l)))
	}
	s.WriteString("\n")
	s.WriteString(tHelpStyle.Render("  ↑/↓ Navigate  Enter Confirm  1/2 Quick-Select  q Quit"))
	s.WriteString("\n")
	return s.String()
}

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
			return m, nil
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

func (m tuiModel) selectedMainDesc(lang string) string {
	item := m.mainItems[m.mainCursor]
	if lang == "zh" {
		return item.descZh
	}
	return item.descEn
}

func (m tuiModel) viewMain() string {
	lang := m.result.language
	var s strings.Builder
	s.WriteString("\n")
	if lang == "zh" {
		s.WriteString(tTitleStyle.Render(fmt.Sprintf("  VPS融合怪 %s", m.config.EcsVersion)))
	} else {
		s.WriteString(tTitleStyle.Render(fmt.Sprintf("  VPS Fusion Monster %s", m.config.EcsVersion)))
	}
	s.WriteString("\n")
	if m.preCheck.Connected && m.cmpVersion == -1 {
		if lang == "zh" {
			s.WriteString(tWarnStyle.Render(fmt.Sprintf("  ! 检测到新版本 %s 如有必要请更新", m.newVersion)))
		} else {
			s.WriteString(tWarnStyle.Render(fmt.Sprintf("  ! New version %s detected", m.newVersion)))
		}
		s.WriteString("\n")
	}
	if m.preCheck.Connected && m.hasStats {
		if lang == "zh" {
			s.WriteString(tInfoStyle.Render(fmt.Sprintf("  总使用量: %s | 今日使用: %s", utils.FormatGoecsNumber(m.statsTotal), utils.FormatGoecsNumber(m.statsDaily))))
		} else {
			s.WriteString(tInfoStyle.Render(fmt.Sprintf("  Total Usage: %s | Daily Usage: %s", utils.FormatGoecsNumber(m.statsTotal), utils.FormatGoecsNumber(m.statsDaily))))
		}
		s.WriteString("\n")
	}
	s.WriteString("\n")
	if lang == "zh" {
		s.WriteString(tSectStyle.Render("  请选择测试方案:"))
	} else {
		s.WriteString(tSectStyle.Render("  Select Test Suite:"))
	}
	s.WriteString("\n\n")
	for i, item := range m.mainItems {
		cursor := "   "
		style := tNormStyle
		if m.mainCursor == i {
			cursor = " > "
			style = tSelStyle
		}
		label := item.en
		if lang == "zh" {
			label = item.zh
		}
		prefix := ""
		switch {
		case item.id == "custom":
			prefix = ""
		case item.id == "0":
			prefix = " 0. "
		default:
			prefix = fmt.Sprintf("%2s. ", item.id)
		}
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
	panelTitle := "  当前选项说明"
	panelBody := m.selectedMainDesc(lang)
	if lang == "en" {
		panelTitle = "  Selected Option Description"
	}
	s.WriteString(tSectStyle.Render(panelTitle) + "\n")
	s.WriteString(tPanelStyle.Width(maxInt(60, m.width-6)).Render(panelBody) + "\n")
	s.WriteString("\n")
	if lang == "zh" {
		s.WriteString(tHelpStyle.Render("  ↑/↓/j/k 移动  Enter 确认  数字 快速选择  q 退出"))
	} else {
		s.WriteString(tHelpStyle.Render("  Up/Down/j/k Navigate  Enter Confirm  Number Quick-Select  q Quit"))
	}
	s.WriteString("\n")
	return s.String()
}

func (m *tuiModel) startEditText(settingIdx int) {
	m.editingText = true
	m.editingIdx = settingIdx
	m.textInput.SetValue(m.advanced[settingIdx].textVal)
	m.textInput.Focus()
}

func (m *tuiModel) stopEditText(save bool) {
	if save {
		m.advanced[m.editingIdx].textVal = strings.TrimSpace(m.textInput.Value())
	}
	m.textInput.Blur()
	m.editingText = false
}

func (m tuiModel) updateCustom(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingText {
		switch msg.String() {
		case "enter":
			m.stopEditText(true)
			return m, nil
		case "esc":
			m.stopEditText(false)
			return m, nil
		case "ctrl+c":
			m.result.quit = true
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

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
	case " ", "enter", "right", "l", "left", "h":
		if m.customCursor < len(m.toggles) {
			t := &m.toggles[m.customCursor]
			if t.needNet && !m.preCheck.Connected {
				break
			}
			t.enabled = !t.enabled
			break
		}
		if m.customCursor == m.customTotal-1 {
			m.result.custom = true
			m.result.choice = "custom"
			m.result.toggles = m.toggles
			m.result.advanced = m.advanced
			return m, tea.Quit
		}
		advIdx := m.customCursor - len(m.toggles)
		if advIdx >= 0 && advIdx < len(m.advanced) {
			a := &m.advanced[advIdx]
			switch a.kind {
			case "bool":
				a.boolVal = !a.boolVal
			case "option":
				if key == "left" || key == "h" {
					a.current = (a.current - 1 + len(a.options)) % len(a.options)
				} else {
					a.current = (a.current + 1) % len(a.options)
				}
			case "text":
				if key == "enter" || key == " " {
					m.startEditText(advIdx)
				}
			}
		}
	case "a":
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
	case "esc":
		m.phase = phaseMain
		return m, nil
	case "q", "ctrl+c":
		m.result.quit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m tuiModel) currentCustomDescription(lang string) string {
	if m.customCursor < len(m.toggles) {
		t := m.toggles[m.customCursor]
		if lang == "zh" {
			return t.descZh
		}
		return t.descEn
	}
	if m.customCursor == m.customTotal-1 {
		if lang == "zh" {
			return "确认当前高级自定义配置并开始测试。"
		}
		return "Confirm current advanced custom configuration and start tests."
	}
	idx := m.customCursor - len(m.toggles)
	a := m.advanced[idx]
	if a.kind == "option" {
		op := a.options[a.current]
		if lang == "zh" {
			return a.descZh + " 当前选项: " + op.labelZh + "。" + op.descZh
		}
		return a.descEn + " Current option: " + op.labelEn + ". " + op.descEn
	}
	if a.kind == "bool" {
		if lang == "zh" {
			state := "关闭"
			if a.boolVal {
				state = "开启"
			}
			return a.descZh + " 当前状态: " + state + "。"
		}
		state := "OFF"
		if a.boolVal {
			state = "ON"
		}
		return a.descEn + " Current state: " + state + "."
	}
	if lang == "zh" {
		return a.descZh + " 当前值: " + a.textVal
	}
	return a.descEn + " Current value: " + a.textVal
}

func (m tuiModel) advDisplayValue(a advSetting, lang string) string {
	switch a.kind {
	case "option":
		op := a.options[a.current]
		if lang == "zh" {
			return op.labelZh
		}
		return op.labelEn
	case "bool":
		if a.boolVal {
			if lang == "zh" {
				return "开启"
			}
			return "ON"
		}
		if lang == "zh" {
			return "关闭"
		}
		return "OFF"
	case "text":
		if strings.TrimSpace(a.textVal) == "" {
			if lang == "zh" {
				return "(默认)"
			}
			return "(default)"
		}
		return a.textVal
	}
	return ""
}

func (m tuiModel) viewCustom() string {
	lang := m.result.language
	var s strings.Builder
	s.WriteString("\n")
	if lang == "zh" {
		s.WriteString(tTitleStyle.Render(fmt.Sprintf("  VPS融合怪 %s  —  高级自定义", m.config.EcsVersion)))
	} else {
		s.WriteString(tTitleStyle.Render(fmt.Sprintf("  VPS Fusion Monster %s  —  Advanced Custom", m.config.EcsVersion)))
	}
	s.WriteString("\n")
	if m.preCheck.Connected && m.cmpVersion == -1 {
		if lang == "zh" {
			s.WriteString(tWarnStyle.Render(fmt.Sprintf("  ! 检测到新版本 %s 如有必要请更新", m.newVersion)))
		} else {
			s.WriteString(tWarnStyle.Render(fmt.Sprintf("  ! New version %s detected", m.newVersion)))
		}
		s.WriteString("\n")
	}
	if m.preCheck.Connected && m.hasStats {
		if lang == "zh" {
			s.WriteString(tInfoStyle.Render(fmt.Sprintf("  总使用量: %s | 今日使用: %s", utils.FormatGoecsNumber(m.statsTotal), utils.FormatGoecsNumber(m.statsDaily))))
		} else {
			s.WriteString(tInfoStyle.Render(fmt.Sprintf("  Total Usage: %s | Daily Usage: %s", utils.FormatGoecsNumber(m.statsTotal), utils.FormatGoecsNumber(m.statsDaily))))
		}
		s.WriteString("\n")
	}
	s.WriteString("\n")
	if lang == "zh" {
		s.WriteString(tSectStyle.Render("  测试开关 (空格切换, a 全选/全不选):"))
	} else {
		s.WriteString(tSectStyle.Render("  Test Toggles (Space to toggle, a all/none):"))
	}
	s.WriteString("\n\n")
	for i, t := range m.toggles {
		cursor := "   "
		style := tNormStyle
		if m.customCursor == i {
			cursor = " > "
			style = tSelStyle
		}
		if t.needNet && !m.preCheck.Connected {
			style = tDimStyle
		}
		check := tChkOffStyle.Render("[ ]")
		if t.enabled {
			check = tChkOnStyle.Render("[x]")
		}
		name := t.nameEn
		if lang == "zh" {
			name = t.nameZh
		}
		s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, style.Render(name)))
	}
	s.WriteString("\n")
	if lang == "zh" {
		s.WriteString(tSectStyle.Render("  参数设置 (Enter/空格切换, ←/→改选项):"))
	} else {
		s.WriteString(tSectStyle.Render("  Parameter Settings (Enter/Space switch, Left/Right cycle):"))
	}
	s.WriteString("\n\n")
	for i, a := range m.advanced {
		idx := len(m.toggles) + i
		cursor := "   "
		if m.customCursor == idx {
			cursor = " > "
		}
		style := tNormStyle
		if m.customCursor == idx {
			style = tSelStyle
		}
		name := a.nameEn
		if lang == "zh" {
			name = a.nameZh
		}
		value := m.advDisplayValue(a, lang)
		if a.kind == "option" {
			value = "< " + value + " >"
		}
		s.WriteString(fmt.Sprintf("%s%-26s %s\n", cursor, style.Render(name+":"), tDimStyle.Render(value)))
	}

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
	panelTitle := "  当前项说明"
	if lang == "en" {
		panelTitle = "  Current Item Description"
	}
	s.WriteString(tSectStyle.Render(panelTitle) + "\n")
	s.WriteString(tPanelStyle.Width(maxInt(60, m.width-6)).Render(m.currentCustomDescription(lang)) + "\n")

	if m.editingText {
		if lang == "zh" {
			s.WriteString("\n" + tWarnStyle.Render("  文本编辑模式: Enter 保存, Esc 取消") + "\n")
		} else {
			s.WriteString("\n" + tWarnStyle.Render("  Text edit mode: Enter save, Esc cancel") + "\n")
		}
		s.WriteString("  " + m.textInput.View() + "\n")
	}

	s.WriteString("\n")
	if lang == "zh" {
		s.WriteString(tHelpStyle.Render("  ↑/↓ 移动  Enter/空格 切换  ←/→ 改选项  a 全选  Esc 返回  q 退出"))
	} else {
		s.WriteString(tHelpStyle.Render("  Up/Down Move  Enter/Space Toggle  Left/Right Cycle  a All  Esc Back  q Quit"))
	}
	s.WriteString("\n")
	return s.String()
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
		case "nt3loc":
			config.Nt3Location = a.options[a.current].value
		case "nt3t":
			config.Nt3CheckType = a.options[a.current].value
		case "spnum":
			if v, err := strconv.Atoi(a.options[a.current].value); err == nil {
				config.SpNum = v
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
		case "width":
			if v, err := strconv.Atoi(a.options[a.current].value); err == nil {
				config.Width = v
			}
		}
	}

	if !config.BasicStatus && !config.CpuTestStatus && !config.MemoryTestStatus && !config.DiskTestStatus {
		config.OnlyIpInfoCheck = true
	}
}
