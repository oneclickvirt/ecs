package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	unlockexecutor "github.com/oneclickvirt/UnlockTests/executor"
	unlockmodel "github.com/oneclickvirt/UnlockTests/model"
	bgptools "github.com/oneclickvirt/backtrace/bgptools"
	backtracemodel "github.com/oneclickvirt/backtrace/model"
	basicmodel "github.com/oneclickvirt/basics/model"
	basicssystem "github.com/oneclickvirt/basics/system"
	"github.com/oneclickvirt/cputest/cpu"
	cputestmodel "github.com/oneclickvirt/cputest/model"
	"github.com/oneclickvirt/disktest/disk"
	gostunmodel "github.com/oneclickvirt/gostun/model"
	"github.com/oneclickvirt/gostun/stuncheck"
	"github.com/oneclickvirt/memorytest/memory"
	nt3model "github.com/oneclickvirt/nt3/model"
	nt3 "github.com/oneclickvirt/nt3/nt"
	pingmodel "github.com/oneclickvirt/pingtest/model"
	pingprobe "github.com/oneclickvirt/pingtest/pt"
	portemail "github.com/oneclickvirt/portchecker/email"
	speedmodel "github.com/oneclickvirt/speedtest/model"
)

// collectPublishedComponentReports calls the structured APIs from the
// published component modules. It deliberately does not call shell scripts;
// the standard CPU, memory, and disk probes share one bounded hardware-stage
// context.
func collectPublishedComponentReports(ctx context.Context, config *Config, inputs componentInputs) []ComponentReport {
	reports, _ := collectPublishedComponentReportsWithTCP(ctx, config, inputs)
	return reports
}

type structuredTaskResult struct {
	components []ComponentReport
	tcp        []TCPReport
	status     ReportStatus
	reason     string
}

type structuredComponentTask struct {
	section string
	run     func(context.Context) structuredTaskResult
}

type structuredCollectionPlan struct {
	basics     func(context.Context) []ComponentReport
	hardware   func(context.Context) []ComponentReport
	concurrent []structuredComponentTask
	speed      *structuredComponentTask
}

func sortStructuredNetworkTasks(tasks []structuredComponentTask) {
	canonicalOrder := map[string]int{
		"media": 0, "security": 1, "email": 2, "backtrace": 3, "routes": 4,
		"ping": 5, "tgdc": 6, "web": 7, "nat": 8, "tcp": 9,
	}
	sort.SliceStable(tasks, func(i, j int) bool {
		left, leftKnown := canonicalOrder[tasks[i].section]
		right, rightKnown := canonicalOrder[tasks[j].section]
		if leftKnown != rightKnown {
			return leftKnown
		}
		if !leftKnown {
			return false
		}
		return left < right
	})
}

func runStructuredCollectionPlan(ctx context.Context, plan structuredCollectionPlan) ([]ComponentReport, []TCPReport) {
	components := make([]ComponentReport, 0, 12)
	if plan.basics != nil {
		components = append(components, plan.basics(ctx)...)
	}
	if plan.hardware != nil {
		components = append(components, plan.hardware(ctx)...)
	}
	var tcp []TCPReport
	for _, result := range runStructuredConcurrentTasks(ctx, plan.concurrent) {
		components = append(components, result.components...)
		if result.tcp != nil {
			tcp = result.tcp
		}
	}
	if plan.speed != nil && ctx.Err() == nil {
		progressStarted(ctx, plan.speed.section)
		result := runStructuredTask(ctx, *plan.speed)
		status, reason := structuredTaskResultStatus(result)
		if contextStatus, done := contextProgressStatus(ctx); done {
			status, reason = contextStatus, ctx.Err().Error()
		}
		progressCompleted(ctx, plan.speed.section, status, reason)
		components = append(components, result.components...)
	}
	return components, tcp
}

func runStructuredConcurrentTasks(ctx context.Context, tasks []structuredComponentTask) []structuredTaskResult {
	if ctx.Err() != nil {
		results := make([]structuredTaskResult, 0, len(tasks))
		for _, task := range tasks {
			progressStarted(ctx, task.section)
			result := runStructuredTask(ctx, task)
			status, _ := contextProgressStatus(ctx)
			progressCompleted(ctx, task.section, status, ctx.Err().Error())
			results = append(results, result)
		}
		return results
	}
	channels := make([]<-chan structuredTaskResult, len(tasks))
	for index := range tasks {
		result := make(chan structuredTaskResult, 1)
		channels[index] = result
		task := tasks[index]
		go func() {
			value := structuredTaskResult{}
			defer func() {
				if recovered := recover(); recovered != nil {
					value = structuredTaskResult{status: ReportStatusError, reason: fmt.Sprintf("%s component panic", task.section)}
				}
				result <- value
				close(result)
			}()
			value = runStructuredTask(ctx, task)
		}()
	}
	results := make([]structuredTaskResult, 0, len(tasks))
	for index, channel := range channels {
		progressStarted(ctx, tasks[index].section)
		select {
		case result := <-channel:
			status, reason := structuredTaskResultStatus(result)
			if contextStatus, done := contextProgressStatus(ctx); done {
				status, reason = contextStatus, ctx.Err().Error()
			}
			progressCompleted(ctx, tasks[index].section, status, reason)
			results = append(results, result)
		case <-ctx.Done():
			status, _ := contextProgressStatus(ctx)
			progressCompleted(ctx, tasks[index].section, status, ctx.Err().Error())
			return results
		}
	}
	return results
}

func runStructuredTask(ctx context.Context, task structuredComponentTask) (result structuredTaskResult) {
	defer func() {
		if recover() != nil {
			result = structuredTaskResult{status: ReportStatusError, reason: fmt.Sprintf("%s component panic", task.section)}
		}
	}()
	if task.run == nil {
		return structuredTaskResult{status: ReportStatusUnavailable, reason: fmt.Sprintf("%s component runner unavailable", task.section)}
	}
	return task.run(ctx)
}

func structuredTaskResultStatus(result structuredTaskResult) (ReportStatus, string) {
	if result.status != "" {
		return result.status, result.reason
	}
	if len(result.components) > 0 {
		return aggregateComponentSectionStatus(result.components)
	}
	if result.tcp != nil {
		return tcpSectionStatus(result.tcp)
	}
	return ReportStatusOK, ""
}

