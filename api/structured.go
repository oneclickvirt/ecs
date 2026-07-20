package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	datarepo "github.com/oneclickvirt/ecs/internal/data"
	"github.com/oneclickvirt/ecs/utils"
)

const StructuredReportSchema = "goecs.report/v1"

type ReportStatus string

const (
	ReportStatusOK          ReportStatus = "ok"
	ReportStatusPartial     ReportStatus = "partial"
	ReportStatusUnavailable ReportStatus = "unavailable"
	ReportStatusTimeout     ReportStatus = "timeout"
	ReportStatusCanceled    ReportStatus = "canceled"
	ReportStatusError       ReportStatus = "error"
	ReportStatusSkipped     ReportStatus = "skipped"
)

type DataVersion struct {
	Schema      string    `json:"schema"`
	GeneratedAt time.Time `json:"generated_at"`
	Source      string    `json:"source"`
	Fallback    string    `json:"fallback,omitempty"`
	File        string    `json:"file"`
	Count       int       `json:"count"`
}

// DataFileVersion describes one validated ecs-data payload. DataVersion is
// intentionally kept unchanged for callers that historically consumed the
// primary TCP target file; StructuredReport.DataFiles carries the complete
// manifest view for newer callers.
type DataFileVersion struct {
	File        string       `json:"file"`
	Schema      string       `json:"schema"`
	GeneratedAt time.Time    `json:"generated_at"`
	Source      string       `json:"source"`
	Fallback    string       `json:"fallback"`
	Count       int          `json:"count"`
	Status      ReportStatus `json:"status"`
	Reason      string       `json:"reason,omitempty"`
}

type SectionReport struct {
	Name    string       `json:"name"`
	Enabled bool         `json:"enabled"`
	Status  ReportStatus `json:"status"`
	Reason  string       `json:"reason,omitempty"`
}

type TCPTarget struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Category string `json:"category,omitempty"`
}

type TCPSample struct {
	DurationMS float64 `json:"duration_ms,omitempty"`
	Status     string  `json:"status"`
}

type TCPReport struct {
	Target             TCPTarget      `json:"target"`
	Attempts           int            `json:"attempts"`
	Successful         int            `json:"successful"`
	SuccessRatePercent float64        `json:"success_rate_percent"`
	LossPercent        float64        `json:"loss_percent"`
	MinMS              float64        `json:"min_ms,omitempty"`
	MaxMS              float64        `json:"max_ms,omitempty"`
	MeanMS             float64        `json:"mean_ms,omitempty"`
	P50MS              float64        `json:"p50_ms,omitempty"`
	P95MS              float64        `json:"p95_ms,omitempty"`
	Samples            []TCPSample    `json:"samples"`
	Errors             map[string]int `json:"errors,omitempty"`
}

type StructuredReport struct {
	SchemaVersion string            `json:"schema_version"`
	ECSVersion    string            `json:"ecs_version"`
	Status        ReportStatus      `json:"status"`
	StartedAt     time.Time         `json:"started_at"`
	FinishedAt    time.Time         `json:"finished_at"`
	DurationMS    int64             `json:"duration_ms"`
	DeepMode      bool              `json:"deep_mode"`
	PrivacyMode   bool              `json:"privacy_mode"`
	Data          *DataVersion      `json:"data,omitempty"`
	DataFiles     []DataFileVersion `json:"data_files,omitempty"`
	Sections      []SectionReport   `json:"sections"`
	Components    []ComponentReport `json:"components,omitempty"`
	TCP           []TCPReport       `json:"tcp,omitempty"`
	Text          string            `json:"text"`
}

// ComponentReport is the cross-repository envelope used by component
// releases and by ecs-gui fixtures. Payload remains versioned by its owner.
type ComponentReport struct {
	Name          string          `json:"name"`
	SchemaVersion string          `json:"schema_version"`
	Status        ReportStatus    `json:"status"`
	Reason        string          `json:"reason,omitempty"`
	DurationMS    int64           `json:"duration_ms,omitempty"`
	Payload       json.RawMessage `json:"payload,omitempty"`
}

type UnifiedReport struct {
	*StructuredReport
	// Components is retained as a Go-level convenience field for callers that
	// used the pre-release envelope. JSON serialization is owned by
	// StructuredReport to avoid duplicate `components` keys.
	Components []ComponentReport `json:"-"`
}

