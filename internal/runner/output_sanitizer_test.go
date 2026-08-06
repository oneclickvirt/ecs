package runner

import "testing"

func TestSanitizeRunnerOutputUsesConfiguredBoundary(t *testing.T) {
	SetOutputSanitizer(func(string) string { return "sanitized" })
	defer SetOutputSanitizer(nil)
	if got := sanitizeRunnerOutput("private"); got != "sanitized" {
		t.Fatalf("sanitizeRunnerOutput() = %q, want sanitized", got)
	}
}

func TestSanitizeRunnerOutputFailsClosedOnPanic(t *testing.T) {
	SetOutputSanitizer(func(string) string { panic("broken sanitizer") })
	defer SetOutputSanitizer(nil)
	if got := sanitizeRunnerOutput("private"); got != "output unavailable" {
		t.Fatalf("sanitizeRunnerOutput() = %q, want fail-closed output", got)
	}
}