func collectPublishedComponentReportsWithTCP(ctx context.Context, config *Config, inputs componentInputs) ([]ComponentReport, []TCPReport) {
	if ctx == nil {
		ctx = context.Background()
	}
	if config == nil {
		config = NewDefaultConfig()
	}
	plan := structuredCollectionPlan{
		basics: func(stageCtx context.Context) []ComponentReport {
			if !config.BasicStatus {
				return nil
			}
			return []ComponentReport{collectComponentStep(stageCtx, "basics", func() ComponentReport {
				started := time.Now()
				report := basicssystem.CollectSystemReport(stageCtx)
				return componentPayload("basics", report.SchemaVersion, componentStatus(string(report.Availability)), started, report, nil)
			})}
		},
		hardware: func(stageCtx context.Context) []ComponentReport {
			return collectHardwareComponentReports(stageCtx, config, defaultHardwareComponentRunners())
		},
	}
	if config.Nt3Status && inputs.Network && len(inputs.ProvinceRoutes) > 0 {
		plan.concurrent = append(plan.concurrent, structuredComponentTask{section: "routes", run: func(taskCtx context.Context) structuredTaskResult {
			result := make([]ComponentReport, 0, 2)
			started := time.Now()
			routes := inputs.ProvinceRoutes
			targets := nt3model.BuildProvinceLatencyTargets(routes, config.Nt3CheckType)
			probeConfig := nt3.StandardProvinceLatencyConfig()
			if config.DeepMode {
				probeConfig = nt3.DeepProvinceLatencyConfig()
			}
			probeCtx, cancel := componentContext(taskCtx, 45*time.Second)
			probes := nt3.RunProvinceLatency(probeCtx, targets, probeConfig)
			probeErr := probeCtx.Err()
			cancel()
			status := ReportStatusOK
			if probeErr != nil {
				if probeErr == context.DeadlineExceeded {
					status = ReportStatusTimeout
				} else {
					status = ReportStatusCanceled
				}
			}
			result = append(result, componentPayload("nt3.province_latency", "goecs.nt3/province-latency-v1", status, started, probes, nil))
			if config.DeepMode && taskCtx.Err() == nil {
				routeStarted := time.Now()
				routeCtx, routeCancel := context.WithTimeout(taskCtx, 3*time.Minute)
				routeConfig := nt3.DeepDetailedProvinceRouteConfig(nt3.NTraceProvinceTracer)
				routeConfig.IPVersion = config.Nt3CheckType
				routeConfig.Concurrency = 3
				routesReport, routeErr := nt3.RunDetailedProvinceRoutes(routeCtx, routes, routeConfig)
				routeStatus := detailedRouteComponentStatus(routeCtx, routesReport, routeErr)
				routeCancel()
				result = append(result, componentPayload("nt3.province_routes", "goecs.nt3/province-routes-v1", routeStatus, routeStarted, routesReport, routeErr))
			}
			return structuredTaskResult{components: result}
		}})
	}
	if config.PingTestStatus && inputs.Network && (len(inputs.ProvinceRoutes) > 0 || config.Language == "en" || config.PingScope == "international") {
		plan.concurrent = append(plan.concurrent, structuredComponentTask{section: "ping", run: func(taskCtx context.Context) structuredTaskResult {
			started := time.Now()
			targets := structuredPingTargets(config, inputs.ProvinceRoutes)
			pingCtx, cancel := componentContext(taskCtx, 30*time.Second)
			defer cancel()
			probes := pingprobe.RunICMPProbes(pingCtx, targets, pingprobe.ICMPProbeConfig{Count: 3, Timeout: 5 * time.Second, Concurrency: 8})
			sortICMPResults(probes, config.PingSortOrder)
			report := componentPayload("ping.icmp", "goecs.ping/icmp-v1", pingComponentStatus(pingCtx, probes), started, probes, nil)
			report.Reason = pingComponentReason(probes, report.Status)
			return structuredTaskResult{components: []ComponentReport{report}}
		}})
	}
	if config.TgdcTestStatus && inputs.Network {
		plan.concurrent = append(plan.concurrent, structuredComponentTask{section: "tgdc", run: func(taskCtx context.Context) structuredTaskResult {
			started := time.Now()
			probeCtx, cancel := componentContext(taskCtx, 30*time.Second)
			defer cancel()
			probes := pingprobe.RunTelegramICMPProbes(probeCtx, pingprobe.ICMPProbeConfig{Count: 3, Timeout: 5 * time.Second, Concurrency: 5})
			report := componentPayload("ping.telegram", "goecs.ping/telegram-v1", pingComponentStatus(probeCtx, probes), started, probes, nil)
			report.Reason = pingComponentReason(probes, report.Status)
			return structuredTaskResult{components: []ComponentReport{report}}
		}})
	}
	if config.WebTestStatus && inputs.Network {
		plan.concurrent = append(plan.concurrent, structuredComponentTask{section: "web", run: func(taskCtx context.Context) structuredTaskResult {
			started := time.Now()
			probeCtx, cancel := componentContext(taskCtx, 45*time.Second)
			defer cancel()
			probes := pingprobe.RunWebsiteTCPProbes(probeCtx, pingprobe.TCPProbeConfig{Attempts: 3, Timeout: 5 * time.Second, Concurrency: 16})
			report := componentPayload("ping.web_tcp", "goecs.ping/web-tcp-v1", pingTCPComponentStatus(probeCtx, probes), started, probes, nil)
			report.Reason = pingTCPComponentReason(probes, report.Status)
			return structuredTaskResult{components: []ComponentReport{report}}
		}})
	}
	if config.UtTestStatus && inputs.Network {
		plan.concurrent = append(plan.concurrent, structuredComponentTask{section: "media", run: func(taskCtx context.Context) structuredTaskResult {
			started := time.Now()
			mediaCtx, cancel := componentContext(taskCtx, 60*time.Second)
			defer cancel()
			report := withComponentDuration(collectMediaComponentWithMetadata(mediaCtx, config, inputs.MediaProviders), started)
			return structuredTaskResult{components: []ComponentReport{report}}
		}})
	}
	if config.SecurityTestStatus && inputs.Network {
		plan.concurrent = append(plan.concurrent, structuredComponentTask{section: "security", run: func(taskCtx context.Context) structuredTaskResult {
			started := time.Now()
			securityCtx, cancel := componentContext(taskCtx, 60*time.Second)
			defer cancel()
			report := withComponentDuration(collectSecurityComponent(securityCtx, inputs.PublicIPv4, inputs.PublicIPv6, inputs.DNSBLZones), started)
			return structuredTaskResult{components: []ComponentReport{report}}
		}})
	}
	if config.BacktraceStatus && inputs.Network {
		plan.concurrent = append(plan.concurrent, structuredComponentTask{section: "backtrace", run: func(taskCtx context.Context) structuredTaskResult {
			started := time.Now()
			backtraceCtx, cancel := componentContext(taskCtx, 45*time.Second)
			defer cancel()
			report := withComponentDuration(collectBacktraceComponentWithRegistry(backtraceCtx, inputs.PublicIPv4, inputs.PublicIPv6, inputs.BGPASNMap), started)
			return structuredTaskResult{components: []ComponentReport{report}}
		}})
	}
	if config.EmailTestStatus && inputs.Network {
		plan.concurrent = append(plan.concurrent, structuredComponentTask{section: "email", run: func(taskCtx context.Context) structuredTaskResult {
			started := time.Now()
			mailCtx, cancel := componentContext(taskCtx, 30*time.Second)
			defer cancel()
			report := withComponentDuration(collectMailComponent(mailCtx, portemail.DefaultPlatformSpecs(), nil, nil, nil), started)
			return structuredTaskResult{components: []ComponentReport{report}}
		}})
	}
	if inputs.Network {
		plan.concurrent = append(plan.concurrent, structuredComponentTask{section: "nat", run: func(taskCtx context.Context) structuredTaskResult {
			started := time.Now()
			stunCtx, cancel := componentContext(taskCtx, 15*time.Second)
			defer cancel()
			servers := gostunmodel.GetDefaultServers(gostunmodel.IPVersion)
			if !config.DeepMode && len(servers) > 1 {
				servers = servers[:1]
			}
			report := collectSTUNComponent(stunCtx, stuncheck.ProbeConfig{
				Servers: servers, IPVersion: gostunmodel.IPVersion,
				Timeout: 3 * time.Second, MaxConcurrent: 1,
			}, stuncheck.ProbeNAT)
			report = withComponentDuration(report, started)
			return structuredTaskResult{components: []ComponentReport{report}}
		}})
	}
	if config.TCPProbeStatus && inputs.Network && len(inputs.TCPTargets) > 0 {
		plan.concurrent = append(plan.concurrent, structuredComponentTask{section: "tcp", run: func(taskCtx context.Context) structuredTaskResult {
			reports := runTCPReports(taskCtx, inputs.TCPTargets, tcpProbeConfig{
				attempts: 3, timeout: 3 * time.Second, concurrency: 16,
				dial: (&net.Dialer{}).DialContext,
			})
			sortTCPReports(reports, config.TCPSortOrder)
			return structuredTaskResult{tcp: reports}
		}})
	}
	if config.SpeedTestStatus && inputs.Network {
		plan.speed = &structuredComponentTask{section: "speed", run: func(taskCtx context.Context) structuredTaskResult {
			started := time.Now()
			speedCtx, cancel := componentContext(taskCtx, 75*time.Second)
			defer cancel()
			privateRunner := privateSpeedRunnerForConfig(config.Language, config.DataOffline)
			report := withComponentDuration(collectSpeedComponentFromRegistryForLanguageWithDependencies(speedCtx, inputs.SpeedtestServers, inputs.TransferTargets, config.Language, config.SpNum, nil, nil, privateRunner), started)
			return structuredTaskResult{components: []ComponentReport{report}}
		}}
	}
	sortStructuredNetworkTasks(plan.concurrent)
	return runStructuredCollectionPlan(ctx, plan)
}

func structuredPingTargets(config *Config, routes []nt3model.ProvinceRoute) []pingprobe.ICMPTarget {
	if config == nil {
		config = NewDefaultConfig()
	}
	if config.Language == "en" || config.PingScope == "international" {
		return pingprobe.InternationalICMPTargets()
	}
	return representativeICMPTargets(nt3model.BuildProvinceLatencyTargets(routes, config.Nt3CheckType), config.DeepMode)
}

func sortICMPResults(results []pingprobe.ICMPResult, order string) {
	sort.SliceStable(results, func(i, j int) bool {
		leftName := strings.ToLower(strings.TrimSpace(results[i].Target.Name))
		rightName := strings.ToLower(strings.TrimSpace(results[j].Target.Name))
		if order == "name" {
			return leftName < rightName
		}
		leftAvailable := results[i].Received > 0
		rightAvailable := results[j].Received > 0
		if leftAvailable != rightAvailable {
			return leftAvailable
		}
		if results[i].Mean != results[j].Mean {
			return results[i].Mean < results[j].Mean
		}
		return leftName < rightName
	})
}

// structuredOwnsHardware reports whether this build has the structured
// component APIs needed to execute hardware tests exactly once.
func structuredOwnsHardware() bool { return true }

func structuredOwnsNetwork() bool { return true }

func configureStructuredLogging(enabled bool) {
	unlockmodel.EnableLoger = enabled
	backtracemodel.EnableLoger = enabled
	basicmodel.EnableLoger = enabled
	cputestmodel.EnableLoger = enabled
	disk.EnableLoger = enabled
	gostunmodel.EnableLoger = enabled
	memory.EnableLoger = enabled
	nt3model.EnableLoger = enabled
	pingmodel.EnableLoger = enabled
	speedmodel.EnableLoger = enabled
}

