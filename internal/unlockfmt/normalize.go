package unlockfmt

import "strings"

// Normalize makes UnlockTests output stable across old and new versions:
// legacy standalone IPV4:/IPV6: lines are removed, and legacy unversioned
// section headers are prefixed with the tested IP stack.
func Normalize(ipVersion, output string) string {
	ipLabel := ipLabel(ipVersion)
	if ipLabel == "" || output == "" {
		return output
	}
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if isStandaloneIPLabel(line) {
			lines[i] = ""
			continue
		}
		lines[i] = normalizeHeaderLine(ipLabel, line)
	}
	return compactBlankLines(strings.Join(lines, "\n"))
}

func ipLabel(ipVersion string) string {
	switch strings.ToLower(strings.TrimSpace(ipVersion)) {
	case "ipv4", "tcp4", "4":
		return "IPV4"
	case "ipv6", "tcp6", "6":
		return "IPV6"
	default:
		return ""
	}
}

func isStandaloneIPLabel(line string) bool {
	trimmed := strings.TrimSpace(stripANSIEscapeSequences(line))
	trimmed = strings.TrimSuffix(trimmed, ":")
	trimmed = strings.ToUpper(strings.TrimSpace(trimmed))
	return trimmed == "IPV4" || trimmed == "IPV6"
}

func stripANSIEscapeSequences(value string) string {
	var builder strings.Builder
	for i := 0; i < len(value); {
		if value[i] == 0x1b && i+1 < len(value) && value[i+1] == '[' {
			i += 2
			for i < len(value) {
				ch := value[i]
				i++
				if ch >= '@' && ch <= '~' {
					break
				}
			}
			continue
		}
		builder.WriteByte(value[i])
		i++
	}
	return builder.String()
}

func normalizeHeaderLine(ipLabel, line string) string {
	open := strings.Index(line, "[")
	close := strings.LastIndex(line, "]")
	if open < 0 || close <= open {
		return line
	}
	prefix := strings.TrimSpace(line[:open])
	suffix := strings.TrimSpace(line[close+1:])
	if !isHeaderBorder(prefix) || !isHeaderBorder(suffix) {
		return line
	}
	title := strings.TrimSpace(line[open+1 : close])
	upperTitle := strings.ToUpper(title)
	if title == "" || strings.HasPrefix(upperTitle, "IPV4 ") || strings.HasPrefix(upperTitle, "IPV6 ") {
		return line
	}
	return centeredHeader(ipLabel + " " + title)
}

func isHeaderBorder(value string) bool {
	if len(value) < 3 {
		return false
	}
	for _, r := range value {
		if r != '=' && r != '-' {
			return false
		}
	}
	return true
}

func centeredHeader(title string) string {
	message := "[ " + strings.TrimSpace(title) + " ]"
	totalLength := 40
	if len(message) > totalLength {
		totalLength = len(message)
	}
	paddingLength := (totalLength - len(message)) / 2
	leftPadding := strings.Repeat("=", paddingLength)
	rightPadding := strings.Repeat("=", totalLength-len(message)-paddingLength)
	return leftPadding + message + rightPadding
}

func compactBlankLines(output string) string {
	lines := strings.Split(output, "\n")
	result := make([]string, 0, len(lines))
	lastBlank := false
	for _, line := range lines {
		blank := strings.TrimSpace(line) == ""
		if blank && lastBlank {
			continue
		}
		result = append(result, line)
		lastBlank = blank
	}
	return strings.TrimLeft(strings.Join(result, "\n"), "\n")
}
