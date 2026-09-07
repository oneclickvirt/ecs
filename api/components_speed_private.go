//go:build !ecs_public

package api

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	privatepst "github.com/oneclickvirt/privatespeedtest/pst"
	privatetransfer "github.com/oneclickvirt/privatespeedtest/transfer"
	speedmodel "github.com/oneclickvirt/speedtest/model"
)

func hasPrivateComponentData() bool { return true }

func loadPrivateSpeedComponentData(ctx context.Context, offline bool) componentDataResult {
	return loadPrivateSpeedComponentDataWithNetwork(ctx, offline, speedmodel.NetworkAuto)
}

func loadPrivateSpeedComponentDataWithNetwork(ctx context.Context, offline bool, network speedmodel.Network) componentDataResult {
	var loaded privatepst.RegistryLoadResult
	var err error
	if offline {
		loaded, err = privatepst.LoadEmbeddedServerList()
	} else {
		loaded, err = privatepst.LoadServerListWithMetadataContextWithNetwork(ctx, privateSpeedNetwork(network))
	}
	if err != nil {
		return failedComponentData(ctx, privateDataFile, err)
	}
	file := stringMetadataFile(privateDataFile, loaded.Metadata.Schema, loaded.Metadata.GeneratedAt, loaded.Source, loaded.Fallback, loaded.Metadata.Count)
	return componentDataResult{file: file, apply: func(inputs *componentInputs) { inputs.PrivateSpeedData = loaded }}
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
	return runPrivateSpeedBenchmarksWithNetwork(ctx, limit, speedmodel.NetworkAuto)
}

func runEmbeddedPrivateSpeedBenchmarks(ctx context.Context, limit int) (any, int, []privateSpeedBenchmark) {
	return runEmbeddedPrivateSpeedBenchmarksWithNetwork(ctx, limit, speedmodel.NetworkAuto)
}

func runInternationalPrivateSpeedBenchmarks(ctx context.Context, limit int) (any, int, []privateSpeedBenchmark) {
	return runInternationalPrivateSpeedBenchmarksWithNetwork(ctx, limit, speedmodel.NetworkAuto)
}

func runEmbeddedInternationalPrivateSpeedBenchmarks(ctx context.Context, limit int) (any, int, []privateSpeedBenchmark) {
	return runEmbeddedInternationalPrivateSpeedBenchmarksWithNetwork(ctx, limit, speedmodel.NetworkAuto)
}

func runPrivateSpeedBenchmarksWithNetwork(ctx context.Context, limit int, network speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarksWithLoaderAndNetwork(ctx, limit, func(ctx context.Context, network speedmodel.Network) (privatepst.RegistryLoadResult, error) {
		return privatepst.LoadServerListWithMetadataContextWithNetwork(ctx, privateSpeedNetwork(network))
	}, network)
}

func runEmbeddedPrivateSpeedBenchmarksWithNetwork(ctx context.Context, limit int, network speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarksWithLoaderAndNetwork(ctx, limit, func(context.Context, speedmodel.Network) (privatepst.RegistryLoadResult, error) {
		return privatepst.LoadEmbeddedServerList()
	}, network)
}

func runInternationalPrivateSpeedBenchmarksWithNetwork(ctx context.Context, limit int, network speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarksWithLoaderAndNetwork(ctx, limit, func(ctx context.Context, network speedmodel.Network) (privatepst.RegistryLoadResult, error) {
		loaded, err := privatepst.LoadServerListWithMetadataContextWithNetwork(ctx, privateSpeedNetwork(network))
		return filterInternationalPrivateRegistry(loaded), err
	}, network)
}

func runEmbeddedInternationalPrivateSpeedBenchmarksWithNetwork(ctx context.Context, limit int, network speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	return runPrivateSpeedBenchmarksWithLoaderAndNetwork(ctx, limit, func(context.Context, speedmodel.Network) (privatepst.RegistryLoadResult, error) {
		loaded, err := privatepst.LoadEmbeddedServerList()
		return filterInternationalPrivateRegistry(loaded), err
	}, network)
}

