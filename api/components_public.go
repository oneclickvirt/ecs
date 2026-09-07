//go:build ecs_public

package api

import (
	"context"
	"fmt"
	"time"

	speedmodel "github.com/oneclickvirt/speedtest/model"
)

func hasPrivateComponentData() bool { return false }

func loadSecurityComponentData(ctx context.Context, _ bool) componentDataResult {
	err := fmt.Errorf("security component unavailable in public build")
	return failedComponentData(ctx, dnsblDataFile, err)
}

func loadPrivateSpeedComponentData(ctx context.Context, _ bool) componentDataResult {
	err := fmt.Errorf("private speed component unavailable in public build")
	return failedComponentData(ctx, privateDataFile, err)
}

func loadPrivateSpeedComponentDataWithNetwork(ctx context.Context, offline bool, _ speedmodel.Network) componentDataResult {
	return loadPrivateSpeedComponentData(ctx, offline)
}

func loadTransferComponentData(ctx context.Context, _ bool) componentDataResult {
	err := fmt.Errorf("transfer component unavailable in public build")
	return failedComponentData(ctx, transferDataFile, err)
}

func newPrivateStructuredSpeedPreload(*Config, componentInputs) *structuredPrivateSpeedPreload {
	return nil
}

func collectSecurityComponent(context.Context, string, string, []dnsblZoneInput) ComponentReport {
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

func runInternationalPrivateSpeedBenchmarks(context.Context, int) (any, int, []privateSpeedBenchmark) {
	return nil, 0, nil
}

func runEmbeddedInternationalPrivateSpeedBenchmarks(context.Context, int) (any, int, []privateSpeedBenchmark) {
	return nil, 0, nil
}

// The public build keeps the same family-aware adapter contract as the
// private build.  Private registry work is intentionally unavailable here,
// so each network-aware entry point remains a no-op compatibility stub.
func runPrivateSpeedBenchmarksWithNetwork(ctx context.Context, limit int, _ speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarks(ctx, limit)
}

func runEmbeddedPrivateSpeedBenchmarksWithNetwork(ctx context.Context, limit int, _ speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	return runEmbeddedPrivateSpeedBenchmarks(ctx, limit)
}

func runInternationalPrivateSpeedBenchmarksWithNetwork(ctx context.Context, limit int, _ speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	return runInternationalPrivateSpeedBenchmarks(ctx, limit)
}

func runEmbeddedInternationalPrivateSpeedBenchmarksWithNetwork(ctx context.Context, limit int, _ speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	return runEmbeddedInternationalPrivateSpeedBenchmarks(ctx, limit)
}

// chineseFullPrivateSpeedRunnerForConfigWithNetwork keeps the complete Chinese
// profile's per-carrier public-node selection buildable without linking the
// private registry. Public builds intentionally return no private benchmarks.
func chineseFullPrivateSpeedRunnerForConfigWithNetwork(bool) privateSpeedRunnerWithNetwork {
	return runPrivateSpeedBenchmarksWithNetwork
}
