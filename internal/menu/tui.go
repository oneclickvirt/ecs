package menu

import (
	"fmt"
	"strings"

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
	tOnStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	tOffStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	tValStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	tCurStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
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
	choice      string
	language    string
	quit        bool
	custom      bool
	toggles     []testToggle
	advanced    []advSetting
	mainAnalyze bool
	mainUpload  bool
}

type tuiModel struct {
	phase      menuPhase
	config     *params.Config
	preCheck   utils.NetCheckResult
	langPreset bool

	langCursor       int
	mainCursor       int
	mainItems        []mainMenuItem
	mainAnalyze      bool
	mainUpload       bool
	mainExtraTotal   int
	mainScrollOffset int

	customCursor       int
	toggles            []testToggle
	advanced           []advSetting
	customTotal        int
	customScrollOffset int

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

func (m tuiModel) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Re-clamp scroll offsets after resize
		m.clampMainScroll()
		m.clampCustomScroll()
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
			cursor = tCurStyle.Render(" > ")
			style = tSelStyle
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(l)))
	}
	s.WriteString("\n")
	s.WriteString(tHelpStyle.Render("  ↑/↓ Navigate  Enter Confirm  1/2 Quick-Select  q Quit"))
	s.WriteString("\n")
	return s.String()
}