// chineseFullPrivateSpeedRegistryReport preserves the separately selected
// carrier registries in the complete Chinese profile. The source registry is
// loaded once; each carrier then gets its historical SpNum quota.
type chineseFullPrivateSpeedRegistryReport struct {
	SchemaVersion string                              `json:"schema_version"`
	Source        string                              `json:"source,omitempty"`
	Fallback      bool                                `json:"fallback"`
	Metadata      privatepst.RegistryMetadata         `json:"metadata"`
	Carriers      []chineseFullPrivateSpeedCarrierSet `json:"carriers"`
	Error         string                              `json:"error,omitempty"`
}

type chineseFullPrivateSpeedCarrierSet struct {
	Carrier  string                    `json:"carrier"`
	Registry privatepst.RegistryReport `json:"registry"`
}

type privateRegistryResolver func(context.Context, privatepst.RegistryLoadResult, int, speedmodel.Network) privatepst.RegistryReport

// chineseFullPrivateSpeedRunnerForConfigWithNetwork selects the same embedded
// or managed registry source as the normal structured speed runner.
func chineseFullPrivateSpeedRunnerForConfigWithNetwork(offline bool) privateSpeedRunnerWithNetwork {
	if offline {
		return runEmbeddedChineseFullPrivateSpeedBenchmarksWithNetwork
	}
	return runChineseFullPrivateSpeedBenchmarksWithNetwork
}

func runChineseFullPrivateSpeedBenchmarksWithNetwork(ctx context.Context, limit int, network speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	return runChineseFullPrivateSpeedBenchmarksWithLoaderAndNetwork(ctx, limit, func(ctx context.Context, network speedmodel.Network) (privatepst.RegistryLoadResult, error) {
		return privatepst.LoadServerListWithMetadataContextWithNetwork(ctx, privateSpeedNetwork(network))
	}, network)
}

func runEmbeddedChineseFullPrivateSpeedBenchmarksWithNetwork(ctx context.Context, limit int, network speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	return runChineseFullPrivateSpeedBenchmarksWithLoaderAndNetwork(ctx, limit, func(context.Context, speedmodel.Network) (privatepst.RegistryLoadResult, error) {
		return privatepst.LoadEmbeddedServerList()
	}, network)
}

func runChineseFullPrivateSpeedBenchmarksWithLoaderAndNetwork(ctx context.Context, limit int, loader func(context.Context, speedmodel.Network) (privatepst.RegistryLoadResult, error), network speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	if ctx == nil {
		ctx = context.Background()
	}
	loaded, err := loader(ctx, network)
	if err != nil {
		return chineseFullPrivateSpeedRegistryReport{
			SchemaVersion: "goecs.speed.private-carriers/v1", Error: err.Error(),
		}, 0, nil
	}
	resolver := func(ctx context.Context, carrierLoaded privatepst.RegistryLoadResult, carrierLimit int, network speedmodel.Network) privatepst.RegistryReport {
		return privatepst.ResolveLoadedServerRegistryWithNetwork(ctx, carrierLoaded, carrierLimit, 2*time.Second, nil, privateSpeedNetwork(network))
	}
	speedTest := func(ctx context.Context, node privatepst.RegistryNode, latencyInfo *privatepst.ServerWithLatencyInfo) privatepst.SpeedTestResult {
		return privatepst.RunSpeedTestContextWithNetwork(ctx, node.Server, false, false, 4, 5*time.Second, latencyInfo, false, privateSpeedNetwork(network))
	}
	return runChineseFullPrivateSpeedBenchmarksFromLoaded(ctx, limit, loaded, network, resolver, speedTest)
}

