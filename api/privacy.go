package api

import (
	"encoding/json"
	"net"
	"net/url"
	"regexp"
	"strings"
)

const privacyRedacted = "[redacted]"

var (
	privacyIPv4Pattern = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	privacyIPv6Pattern = regexp.MustCompile(`(?i)(?:[0-9a-f]{0,4}:){2,}[0-9a-f:.%]*`)
)

// applyStructuredPrivacy removes host identity from the machine-readable
// report while preserving schemas, statuses, measurements, and data versions.
func applyStructuredPrivacy(report *StructuredReport) {
	if report == nil {
		return
	}
	report.Text = ""
	for index := range report.Sections {
		report.Sections[index].Reason = redactSensitiveText(report.Sections[index].Reason)
	}
	for index := range report.DataFiles {
		report.DataFiles[index].Reason = redactSensitiveText(report.DataFiles[index].Reason)
	}
	for index := range report.Components {
		report.Components[index].Reason = redactSensitiveText(report.Components[index].Reason)
		report.Components[index].Payload = redactJSONPayload(report.Components[index].Payload)
	}
	for index := range report.TCP {
		report.TCP[index].Target.Host = privacyRedacted
		report.TCP[index].Target.Name = redactSensitiveText(report.TCP[index].Target.Name)
	}
}

func redactJSONPayload(payload json.RawMessage) json.RawMessage {
	if len(payload) == 0 {
		return payload
	}
	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		// Invalid component JSON is already reported by componentPayload. Do not
		// risk returning the original bytes from a privacy-mode report.
		return json.RawMessage(`null`)
	}
	redactJSONValue(value, "")
	encoded, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return encoded
}

func redactJSONValue(value any, key string) {
	switch typed := value.(type) {
	case map[string]any:
		for childKey, child := range typed {
			if privacySensitiveKey(childKey) {
				// Structured probe targets carry useful non-identifying labels
				// alongside a host. Preserve the object shape and redact only its
				// sensitive descendants; scalar device targets stay redacted.
				if strings.EqualFold(strings.TrimSpace(childKey), "target") {
					switch child.(type) {
					case map[string]any, []any:
						redactJSONValue(child, childKey)
						continue
					}
				}
				typed[childKey] = privacyRedacted
				continue
			}
			switch childValue := child.(type) {
			case string:
				typed[childKey] = redactSensitiveText(childValue)
			default:
				redactJSONValue(child, childKey)
			}
		}
	case []any:
		for index, child := range typed {
			switch childValue := child.(type) {
			case string:
				typed[index] = redactSensitiveText(childValue)
			default:
				redactJSONValue(child, key)
			}
		}
	}
}

func privacySensitiveKey(key string) bool {
	normalized := strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(strings.TrimSpace(key)))
	switch normalized {
	case "ip", "ipv4", "ipv6", "publicip", "publicipv4", "publicipv6",
		"address", "localaddress", "mappedaddress", "xormappedaddress",
		"hostname", "host", "interface", "sourceinterface", "query", "target":
		return true
	}
	return strings.Contains(normalized, "serial") ||
		strings.Contains(normalized, "devicepath") ||
		strings.Contains(normalized, "filepath") ||
		strings.HasSuffix(normalized, "proxy")
}

func redactSensitiveText(value string) string {
	if value == "" {
		return value
	}
	value = privacyIPv4Pattern.ReplaceAllStringFunc(value, func(candidate string) string {
		if parsed := net.ParseIP(candidate); parsed != nil && parsed.To4() != nil {
			return "[redacted-ip]"
		}
		return candidate
	})
	value = privacyIPv6Pattern.ReplaceAllStringFunc(value, func(candidate string) string {
		trimmed := strings.Trim(candidate, "[](),;")
		address := trimmed
		if zoneIndex := strings.LastIndexByte(address, '%'); zoneIndex >= 0 {
			address = address[:zoneIndex]
		}
		if parsed := net.ParseIP(address); parsed != nil && parsed.To4() == nil {
			return "[redacted-ip]"
		}
		return candidate
	})
	if parsed, err := url.Parse(value); err == nil && parsed.User != nil {
		parsed.User = url.User(privacyRedacted)
		value = parsed.String()
	}
	return value
}
