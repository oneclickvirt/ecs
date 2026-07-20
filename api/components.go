package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// structuredHardwareContextKey marks a structured collection that follows an
// already completed legacy text run. Legacy CPU, memory, and disk tests are
// intentionally not repeated for that report: they may write files or consume
// significant time, and the published adapter cannot reconstruct their
// payloads without a second execution.
type structuredHardwareContextKey struct{}

func skipStructuredHardware(ctx context.Context) context.Context {
	return context.WithValue(ctx, structuredHardwareContextKey{}, true)
}

func shouldSkipStructuredHardware(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	value, _ := ctx.Value(structuredHardwareContextKey{}).(bool)
	return value
}

// UsesStructuredComponents reports whether this build has the local
// structured adapters that must own component execution. CLI callers use it
// to select the same bounded orchestration path as API and GUI callers.
func UsesStructuredComponents() bool {
	return structuredOwnsHardware() || structuredOwnsNetwork()
}

// collectComponentReports is implemented by the release-compatible adapter
// and by the local-components adapter. Keeping the adapter behind a build
// tag lets ecs be built from published module versions while component repos
// are being released in the documented order.
type componentInputs struct {
	TCPTargets          []TCPTarget
	ProvinceRoutes      []byte
	SpeedtestServers    []byte
	OpenSpeedtestServer []byte
	DNSBLZones          []byte
	MediaProviders      []byte
	BGPASNMap           []byte
	PublicIPv4          string
	PublicIPv6          string
	Network             bool
}

func collectComponentReports(ctx context.Context, config *Config, inputs componentInputs) []ComponentReport {
	return collectPublishedComponentReports(ctx, config, inputs)
}

func componentPayload(name, schema string, status ReportStatus, started time.Time, payload any, err error) ComponentReport {
	report := ComponentReport{
		Name: name, SchemaVersion: schema, Status: status,
		DurationMS: time.Since(started).Milliseconds(),
	}
	if err != nil {
		if report.Status == ReportStatusOK {
			report.Status = ReportStatusError
		}
		report.Reason = err.Error()
	}
	if payload != nil {
		encoded, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			report.Status = ReportStatusError
			report.Reason = fmt.Sprintf("encode component payload: %v", marshalErr)
		} else {
			report.Payload = encoded
		}
	}
	return report
}

func componentStatus(raw string) ReportStatus {
	switch raw {
	case "ok", "available", "success":
		return ReportStatusOK
	case "partial":
		return ReportStatusPartial
	case "timeout", "timed_out":
		return ReportStatusTimeout
	case "canceled", "cancelled":
		return ReportStatusCanceled
	case "unavailable", "unsupported":
		return ReportStatusUnavailable
	case "skipped":
		return ReportStatusSkipped
	default:
		return ReportStatusError
	}
}
