//go:build !ecs_public

package api

import (
	"context"
	"errors"
	"time"

	privatepst "github.com/oneclickvirt/privatespeedtest/pst"
)

func runPrivateSpeedBenchmarks(ctx context.Context, limit int) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarksWithLoader(ctx, limit, privatepst.LoadServerListWithMetadataContext)
}

func runEmbeddedPrivateSpeedBenchmarks(ctx context.Context, limit int) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarksWithLoader(ctx, limit, func(context.Context) (privatepst.RegistryLoadResult, error) {
		return privatepst.LoadEmbeddedServerList()
	})
}

func runPrivateSpeedBenchmarksWithLoader(ctx context.Context, limit int, loader func(context.Context) (privatepst.RegistryLoadResult, error)) (any, int, []privateSpeedBenchmark) {
	if ctx == nil {
		ctx = context.Background()
	}
	loaded, err := loader(ctx)
	if err != nil {
		return privatepst.RegistryReport{
			SchemaVersion: "privatespeedtest.registry/v1", Fallback: true,
			Availability: privatepst.ServerUnavailable, Servers: []privatepst.RegistryNode{}, Error: err.Error(),
		}, 0, nil
	}
	registry := privatepst.ResolveLoadedServerRegistry(ctx, loaded, limit, 2*time.Second, nil)
	benchmarks := make([]privateSpeedBenchmark, 0, len(registry.Selected))
	for _, selected := range registry.Selected {
		if err := ctx.Err(); err != nil {
			benchmarks = append(benchmarks, privateSpeedBenchmark{ID: selected.ID, Name: selected.Name, Source: "privatespeedtest", Status: speedContextStatus(err), Error: err.Error()})
			continue
		}
		latency := time.Duration(selected.LatencyMS) * time.Millisecond
		latencyInfo := &privatepst.ServerWithLatencyInfo{Server: selected.Server, Latency: latency, MinLatency: latency, MaxLatency: latency}
		result := privatepst.RunSpeedTestContext(ctx, selected.Server, false, false, 4, 5*time.Second, latencyInfo, false)
		status := "unavailable"
		switch {
		case errors.Is(ctx.Err(), context.Canceled):
			status = "canceled"
		case errors.Is(ctx.Err(), context.DeadlineExceeded):
			status = "timeout"
		case result.Success:
			status = "available"
		case result.DownloadMbps > 0 || result.UploadMbps > 0:
			status = "partial"
		}
		benchmarks = append(benchmarks, privateSpeedBenchmark{
			ID: selected.ID, Name: selected.Name, Source: "privatespeedtest", Status: status,
			LatencyMS: float64(result.PingLatency) / float64(time.Millisecond), DownloadMbps: result.DownloadMbps,
			UploadMbps: result.UploadMbps, DurationMS: result.Duration.Milliseconds(), Error: result.Error,
		})
	}
	return registry, len(registry.Selected), benchmarks
}
