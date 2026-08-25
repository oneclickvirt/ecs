package api

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/oneclickvirt/ecs/internal/runner"
)

var publicLogSanitizeMu sync.Mutex

func init() {
	// API and GUI callers do not pass through main.go, so install the same
	// line-level boundary used by the CLI before any legacy formatter runs.
	runner.SetOutputSanitizer(SanitizeOutput)
}

var (
	publicErrorURLPattern  = regexp.MustCompile(`(?i)\b(?:https?|ftp)://[^\s)]+`)
	publicErrorGitPattern  = regexp.MustCompile(`(?i)\bgit@[^\s:]+:[^\s]+`)
	publicErrorScpPattern  = regexp.MustCompile(`(?i)\b(?:ssh|git|https?)://[^\s]+`)
	publicErrorAuthHeader  = regexp.MustCompile(`(?i)\bauthorization\s*[:=]\s*(?:bearer|basic)?\s*[^\s,;]+`)
	publicErrorBearer      = regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/=-]+`)
	publicErrorSecret      = regexp.MustCompile(`(?i)(token|api[_-]?key|secret|password|passwd|auth)\s*[:=]\s*["']?[^\s,;"'&#]+`)
	publicErrorPathPattern = regexp.MustCompile(`(?:^|\s)(?:/Users/|/Volumes/|/home/|/root/|[A-Za-z]:\\)[^\s]+`)
	publicErrorQuerySecret = regexp.MustCompile(`(?i)([?&](?:access[_-]?token|token|api[_-]?key|key|secret|password|passwd|auth|authorization)=)[^&#\s]+`)
	publicErrorRepoPath    = regexp.MustCompile(`(?i)\b(?:github\.com|gitlab\.com|bitbucket\.org)/[^\s/]+/[^\s]+`)
	publicURLUserInfo      = regexp.MustCompile(`(?i)\b((?:https?|ftp)://)[^/@\s]+@`)
)

// sanitizePublicText is applied only to user-visible status/error text. It
// does not remove ordinary probe targets from successful structured payloads.
func sanitizePublicText(value string) string {
	value = sanitizePublicCredentials(value)
	value = publicErrorURLPattern.ReplaceAllString(value, "[remote-url]")
	value = publicErrorGitPattern.ReplaceAllString(value, "[remote-source]")
	value = publicErrorScpPattern.ReplaceAllString(value, "[remote-source]")
	value = publicErrorRepoPath.ReplaceAllString(value, "[remote-source]")
	value = publicErrorPathPattern.ReplaceAllString(value, " [local-path]")
	return strings.TrimSpace(value)
}

// sanitizePublicReason converts arbitrary dependency errors into text that is
// suitable for a report or progress event. In particular it prevents registry
// endpoints, repository paths and query credentials from crossing the public
// output boundary.
func sanitizePublicReason(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return sanitizePublicText(value)
}

// sanitizePublicOutput preserves the established result layout and public
// project links while redacting URLs and paths that appear in loader/source
// diagnostics. It is intentionally line based so ordinary benchmark targets
// and project footer links are not rewritten.
func sanitizePublicOutput(value string) string {
	lines := strings.Split(value, "\n")
	for index, line := range lines {
		if containsPublicCredential(line) {
			line = sanitizePublicCredentials(line)
			lines[index] = line
		}
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "://") && !strings.Contains(lower, "git@") && !publicErrorPathPattern.MatchString(line) && !publicErrorSecret.MatchString(line) {
			continue
		}
		if containsAny(lower,
			"data source", "datasource", "registry", "manifest", "fallback", "source url", "source:",
			"fetch", "download", "load ", "loading", "error", "failed", "warning", "warn",
			"数据源", "来源", "回退", "清单", "下载", "加载", "获取", "错误", "失败", "警告",
		) {
			lines[index] = sanitizePublicText(line)
		}
	}
	return strings.Join(lines, "\n")
}