func (report *StructuredReport) WithComponents(components ...ComponentReport) UnifiedReport {
	copyReport := *report
	copyReport.Components = append(append([]ComponentReport(nil), report.Components...), components...)
	return UnifiedReport{StructuredReport: &copyReport, Components: append([]ComponentReport(nil), components...)}
}

func (report *StructuredReport) JSON() ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

type tcpProbeConfig struct {
	attempts    int
	timeout     time.Duration
	concurrency int
	dial        func(context.Context, string, string) (net.Conn, error)
}

type structuredExtras struct {
	data       *DataVersion
	dataFiles  []DataFileVersion
	tcp        []TCPReport
	components []ComponentReport
	err        error
}

// CollectStructuredReport builds the JSON-facing report for callers that
// already executed the text workflow, such as the CLI and desktop GUI.
func CollectStructuredReport(ctx context.Context, preCheck utils.NetCheckResult, config *Config, text string, startedAt, finishedAt time.Time) *StructuredReport {
	if ctx == nil {
		ctx = context.Background()
	}
	if config == nil {
		config = NewDefaultConfig()
	}
	// This API is used after the legacy text workflow by the CLI and GUI. The
	// legacy path has already executed destructive/expensive hardware tests, so
	// do not run a second benchmark just to populate a JSON envelope. Missing
	// component payloads remain partial instead of being represented as fake
	// successes.
	extras := collectStructuredExtras(skipStructuredHardware(ctx), preCheck, config)
	status, reason := ReportStatusOK, ""
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		status, reason = ReportStatusTimeout, ctx.Err().Error()
	} else if errors.Is(ctx.Err(), context.Canceled) {
		status, reason = ReportStatusCanceled, ctx.Err().Error()
	} else if extras.err != nil {
		status, reason = ReportStatusPartial, extras.err.Error()
	}
	if config.PrivacyMode {
		text = ""
	}
	sections := sectionReports(config, preCheck, extras, status, reason)
	status = aggregateReportStatus(status, sections)
	report := &StructuredReport{
		SchemaVersion: StructuredReportSchema, ECSVersion: config.EcsVersion,
		Status: status, StartedAt: startedAt, FinishedAt: finishedAt,
		DurationMS: finishedAt.Sub(startedAt).Milliseconds(), DeepMode: config.DeepMode,
		PrivacyMode: config.PrivacyMode, Data: extras.data, DataFiles: extras.dataFiles,
		Components: extras.components, TCP: extras.tcp,
		Sections: sections, Text: text,
	}
	if config.PrivacyMode {
		applyStructuredPrivacy(report)
	}
	return report
}

func collectStructuredExtras(ctx context.Context, preCheck utils.NetCheckResult, config *Config) structuredExtras {
	loader := datarepo.NewLoader(nil, config.DataCDNBase)
	if config.DataOffline {
		loader.CDNBase = ""
		loader.RawBase = ""
	}
	loadedFiles, dataFiles, loadErr := loadKnownDataFiles(ctx, loader)
	extras := structuredExtras{dataFiles: dataFiles, err: loadErr}
	loaded, ok := loadedFiles["tcp-targets.json"]
	if !ok {
		return extras
	}
	meta, ok := loaded.Manifest.Files["tcp-targets.json"]
	if !ok {
		return extras
	}
	extras.data = &DataVersion{
		Schema: loaded.Manifest.Schema, GeneratedAt: loaded.Manifest.GeneratedAt,
		Source: loaded.Source, Fallback: loaded.Fallback, File: loaded.Name, Count: meta.Count,
	}
	var targets []TCPTarget
	if err := json.Unmarshal(loaded.Data, &targets); err != nil {
		extras.err = errors.Join(extras.err, fmt.Errorf("decode TCP targets: %w", err))
		return extras
	}
	targets = mergeComponentTCPTargets(targets)
	publicIPv4, publicIPv6 := GetIPv4Address(), GetIPv6Address()
	if preCheck.Connected && publicIPv4 == "" && publicIPv6 == "" {
		publicIPv4, publicIPv6 = structuredIdentity(ctx)
	}
	inputs := componentInputs{
		TCPTargets: targets, Network: preCheck.Connected && ctx.Err() == nil,
		PublicIPv4: publicIPv4, PublicIPv6: publicIPv6,
	}
	if province, exists := loadedFiles["province-routes.json"]; exists {
		inputs.ProvinceRoutes = province.Data
	}
	if speedtest, exists := loadedFiles["speedtest-servers.json"]; exists {
		inputs.SpeedtestServers = speedtest.Data
	}
	if openspeedtest, exists := loadedFiles["openspeedtest-servers.json"]; exists {
		inputs.OpenSpeedtestServer = openspeedtest.Data
	}
	if dnsbl, exists := loadedFiles["dnsbl-zones.json"]; exists {
		inputs.DNSBLZones = dnsbl.Data
	}
	if media, exists := loadedFiles["media-providers.json"]; exists {
		inputs.MediaProviders = media.Data
	}
	if bgp, exists := loadedFiles["bgp-asn-map.json"]; exists {
		inputs.BGPASNMap = bgp.Data
	}
	extras.components = collectComponentReports(ctx, config, inputs)
	if !config.TCPProbeStatus || !preCheck.Connected || ctx.Err() != nil {
		return extras
	}
	progressStarted(ctx, "tcp")
	extras.tcp = runTCPReports(ctx, targets, tcpProbeConfig{
		attempts: 3, timeout: 3 * time.Second, concurrency: 16,
		dial: (&net.Dialer{}).DialContext,
	})
	tcpStatus, tcpReason := tcpSectionStatus(extras.tcp)
	if status, done := contextProgressStatus(ctx); done {
		tcpStatus, tcpReason = status, ctx.Err().Error()
	}
	progressCompleted(ctx, "tcp", tcpStatus, tcpReason)
	return extras
}