func runChineseFullPrivateSpeedBenchmarksFromLoaded(ctx context.Context, limit int, loaded privatepst.RegistryLoadResult, network speedmodel.Network, resolver privateRegistryResolver, speedTest privateSpeedTestFunc) (any, int, []privateSpeedBenchmark) {
	if ctx == nil {
		ctx = context.Background()
	}
	if limit <= 0 {
		limit = 2
	}
	if resolver == nil {
		resolver = func(ctx context.Context, carrierLoaded privatepst.RegistryLoadResult, carrierLimit int, network speedmodel.Network) privatepst.RegistryReport {
			return privatepst.ResolveLoadedServerRegistryWithNetwork(ctx, carrierLoaded, carrierLimit, 2*time.Second, nil, privateSpeedNetwork(network))
		}
	}
	if speedTest == nil {
		speedTest = func(ctx context.Context, node privatepst.RegistryNode, latencyInfo *privatepst.ServerWithLatencyInfo) privatepst.SpeedTestResult {
			return privatepst.RunSpeedTestContextWithNetwork(ctx, node.Server, false, false, 4, 5*time.Second, latencyInfo, false, privateSpeedNetwork(network))
		}
	}

	report := chineseFullPrivateSpeedRegistryReport{
		SchemaVersion: "goecs.speed.private-carriers/v1",
		Source:        loaded.Source,
		Fallback:      loaded.Fallback,
		Metadata:      loaded.Metadata,
		Carriers:      make([]chineseFullPrivateSpeedCarrierSet, 0, 3),
	}
	benchmarks := make([]privateSpeedBenchmark, 0, 3*limit)
	selected := 0
	// This is the same carrier ordering as the old complete terminal profile.
	for _, carrier := range []struct {
		label string
		isp   string
	}{
		{label: "unicom", isp: "Unicom"},
		{label: "telecom", isp: "Telecom"},
		{label: "mobile", isp: "Mobile"},
	} {
		carrierLoaded := filterPrivateRegistryByCarrier(loaded, carrier.isp)
		registry := resolver(ctx, carrierLoaded, limit, network)
		count, carrierBenchmarks := runPrivateSpeedBenchmarksFromRegistry(ctx, limit, registry, speedTest)
		report.Carriers = append(report.Carriers, chineseFullPrivateSpeedCarrierSet{Carrier: carrier.label, Registry: registry})
		selected += count
		benchmarks = append(benchmarks, carrierBenchmarks...)
	}
	return report, selected, benchmarks
}

func filterPrivateRegistryByCarrier(loaded privatepst.RegistryLoadResult, carrier string) privatepst.RegistryLoadResult {
	if loaded.List == nil {
		return loaded
	}
	copyList := *loaded.List
	copyList.Servers = privatepst.FilterServersByISP(loaded.List.Servers, carrier)
	copyList.TotalServers = len(copyList.Servers)
	loaded.List = &copyList
	return loaded
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
	return runPrivateSpeedBenchmarksWithLoaderAndNetwork(ctx, limit, func(ctx context.Context, _ speedmodel.Network) (privatepst.RegistryLoadResult, error) {
		return loader(ctx)
	}, speedmodel.NetworkAuto)
}

func runPrivateSpeedBenchmarksWithLoaderAndNetwork(ctx context.Context, limit int, loader func(context.Context, speedmodel.Network) (privatepst.RegistryLoadResult, error), network speedmodel.Network) (any, int, []privateSpeedBenchmark) {
	if ctx == nil {
		ctx = context.Background()
	}
	if limit <= 0 {
		limit = 2
	}
	loaded, err := loader(ctx, network)
	if err != nil {
		return privatepst.RegistryReport{
			SchemaVersion: "privatespeedtest.registry/v1", Fallback: true,
			Availability: privatepst.ServerUnavailable, Servers: []privatepst.RegistryNode{}, Error: err.Error(),
		}, 0, nil
	}
	privateNetwork := privateSpeedNetwork(network)
	registry := privatepst.ResolveLoadedServerRegistryWithNetwork(ctx, loaded, limit, 2*time.Second, nil, privateNetwork)
	selected, benchmarks := runPrivateSpeedBenchmarksFromRegistry(ctx, limit, registry, func(ctx context.Context, node privatepst.RegistryNode, latencyInfo *privatepst.ServerWithLatencyInfo) privatepst.SpeedTestResult {
		return privatepst.RunSpeedTestContextWithNetwork(ctx, node.Server, false, false, 4, 5*time.Second, latencyInfo, false, privateNetwork)
	})
	return registry, selected, benchmarks
}