func structuredIdentity(ctx context.Context) (string, string) {
	if ctx == nil {
		ctx = context.Background()
	}
	type identityResult struct {
		version string
		address string
	}
	results := make(chan identityResult, 2)
	for _, probe := range []struct {
		version string
		network string
		url     string
	}{
		{version: "ipv4", network: "tcp4", url: "https://api4.ipify.org"},
		{version: "ipv6", network: "tcp6", url: "https://api6.ipify.org"},
	} {
		go func(probe struct{ version, network, url string }) {
			probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			transport := &http.Transport{DialContext: (&net.Dialer{}).DialContext}
			transport.DialContext = func(dialCtx context.Context, _, address string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(dialCtx, probe.network, address)
			}
			defer transport.CloseIdleConnections()
			request, err := http.NewRequestWithContext(probeCtx, http.MethodGet, probe.url, nil)
			if err != nil {
				results <- identityResult{version: probe.version}
				return
			}
			response, err := (&http.Client{Transport: transport}).Do(request)
			if err != nil {
				results <- identityResult{version: probe.version}
				return
			}
			defer response.Body.Close()
			buffer := make([]byte, 128)
			count, _ := response.Body.Read(buffer)
			address := strings.TrimSpace(string(buffer[:count]))
			parsed := net.ParseIP(address)
			if response.StatusCode != http.StatusOK || parsed == nil || (probe.version == "ipv4") != (parsed.To4() != nil) {
				address = ""
			}
			results <- identityResult{version: probe.version, address: address}
		}(probe)
	}
	var ipv4, ipv6 string
	for range 2 {
		select {
		case result := <-results:
			if result.version == "ipv4" {
				ipv4 = result.address
			} else {
				ipv6 = result.address
			}
		case <-ctx.Done():
			return ipv4, ipv6
		}
	}
	return ipv4, ipv6
}

func mergeComponentTCPTargets(targets []TCPTarget) []TCPTarget {
	merged := append([]TCPTarget(nil), targets...)
	seen := make(map[string]struct{}, len(merged))
	for _, target := range merged {
		key := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(target.Host), ".")) + fmt.Sprintf(":%d", target.Port)
		seen[key] = struct{}{}
	}
	for _, target := range pingmodel.AllTCPTargets() {
		key := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(target.Host), ".")) + fmt.Sprintf(":%d", target.Port)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, TCPTarget{
			ID: targetID(target.Name, target.Host, target.Port), Name: target.Name,
			Host: target.Host, Port: target.Port, Category: target.Category,
		})
	}
	return merged
}

func targetID(name, host string, port int) string {
	value := strings.ToLower(strings.TrimSpace(name))
	value = strings.NewReplacer(" ", "-", "/", "-", "_", "-").Replace(value)
	value = strings.Trim(value, "-")
	if value != "" {
		return value
	}
	return strings.ToLower(strings.TrimSpace(host)) + fmt.Sprintf("-%d", port)
}

func representativeICMPTargets(targets []nt3model.ProvinceLatencyTarget, deep bool) []pingprobe.ICMPTarget {
	result := make([]pingprobe.ICMPTarget, 0, len(targets))
	seen := make(map[string]struct{})
	for _, target := range targets {
		key := target.Carrier + ":" + target.IPVersion
		if !deep {
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
		}
		result = append(result, pingprobe.ICMPTarget{
			ID:   target.ProvinceCode + "-" + target.Carrier + "-" + target.IPVersion,
			Name: target.ProvinceName + " " + target.Carrier,
			Host: target.Host, IPVersion: target.IPVersion,
		})
	}
	return result
}

func pingComponentStatus(ctx context.Context, results []pingprobe.ICMPResult) ReportStatus {
	if status, done := contextComponentStatus(ctx); done {
		return status
	}
	if len(results) == 0 {
		return ReportStatusUnavailable
	}
	ok, partial := 0, 0
	for _, result := range results {
		if result.Status == "ok" {
			ok++
		} else if result.Status == "partial" {
			partial++
		}
	}
	if ok == len(results) {
		return ReportStatusOK
	}
	if ok+partial > 0 {
		return ReportStatusPartial
	}
	return ReportStatusUnavailable
}

func pingTCPComponentStatus(ctx context.Context, results []pingprobe.TCPResult) ReportStatus {
	if status, done := contextComponentStatus(ctx); done {
		return status
	}
	if len(results) == 0 {
		return ReportStatusUnavailable
	}
	successful, attempts := 0, 0
	for _, result := range results {
		successful += result.Successful
		attempts += result.Attempts
	}
	if attempts == 0 || successful == 0 {
		return ReportStatusUnavailable
	}
	if successful < attempts {
		return ReportStatusPartial
	}
	return ReportStatusOK
}

func pingComponentReason(results []pingprobe.ICMPResult, status ReportStatus) string {
	if status == ReportStatusOK {
		return ""
	}
	received, sent := 0, 0
	for _, result := range results {
		received += result.Received
		sent += result.Sent
	}
	if sent == 0 {
		return string(status)
	}
	return fmt.Sprintf("%d/%d ICMP replies received", received, sent)
}

func pingTCPComponentReason(results []pingprobe.TCPResult, status ReportStatus) string {
	if status == ReportStatusOK {
		return ""
	}
	successful, attempts := 0, 0
	for _, result := range results {
		successful += result.Successful
		attempts += result.Attempts
	}
	if attempts == 0 {
		return string(status)
	}
	return fmt.Sprintf("%d/%d TCP handshakes succeeded", successful, attempts)
}

type hardwareComponentRunners struct {
	CPU               func(context.Context, cpu.StructuredConfig) cpu.StructuredResult
	Memory            func(context.Context, memory.BenchmarkConfig) (memory.BenchmarkResult, error)
	Disk              func(context.Context, disk.MatrixConfig) disk.MatrixResult
	DeepDisk          func(context.Context, disk.MatrixConfig) disk.MatrixResult
	DiscoverDiskPaths func() (disk.TestPathInfo, error)
	DeepMultiDisk     func(context.Context, []string, disk.MatrixConfig) disk.MultiPathResult
	Burn              func(context.Context, cpu.BurnConfig) cpu.BurnResult
}

func defaultHardwareComponentRunners() hardwareComponentRunners {
	return hardwareComponentRunners{
		CPU: cpu.RunStructured, Memory: memory.RunBenchmark,
		Disk: disk.RunStandardFioMatrix, DeepDisk: disk.RunDeepFioMatrix,
		DiscoverDiskPaths: disk.DiscoverTestPaths, DeepMultiDisk: disk.RunDeepMultiPathMatrix,
		Burn: cpu.RunBurn,
	}
}

func withDeepHardwareRunnerDefaults(runners hardwareComponentRunners) hardwareComponentRunners {
	if runners.DiscoverDiskPaths == nil {
		runners.DiscoverDiskPaths = disk.DiscoverTestPaths
	}
	if runners.DeepMultiDisk == nil {
		runners.DeepMultiDisk = disk.RunDeepMultiPathMatrix
	}
	if runners.Burn == nil {
		runners.Burn = cpu.RunBurn
	}
	return runners
}

func collectHardwareComponentReports(parent context.Context, config *Config, runners hardwareComponentRunners) []ComponentReport {
	if config == nil || shouldSkipStructuredHardware(parent) || (!config.CpuTestStatus && !config.MemoryTestStatus && !config.DiskTestStatus && !hasExplicitDeepHardware(config)) {
		return nil
	}
	hardwareCtx, cancel := hardwareStageContext(parent, config.HardwareBudget)
	defer cancel()
	runners = withDeepHardwareRunnerDefaults(runners)
	reports := make([]ComponentReport, 0, 3)
	if config.CpuTestStatus {
		progressStarted(parent, "cpu")
		started := time.Now()
		if report, complete := canceledHardwareComponent(hardwareCtx, "cputest", "goecs.cpu/v1", started); complete {
			reports = append(reports, report)
		} else if runners.CPU == nil {
			reports = append(reports, componentPayload("cputest", "goecs.cpu/v1", ReportStatusUnavailable, started, nil, errors.New("CPU structured runner unavailable")))
		} else {
			cpuConfig := cpu.StructuredConfig{Threads: runtime.NumCPU(), Duration: 5 * time.Second, MaxPrime: 10000}
			if config.DeepMode {
				cpuConfig.Duration = 20 * time.Second
				cpuConfig.MaxPrime = 50000
				if config.DeepBurnDuration > 0 {
					cpuConfig.Duration = config.DeepBurnDuration
				}
			}
			benchmark := runners.CPU(hardwareCtx, cpuConfig)
			reports = append(reports, hardwareComponentPayload(hardwareCtx, "cputest", benchmark.SchemaVersion, benchmark.Status, started, benchmark, nil))
		}
		last := reports[len(reports)-1]
		progressCompleted(parent, "cpu", last.Status, last.Reason)
	}
	if config.MemoryTestStatus {
		progressStarted(parent, "memory")
		started := time.Now()
		if report, complete := canceledHardwareComponent(hardwareCtx, "memorytest", "goecs.memory/v1", started); complete {
			reports = append(reports, report)
		} else if runners.Memory == nil {
			reports = append(reports, componentPayload("memorytest", "goecs.memory/v1", ReportStatusUnavailable, started, nil, errors.New("memory structured runner unavailable")))
		} else {
			memoryConfig := memory.DefaultBenchmarkConfig()
			if config.DeepMode {
				memoryConfig.WorkingSetBytes = 256 << 20
				memoryConfig.Iterations = 8
			}
			benchmark, err := runners.Memory(hardwareCtx, memoryConfig)
			reports = append(reports, hardwareComponentPayload(hardwareCtx, "memorytest", benchmark.SchemaVersion, string(benchmark.Status), started, benchmark, err))
		}
		last := reports[len(reports)-1]
		progressCompleted(parent, "memory", last.Status, last.Reason)
	}
	if config.DiskTestStatus {
		progressStarted(parent, "disk")
		started := time.Now()
		if report, complete := canceledHardwareComponent(hardwareCtx, "disktest", "goecs.disk/v1", started); complete {
			reports = append(reports, report)
		} else if runners.Disk == nil {
			reports = append(reports, componentPayload("disktest", "goecs.disk/v1", ReportStatusUnavailable, started, nil, errors.New("disk structured runner unavailable")))
		} else {
			path := config.DiskTestPath
			if path == "" {
				path = os.TempDir()
			}
			if absolute, err := filepath.Abs(path); err == nil {
				path = absolute
			}
			diskBudget := 45 * time.Second
			matrixRuntime := time.Second
			sizeBytes := int64(16 << 20)
			diskRunner := runners.Disk
			if config.DeepMode {
				diskBudget = min(3*time.Minute, config.HardwareBudget)
				matrixRuntime = 2 * time.Second
				sizeBytes = 256 << 20
				if runners.DeepDisk != nil {
					diskRunner = runners.DeepDisk
				}
			}
			matrix := diskRunner(hardwareCtx, disk.MatrixConfig{Path: path, SizeBytes: sizeBytes, Runtime: matrixRuntime, MaxDuration: diskBudget})
			reports = append(reports, hardwareComponentPayload(hardwareCtx, "disktest", matrix.SchemaVersion, matrix.Status, started, matrix, nil))
		}
		last := reports[len(reports)-1]
		progressCompleted(parent, "disk", last.Status, last.Reason)
	}
	if config.DeepMode {
		progressStarted(parent, "deep_hardware")
		deepReports := collectExplicitDeepHardwareReportsWithRunners(hardwareCtx, config, runners)
		reports = append(reports, deepReports...)
		deepStatus, deepReason := aggregateComponentSectionStatus(deepReports)
		progressCompleted(parent, "deep_hardware", deepStatus, deepReason)
	}
	return reports
}

