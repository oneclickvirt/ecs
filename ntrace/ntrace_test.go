package ntrace

import "testing"

// https://github.com/nxtrace/NTrace-core/blob/main/fast_trace/fast_trace.go
func TestTraceRoute(t *testing.T) {
	TraceRoute3("en", "GZ", "ipv4")
}