func privateSpeedNetwork(network speedmodel.Network) privatepst.Network {
	switch network {
	case speedmodel.NetworkIPv4:
		return privatepst.NetworkIPv4
	case speedmodel.NetworkIPv6:
		return privatepst.NetworkIPv6
	default:
		return privatepst.NetworkAuto
	}
}

// newPrivateStructuredSpeedPreload resolves candidate nodes before the final
// throughput stage. It starts only when the common structured plan reaches
// its preload boundary, so option 1 never overlaps this work with hardware
// benchmarks. The returned runner reuses those selected candidates and only
// performs download/upload work when the speed section starts.
func newPrivateStructuredSpeedPreload(config *Config, inputs componentInputs) *structuredPrivateSpeedPreload {
	if config == nil {
		config = NewDefaultConfig()
	}
	configSnapshot := *config
	network := inputs.SpeedNetwork
	done := make(chan struct{})
	var once sync.Once
	var runner privateSpeedRunnerWithNetwork

	start := func(ctx context.Context) {
		once.Do(func() {
			if ctx == nil {
				ctx = context.Background()
			}
			go func() {
				defer close(done)
				runner = preloadStructuredPrivateSpeedRunner(ctx, &configSnapshot, inputs, network)
			}()
		})
	}
	return &structuredPrivateSpeedPreload{
		start: start,
		wait: func(ctx context.Context) privateSpeedRunnerWithNetwork {
			start(ctx)
			if ctx == nil {
				ctx = context.Background()
			}
			select {
			case <-done:
				return runner
			case <-ctx.Done():
				return nil
			}
		},
	}
}

func preloadStructuredPrivateSpeedRunner(ctx context.Context, config *Config, inputs componentInputs, network speedmodel.Network) privateSpeedRunnerWithNetwork {
	loaded, err := structuredPrivateSpeedRegistry(ctx, config, inputs, network)
	if err != nil {
		return unavailablePreloadedPrivateSpeedRunner(err)
	}
	limit := structuredPrivateSpeedLimit(config)
	if config != nil && !strings.EqualFold(strings.TrimSpace(config.Language), "en") && usesChineseFullStructuredSpeedProfile(config) {
		return preloadChineseFullPrivateSpeedRunner(ctx, loaded, limit, network)
	}
	if config != nil && strings.EqualFold(strings.TrimSpace(config.Language), "en") {
		loaded = filterInternationalPrivateRegistry(loaded)
	}
	registry := privatepst.ResolveLoadedServerRegistryWithConcurrencyAndNetwork(ctx, loaded, limit, 2*time.Second, 8, nil, privateSpeedNetwork(network))
	return preloadedPrivateSpeedRunner(registry, network)
}

func structuredPrivateSpeedRegistry(ctx context.Context, config *Config, inputs componentInputs, network speedmodel.Network) (privatepst.RegistryLoadResult, error) {
	if loaded, ok := inputs.PrivateSpeedData.(privatepst.RegistryLoadResult); ok {
		return loaded, nil
	}
	if config != nil && config.DataOffline {
		return privatepst.LoadEmbeddedServerList()
	}
	return privatepst.LoadServerListWithMetadataContextWithNetwork(ctx, privateSpeedNetwork(network))
}

func structuredPrivateSpeedLimit(config *Config) int {
	if config == nil || config.SpNum <= 0 {
		return 2
	}
	if usesChineseNearbyCarrierStructuredSpeedProfile(config) {
		return 1
	}
	return config.SpNum
}

