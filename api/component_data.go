package api

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	unlockexecutor "github.com/oneclickvirt/UnlockTests/executor"
	bgptools "github.com/oneclickvirt/backtrace/bgptools"
	nt3model "github.com/oneclickvirt/nt3/model"
	pingmodel "github.com/oneclickvirt/pingtest/model"
	speedmodel "github.com/oneclickvirt/speedtest/model"
)

const (
	tcpDataFile       = "tcp-targets.json"
	provinceDataFile  = "province-routes.json"
	speedDataFile     = "speedtest-servers.json"
	privateDataFile   = "private-speedtest-servers.json"
	transferDataFile  = "openspeedtest-servers.json"
	dnsblDataFile     = "dnsbl-zones.json"
	asnDataFile       = "bgp-asn-map.json"
	providerDataFile  = "media-providers.json"
	componentDataWait = 15 * time.Second
)

type componentDataResult struct {
	file    DataFileVersion
	primary *DataVersion
	apply   func(*componentInputs)
	err     error
}

// loadComponentData asks each component for its own validated registry. A
// failed registry is isolated to its owner so unrelated probes still run.
func loadComponentData(ctx context.Context, offline bool) (componentInputs, []DataFileVersion, *DataVersion, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	loaders := []func(context.Context, bool) componentDataResult{
		loadTCPComponentData,
		loadProvinceComponentData,
		loadSpeedComponentData,
		loadPrivateSpeedComponentData,
		loadTransferComponentData,
		loadSecurityComponentData,
		loadASNComponentData,
		loadProviderComponentData,
	}
	results := make(chan componentDataResult, len(loaders))
	for _, loader := range loaders {
		go func(loader func(context.Context, bool) componentDataResult) {
			loadCtx, cancel := context.WithTimeout(ctx, componentDataWait)
			defer cancel()
			results <- loader(loadCtx, offline)
		}(loader)
	}

	var inputs componentInputs
	files := make([]DataFileVersion, 0, len(loaders))
	var primary *DataVersion
	var loadErr error
	for range loaders {
		result := <-results
		files = append(files, result.file)
		if result.apply != nil {
			result.apply(&inputs)
		}
		if result.primary != nil {
			primary = result.primary
		}
		if result.err != nil {
			loadErr = errors.Join(loadErr, fmt.Errorf("load %s: %w", result.file.File, result.err))
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].File < files[j].File })
	return inputs, files, primary, loadErr
}

func loadTCPComponentData(ctx context.Context, offline bool) componentDataResult {
	sources := pingmodel.DefaultTCPTargetRegistrySources()
	if offline {
		sources = nil
	}
	loaded, err := pingmodel.LoadMergedTCPTargets(ctx, nil, sources, 1)
	if err != nil {
		return failedComponentData(ctx, tcpDataFile, err)
	}
	file := stringMetadataFile(tcpDataFile, loaded.Metadata.Schema, loaded.Metadata.GeneratedAt, loaded.Source, loaded.Fallback, loaded.Metadata.Count)
	targets := make([]TCPTarget, 0, len(loaded.Targets))
	for _, target := range loaded.Targets {
		targets = append(targets, TCPTarget{
			ID: targetID(target.Name, target.Host, target.Port), Name: target.Name,
			Host: target.Host, Port: target.Port, Category: target.Category,
		})
	}
	primary := &DataVersion{
		Schema: file.Schema, GeneratedAt: file.GeneratedAt, Source: file.Source,
		Fallback: file.Fallback, File: file.File, Count: file.Count,
	}
	return componentDataResult{file: file, primary: primary, apply: func(inputs *componentInputs) { inputs.TCPTargets = targets }}
}

func loadProvinceComponentData(ctx context.Context, offline bool) componentDataResult {
	sources := nt3model.DefaultProvinceRouteRegistrySources()
	if offline {
		sources = nil
	}
	loaded, err := nt3model.LoadProvinceRoutes(ctx, nil, sources)
	if err != nil {
		return failedComponentData(ctx, provinceDataFile, err)
	}
	file := stringMetadataFile(provinceDataFile, loaded.Metadata.Schema, loaded.Metadata.GeneratedAt, loaded.Source, loaded.Fallback, loaded.Metadata.Count)
	return componentDataResult{file: file, apply: func(inputs *componentInputs) { inputs.ProvinceRoutes = loaded.Routes }}
}

func loadSpeedComponentData(ctx context.Context, offline bool) componentDataResult {
	sources := speedmodel.DefaultRegistrySources()
	if offline {
		sources = nil
	}
	loaded, err := speedmodel.LoadServerRegistry(ctx, nil, sources, 10)
	if err != nil {
		return failedComponentData(ctx, speedDataFile, err)
	}
	file := stringMetadataFile(speedDataFile, loaded.Metadata.Schema, loaded.Metadata.GeneratedAt, loaded.Source, loaded.Fallback, loaded.Metadata.Count)
	return componentDataResult{file: file, apply: func(inputs *componentInputs) { inputs.SpeedtestServers = loaded.Servers }}
}

func loadASNComponentData(ctx context.Context, offline bool) componentDataResult {
	var entries []bgptools.ASNMetadata
	var source bgptools.ASNMetadataSource
	var err error
	if offline {
		entries, source, err = bgptools.EmbeddedASNMetadata()
	} else {
		entries, source, err = bgptools.LoadASNMetadata(ctx, nil)
	}
	if err != nil {
		return failedComponentData(ctx, asnDataFile, err)
	}
	file := timeMetadataFile(asnDataFile, source.Schema, source.GeneratedAt, source.Source, source.Fallback, source.Count)
	return componentDataResult{file: file, apply: func(inputs *componentInputs) { inputs.BGPASNMap = entries }}
}

func loadProviderComponentData(ctx context.Context, offline bool) componentDataResult {
	var providers []unlockexecutor.ProviderMetadata
	var source unlockexecutor.ProviderMetadataSource
	var err error
	if offline {
		providers, source, err = unlockexecutor.EmbeddedProviderMetadataSnapshot()
	} else {
		providers, source, err = unlockexecutor.LoadProviderMetadata(ctx, nil)
	}
	if err != nil {
		return failedComponentData(ctx, providerDataFile, err)
	}
	file := timeMetadataFile(providerDataFile, source.Schema, source.GeneratedAt, source.Source, source.Fallback, source.Count)
	return componentDataResult{file: file, apply: func(inputs *componentInputs) { inputs.MediaProviders = providers }}
}

func stringMetadataFile(file, schema, generatedAt, source string, fallback bool, count int) DataFileVersion {
	parsed, _ := time.Parse(time.RFC3339, generatedAt)
	return timeMetadataFile(file, schema, parsed, source, fallback, count)
}

func timeMetadataFile(file, schema string, generatedAt time.Time, source string, fallback bool, count int) DataFileVersion {
	fallbackSource := ""
	if fallback {
		fallbackSource = source
	}
	return DataFileVersion{
		File: file, Schema: schema, GeneratedAt: generatedAt, Source: source,
		Fallback: fallbackSource, Count: count, Status: ReportStatusOK,
	}
}

func failedComponentData(ctx context.Context, file string, err error) componentDataResult {
	return componentDataResult{
		file: DataFileVersion{File: file, Status: dataFileStatus(ctx, err), Reason: err.Error()},
		err:  err,
	}
}
