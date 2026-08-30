package api

import "regexp"

// A stop-reason line from NTrace-core can contain ICMP and may be followed by
// the next carrier header without a newline when an older nt3 formatter is in
// use. Repair that boundary at the public text boundary as a final defense.
// The first expression handles parenthesized responses; the second covers
// custom reasons without a response block. The pass repeats because an older
// formatter can concatenate more than one route section onto one line.
var (
	traceStopHeaderBoundaryPattern = regexp.MustCompile(`(?i)(Trace Stopped:[^\r\n]*\))((?:\x1b\[[0-?]*[ -/]*[@-~])*[^\r\n]*?[ \t]+-[ \t]+ICMP[ \t]+v[46][ \t]+-[ \t]*)`)
	traceStopHeaderNoParenPattern  = regexp.MustCompile(`(?i)(Trace Stopped:[^\r\n]*?\bat[ \t]+Hop[ \t]+[0-9]+)((?:\x1b\[[0-?]*[ -/]*[@-~])*[^)\r\n]*?[ \t]+-[ \t]+ICMP[ \t]+v[46][ \t]+-[ \t]*)`)
)

func normalizeTraceOutputBoundaries(value string) string {
	for {
		normalized := traceStopHeaderBoundaryPattern.ReplaceAllString(value, "$1\n$2")
		normalized = traceStopHeaderNoParenPattern.ReplaceAllString(normalized, "$1\n$2")
		if normalized == value {
			return value
		}
		value = normalized
	}
}
