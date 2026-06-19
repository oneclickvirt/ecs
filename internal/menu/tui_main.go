package menu

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oneclickvirt/ecs/utils"
)

func (m tuiModel) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	maxCursor := len(m.mainItems) + m.mainExtraTotal - 1
	switch key {
	case "up", "k":
		if m.mainCursor > 0 {
			m.mainCursor--
		}
		m.ensureMainCursorVisible()
	case "down", "j":
		if m.mainCursor < maxCursor {
			m.mainCursor++
		}
		m.ensureMainCursorVisible()
	case "home":
		m.mainCursor = 0
		m.ensureMainCursorVisible()
	case "end":
		m.mainCursor = maxCursor
		m.ensureMainCursorVisible()
	case " ":
		if m.mainCursor >= len(m.mainItems) {
			switch m.mainCursor - len(m.mainItems) {
			case 0:
				m.mainAnalyze = !m.mainAnalyze
			case 1:
				m.mainUpload = !m.mainUpload
			}
		}
	case "enter":
		if m.mainCursor >= len(m.mainItems) {
			switch m.mainCursor - len(m.mainItems) {
			case 0:
				m.mainAnalyze = !m.mainAnalyze
			case 1:
				m.mainUpload = !m.mainUpload
			}
			return m, nil
		}
		item := m.mainItems[m.mainCursor]
		if item.needNet && !m.preCheck.Connected {
			return m, nil
		}
		if item.id == "custom" {
			m.syncMainQuickOptionsToAdvanced()
			m.phase = phaseCustom
			m.customCursor = 0
			m.customScrollOffset = 0
			return m, nil
		}
		m.result.mainAnalyze = m.mainAnalyze
		m.result.mainUpload = m.mainUpload
		m.result.choice = item.id
		return m, tea.Quit
	case "esc":
		m.phase = phaseLang
		return m, nil
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
					m.syncMainQuickOptionsToAdvanced()
					m.phase = phaseCustom
					m.customCursor = 0
					m.customScrollOffset = 0
					return m, nil
				}
				m.mainCursor = i
				m.ensureMainCursorVisible()
				m.result.mainAnalyze = m.mainAnalyze
				m.result.mainUpload = m.mainUpload
				m.result.choice = item.id
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *tuiModel) syncMainQuickOptionsToAdvanced() {
	for i := range m.advanced {
		switch m.advanced[i].key {
		case "analysis":
			m.advanced[i].boolVal = m.mainAnalyze
		case "upload":
			m.advanced[i].boolVal = m.mainUpload
		}
	}
}

func (m tuiModel) selectedMainDesc(lang string) string {
	if m.mainCursor >= len(m.mainItems) {
		switch m.mainCursor - len(m.mainItems) {
		case 0:
			if lang == "zh" {
				return "测试结束后输出简明总结（含CPU排名、带宽和延迟数据）。默认关闭。"
			}
			return "Output a concise summary after tests (CPU rank, bandwidth, latency scores). Disabled by default."
		case 1:
			if lang == "zh" {
				return "上传测试结果到服务端并生成可分享链接。默认启用。"
			}
			return "Upload test results to the server and generate a shareable link. Enabled by default."
		}
		return ""
	}
	item := m.mainItems[m.mainCursor]
	if lang == "zh" {
		return item.descZh
	}
	return item.descEn
}

