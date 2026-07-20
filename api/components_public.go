//go:build ecs_public

package api

import (
	"context"
	"time"
)

func collectSecurityComponent(context.Context, string, string, []byte) ComponentReport {
	report := componentPayload("security.evidence", "goecs.security/v1", ReportStatusUnavailable, time.Now(), nil, nil)
	report.Reason = "security component unavailable in public build"
	return report
}

func runPrivateSpeedBenchmarks(context.Context, int) (any, int, []privateSpeedBenchmark) {
	return nil, 0, nil
}

func runEmbeddedPrivateSpeedBenchmarks(context.Context, int) (any, int, []privateSpeedBenchmark) {
	return nil, 0, nil
}