func contextProgressStatus(ctx context.Context) (ReportStatus, bool) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ReportStatusTimeout, true
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return ReportStatusCanceled, true
	}
	return "", false
}

// loadKnownDataFiles validates every payload understood by this build. Each
// file is verified against one manifest candidate, preventing a report from
// mixing payload generations when a CDN is only partially updated.
func loadKnownDataFiles(ctx context.Context, loader *datarepo.Loader) (map[string]datarepo.Result, []DataFileVersion, error) {
	names := datarepo.KnownFiles()
	loaded, loadErr := loader.LoadMany(ctx, names)
	versions := make([]DataFileVersion, len(names))
	if loadErr != nil {
		for index, name := range names {
			versions[index] = DataFileVersion{
				File: name, Status: dataFileStatus(ctx, loadErr),
				Reason: loadErr.Error(),
			}
		}
		return nil, versions, loadErr
	}
	for index, name := range names {
		result, ok := loaded[name]
		if !ok {
			loadErr = errors.Join(loadErr, fmt.Errorf("load %s: result missing", name))
			versions[index] = DataFileVersion{File: name, Status: ReportStatusError, Reason: "result missing"}
			continue
		}
		meta, ok := result.Manifest.Files[name]
		if !ok {
			err := fmt.Errorf("manifest metadata missing")
			loadErr = errors.Join(loadErr, fmt.Errorf("load %s: %w", name, err))
			versions[index] = DataFileVersion{File: name, Status: ReportStatusError, Reason: err.Error()}
			continue
		}
		versions[index] = DataFileVersion{
			File: name, Schema: result.Manifest.Schema,
			GeneratedAt: result.Manifest.GeneratedAt, Source: result.Source,
			Fallback: result.Fallback, Count: meta.Count, Status: ReportStatusOK,
		}
	}
	return loaded, versions, loadErr
}

func dataFileStatus(ctx context.Context, err error) ReportStatus {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return ReportStatusTimeout
	}
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
		return ReportStatusCanceled
	}
	return ReportStatusError
}

func runTCPReports(ctx context.Context, targets []TCPTarget, config tcpProbeConfig) []TCPReport {
	if len(targets) == 0 {
		return nil
	}
	if config.attempts <= 0 {
		config.attempts = 1
	}
	if config.timeout <= 0 {
		config.timeout = 3 * time.Second
	}
	if config.concurrency <= 0 {
		config.concurrency = 1
	}
	if config.dial == nil {
		config.dial = (&net.Dialer{}).DialContext
	}
	results := make([]TCPReport, len(targets))
	jobs := make(chan int)
	workerCount := min(config.concurrency, len(targets))
	var wg sync.WaitGroup
	wg.Add(workerCount)
	for range workerCount {
		go func() {
			defer wg.Done()
			for index := range jobs {
				results[index] = runOneTCPReport(ctx, targets[index], config)
			}
		}()
	}
	for index := range targets {
		select {
		case jobs <- index:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return results
		}
	}
	close(jobs)
	wg.Wait()
	return results
}