func (m tuiModel) viewMain() string {
	lang := m.result.language

	// === FIXED HEADER (always visible at top) ===
	var hdr strings.Builder
	hdr.WriteString("\n")
	if lang == "zh" {
		hdr.WriteString(tTitleStyle.Render(fmt.Sprintf("  VPS融合怪 %s", m.config.EcsVersion)))
	} else {
		hdr.WriteString(tTitleStyle.Render(fmt.Sprintf("  VPS Fusion Monster %s", m.config.EcsVersion)))
	}
	hdr.WriteString("\n")
	if m.preCheck.Connected && m.cmpVersion == -1 {
		if lang == "zh" {
			hdr.WriteString(tWarnStyle.Render(fmt.Sprintf("  ! 检测到新版本 %s 如有必要请更新", m.newVersion)))
		} else {
			hdr.WriteString(tWarnStyle.Render(fmt.Sprintf("  ! New version %s detected", m.newVersion)))
		}
		hdr.WriteString("\n")
	}
	if m.preCheck.Connected && m.hasStats {
		if lang == "zh" {
			hdr.WriteString(tInfoStyle.Render(fmt.Sprintf("  总使用量: %s | 今日使用: %s", utils.FormatGoecsNumber(m.statsTotal), utils.FormatGoecsNumber(m.statsDaily))))
		} else {
			hdr.WriteString(tInfoStyle.Render(fmt.Sprintf("  Total Usage: %s | Daily Usage: %s", utils.FormatGoecsNumber(m.statsTotal), utils.FormatGoecsNumber(m.statsDaily))))
		}
		hdr.WriteString("\n")
	}
	hdr.WriteString("\n")
	if lang == "zh" {
		hdr.WriteString(tSectStyle.Render("  请选择测试方案:"))
	} else {
		hdr.WriteString(tSectStyle.Render("  Select Test Suite:"))
	}
	hdr.WriteString("\n\n")
	headerStr := hdr.String()
	headerLines := strings.Count(headerStr, "\n")

	// === FIXED FOOTER (always visible at bottom) ===
	var ftr strings.Builder
	ftr.WriteString("\n")
	panelTitle := "  当前选项说明"
	panelBody := m.selectedMainDesc(lang)
	if lang == "en" {
		panelTitle = "  Selected Option Description"
	}
	renderedPanel := tPanelStyle.Width(maxInt(60, m.width-6)).Render(panelBody)
	ftr.WriteString(tSectStyle.Render(panelTitle) + "\n")
	ftr.WriteString(renderedPanel + "\n")
	ftr.WriteString("\n")
	if lang == "zh" {
		ftr.WriteString(tHelpStyle.Render("  ↑/↓/j/k 移动  Enter 确认  Space 切换  数字 快速选择  Esc 返回语言  q 退出"))
	} else {
		ftr.WriteString(tHelpStyle.Render("  Up/Down/j/k Move  Enter Confirm  Space Toggle  Number Quick-Select  Esc Lang  q Quit"))
	}
	ftr.WriteString("\n")
	footerStr := ftr.String()
	footerLines := strings.Count(footerStr, "\n")

	// === SCROLLABLE BODY: items + quick options ===
	var bodyLines []string
	for i, item := range m.mainItems {
		cursor := "   "
		style := tNormStyle
		if m.mainCursor == i {
			cursor = tCurStyle.Render(" > ")
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
		bodyLines = append(bodyLines, fmt.Sprintf("%s%s%s\n", cursor, style.Render(prefix+label), tDimStyle.Render(suffix)))
	}
	// Separator + quick options section
	bodyLines = append(bodyLines, "\n")
	if lang == "zh" {
		bodyLines = append(bodyLines, tSectStyle.Render("  快速选项:")+"  "+tDimStyle.Render("Space/Enter 切换")+"\n")
	} else {
		bodyLines = append(bodyLines, tSectStyle.Render("  Quick Options:")+"  "+tDimStyle.Render("Space/Enter to toggle")+"\n")
	}
	for qi, qState := range []bool{m.mainAnalyze, m.mainUpload} {
		qIdx := len(m.mainItems) + qi
		cur := "   "
		nameStyle := tNormStyle
		if m.mainCursor == qIdx {
			cur = tCurStyle.Render(" > ")
			nameStyle = tSelStyle
		}
		chk := tChkOffStyle.Render("[ ]")
		if qState {
			chk = tChkOnStyle.Render("[x]")
		}
		var qName, qVal string
		if qi == 0 {
			if lang == "zh" {
				qName = "测试后自动总结分析"
			} else {
				qName = "Post-test Summary Analysis"
			}
		} else {
			if lang == "zh" {
				qName = "上传结果并生成分享链接"
			} else {
				qName = "Upload Result & Share Link"
			}
		}
		if qState {
			if lang == "zh" {
				qVal = tOnStyle.Render("开启")
			} else {
				qVal = tOnStyle.Render("ON")
			}
		} else {
			if lang == "zh" {
				qVal = tOffStyle.Render("关闭")
			} else {
				qVal = tOffStyle.Render("OFF")
			}
		}
		bodyLines = append(bodyLines, fmt.Sprintf("%s%s %s  %s\n", cur, chk, nameStyle.Render(qName), qVal))
	}

	// === VIEWPORT: show only what fits between header and footer ===
	totalBodyLines := len(bodyLines)
	avail := m.height - headerLines - footerLines - 1 // -1 for scroll indicator
	if avail < 4 || m.height == 0 {
		avail = totalBodyLines // terminal too small or unknown: show all
	}
	startLine := m.mainScrollOffset
	if startLine < 0 {
		startLine = 0
	}
	if startLine > totalBodyLines-1 {
		startLine = totalBodyLines - 1
	}
	endLine := startLine + avail
	if endLine > totalBodyLines {
		endLine = totalBodyLines
	}

	// === ASSEMBLE OUTPUT ===
	var s strings.Builder
	s.WriteString(headerStr)
	if startLine > 0 {
		if lang == "zh" {
			s.WriteString(tDimStyle.Render("  ↑ 向上滚动查看更多") + "\n")
		} else {
			s.WriteString(tDimStyle.Render("  ↑ Scroll up for more") + "\n")
		}
	}
	for _, line := range bodyLines[startLine:endLine] {
		s.WriteString(line)
	}
	if endLine < totalBodyLines {
		if lang == "zh" {
			s.WriteString(tDimStyle.Render("  ↓ 向下滚动查看更多") + "\n")
		} else {
			s.WriteString(tDimStyle.Render("  ↓ Scroll down for more") + "\n")
		}
	}
	s.WriteString(footerStr)
	return s.String()
}
