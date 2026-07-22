package api

import (
	"regexp"
	"strings"
)

var (
	publicErrorURLPattern  = regexp.MustCompile(`(?i)\bhttps?://[^\s)]+`)
	publicErrorGitPattern  = regexp.MustCompile(`(?i)\bgit@[^\s:]+:[^\s]+`)
	publicErrorSecret      = regexp.MustCompile(`(?i)(authorization|bearer|token|api[_-]?key|secret|password|passwd)\s*[:=]\s*[^\s,;]+`)
	publicErrorPathPattern = regexp.MustCompile(`(?:^|\s)(?:/Users/|/Volumes/|/home/|/root/|[A-Za-z]:\\)[^\s]+`)
)

// sanitizePublicText is applied only to user-visible status/error text. It
// does not remove ordinary probe targets from successful structured payloads.
func sanitizePublicText(value string) string {
	value = publicErrorURLPattern.ReplaceAllString(value, "[remote-url]")
	value = publicErrorGitPattern.ReplaceAllString(value, "[remote-source]")
	value = publicErrorSecret.ReplaceAllString(value, "$1=[redacted]")
	value = publicErrorPathPattern.ReplaceAllString(value, " [local-path]")
	return strings.TrimSpace(value)
}