func hasExplicitDeepHardware(config *Config) bool {
	return config != nil && config.DeepMode && (config.DiskMultiCheck || strings.TrimSpace(config.DeepDiskPaths) != "" ||
		strings.TrimSpace(config.DeepSMARTDevices) != "" || config.DeepBurnDuration > 0 || strings.TrimSpace(config.DeepGPUDevice) != "")
}

func hardwareComponentStatus(ctx context.Context, raw string) ReportStatus {
	if status, done := contextComponentStatus(ctx); done {
		return status
	}
	return componentStatus(raw)
}

type smartSelfTestPayload struct {
	SchemaVersion string                        `json:"schema_version"`
	Results       []basicssystem.DeepToolResult `json:"results"`
}

func collectExplicitDeepHardwareReports(ctx context.Context, config *Config) []ComponentReport {
	return collectExplicitDeepHardwareReportsWithRunners(ctx, config, defaultHardwareComponentRunners())
}

func collectExplicitDeepHardwareReportsWithRunners(ctx context.Context, config *Config, runners hardwareComponentRunners) []ComponentReport {
	if config == nil || !config.DeepMode {
		return nil
	}
	runners = withDeepHardwareRunnerDefaults(runners)
	result := make([]ComponentReport, 0, 4)

	started := time.Now()
	paths := splitExplicitTargets(config.DeepDiskPaths)
	if len(paths) == 0 && config.DiskMultiCheck {
		if discovered, err := runners.DiscoverDiskPaths(); err == nil {
			paths = discovered.MountPoints
		}
	}
	if len(paths) == 0 {
		result = append(result, skippedDeepComponent("disktest.deep_multi", "goecs.disk/deep-multi-v1", started))
	} else {
		matrix := runners.DeepMultiDisk(ctx, paths, disk.MatrixConfig{SizeBytes: 256 << 20, Runtime: 2 * time.Second, MaxDuration: min(3*time.Minute, config.HardwareBudget)})
		result = append(result, hardwareComponentPayload(ctx, "disktest.deep_multi", matrix.SchemaVersion, matrix.Status, started, matrix, nil))
	}

	started = time.Now()
	devices := splitExplicitTargets(config.DeepSMARTDevices)
	if len(devices) == 0 {
		result = append(result, skippedDeepComponent("basics.smart_selftest", "goecs.smart/selftest-v1", started))
	} else {
		payload := smartSelfTestPayload{SchemaVersion: "goecs.smart/selftest-v1", Results: make([]basicssystem.DeepToolResult, 0, len(devices))}
		for _, device := range devices {
			if ctx.Err() != nil {
				break
			}
			payload.Results = append(payload.Results, basicssystem.RunSMARTSelfTest(ctx, device))
		}
		status := aggregateDeepToolStatus(ctx, payload.Results)
		result = append(result, hardwareComponentPayload(ctx, "basics.smart_selftest", payload.SchemaVersion, string(status), started, payload, nil))
	}

	started = time.Now()
	if !config.CpuTestStatus && config.DeepBurnDuration > 0 {
		burn := runners.Burn(ctx, cpu.BurnConfig{Threads: runtime.NumCPU(), Duration: config.DeepBurnDuration, MaxPrime: 50000})
		result = append(result, hardwareComponentPayload(ctx, "cputest.burn", burn.SchemaVersion, burn.Status, started, burn, nil))
	}

	started = time.Now()
	if strings.TrimSpace(config.DeepGPUDevice) == "" {
		result = append(result, skippedDeepComponent("basics.gpu_compute", "goecs.gpu/compute-v1", started))
	} else {
		compute := basicssystem.RunGPUCompute(ctx, config.DeepGPUDevice)
		result = append(result, hardwareComponentPayload(ctx, "basics.gpu_compute", compute.SchemaVersion, compute.Status, started, compute, nil))
	}
	return result
}

func skippedDeepComponent(name, schema string, started time.Time) ComponentReport {
	report := componentPayload(name, schema, ReportStatusSkipped, started, map[string]any{"requested": false}, nil)
	report.Reason = "explicit deep target not configured"
	return report
}

func splitExplicitTargets(value string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func aggregateDeepToolStatus(ctx context.Context, results []basicssystem.DeepToolResult) ReportStatus {
	if status, done := contextComponentStatus(ctx); done {
		return status
	}
	if len(results) == 0 {
		return ReportStatusSkipped
	}
	ok, skipped := 0, 0
	for _, result := range results {
		if result.Status == "ok" {
			ok++
		} else if result.Status == "skipped" {
			skipped++
		}
	}
	if ok == len(results) {
		return ReportStatusOK
	}
	if skipped == len(results) {
		return ReportStatusSkipped
	}
	if ok > 0 {
		return ReportStatusPartial
	}
	return ReportStatusUnavailable
}

func hardwareComponentPayload(ctx context.Context, name, schema, raw string, started time.Time, payload any, err error) ComponentReport {
	report := componentPayload(name, schema, hardwareComponentStatus(ctx, raw), started, payload, err)
	if report.Reason == "" && report.Status != ReportStatusOK && ctx != nil && ctx.Err() != nil {
		report.Reason = ctx.Err().Error()
	}
	return report
}

// hardwareStageContext is deliberately separate from componentContext. The
// latter is appropriate for independent network probes and caps each probe at
// one minute; CPU, memory, and disk are one standard stage and must consume a
// single caller-configured budget (two minutes by default).
func hardwareStageContext(parent context.Context, budget time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if budget <= 0 {
		budget = 2 * time.Minute
	}
	return context.WithTimeout(parent, budget)
}

// canceledHardwareComponent records components that were enabled but could
// not start because the shared stage budget or parent context had expired.
// The second return value tells the caller to skip the underlying benchmark.
func canceledHardwareComponent(ctx context.Context, name, schema string, started time.Time) (ComponentReport, bool) {
	if ctx == nil || ctx.Err() == nil {
		return ComponentReport{}, false
	}
	status, _ := contextComponentStatus(ctx)
	if status == ReportStatusOK {
		status = ReportStatusCanceled
	}
	report := componentPayload(name, schema, status, started, nil, nil)
	report.Reason = ctx.Err().Error()
	return report, true
}

func detailedRouteComponentStatus(ctx context.Context, results []nt3.DetailedProvinceRouteResult, err error) ReportStatus {
	if err != nil {
		return ReportStatusError
	}
	if ctx != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ReportStatusTimeout
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return ReportStatusCanceled
		}
	}
	if len(results) == 0 {
		return ReportStatusUnavailable
	}
	ok, degraded := 0, 0
	for _, result := range results {
		if result.Status == nt3.ProvinceRouteStatusOK {
			ok++
		} else {
			degraded++
		}
	}
	if ok == len(results) {
		return ReportStatusOK
	}
	if ok > 0 || degraded > 0 {
		return ReportStatusPartial
	}
	return ReportStatusUnavailable
}