func runOneTCPReport(ctx context.Context, target TCPTarget, config tcpProbeConfig) TCPReport {
	report := TCPReport{
		Target: target, Attempts: config.attempts,
		Samples: make([]TCPSample, 0, config.attempts), Errors: make(map[string]int),
	}
	address := net.JoinHostPort(target.Host, fmt.Sprintf("%d", target.Port))
	latencies := make([]float64, 0, config.attempts)
	for range config.attempts {
		attemptCtx, cancel := context.WithTimeout(ctx, config.timeout)
		started := time.Now()
		conn, err := config.dial(attemptCtx, "tcp", address)
		elapsed := float64(time.Since(started).Microseconds()) / 1000
		cancel()
		if err != nil {
			status := classifyTCPError(err)
			report.Errors[status]++
			report.Samples = append(report.Samples, TCPSample{Status: status})
			continue
		}
		if conn != nil {
			_ = conn.Close()
		}
		report.Successful++
		latencies = append(latencies, elapsed)
		report.Samples = append(report.Samples, TCPSample{DurationMS: elapsed, Status: "ok"})
	}
	report.SuccessRatePercent = float64(report.Successful) * 100 / float64(report.Attempts)
	report.LossPercent = float64(report.Attempts-report.Successful) * 100 / float64(report.Attempts)
	if len(latencies) == 0 {
		return report
	}
	sort.Float64s(latencies)
	report.MinMS, report.MaxMS = latencies[0], latencies[len(latencies)-1]
	for _, value := range latencies {
		report.MeanMS += value
	}
	report.MeanMS /= float64(len(latencies))
	report.P50MS = percentileFloat(latencies, 0.50)
	report.P95MS = percentileFloat(latencies, 0.95)
	return report
}

func percentileFloat(values []float64, quantile float64) float64 {
	position := quantile * float64(len(values)-1)
	lower, upper := int(math.Floor(position)), int(math.Ceil(position))
	if lower == upper {
		return values[lower]
	}
	return values[lower] + (values[upper]-values[lower])*(position-float64(lower))
}

func classifyTCPError(err error) string {
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "dns"
	}
	if errors.Is(err, syscall.ECONNREFUSED) || strings.Contains(strings.ToLower(err.Error()), "connection refused") {
		return "refused"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	return "network"
}