func unavailablePreloadedPrivateSpeedRunner(err error) privateSpeedRunnerWithNetwork {
	return func(context.Context, int, speedmodel.Network) (any, int, []privateSpeedBenchmark) {
		return privatepst.RegistryReport{
			SchemaVersion: "privatespeedtest.registry/v1",
			Fallback:      true,
			Availability:  privatepst.ServerUnavailable,
			Servers:       []privatepst.RegistryNode{},
			Error:         err.Error(),
		}, 0, nil
	}
}

func preloadedPrivateSpeedRunner(registry privatepst.RegistryReport, network speedmodel.Network) privateSpeedRunnerWithNetwork {
	return func(ctx context.Context, limit int, requestedNetwork speedmodel.Network) (any, int, []privateSpeedBenchmark) {
		effectiveNetwork := network
		if requestedNetwork != speedmodel.NetworkAuto {
			effectiveNetwork = requestedNetwork
		}
		selected, benchmarks := runPrivateSpeedBenchmarksFromRegistry(ctx, limit, registry, privateSpeedTestForNetwork(effectiveNetwork))
		return registry, selected, benchmarks
	}
}

func privateSpeedTestForNetwork(network speedmodel.Network) privateSpeedTestFunc {
	privateNetwork := privateSpeedNetwork(network)
	return func(ctx context.Context, node privatepst.RegistryNode, latencyInfo *privatepst.ServerWithLatencyInfo) privatepst.SpeedTestResult {
		return privatepst.RunSpeedTestContextWithNetwork(ctx, node.Server, false, false, 4, 5*time.Second, latencyInfo, false, privateNetwork)
	}
}

func preloadChineseFullPrivateSpeedRunner(ctx context.Context, loaded privatepst.RegistryLoadResult, limit int, network speedmodel.Network) privateSpeedRunnerWithNetwork {
	if limit <= 0 {
		limit = 2
	}
	type carrierDefinition struct {
		label string
		isp   string
	}
	carriers := []carrierDefinition{
		{label: "unicom", isp: "Unicom"},
		{label: "telecom", isp: "Telecom"},
		{label: "mobile", isp: "Mobile"},
	}
	registries := make([]privatepst.RegistryReport, len(carriers))
	var wait sync.WaitGroup
	for index, carrier := range carriers {
		index, carrier := index, carrier
		wait.Add(1)
		go func() {
			defer wait.Done()
			carrierLoaded := filterPrivateRegistryByCarrier(loaded, carrier.isp)
			// Three bounded pools cap this full-suite preload at nine candidate
			// probes, leaving headroom for the concurrent non-throughput checks.
			registries[index] = privatepst.ResolveLoadedServerRegistryWithConcurrencyAndNetwork(ctx, carrierLoaded, limit, 2*time.Second, 3, nil, privateSpeedNetwork(network))
		}()
	}
	wait.Wait()

	return func(runnerCtx context.Context, runnerLimit int, requestedNetwork speedmodel.Network) (any, int, []privateSpeedBenchmark) {
		if runnerLimit <= 0 {
			runnerLimit = limit
		}
		effectiveNetwork := network
		if requestedNetwork != speedmodel.NetworkAuto {
			effectiveNetwork = requestedNetwork
		}
		report := chineseFullPrivateSpeedRegistryReport{
			SchemaVersion: "goecs.speed.private-carriers/v1",
			Source:        loaded.Source,
			Fallback:      loaded.Fallback,
			Metadata:      loaded.Metadata,
			Carriers:      make([]chineseFullPrivateSpeedCarrierSet, 0, len(carriers)),
		}
		benchmarks := make([]privateSpeedBenchmark, 0, len(carriers)*runnerLimit)
		selected := 0
		for index, carrier := range carriers {
			registry := registries[index]
			count, carrierBenchmarks := runPrivateSpeedBenchmarksFromRegistry(runnerCtx, runnerLimit, registry, privateSpeedTestForNetwork(effectiveNetwork))
			report.Carriers = append(report.Carriers, chineseFullPrivateSpeedCarrierSet{Carrier: carrier.label, Registry: registry})
			selected += count
			benchmarks = append(benchmarks, carrierBenchmarks...)
		}
		return report, selected, benchmarks
	}
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
