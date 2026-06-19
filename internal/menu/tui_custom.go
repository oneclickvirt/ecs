package menu

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/oneclickvirt/ecs/utils"
)

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
		m.ensureCustomCursorVisible()
	case "down", "j":
		if m.customCursor < m.customTotal-1 {
			m.customCursor++
		}
		m.ensureCustomCursorVisible()
	case "home":
		m.customCursor = 0
		m.ensureCustomCursorVisible()
	case "end":
		m.customCursor = m.customTotal - 1
		m.ensureCustomCursorVisible()
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

	// === FIXED HEADER (always visible at top) ===
	var hdr strings.Builder
	hdr.WriteString("\n")
	if lang == "zh" {
		hdr.WriteString(tTitleStyle.Render(fmt.Sprintf("  VPS融合怪 %s  —  高级自定义", m.config.EcsVersion)))
	} else {
		hdr.WriteString(tTitleStyle.Render(fmt.Sprintf("  VPS Fusion Monster %s  —  Advanced Custom", m.config.EcsVersion)))
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
	headerStr := hdr.String()
	headerLines := strings.Count(headerStr, "\n")

	// === FIXED FOOTER (always visible at bottom) ===
	var ftr strings.Builder
	ftr.WriteString("\n")
	panelTitle := "  当前项说明"
	if lang == "en" {
		panelTitle = "  Current Item Description"
	}
	renderedPanel := tPanelStyle.Width(maxInt(60, m.width-6)).Render(m.currentCustomDescription(lang))
	ftr.WriteString(tSectStyle.Render(panelTitle) + "\n")
	ftr.WriteString(renderedPanel + "\n")
	if m.editingText {
		if lang == "zh" {
			ftr.WriteString("\n" + tWarnStyle.Render("  文本编辑模式: Enter 保存, Esc 取消") + "\n")
		} else {
			ftr.WriteString("\n" + tWarnStyle.Render("  Text edit mode: Enter save, Esc cancel") + "\n")
		}
		ftr.WriteString("  " + m.textInput.View() + "\n")
	}
	ftr.WriteString("\n")
	if lang == "zh" {
		ftr.WriteString(tHelpStyle.Render("  ↑/↓ 移动  Enter/空格 切换  ←/→ 改选项  a 全选  Esc 返回  q 退出"))
	} else {
		ftr.WriteString(tHelpStyle.Render("  Up/Down Move  Enter/Space Toggle  Left/Right Cycle  a All  Esc Back  q Quit"))
	}
	ftr.WriteString("\n")
	footerStr := ftr.String()
	footerLines := strings.Count(footerStr, "\n")

	// === SCROLLABLE BODY ===
	var bodyLines []string

	// Toggles section header
	if lang == "zh" {
		bodyLines = append(bodyLines, tSectStyle.Render("  测试开关 (空格切换, a 全选/全不选):")+"\n")
	} else {
		bodyLines = append(bodyLines, tSectStyle.Render("  Test Toggles (Space to toggle, a all/none):")+"\n")
	}
	bodyLines = append(bodyLines, "\n")
	for i, t := range m.toggles {
		cursor := "   "
		style := tNormStyle
		if m.customCursor == i {
			cursor = tCurStyle.Render(" > ")
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
		bodyLines = append(bodyLines, fmt.Sprintf("%s%s %s\n", cursor, check, style.Render(name)))
	}
	bodyLines = append(bodyLines, "\n")

	// Advanced section header
	if lang == "zh" {
		bodyLines = append(bodyLines, tSectStyle.Render("  参数设置 (Enter/空格切换, ←/→改选项):")+"\n")
	} else {
		bodyLines = append(bodyLines, tSectStyle.Render("  Parameter Settings (Enter/Space switch, Left/Right cycle):")+"\n")
	}
	bodyLines = append(bodyLines, "\n")

	// Advanced settings — alignment fix: compute column width using runewidth
	nameColW := advNameColWidth(m.advanced, lang)
	for i, a := range m.advanced {
		idx := len(m.toggles) + i
		cursor := "   "
		style := tNormStyle
		if m.customCursor == idx {
			cursor = tCurStyle.Render(" > ")
			style = tSelStyle
		}
		name := a.nameEn
		if lang == "zh" {
			name = a.nameZh
		}
		var valueRendered string
		switch a.kind {
		case "bool":
			if a.boolVal {
				if lang == "zh" {
					valueRendered = tOnStyle.Render("开启")
				} else {
					valueRendered = tOnStyle.Render("ON")
				}
			} else {
				if lang == "zh" {
					valueRendered = tOffStyle.Render("关闭")
				} else {
					valueRendered = tOffStyle.Render("OFF")
				}
			}
		case "option":
			op := a.options[a.current]
			lbl := op.labelEn
			if lang == "zh" {
				lbl = op.labelZh
			}
			valueRendered = tDimStyle.Render("< ") + tValStyle.Render(lbl) + tDimStyle.Render(" >")
		case "text":
			v := strings.TrimSpace(a.textVal)
			if v == "" {
				if lang == "zh" {
					valueRendered = tDimStyle.Render("(默认)")
				} else {
					valueRendered = tDimStyle.Render("(default)")
				}
			} else {
				valueRendered = tValStyle.Render(v)
			}
		}
		// Alignment: pad the name column to nameColW visible cells
		nameWithColon := name + ":"
		visW := runewidth.StringWidth(nameWithColon)
		padLen := nameColW - visW
		if padLen < 1 {
			padLen = 1
		}
		padding := strings.Repeat(" ", padLen)
		bodyLines = append(bodyLines, fmt.Sprintf("%s%s%s %s\n", cursor, style.Render(nameWithColon), padding, valueRendered))
	}
	bodyLines = append(bodyLines, "\n")

	// Confirm button
	confirmIdx := m.customTotal - 1
	if m.customCursor == confirmIdx {
		if lang == "zh" {
			bodyLines = append(bodyLines, fmt.Sprintf("   %s\n", tBtnStyle.Render(">> 开始测试 <<")))
		} else {
			bodyLines = append(bodyLines, fmt.Sprintf("   %s\n", tBtnStyle.Render(">> Start Test <<")))
		}
	} else {
		if lang == "zh" {
			bodyLines = append(bodyLines, fmt.Sprintf("   %s\n", tBtnDimStyle.Render(">> 开始测试 <<")))
		} else {
			bodyLines = append(bodyLines, fmt.Sprintf("   %s\n", tBtnDimStyle.Render(">> Start Test <<")))
		}
	}

	// === VIEWPORT: show only what fits between header and footer ===
	totalBodyLines := len(bodyLines)
	avail := m.height - headerLines - footerLines - 1 // -1 for scroll indicator
	if avail < 4 || m.height == 0 {
		avail = totalBodyLines
	}
	startLine := m.customScrollOffset
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

// advNameColWidth returns the visible-cell width needed for the name column
// in the advanced settings panel, computed from the widest name + colon + 1 space.