type backtraceRunner func(context.Context, string, bgptools.IPBGPReportConfig) (*bgptools.IPBGPReport, error)

type backtraceConfigFactory func() bgptools.IPBGPReportConfig

func defaultBacktraceConfig() bgptools.IPBGPReportConfig {
	return bgptools.IPBGPReportConfig{
		Timeout: 20 * time.Second, FetchGeofeed: true,
		EnableWHOISFallback: true, WHOISTimeout: 3 * time.Second,
		ResolveASN: bgptools.ResolveOriginASN,
	}
}

func collectBacktraceComponent(ctx context.Context, ipv4, ipv6 string) ComponentReport {
	return collectBacktraceComponentWithData(ctx, ipv4, ipv6, nil)
}

func collectBacktraceComponentWithRunner(ctx context.Context, ipv4, ipv6 string, runner backtraceRunner) ComponentReport {
	return collectBacktraceComponentWithRunnerAndData(ctx, ipv4, ipv6, nil, runner)
}

func collectBacktraceComponentWithData(ctx context.Context, ipv4, ipv6 string, asnData []byte) ComponentReport {
	return collectBacktraceComponentWithRunnerAndData(ctx, ipv4, ipv6, asnData, bgptools.QueryIPBGPReport)
}

func collectBacktraceComponentWithRegistry(ctx context.Context, ipv4, ipv6 string, entries []bgptools.ASNMetadata) ComponentReport {
	registry := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.ASN != 0 && strings.TrimSpace(entry.Name) != "" {
			registry[strconv.FormatUint(uint64(entry.ASN), 10)] = strings.TrimSpace(entry.Name)
		}
	}
	return collectBacktraceComponentWithRegistryAndDependencies(ctx, ipv4, ipv6, registry, bgptools.QueryIPBGPReport, nil)
}

type asnMetadata struct {
	ASN  string `json:"asn"`
	Name string `json:"name"`
}

type asnRegistryUsage struct {
	SourceCount int           `json:"source_count"`
	Matched     []asnMetadata `json:"matched,omitempty"`
}

func collectBacktraceComponentWithRunnerAndData(ctx context.Context, ipv4, ipv6 string, asnData []byte, runner backtraceRunner) ComponentReport {
	return collectBacktraceComponentWithDependencies(ctx, ipv4, ipv6, asnData, runner, nil)
}

func collectBacktraceComponentWithDependencies(ctx context.Context, ipv4, ipv6 string, asnData []byte, runner backtraceRunner, configFactory backtraceConfigFactory) ComponentReport {
	asnRegistry, registryErr := parseASNRegistry(asnData)
	if registryErr != nil {
		return componentPayload("backtrace.ip_bgp", "goecs.backtrace/v1", ReportStatusError, time.Now(), nil, registryErr)
	}
	return collectBacktraceComponentWithRegistryAndDependencies(ctx, ipv4, ipv6, asnRegistry, runner, configFactory)
}

func collectBacktraceComponentWithRegistryAndDependencies(ctx context.Context, ipv4, ipv6 string, asnRegistry map[string]string, runner backtraceRunner, configFactory backtraceConfigFactory) ComponentReport {
	started := time.Now()
	if ctx == nil {
		ctx = context.Background()
	}
	if runner == nil {
		runner = bgptools.QueryIPBGPReport
	}
	if configFactory == nil {
		configFactory = defaultBacktraceConfig
	}
	addresses := make([]string, 0, 2)
	for _, ip := range []string{strings.TrimSpace(ipv4), strings.TrimSpace(ipv6)} {
		if net.ParseIP(ip) != nil {
			addresses = append(addresses, ip)
		}
	}
	if len(addresses) == 0 {
		payload := map[string]any{"schema_version": "goecs.backtrace/v1", "reports": []any{}}
		report := componentPayload("backtrace.ip_bgp", "goecs.backtrace/v1", ReportStatusUnavailable, started, payload, nil)
		report.Reason = "no valid public IP address"
		return report
	}
	reports := make([]*bgptools.IPBGPReport, len(addresses))
	errorsByIndex := make([]error, len(addresses))
	var wg sync.WaitGroup
	for index, ip := range addresses {
		wg.Add(1)
		go func(index int, ip string) {
			defer wg.Done()
			reports[index], errorsByIndex[index] = runner(ctx, ip, configFactory())
		}(index, ip)
	}
	wg.Wait()
	payload := struct {
		SchemaVersion string                  `json:"schema_version"`
		Reports       []*bgptools.IPBGPReport `json:"reports"`
		ASNRegistry   asnRegistryUsage        `json:"asn_registry"`
	}{SchemaVersion: "goecs.backtrace/v1", Reports: reports, ASNRegistry: buildASNRegistryUsage(asnRegistry, reports)}
	valid, degraded := 0, 0
	for index, report := range reports {
		if report == nil {
			degraded++
			if errorsByIndex[index] != nil {
				if errors.Is(errorsByIndex[index], context.DeadlineExceeded) {
					degraded++
				}
			}
			continue
		}
		switch report.Status {
		case bgptools.ReportAvailable:
			valid++
		default:
			degraded++
		}
	}
	status := ReportStatusUnavailable
	if valid > 0 && degraded == 0 {
		status = ReportStatusOK
	} else if valid > 0 {
		status = ReportStatusPartial
	} else if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			status = ReportStatusTimeout
		} else {
			status = ReportStatusCanceled
		}
	} else if degraded > 0 {
		status = ReportStatusPartial
	}
	result := componentPayload("backtrace.ip_bgp", payload.SchemaVersion, status, started, payload, nil)
	if result.Status != ReportStatusOK {
		result.Reason = fmt.Sprintf("%d/%d IP/BGP reports available", valid, len(reports))
	}
	return result
}