func sectionReports(config *Config, preCheck utils.NetCheckResult, extras structuredExtras, status ReportStatus, reason string) []SectionReport {
	sections := []struct {
		name    string
		enabled bool
		network bool
	}{
		{"basics", config.BasicStatus, false}, {"cpu", config.CpuTestStatus, false},
		{"memory", config.MemoryTestStatus, false}, {"disk", config.DiskTestStatus, false},
		{"media", config.UtTestStatus, true}, {"security", config.SecurityTestStatus, true},
		{"email", config.EmailTestStatus, true}, {"backtrace", config.BacktraceStatus, true},
		{"routes", config.Nt3Status, true}, {"ping", config.PingTestStatus, true},
		{"tgdc", config.TgdcTestStatus, true}, {"web", config.WebTestStatus, true},
		{"tcp", config.TCPProbeStatus, true}, {"speed", config.SpeedTestStatus, true},
		{"nat", true, true},
	}
	result := make([]SectionReport, 0, len(sections))
	componentsBySection := make(map[string][]ComponentReport, len(extras.components))
	componentSections := map[string]string{
		"basics": "basics", "cputest": "cpu", "memorytest": "memory",
		"disktest": "disk", "disktest.deep_multi": "disk", "basics.smart_selftest": "disk",
		"cputest.burn": "cpu", "basics.gpu_compute": "basics",
		"nt3.province_latency": "routes", "nt3.province_routes": "routes",
		"unlocktests.media": "media", "security.evidence": "security",
		"backtrace.ip_bgp": "backtrace", "portchecker.email": "email", "speed.registry": "speed",
		"gostun.nat": "nat", "ping.icmp": "ping",
		"ping.telegram": "tgdc", "ping.web_tcp": "web",
	}
	// A section without a structured component must not inherit the overall
	// report's ok status. This is especially important for release builds while
	// a new component module is still waiting to be published.
	structuredSections := map[string]bool{
		"basics": true, "cpu": true, "memory": true, "disk": true,
		"media": true, "security": true, "email": true, "backtrace": true,
		"routes": true, "ping": true, "tgdc": true, "web": true,
		"tcp": true, "speed": true, "nat": true,
	}
	for _, component := range extras.components {
		if sectionName := componentSections[component.Name]; sectionName != "" {
			componentsBySection[sectionName] = append(componentsBySection[sectionName], component)
		}
	}
	for _, section := range sections {
		sectionStatus, sectionReason := status, reason
		if !section.enabled {
			sectionStatus, sectionReason = ReportStatusSkipped, "disabled"
		} else if section.network && !preCheck.Connected {
			sectionStatus, sectionReason = ReportStatusUnavailable, "network unavailable"
		} else if components := componentsBySection[section.name]; len(components) > 0 {
			sectionStatus, sectionReason = aggregateComponentSectionStatus(components)
		} else if section.name == "tcp" && extras.err != nil {
			sectionStatus, sectionReason = ReportStatusError, extras.err.Error()
		} else if section.name == "tcp" && len(extras.tcp) == 0 {
			sectionStatus, sectionReason = ReportStatusUnavailable, "no TCP results"
		} else if section.name == "tcp" {
			sectionStatus, sectionReason = tcpSectionStatus(extras.tcp)
		} else if structuredSections[section.name] {
			sectionStatus, sectionReason = ReportStatusPartial, "structured component unavailable"
		}
		result = append(result, SectionReport{Name: section.name, Enabled: section.enabled, Status: sectionStatus, Reason: sectionReason})
	}
	return result
}

func aggregateComponentSectionStatus(components []ComponentReport) (ReportStatus, string) {
	if len(components) == 0 {
		return ReportStatusUnavailable, "structured component unavailable"
	}
	if len(components) == 1 {
		reason := strings.TrimSpace(components[0].Reason)
		if reason == "" && components[0].Status != ReportStatusOK && components[0].Status != ReportStatusSkipped {
			reason = string(components[0].Status)
		}
		return components[0].Status, reason
	}
	priority := map[ReportStatus]int{
		ReportStatusOK: 0, ReportStatusSkipped: 1, ReportStatusUnavailable: 2,
		ReportStatusPartial: 3, ReportStatusError: 4, ReportStatusCanceled: 5,
		ReportStatusTimeout: 6,
	}
	status := ReportStatusOK
	executed := 0
	reasons := make([]string, 0, len(components))
	for _, component := range components {
		if component.Status == ReportStatusSkipped {
			continue
		}
		executed++
		componentPriority, known := priority[component.Status]
		if !known {
			componentPriority = priority[ReportStatusError]
			component.Status = ReportStatusError
		}
		if componentPriority > priority[status] {
			status = component.Status
		}
		if component.Status == ReportStatusOK {
			continue
		}
		reason := strings.TrimSpace(component.Reason)
		if reason == "" {
			reason = string(component.Status)
		}
		reasons = append(reasons, component.Name+": "+reason)
	}
	if executed == 0 {
		return ReportStatusSkipped, "all components skipped"
	}
	return status, strings.Join(reasons, "; ")
}

func tcpSectionStatus(reports []TCPReport) (ReportStatus, string) {
	if len(reports) == 0 {
		return ReportStatusUnavailable, "no TCP results"
	}
	successful, attempts := 0, 0
	for _, report := range reports {
		successful += report.Successful
		attempts += report.Attempts
	}
	if attempts == 0 || successful == 0 {
		return ReportStatusUnavailable, "no TCP handshakes succeeded"
	}
	if successful < attempts {
		return ReportStatusPartial, fmt.Sprintf("%d/%d TCP handshakes succeeded", successful, attempts)
	}
	return ReportStatusOK, ""
}

func aggregateReportStatus(current ReportStatus, sections []SectionReport) ReportStatus {
	if current != ReportStatusOK {
		return current
	}
	for _, section := range sections {
		if !section.Enabled || section.Status == ReportStatusSkipped || section.Status == ReportStatusOK {
			continue
		}
		return ReportStatusPartial
	}
	return ReportStatusOK
}