// SanitizeOutput applies the public text boundary to a completed classic run.
// It preserves headings, project links and benchmark target rows.
func SanitizeOutput(value string) string {
	return sanitizePublicOutput(value)
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

// sanitizeComponentPayload removes registry provenance from a machine-readable
// component payload. Probe targets, speed-test node hosts/URLs, business URLs,
// and measurements remain intact. Remote addresses are redacted only when they
// occur inside error-like text.
func sanitizeComponentPayload(payload json.RawMessage) json.RawMessage {
	if len(payload) == 0 {
		return payload
	}
	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		return json.RawMessage(`null`)
	}
	sanitizePublicJSONValue(value)
	encoded, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return encoded
}

func sanitizeStructuredReport(report *StructuredReport) {
	if report == nil {
		return
	}
	// Provenance is retained only in the private loader stack. It is not part
	// of the public Go result, JSON schema, text report, GUI or upload payload.
	report.Data = nil
	report.DataFiles = nil
	report.Text = sanitizePublicOutput(report.Text)
	if report.DNS != nil {
		report.DNS.Requested = sanitizePublicText(report.DNS.Requested)
		report.DNS.Active = sanitizePublicText(report.DNS.Active)
		report.DNS.Provider = sanitizePublicText(report.DNS.Provider)
		report.DNS.Reason = sanitizePublicReason(report.DNS.Reason)
	}
	for index := range report.Sections {
		report.Sections[index].Reason = sanitizePublicReason(report.Sections[index].Reason)
	}
	for index := range report.Components {
		report.Components[index].Reason = sanitizePublicReason(report.Components[index].Reason)
		report.Components[index].Payload = sanitizeComponentPayload(report.Components[index].Payload)
	}
}

func sanitizePublicJSONValue(value any) {
	sanitizePublicJSONValueScoped(value, "")
}

func sanitizePublicJSONValueScoped(value any, endpointScope string) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			normalized := normalizePublicJSONKey(key)
			if isProvenanceJSONKey(normalized, endpointScope) || isCredentialJSONKey(normalized) || (endpointScope != "" && isEndpointJSONKey(normalized)) {
				delete(typed, key)
				continue
			}
			childScope := endpointScope
			if normalized == "rdap" || normalized == "whois" || normalized == "geofeed" || normalized == "geofeeds" {
				childScope = normalized
			}
			if isProvenanceContainer(normalized) {
				childScope = "provenance"
			}
			switch childValue := child.(type) {
			case string:
				if isErrorJSONKey(normalized) {
					typed[key] = sanitizePublicReason(childValue)
				} else if containsPublicCredential(childValue) {
					typed[key] = sanitizePublicCredentials(childValue)
				}
			default:
				sanitizePublicJSONValueScoped(child, childScope)
			}
		}
	case []any:
		for _, child := range typed {
			sanitizePublicJSONValueScoped(child, endpointScope)
		}
	}
}

func normalizePublicJSONKey(key string) string {
	return strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(strings.TrimSpace(key)))
}

func isProvenanceJSONKey(normalized, scope string) bool {
	switch normalized {
	case "privateregistry", "datafiles", "datasource",
		"sourceurl", "fallbackurl", "manifesturl", "registryurl", "registrysource", "registryfallback",
		"geofeedurl", "geofeedurls", "whoisserver", "whoisurl", "whoisendpoint",
		"rdapserver", "rdapurl", "rdapendpoint":
		return true
	}
	if scope == "provenance" && (normalized == "source" || normalized == "fallback" || normalized == "manifest" || normalized == "registry") {
		return true
	}
	if strings.HasPrefix(normalized, "geofeed") && containsAny(normalized, "url", "uri", "endpoint", "server", "source") {
		return true
	}
	if (strings.HasPrefix(normalized, "whois") || strings.HasPrefix(normalized, "rdap")) &&
		containsAny(normalized, "url", "uri", "endpoint", "server", "source") {
		return true
	}
	return false
}

func isProvenanceContainer(normalized string) bool {
	return normalized == "registry" || normalized == "registryreport" || normalized == "registryresolution" ||
		normalized == "serverregistry" || normalized == "providermetadata" || normalized == "metadataregistry" ||
		normalized == "manifest"
}