func parseASNRegistry(data []byte) (map[string]string, error) {
	registry := make(map[string]string)
	if len(data) == 0 {
		return registry, nil
	}
	var entries []struct {
		ASN  json.Number `json:"asn"`
		Name string      `json:"name"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err := decoder.Decode(&entries); err != nil {
		return nil, fmt.Errorf("decode BGP ASN metadata: %w", err)
	}
	for _, entry := range entries {
		asn := strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(entry.ASN.String())), "AS")
		if asn == "" || strings.TrimSpace(entry.Name) == "" {
			continue
		}
		if number, err := strconv.ParseUint(asn, 10, 32); err == nil && number > 0 {
			registry[strconv.FormatUint(number, 10)] = strings.TrimSpace(entry.Name)
		}
	}
	return registry, nil
}

func buildASNRegistryUsage(registry map[string]string, reports []*bgptools.IPBGPReport) asnRegistryUsage {
	usage := asnRegistryUsage{SourceCount: len(registry)}
	seen := make(map[string]struct{})
	add := func(asn string) {
		asn = strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(asn)), "AS")
		name, exists := registry[asn]
		if !exists {
			return
		}
		if _, exists := seen[asn]; exists {
			return
		}
		seen[asn] = struct{}{}
		usage.Matched = append(usage.Matched, asnMetadata{ASN: asn, Name: name})
	}
	for _, report := range reports {
		if report == nil {
			continue
		}
		add(report.ASN)
		if report.Relationships == nil {
			continue
		}
		for _, entries := range [][]bgptools.ASNRelationship{report.Relationships.Upstreams, report.Relationships.Peers, report.Relationships.IXPs} {
			for _, entry := range entries {
				add(entry.ASN)
			}
		}
	}
	sort.Slice(usage.Matched, func(i, j int) bool { return usage.Matched[i].ASN < usage.Matched[j].ASN })
	return usage
}

type mediaComponentPayload struct {
	SchemaVersion string                            `json:"schema_version"`
	Selection     string                            `json:"selection"`
	IPVersion     string                            `json:"ip_version"`
	Results       []unlockexecutor.StructuredResult `json:"results"`
	Registry      mediaRegistryUsage                `json:"registry"`
	Error         string                            `json:"error,omitempty"`
}

type mediaProviderMetadata struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Groups       []string `json:"groups"`
	SupportsIPv6 bool     `json:"supports_ipv6"`
	Category     string   `json:"category,omitempty"`
	Aliases      []string `json:"aliases,omitempty"`
}

type mediaRegistryUsage struct {
	SourceCount int                     `json:"source_count"`
	Matched     []mediaProviderMetadata `json:"matched,omitempty"`
	Missing     []string                `json:"missing,omitempty"`
}

func collectMediaComponent(ctx context.Context, config *Config) ComponentReport {
	return collectMediaComponentWithRegistry(ctx, config, nil)
}

func collectMediaComponentWithRegistry(ctx context.Context, config *Config, metadata []byte) ComponentReport {
	if config == nil {
		config = NewDefaultConfig()
	}
	return collectMediaComponentWithRegistryRunner(ctx, unlockexecutor.RunOptions{
		Selection: config.UnlockTestRegion, IPVersion: config.UnlockTestIPVersion,
		Interface: config.UnlockTestInterface, DNSServers: config.UnlockTestDNSServers,
		HTTPProxy: config.UnlockTestHTTPProxy, SOCKSProxy: config.UnlockTestSOCKSProxy,
		Concurrency: config.UnlockTestConcurrency, IncludeHeads: false,
	}, metadata, unlockexecutor.RunStructured)
}

func collectMediaComponentWithMetadata(ctx context.Context, config *Config, metadata []unlockexecutor.ProviderMetadata) ComponentReport {
	if config == nil {
		config = NewDefaultConfig()
	}
	registry := make(map[string]mediaProviderMetadata, len(metadata))
	for _, entry := range metadata {
		converted := mediaProviderMetadata{Name: strings.TrimSpace(entry.Name), Category: strings.TrimSpace(entry.Category), Aliases: append([]string(nil), entry.Aliases...)}
		if converted.Name == "" {
			continue
		}
		registry[strings.ToLower(converted.Name)] = converted
		for _, alias := range converted.Aliases {
			alias = strings.ToLower(strings.TrimSpace(alias))
			if alias != "" {
				registry[alias] = converted
			}
		}
	}
	return collectMediaComponentWithParsedRegistryRunner(ctx, unlockexecutor.RunOptions{
		Selection: config.UnlockTestRegion, IPVersion: config.UnlockTestIPVersion,
		Interface: config.UnlockTestInterface, DNSServers: config.UnlockTestDNSServers,
		HTTPProxy: config.UnlockTestHTTPProxy, SOCKSProxy: config.UnlockTestSOCKSProxy,
		Concurrency: config.UnlockTestConcurrency, IncludeHeads: false,
	}, registry, unlockexecutor.RunStructured)
}

type mediaRunner func(context.Context, unlockexecutor.RunOptions) ([]unlockexecutor.StructuredResult, error)

func collectMediaComponentWithRunner(ctx context.Context, options unlockexecutor.RunOptions, runner mediaRunner) ComponentReport {
	return collectMediaComponentWithRegistryRunner(ctx, options, nil, runner)
}

func collectMediaComponentWithRegistryRunner(ctx context.Context, options unlockexecutor.RunOptions, metadata []byte, runner mediaRunner) ComponentReport {
	registry, registryErr := parseMediaRegistry(metadata)
	if registryErr != nil {
		return componentPayload("unlocktests.media", "goecs.unlocktests/media-v1", ReportStatusError, time.Now(), nil, registryErr)
	}
	return collectMediaComponentWithParsedRegistryRunner(ctx, options, registry, runner)
}

func collectMediaComponentWithParsedRegistryRunner(ctx context.Context, options unlockexecutor.RunOptions, registry map[string]mediaProviderMetadata, runner mediaRunner) ComponentReport {
	started := time.Now()
	if ctx == nil {
		ctx = context.Background()
	}
	if status, canceled := contextComponentStatus(ctx); canceled {
		return componentPayload("unlocktests.media", "goecs.unlocktests/media-v1", status, started, mediaComponentPayload{
			SchemaVersion: "goecs.unlocktests/media-v1", Selection: options.Selection, IPVersion: options.IPVersion,
			Registry: mediaRegistryUsage{SourceCount: len(registry)},
		}, nil)
	}
	if runner == nil {
		runner = unlockexecutor.RunStructured
	}
	results, err := runner(ctx, options)
	payload := mediaComponentPayload{
		SchemaVersion: "goecs.unlocktests/media-v1", Selection: options.Selection,
		IPVersion: options.IPVersion, Results: results,
		Registry: buildMediaRegistryUsage(registry, results),
	}
	if err != nil {
		payload.Error = err.Error()
	}
	status := mediaComponentStatus(ctx, results, err)
	report := componentPayload("unlocktests.media", payload.SchemaVersion, status, started, payload, nil)
	if report.Status != ReportStatusOK {
		report.Reason = mediaComponentReason(results, err)
	}
	return report
}

func parseMediaRegistry(data []byte) (map[string]mediaProviderMetadata, error) {
	registry := make(map[string]mediaProviderMetadata)
	if len(data) == 0 {
		return registry, nil
	}
	var entries []mediaProviderMetadata
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("decode media provider metadata: %w", err)
	}
	for _, entry := range entries {
		entry.ID = strings.TrimSpace(entry.ID)
		entry.Name = strings.TrimSpace(entry.Name)
		if entry.ID == "" || entry.Name == "" {
			continue
		}
		registry[strings.ToLower(entry.Name)] = entry
	}
	return registry, nil
}

func buildMediaRegistryUsage(registry map[string]mediaProviderMetadata, results []unlockexecutor.StructuredResult) mediaRegistryUsage {
	usage := mediaRegistryUsage{SourceCount: len(registry)}
	matched := make(map[string]mediaProviderMetadata)
	missing := make(map[string]struct{})
	for _, result := range results {
		name := strings.TrimSpace(result.Name)
		if name == "" {
			continue
		}
		if entry, exists := registry[strings.ToLower(name)]; exists {
			key := strings.TrimSpace(entry.ID)
			if key == "" {
				key = strings.ToLower(strings.TrimSpace(entry.Name))
			}
			matched[key] = entry
		} else {
			missing[name] = struct{}{}
		}
	}
	for _, entry := range matched {
		usage.Matched = append(usage.Matched, entry)
	}
	for name := range missing {
		usage.Missing = append(usage.Missing, name)
	}
	sort.Slice(usage.Matched, func(i, j int) bool {
		if usage.Matched[i].ID != usage.Matched[j].ID {
			return usage.Matched[i].ID < usage.Matched[j].ID
		}
		return usage.Matched[i].Name < usage.Matched[j].Name
	})
	sort.Strings(usage.Missing)
	return usage
}

func mediaComponentStatus(ctx context.Context, results []unlockexecutor.StructuredResult, err error) ReportStatus {
	if ctx != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ReportStatusTimeout
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return ReportStatusCanceled
		}
	}
	if len(results) == 0 {
		if err != nil {
			return ReportStatusError
		}
		return ReportStatusUnavailable
	}
	failed, limited, available := 0, 0, 0
	for _, result := range results {
		switch result.Status {
		case unlockmodel.StatusRateLimited:
			limited++
		case unlockmodel.StatusYes, unlockmodel.StatusNo, unlockmodel.StatusRestricted, unlockmodel.StatusBanned, unlockmodel.StatusCDNRelay:
			available++
		default:
			failed++
		}
	}
	if failed == 0 && limited == 0 {
		return ReportStatusOK
	}
	if available == 0 && limited == len(results) {
		return ReportStatusPartial
	}
	if available > 0 {
		return ReportStatusPartial
	}
	return ReportStatusError
}

func mediaComponentReason(results []unlockexecutor.StructuredResult, err error) string {
	if err != nil {
		return sanitizePublicText(err.Error())
	}
	failed, limited := 0, 0
	for _, result := range results {
		if result.Status == unlockmodel.StatusRateLimited {
			limited++
		} else if result.Status != unlockmodel.StatusYes && result.Status != unlockmodel.StatusNo && result.Status != unlockmodel.StatusRestricted && result.Status != unlockmodel.StatusBanned && result.Status != unlockmodel.StatusCDNRelay {
			failed++
		}
	}
	return fmt.Sprintf("%d media results failed, %d rate limited", failed, limited)
}

type speedNodeInput struct {
	ID       string `json:"id"`
	Host     string `json:"host"`
	URL      string `json:"url"`
	PortFrom int    `json:"port_from"`
	Provider string `json:"provider"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Status   string `json:"status"`
}

type speedNodeResult struct {
	ID           string `json:"id"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Provider     string `json:"provider,omitempty"`
	Country      string `json:"country,omitempty"`
	City         string `json:"city,omitempty"`
	Source       string `json:"source"`
	Availability string `json:"availability"`
	LatencyMS    int64  `json:"latency_ms,omitempty"`
	Error        string `json:"error,omitempty"`
	URL          string `json:"-"`
}

type speedComponentPayload struct {
	SchemaVersion     string                        `json:"schema_version"`
	Selected          []speedNodeResult             `json:"selected,omitempty"`
	Nodes             []speedNodeResult             `json:"nodes"`
	Benchmarks        []speedmodel.ThroughputResult `json:"benchmarks,omitempty"`
	PrivateRegistry   any                           `json:"private_registry,omitempty"`
	PrivateBenchmarks []privateSpeedBenchmark       `json:"private_benchmarks,omitempty"`
}

func collectSpeedComponent(ctx context.Context, speedtestData, openData []byte, limit int) ComponentReport {
	return collectSpeedComponentWithAllDependencies(ctx, speedtestData, openData, limit, nil, nil, runPrivateSpeedBenchmarks)
}

func collectSpeedComponentWithOffline(ctx context.Context, speedtestData, openData []byte, limit int, offline bool) ComponentReport {
	privateRunner := runPrivateSpeedBenchmarks
	if offline {
		privateRunner = runEmbeddedPrivateSpeedBenchmarks
	}
	return collectSpeedComponentWithAllDependencies(ctx, speedtestData, openData, limit, nil, nil, privateRunner)
}

type speedDialFunc func(context.Context, string, string) (net.Conn, error)

func collectSpeedComponentWithDial(ctx context.Context, speedtestData, openData []byte, limit int, dial speedDialFunc) ComponentReport {
	return collectSpeedComponentWithDependencies(ctx, speedtestData, openData, limit, dial, nil)
}

func collectSpeedComponentWithDependencies(ctx context.Context, speedtestData, openData []byte, limit int, dial speedDialFunc, throughput speedmodel.ThroughputProbe) ComponentReport {
	return collectSpeedComponentWithAllDependencies(ctx, speedtestData, openData, limit, dial, throughput, nil)
}

type privateSpeedBenchmark struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Source       string  `json:"source"`
	Status       string  `json:"status"`
	LatencyMS    float64 `json:"latency_ms,omitempty"`
	DownloadMbps float64 `json:"download_mbps,omitempty"`
	UploadMbps   float64 `json:"upload_mbps,omitempty"`
	DurationMS   int64   `json:"duration_ms"`
	Error        string  `json:"error,omitempty"`
}

type privateSpeedRunner func(context.Context, int) (any, int, []privateSpeedBenchmark)

func privateSpeedRunnerForConfig(language string, offline bool) privateSpeedRunner {
	if language == "en" {
		if offline {
			return runEmbeddedInternationalPrivateSpeedBenchmarks
		}
		return runInternationalPrivateSpeedBenchmarks
	}
	if offline {
		return runEmbeddedPrivateSpeedBenchmarks
	}
	return runPrivateSpeedBenchmarks
}

func isMainlandChinaCountry(country string) bool {
	normalized := strings.ToLower(strings.TrimSpace(country))
	normalized = strings.Join(strings.Fields(strings.NewReplacer(".", "", "_", " ", "-", " ").Replace(normalized)), " ")
	switch normalized {
	case "cn", "china", "mainland china", "china mainland", "prc", "people's republic of china", "peoples republic of china", "people s republic of china", "中国", "中国大陆", "中华人民共和国":
		return true
	default:
		return strings.Contains(normalized, "mainland china")
	}
}

func internationalSpeedServers(servers []speedmodel.ServerMetadata) []speedmodel.ServerMetadata {
	return speedmodel.FilterServersForLanguage(servers, "en")
}

func internationalTransferTargets(targets []transferTargetInput) []transferTargetInput {
	result := make([]transferTargetInput, 0, len(targets))
	for _, target := range targets {
		if strings.TrimSpace(target.Country) != "" && !isMainlandChinaCountry(target.Country) {
			result = append(result, target)
		}
	}
	return result
}

func collectSpeedComponentFromRegistry(ctx context.Context, servers []speedmodel.ServerMetadata, transfers []transferTargetInput, limit int, offline bool) ComponentReport {
	privateRunner := runPrivateSpeedBenchmarks
	if offline {
		privateRunner = runEmbeddedPrivateSpeedBenchmarks
	}
	return collectSpeedComponentFromRegistryWithDependencies(ctx, servers, transfers, limit, nil, nil, privateRunner)
}

func collectSpeedComponentFromRegistryWithDependencies(ctx context.Context, servers []speedmodel.ServerMetadata, transfers []transferTargetInput, limit int, dial speedDialFunc, throughput speedmodel.ThroughputProbe, privateRunner privateSpeedRunner) ComponentReport {
	return collectSpeedComponentFromRegistryWithSelection(ctx, servers, transfers, limit, dial, throughput, privateRunner, false)
}

func collectSpeedComponentFromRegistryForLanguageWithDependencies(ctx context.Context, servers []speedmodel.ServerMetadata, transfers []transferTargetInput, language string, limit int, dial speedDialFunc, throughput speedmodel.ThroughputProbe, privateRunner privateSpeedRunner) ComponentReport {
	international := strings.EqualFold(strings.TrimSpace(language), "en")
	if international {
		servers = internationalSpeedServers(servers)
		transfers = internationalTransferTargets(transfers)
	}
	return collectSpeedComponentFromRegistryWithSelection(ctx, servers, transfers, limit, dial, throughput, privateRunner, international)
}

func collectSpeedComponentFromRegistryWithSelection(ctx context.Context, servers []speedmodel.ServerMetadata, transfers []transferTargetInput, limit int, dial speedDialFunc, throughput speedmodel.ThroughputProbe, privateRunner privateSpeedRunner, representative bool) ComponentReport {
	started := time.Now()
	payload := speedComponentPayload{SchemaVersion: "goecs.speed/v1", Nodes: make([]speedNodeResult, 0, len(servers))}
	for _, server := range servers {
		host := strings.TrimSpace(server.Host)
		port := 0
		if _, parsedPort, err := net.SplitHostPort(host); err == nil {
			fmt.Sscanf(parsedPort, "%d", &port)
		}
		availability := string(server.Availability)
		if availability == "" {
			availability = "candidate"
		}
		payload.Nodes = append(payload.Nodes, speedNodeResult{
			ID: server.ID, Host: host, Port: port, Provider: server.Provider,
			Country: server.Country, City: server.City, Source: "speedtest",
			Availability: availability, LatencyMS: server.LatencyMS,
			Error: server.Error, URL: server.URL,
		})
	}
	for _, target := range transfers {
		availability := strings.ToLower(strings.TrimSpace(target.Status))
		if availability == "" || availability == "available" {
			availability = "candidate"
		}
		host := net.JoinHostPort(strings.Trim(target.Host, "[]"), strconv.Itoa(target.PortFrom))
		payload.Nodes = append(payload.Nodes, speedNodeResult{
			ID: target.ID, Host: host, Port: target.PortFrom, Provider: target.Provider,
			Country: target.Country, City: target.City, Source: "openspeedtest",
			Availability: availability,
		})
	}
	return collectSpeedComponentFromNodes(ctx, payload, limit, dial, throughput, privateRunner, representative, started)
}

func collectSpeedComponentWithAllDependencies(ctx context.Context, speedtestData, openData []byte, limit int, dial speedDialFunc, throughput speedmodel.ThroughputProbe, privateRunner privateSpeedRunner) ComponentReport {
	started := time.Now()
	payload := speedComponentPayload{SchemaVersion: "goecs.speed/v1"}
	for _, sourceData := range []struct {
		source string
		data   []byte
	}{
		{source: "speedtest", data: speedtestData},
		{source: "openspeedtest", data: openData},
	} {
		data, source := sourceData.data, sourceData.source
		if len(data) == 0 {
			continue
		}
		var inputs []speedNodeInput
		if err := json.Unmarshal(data, &inputs); err != nil {
			return componentPayload("speed.registry", payload.SchemaVersion, ReportStatusError, started, payload, fmt.Errorf("decode %s nodes: %w", source, err))
		}
		for _, input := range inputs {
			if strings.TrimSpace(input.Host) == "" || strings.EqualFold(input.Status, "unavailable") {
				payload.Nodes = append(payload.Nodes, speedNodeResult{ID: input.ID, Host: input.Host, Port: input.PortFrom, Provider: input.Provider, Country: input.Country, City: input.City, Source: source, Availability: "unavailable", Error: "static node unavailable", URL: input.URL})
				continue
			}
			port := input.PortFrom
			host := strings.TrimSpace(input.Host)
			if _, _, err := net.SplitHostPort(host); err != nil {
				if port <= 0 {
					port = 80
				}
				host = net.JoinHostPort(strings.Trim(host, "[]"), fmt.Sprintf("%d", port))
			} else if port <= 0 {
				if _, parsedPort, err := net.SplitHostPort(host); err == nil {
					fmt.Sscanf(parsedPort, "%d", &port)
				}
			}
			payload.Nodes = append(payload.Nodes, speedNodeResult{ID: input.ID, Host: host, Port: port, Provider: input.Provider, Country: input.Country, City: input.City, Source: source, Availability: "candidate", URL: input.URL})
		}
	}
	return collectSpeedComponentFromNodes(ctx, payload, limit, dial, throughput, privateRunner, false, started)
}

func collectSpeedComponentFromNodes(ctx context.Context, payload speedComponentPayload, limit int, dial speedDialFunc, throughput speedmodel.ThroughputProbe, privateRunner privateSpeedRunner, representative bool, started time.Time) ComponentReport {
	probeSpeedNodes(ctx, payload.Nodes, dial)
	available := make([]speedNodeResult, 0, len(payload.Nodes))
	for _, node := range payload.Nodes {
		if node.Availability == "available" && node.Source == "speedtest" && strings.TrimSpace(node.URL) != "" {
			available = append(available, node)
		}
	}
	sort.SliceStable(available, func(i, j int) bool {
		if available[i].LatencyMS == available[j].LatencyMS {
			return available[i].ID < available[j].ID
		}
		return available[i].LatencyMS < available[j].LatencyMS
	})
	if limit <= 0 {
		limit = 2
	}
	if representative {
		available = selectRepresentativeSpeedNodes(available, limit)
	} else if len(available) > limit {
		available = available[:limit]
	}
	payload.Selected = append(payload.Selected, available...)
	benchmarkServers := make([]speedmodel.ServerMetadata, 0, len(payload.Selected))
	for _, selected := range payload.Selected {
		benchmarkServers = append(benchmarkServers, speedmodel.ServerMetadata{
			ID: selected.ID, Name: selected.City, Host: selected.Host, URL: selected.URL,
			Provider: selected.Provider, Country: selected.Country, City: selected.City,
		})
	}
	payload.Benchmarks = speedmodel.BenchmarkServers(ctx, benchmarkServers, len(benchmarkServers), throughput)
	completed := 0
	for _, benchmark := range payload.Benchmarks {
		if benchmark.Status == speedmodel.ThroughputAvailable {
			completed++
		}
	}
	privateSelected := 0
	if privateRunner != nil && (ctx == nil || ctx.Err() == nil) {
		registry, selected, benchmarks := privateRunner(ctx, limit)
		payload.PrivateRegistry = registry
		payload.PrivateBenchmarks = benchmarks
		privateSelected = selected
		for _, benchmark := range benchmarks {
			if benchmark.Status == "available" {
				completed++
			}
		}
	}
	totalSelected := len(payload.Selected) + privateSelected
	status := ReportStatusOK
	if totalSelected == 0 || completed == 0 {
		status = ReportStatusUnavailable
	} else if completed != len(payload.Benchmarks)+len(payload.PrivateBenchmarks) {
		status = ReportStatusPartial
	}
	if ctx != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		status = ReportStatusTimeout
	}
	if ctx != nil && errors.Is(ctx.Err(), context.Canceled) {
		status = ReportStatusCanceled
	}
	report := componentPayload("speed.registry", payload.SchemaVersion, status, started, payload, nil)
	if report.Status != ReportStatusOK {
		report.Reason = fmt.Sprintf("%d/%d selected nodes completed throughput", completed, totalSelected)
	}
	return report
}

func selectRepresentativeSpeedNodes(nodes []speedNodeResult, limit int) []speedNodeResult {
	servers := make([]speedmodel.ServerMetadata, 0, len(nodes))
	for _, node := range nodes {
		servers = append(servers, speedmodel.ServerMetadata{
			ID: node.ID, Name: node.City, Host: node.Host, URL: node.URL,
			Provider: node.Provider, Country: node.Country, City: node.City,
			Availability: speedmodel.ServerAvailable, LatencyMS: node.LatencyMS,
		})
	}
	selected, err := speedmodel.SelectRepresentativeServers(servers, limit)
	if err != nil {
		return nil
	}
	result := make([]speedNodeResult, 0, len(selected))
	for _, server := range selected {
		for _, node := range nodes {
			if node.ID == server.ID && node.Host == server.Host && node.URL == server.URL {
				result = append(result, node)
				break
			}
		}
	}
	return result
}

func speedContextStatus(err error) string {
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	return "unavailable"
}

func probeSpeedNodes(ctx context.Context, nodes []speedNodeResult, dial speedDialFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if dial == nil {
		dial = (&net.Dialer{}).DialContext
	}
	if len(nodes) == 0 {
		return
	}
	workers := min(8, len(nodes))
	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for index := range jobs {
				node := &nodes[index]
				if node.Availability != "candidate" {
					continue
				}
				started := time.Now()
				probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				conn, err := dial(probeCtx, "tcp", node.Host)
				node.LatencyMS = time.Since(started).Milliseconds()
				cancel()
				if err != nil {
					node.Availability, node.Error = "unavailable", classifySpeedProbeError(err)
					continue
				}
				node.Availability = "available"
				if conn != nil {
					_ = conn.Close()
				}
			}
		}()
	}
	for index := range nodes {
		select {
		case jobs <- index:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		}
	}
	close(jobs)
	wg.Wait()
}

func classifySpeedProbeError(err error) string {
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	return "connection_error"
}

// withComponentDuration preserves the component adapter's status and payload
// while adding the elapsed time measured by the ecs orchestration layer.
func withComponentDuration(report ComponentReport, started time.Time) ComponentReport {
	report.DurationMS = time.Since(started).Milliseconds()
	return report
}

type stunProbeFunc func(context.Context, stuncheck.ProbeConfig) stuncheck.NATSummary

// collectMailComponent adapts portchecker's dependency-injectable structured
// API. Keeping the dependencies as arguments makes cancellation and offline
// fixture tests deterministic without changing the production path.
func collectMailComponent(ctx context.Context, specs []portemail.PlatformSpec, resolver portemail.MXResolver, dialer portemail.Dialer, listener portemail.LocalListener) ComponentReport {
	started := time.Now()
	if ctx == nil {
		ctx = context.Background()
	}
	if status, canceled := contextComponentStatus(ctx); canceled {
		empty := portemail.MailReport{GeneratedAt: time.Now().UTC()}
		report := componentPayload("portchecker.email", "goecs.portchecker/mail-v1", status, started, empty, nil)
		report.Reason = ctx.Err().Error()
		return report
	}
	report := portemail.CheckMail(ctx, specs, resolver, dialer, listener)
	status := mailComponentStatus(ctx, report)
	result := componentPayload("portchecker.email", "goecs.portchecker/mail-v1", status, started, report, nil)
	if result.Status != ReportStatusOK {
		result.Reason = mailComponentReason(report)
	}
	return result
}

func contextComponentStatus(ctx context.Context) (ReportStatus, bool) {
	if ctx == nil {
		return ReportStatusOK, false
	}
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		return ReportStatusTimeout, true
	case errors.Is(ctx.Err(), context.Canceled):
		return ReportStatusCanceled, true
	default:
		return ReportStatusOK, false
	}
}

func mailComponentReason(report portemail.MailReport) string {
	total, available := 0, 0
	for _, group := range [][]portemail.EndpointResult{report.Local, report.OutboundSMTP25, report.MX, report.Fixed} {
		for _, endpoint := range group {
			total++
			if endpoint.Status == portemail.MailAvailable {
				available++
			}
		}
	}
	if total == 0 {
		return "no mail endpoints were tested"
	}
	return fmt.Sprintf("%d/%d mail capabilities available", available, total)
}

func mailComponentStatus(ctx context.Context, report portemail.MailReport) ReportStatus {
	if ctx != nil {
		switch {
		case errors.Is(ctx.Err(), context.DeadlineExceeded):
			return ReportStatusTimeout
		case errors.Is(ctx.Err(), context.Canceled):
			return ReportStatusCanceled
		}
	}
	total, available, timedOut := 0, 0, 0
	for _, group := range [][]portemail.EndpointResult{report.Local, report.OutboundSMTP25, report.MX, report.Fixed} {
		for _, endpoint := range group {
			total++
			switch endpoint.Status {
			case portemail.MailAvailable:
				available++
			case portemail.MailTimeout:
				timedOut++
			}
		}
	}
	if total == 0 {
		return ReportStatusUnavailable
	}
	if available == total {
		return ReportStatusOK
	}
	if timedOut == total {
		return ReportStatusTimeout
	}
	return ReportStatusPartial
}

// collectSTUNComponent adapts gostun's context-aware binding/hairpin probe to
// the common ecs component envelope. A probe function is accepted so local
// tests can use a deterministic fixture instead of a public STUN server.
func collectSTUNComponent(ctx context.Context, config stuncheck.ProbeConfig, probe stunProbeFunc) ComponentReport {
	started := time.Now()
	if ctx == nil {
		ctx = context.Background()
	}
	if status, canceled := contextComponentStatus(ctx); canceled {
		report := stuncheck.NATSummary{SchemaVersion: "goecs.stun/v1", IPVersion: config.IPVersion, Error: ctx.Err().Error()}
		result := componentPayload("gostun.nat", report.SchemaVersion, status, started, report, nil)
		result.Reason = ctx.Err().Error()
		return result
	}
	if probe == nil {
		probe = stuncheck.ProbeNAT
	}
	if len(config.Servers) == 0 {
		return componentPayload("gostun.nat", "goecs.stun/v1", ReportStatusUnavailable, started, nil, fmt.Errorf("no STUN server configured"))
	}
	report := probe(ctx, config)
	status := stunComponentStatus(ctx, report)
	result := componentPayload("gostun.nat", report.SchemaVersion, status, started, report, nil)
	if result.Status != ReportStatusOK && result.Reason == "" {
		if report.Error != "" {
			result.Reason = report.Error
		} else {
			result.Reason = fmt.Sprintf("%d/%d STUN probes succeeded", report.Successful, len(report.Results))
		}
	}
	return result
}

func stunComponentStatus(ctx context.Context, report stuncheck.NATSummary) ReportStatus {
	if ctx != nil {
		switch {
		case errors.Is(ctx.Err(), context.DeadlineExceeded):
			return ReportStatusTimeout
		case errors.Is(ctx.Err(), context.Canceled):
			return ReportStatusCanceled
		}
	}
	if report.Partial {
		return ReportStatusPartial
	}
	switch report.Status {
	case stuncheck.CapabilityAvailable:
		return ReportStatusOK
	case stuncheck.CapabilityTimeout:
		return ReportStatusTimeout
	case stuncheck.CapabilityUnavailable, stuncheck.CapabilityUnsupported:
		return ReportStatusUnavailable
	default:
		return ReportStatusError
	}
}

func componentContext(parent context.Context, budget time.Duration) (context.Context, context.CancelFunc) {
	if budget <= 0 {
		budget = 30 * time.Second
	}
	if budget > 60*time.Second {
		budget = 60 * time.Second
	}
	return context.WithTimeout(parent, budget)
}
