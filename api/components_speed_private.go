//go:build !ecs_public

package api

import (
	"context"
	"errors"
	"strings"
	"time"

	privatepst "github.com/oneclickvirt/privatespeedtest/pst"
	privatetransfer "github.com/oneclickvirt/privatespeedtest/transfer"
)

func hasPrivateComponentData() bool { return true }

func loadPrivateSpeedComponentData(ctx context.Context, offline bool) componentDataResult {
	var loaded privatepst.RegistryLoadResult
	var err error
	if offline {
		loaded, err = privatepst.LoadEmbeddedServerList()
	} else {
		loaded, err = privatepst.LoadServerListWithMetadataContext(ctx)
	}
	if err != nil {
		return failedComponentData(ctx, privateDataFile, err)
	}
	file := stringMetadataFile(privateDataFile, loaded.Metadata.Schema, loaded.Metadata.GeneratedAt, loaded.Source, loaded.Fallback, loaded.Metadata.Count)
	return componentDataResult{file: file}
}

func loadTransferComponentData(ctx context.Context, offline bool) componentDataResult {
	var loaded privatetransfer.RegistryLoadResult
	var err error
	if offline {
		loaded, err = privatetransfer.LoadEmbeddedRegistry()
	} else {
		loaded, err = privatetransfer.LoadDefaultRegistry(ctx)
	}
	if err != nil {
		return failedComponentData(ctx, transferDataFile, err)
	}
	file := stringMetadataFile(transferDataFile, loaded.Metadata.Schema, loaded.Metadata.GeneratedAt, loaded.Source, loaded.Fallback, loaded.Metadata.Count)
	targets := make([]transferTargetInput, 0, len(loaded.Targets))
	for _, target := range loaded.Targets {
		targets = append(targets, transferTargetInput{
			ID: target.ID, Host: target.Host, PortFrom: target.PortFrom, PortTo: target.PortTo,
			Provider: target.Provider, Country: target.Country, City: target.City, Status: target.Status,
		})
	}
	return componentDataResult{file: file, apply: func(inputs *componentInputs) { inputs.TransferTargets = targets }}
}

func runPrivateSpeedBenchmarks(ctx context.Context, limit int) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarksWithLoader(ctx, limit, privatepst.LoadServerListWithMetadataContext)
}

func runEmbeddedPrivateSpeedBenchmarks(ctx context.Context, limit int) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarksWithLoader(ctx, limit, func(context.Context) (privatepst.RegistryLoadResult, error) {
		return privatepst.LoadEmbeddedServerList()
	})
}

func runInternationalPrivateSpeedBenchmarks(ctx context.Context, limit int) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarksWithLoader(ctx, limit, func(ctx context.Context) (privatepst.RegistryLoadResult, error) {
		loaded, err := privatepst.LoadServerListWithMetadataContext(ctx)
		return filterInternationalPrivateRegistry(loaded), err
	})
}

func runEmbeddedInternationalPrivateSpeedBenchmarks(ctx context.Context, limit int) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarksWithLoader(ctx, limit, func(context.Context) (privatepst.RegistryLoadResult, error) {
		loaded, err := privatepst.LoadEmbeddedServerList()
		return filterInternationalPrivateRegistry(loaded), err
	})
}

func filterInternationalPrivateRegistry(loaded privatepst.RegistryLoadResult) privatepst.RegistryLoadResult {
	if loaded.List == nil {
		return loaded
	}
	copyList := *loaded.List
	copyList.Servers = make([]privatepst.ServerConfig, 0, len(loaded.List.Servers))
	for _, server := range loaded.List.Servers {
		if strings.TrimSpace(server.Country) != "" && !isMainlandChinaCountry(server.Country) {
			copyList.Servers = append(copyList.Servers, server)
		}
	}
	copyList.TotalServers = len(copyList.Servers)
	loaded.List = &copyList
	return loaded
}

func runPrivateSpeedBenchmarksWithLoader(ctx context.Context, limit int, loader func(context.Context) (privatepst.RegistryLoadResult, error)) (any, int, []privateSpeedBenchmark) {
	if ctx == nil {
		ctx = context.Background()
	}
	if limit <= 0 {
		limit = 2
	}
	loaded, err := loader(ctx)
	if err != nil {
		return privatepst.RegistryReport{
			SchemaVersion: "privatespeedtest.registry/v1", Fallback: true,
			Availability: privatepst.ServerUnavailable, Servers: []privatepst.RegistryNode{}, Error: err.Error(),
		}, 0, nil
	}
	registry := privatepst.ResolveLoadedServerRegistry(ctx, loaded, limit, 2*time.Second, nil)
	selected, benchmarks := runPrivateSpeedBenchmarksFromRegistry(ctx, limit, registry, func(ctx context.Context, node privatepst.RegistryNode, latencyInfo *privatepst.ServerWithLatencyInfo) privatepst.SpeedTestResult {
		return privatepst.RunSpeedTestContext(ctx, node.Server, false, false, 4, 5*time.Second, latencyInfo, false)
	})
	return registry, selected, benchmarks
}

type privateSpeedTestFunc func(context.Context, privatepst.RegistryNode, *privatepst.ServerWithLatencyInfo) privatepst.SpeedTestResult

func runPrivateSpeedBenchmarksFromRegistry(ctx context.Context, limit int, registry privatepst.RegistryReport, speedTest privateSpeedTestFunc) (int, []privateSpeedBenchmark) {
	if ctx == nil {
		ctx = context.Background()
	}
	attempts := make([]privatepst.RegistryNode, 0, len(registry.Selected)+len(registry.Standby))
	attempts = append(attempts, registry.Selected...)
	attempts = append(attempts, registry.Standby...)
	benchmarks := make([]privateSpeedBenchmark, 0, len(attempts))
	usable := 0
	for _, selected := range attempts {
		if usable >= limit {
			break
		}
		if err := ctx.Err(); err != nil {
			benchmarks = append(benchmarks, privateSpeedBenchmark{ID: selected.ID, Name: selected.Name, Source: "privatespeedtest", Status: speedContextStatus(err), Error: err.Error()})
			break
		}
		latency := time.Duration(selected.LatencyMS) * time.Millisecond
		latencyInfo := &privatepst.ServerWithLatencyInfo{
			Server: selected.Server, Latency: latency, MinLatency: latency, MaxLatency: latency,
			Availability: selected.Availability, ProbeMethod: selected.ProbeMethod, ProbeResults: selected.ProbeResults,
		}
		result := speedTest(ctx, selected, latencyInfo)
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
		if result.DownloadMbps > 0 || result.UploadMbps > 0 {
			usable++
		}
		benchmarks = append(benchmarks, privateSpeedBenchmark{
			ID: selected.ID, Name: selected.Name, Source: "privatespeedtest", Status: status,
			LatencyMS: float64(result.PingLatency) / float64(time.Millisecond), DownloadMbps: result.DownloadMbps,
			UploadMbps: result.UploadMbps, DurationMS: result.Duration.Milliseconds(), Error: result.Error,
		})
	}
	return len(benchmarks), benchmarks
}