func isEndpointJSONKey(normalized string) bool {
	return normalized == "url" || normalized == "uri" || normalized == "href" || normalized == "server" ||
		normalized == "endpoint" || normalized == "baseurl"
}

func isCredentialJSONKey(normalized string) bool {
	return normalized == "authorization" || normalized == "bearer" || normalized == "token" || normalized == "apikey" ||
		normalized == "secret" || normalized == "password" || normalized == "passwd" || normalized == "credential" ||
		strings.HasSuffix(normalized, "token") || strings.HasSuffix(normalized, "apikey") || strings.HasSuffix(normalized, "secret") ||
		strings.HasSuffix(normalized, "password") || strings.HasSuffix(normalized, "credential")
}

func isErrorJSONKey(normalized string) bool {
	return normalized == "error" || normalized == "reason" || normalized == "message" || normalized == "detail" ||
		strings.HasSuffix(normalized, "error") || strings.HasSuffix(normalized, "reason")
}

// SanitizeLogFile closes the public ecs.log boundary used by component
// loggers. It rewrites only provenance/loading diagnostics and leaves project
// links and ordinary probe/target URLs untouched.
func SanitizeLogFile(path string) error {
	publicLogSanitizeMu.Lock()
	defer publicLogSanitizeMu.Unlock()

	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	cleaned := sanitizePublicLog(string(data))
	if cleaned != string(data) {
		if err := os.WriteFile(path, []byte(cleaned), 0o600); err != nil {
			return err
		}
	}
	return os.Chmod(path, 0o600)
}

func sanitizeConfiguredLog(config *Config) {
	if config != nil && config.EnableLogger {
		_ = SanitizeLogFile("ecs.log")
	}
}

func finalizePublicText(config *Config, output string) string {
	output = sanitizePublicOutput(output)
	sanitizeConfiguredLog(config)
	return output
}

func sanitizePublicLog(value string) string {
	lines := strings.Split(value, "\n")
	for index, line := range lines {
		if containsPublicCredential(line) {
			line = sanitizePublicCredentials(line)
			lines[index] = line
		}
		lower := strings.ToLower(line)
		if !containsAny(lower, "://", "git@", "github.com/", "gitlab.com/", "bitbucket.org/") &&
			!publicErrorPathPattern.MatchString(line) && !publicErrorSecret.MatchString(line) {
			continue
		}
		if isProvenanceDiagnostic(lower) {
			lines[index] = sanitizePublicText(line)
		}
	}
	return strings.Join(lines, "\n")
}

func containsPublicCredential(value string) bool {
	return publicErrorQuerySecret.MatchString(value) || publicErrorAuthHeader.MatchString(value) ||
		publicErrorBearer.MatchString(value) || publicErrorSecret.MatchString(value) || publicURLUserInfo.MatchString(value)
}

func sanitizePublicCredentials(value string) string {
	value = publicURLUserInfo.ReplaceAllString(value, "$1[redacted]@")
	value = publicErrorQuerySecret.ReplaceAllString(value, "$1[redacted]")
	value = publicErrorAuthHeader.ReplaceAllString(value, "authorization=[redacted]")
	value = publicErrorBearer.ReplaceAllString(value, "bearer=[redacted]")
	return publicErrorSecret.ReplaceAllString(value, "$1=[redacted]")
}

func isProvenanceDiagnostic(lower string) bool {
	return containsAny(lower,
		"data source", "datasource", "source url", "source:", "registry", "manifest", "fallback",
		"metadata source", "provider metadata", "snapshot", "dataset", "server list", "target list",
		"geofeed url", "geofeed endpoint", "whois server", "whois endpoint", "rdap server", "rdap endpoint",
		"load ", "loading ", "fetch ", "download ",
		"数据源", "来源地址", "回退地址", "清单", "注册表", "元数据地址", "快照地址", "加载 ", "获取 ", "下载 ",
	)
}
