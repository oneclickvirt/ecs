package tests

import "github.com/oneclickvirt/nt3/nt"

// Keep the existing goecs formatter seam while delegating the boundary rules
// to the released nt3 component. This prevents the two projects from drifting
// when NTrace output changes again.
func formatNextTraceOutput(lines []string) string {
	return nt.FormatTraceOutput(lines)
}
