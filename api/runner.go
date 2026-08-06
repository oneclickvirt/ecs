package api

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/oneclickvirt/ecs/internal/runner"
	"github.com/oneclickvirt/ecs/utils"
)

// RunResult 运行结果
type RunResult struct {
	Output           string        // 完整输出
	StructuredOutput string        // 结构化组件追加的文本输出
	Duration         time.Duration // 运行时长
	StartTime        time.Time     // 开始时间
	EndTime          time.Time     // 结束时间
	Report           *StructuredReport
	JSON             []byte
}

// NewTextRunResult wraps an already completed classic text run without
// starting any component, registry, hardware, network, or speed test again.
// It is used by CLI/GUI callers that need JSON alongside the established
// real-time output.
func NewTextRunResult(ctx context.Context, preCheck utils.NetCheckResult, config *Config, output string, startedAt, finishedAt time.Time) *RunResult {
	if ctx == nil {
		ctx = context.Background()
	}
	config = validatedStructuredConfig(config)
	status, reason := ReportStatusOK, ""
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		status, reason = ReportStatusTimeout, sanitizePublicReason(ctx.Err().Error())
	} else if errors.Is(ctx.Err(), context.Canceled) {
		status, reason = ReportStatusCanceled, sanitizePublicReason(ctx.Err().Error())
	}
	output = sanitizePublicOutput(output)
	sections := classicSectionReports(config, preCheck, status, reason)
	reportStatus := aggregateReportStatus(status, sections)
	report := &StructuredReport{
		SchemaVersion: StructuredReportSchema, ECSVersion: config.EcsVersion,
		Status: reportStatus, StartedAt: startedAt, FinishedAt: finishedAt,
		DurationMS: finishedAt.Sub(startedAt).Milliseconds(), DeepMode: config.DeepMode,
		PrivacyMode: config.PrivacyMode, Sections: sections, Text: output,
	}
	if config.PrivacyMode {
		applyStructuredPrivacy(report)
		output = ""
	}
	sanitizeStructuredReport(report)
	jsonData, _ := report.JSON()
	return &RunResult{
		Output: output, Duration: finishedAt.Sub(startedAt), StartTime: startedAt,
		EndTime: finishedAt, Report: report, JSON: jsonData,
	}
}

func classicSectionReports(config *Config, preCheck utils.NetCheckResult, status ReportStatus, reason string) []SectionReport {
	definitions := []struct {
		name    string
		enabled bool
		network bool
	}{
		{"basics", config.BasicStatus, false}, {"cpu", config.CpuTestStatus || (config.DeepMode && config.DeepBurnDuration > 0), false},
		{"memory", config.MemoryTestStatus, false}, {"disk", config.DiskTestStatus, false},
		{"media", config.UtTestStatus, true}, {"security", config.SecurityTestStatus, true},
		{"email", config.EmailTestStatus, true}, {"backtrace", config.BacktraceStatus, true},
		{"routes", config.Nt3Status, true}, {"ping", config.PingTestStatus, true},
		{"tgdc", config.TgdcTestStatus, true}, {"web", config.WebTestStatus, true},
		{"tcp", config.TCPProbeStatus, true}, {"speed", config.SpeedTestStatus, true},
	}
	sections := make([]SectionReport, 0, len(definitions))
	for _, definition := range definitions {
		sectionStatus, sectionReason := status, reason
		if !definition.enabled {
			sectionStatus, sectionReason = ReportStatusSkipped, "disabled"
		} else if definition.network && !preCheck.Connected {
			sectionStatus, sectionReason = ReportStatusUnavailable, "network unavailable"
		}
		sections = append(sections, SectionReport{
			Name: definition.name, Enabled: definition.enabled,
			Status: sectionStatus, Reason: sanitizePublicReason(sectionReason),
		})
	}
	return sections
}

func applyLanguageAndUploadRules(preCheck utils.NetCheckResult, config *Config) {
	if config.Language == "en" {
		config.BacktraceStatus = false
		config.Nt3Status = false
	}
	if !preCheck.Connected {
		config.EnableUpload = false
	}
}

// RunAllTests 执行所有测试（高级接口）
// preCheck: 网络检查结果
// config: 配置对象
// 返回: 运行结果
func RunAllTests(preCheck utils.NetCheckResult, config *Config) *RunResult {
	return RunAllTestsContext(context.Background(), preCheck, config)
}

// RunAllTestsContextWithProgress executes the structured workflow and reports
// real section transitions to observer.
func RunAllTestsContextWithProgress(parent context.Context, preCheck utils.NetCheckResult, config *Config, observer ProgressObserver) *RunResult {
	return RunAllTestsContext(WithProgressObserver(parent, observer), preCheck, config)
}

// RunAllTestsContext executes the existing text workflow and returns a
// versioned structured report. A positive Config.MaxDuration bounds the whole
// run; zero leaves the caller context unchanged.
func RunAllTestsContext(parent context.Context, preCheck utils.NetCheckResult, config *Config) *RunResult {
	if parent == nil {
		parent = context.Background()
	}
	if config == nil {
		config = NewDefaultConfig()
	}
	config.ValidateParams()
	defer sanitizeConfiguredLog(config)
	ctx := parent
	cancel := func() {}
	if config.MaxDuration > 0 {
		ctx, cancel = context.WithTimeout(parent, config.MaxDuration)
	}
	defer cancel()
	var (
		wg1, wg2, wg3                                         sync.WaitGroup
		basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo string
		output, tempOutput                                    string
		outputMutex                                           sync.Mutex
		infoMutex                                             sync.Mutex
	)

	startTime := time.Now()
	applyLanguageAndUploadRules(preCheck, config)
	structuredConfig := *config
	// The classic runner remains the sole human-facing text formatter. Local
	// structured adapters still collect machine-readable payloads below, but
	// they must never replace the established option-1 sections with a second
	// table layout.
	legacyConfig := config
	if UsesStructuredComponents() {
		configureStructuredLogging(config.EnableLogger)
	}
	identityReady := make(chan struct{}, 1)
	workflowCtx := runner.WithIdentityReady(ctx, identityReady)
	workflowDone := make(chan struct{})
	go func() {
		defer close(workflowDone)
		switch legacyConfig.Language {
		case "zh":
			runner.RunChineseTests(workflowCtx, preCheck, legacyConfig, &wg1, &wg2, &wg3,
				&basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo,
				&output, tempOutput, startTime, &outputMutex, &infoMutex)
		case "en":
			runner.RunEnglishTests(workflowCtx, preCheck, legacyConfig, &wg1, &wg2, &wg3,
				&basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo,
				&output, tempOutput, startTime, &outputMutex, &infoMutex)
		default:
			runner.RunChineseTests(workflowCtx, preCheck, legacyConfig, &wg1, &wg2, &wg3,
				&basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo,
				&output, tempOutput, startTime, &outputMutex, &infoMutex)
		}
	}()
	extrasChannel := make(chan structuredExtras, 1)
	go func() {
		// Channel synchronization also establishes the happens-before edge for
		// tests.IPV4/IPV6 written by the legacy basic/IP-info stage.
		select {
		case <-identityReady:
		case <-workflowDone:
		case <-ctx.Done():
			return
		}
		extrasChannel <- collectStructuredExtras(ctx, preCheck, &structuredConfig)
	}()
	workflowFinished := true
	select {
	case <-workflowDone:
	case <-ctx.Done():
		workflowFinished = false
	}
	if workflowFinished && config.AnalyzeResult {
		output = runner.AppendAnalysisSummary(config, output, tempOutput, &outputMutex)
	}

	var extras structuredExtras
	if workflowFinished {
		select {
		case extras = <-extrasChannel:
		case <-ctx.Done():
		}
	} else {
		// Do not wait for legacy synchronous providers after the global
		// deadline. The structured result intentionally omits in-flight text
		// and payloads rather than racing their output buffers.
		output = ""
		select {
		case extras = <-extrasChannel:
		default:
		}
	}
	status, reason := ReportStatusOK, ""
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		status, reason = ReportStatusTimeout, ctx.Err().Error()
	} else if errors.Is(ctx.Err(), context.Canceled) {
		status, reason = ReportStatusCanceled, ctx.Err().Error()
	}
	if extras.err != nil && status == ReportStatusOK {
		status, reason = ReportStatusPartial, sanitizePublicReason(extras.err.Error())
	}
	// Structured payloads are returned in report.JSON; human-readable output is
	// intentionally the legacy option-1 stream above.
	structuredOutput := ""
	endTime := time.Now()
	sections := sectionReports(config, preCheck, extras, status, reason)
	status = aggregateReportStatus(status, sections)
	report := &StructuredReport{
		SchemaVersion: StructuredReportSchema, ECSVersion: config.EcsVersion,
		Status: status, StartedAt: startTime, FinishedAt: endTime,
		DurationMS: endTime.Sub(startTime).Milliseconds(), DeepMode: config.DeepMode,
		PrivacyMode: config.PrivacyMode, Data: extras.data, DataFiles: extras.dataFiles,
		Components: extras.components, TCP: extras.tcp,
		Sections: sections, Text: output,
	}
	if config.PrivacyMode {
		applyStructuredPrivacy(report)
		output = report.Text
		structuredOutput = ""
	}
	if !workflowFinished && status == ReportStatusOK {
		status = ReportStatusTimeout
		report.Status = status
		report.Text = ""
	}
	output = sanitizePublicOutput(output)
	structuredOutput = sanitizePublicOutput(structuredOutput)
	if !config.PrivacyMode {
		report.Text = output
	}
	sanitizeStructuredReport(report)
	jsonData, _ := report.JSON()
	return &RunResult{
		Output:           output,
		StructuredOutput: structuredOutput,
		Duration:         endTime.Sub(startTime),
		StartTime:        startTime,
		EndTime:          endTime,
		Report:           report,
		JSON:             jsonData,
	}
}

func structuredRunStatus(ctx context.Context, runErr error) (ReportStatus, string) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ReportStatusTimeout, sanitizePublicReason(ctx.Err().Error())
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return ReportStatusCanceled, sanitizePublicReason(ctx.Err().Error())
	}
	if runErr != nil {
		return ReportStatusPartial, sanitizePublicReason(runErr.Error())
	}
	return ReportStatusOK, ""
}

func legacyConfigForStructured(config *Config) *Config {
	if config == nil || (!structuredOwnsHardware() && !structuredOwnsNetwork()) {
		return config
	}
	legacyCopy := *config
	if structuredOwnsHardware() || structuredOwnsNetwork() {
		if structuredOwnsHardware() {
			// Local component builds make the context-aware structured adapters
			// the single owner of CPU, memory, and disk execution.
			legacyCopy.CpuTestStatus = false
			legacyCopy.MemoryTestStatus = false
			legacyCopy.DiskTestStatus = false
		}
		if structuredOwnsNetwork() {
			// These sections have context-aware structured implementations. Do not
			// execute media, IP evidence, mail, BGP, route, ping, or speed stages
			// through the legacy runner a second time.
			needsIdentity := config.SecurityTestStatus || config.BacktraceStatus
			legacyCopy.UtTestStatus = false
			legacyCopy.SecurityTestStatus = false
			legacyCopy.EmailTestStatus = false
			legacyCopy.BacktraceStatus = false
			legacyCopy.Nt3Status = false
			legacyCopy.PingTestStatus = false
			legacyCopy.SpeedTestStatus = false
			if needsIdentity && !legacyCopy.BasicStatus {
				legacyCopy.OnlyIpInfoCheck = true
			}
		}
	}
	return &legacyCopy
}

// RunBasicTests 运行基础信息测试
func RunBasicTests(preCheck utils.NetCheckResult, config *Config) string {
	var (
		basicInfo, securityInfo string
		output, tempOutput      string
		outputMutex             sync.Mutex
	)
	return finalizePublicText(config, runner.RunBasicTests(context.Background(), preCheck, config, &basicInfo, &securityInfo, output, tempOutput, &outputMutex))
}

// RunCPUTest 运行CPU测试
func RunCPUTest(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return finalizePublicText(config, runner.RunCPUTest(context.Background(), config, output, tempOutput, &outputMutex))
}

// RunMemoryTest 运行内存测试
func RunMemoryTest(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return finalizePublicText(config, runner.RunMemoryTest(context.Background(), config, output, tempOutput, &outputMutex))
}

// RunDiskTest 运行硬盘测试
func RunDiskTest(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return finalizePublicText(config, runner.RunDiskTest(context.Background(), config, output, tempOutput, &outputMutex))
}

// RunIpInfoCheck 执行IP信息检测
func RunIpInfoCheck(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return finalizePublicText(config, runner.RunIpInfoCheck(context.Background(), config, output, tempOutput, &outputMutex))
}

// RunStreamingTests 运行流媒体测试
func RunStreamingTests(config *Config, mediaInfo string) string {
	var (
		wg1                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return finalizePublicText(config, runner.RunStreamingTests(context.Background(), config, &wg1, &mediaInfo, output, tempOutput, &outputMutex, &infoMutex))
}

// RunSecurityTests 运行安全测试
func RunSecurityTests(config *Config, securityInfo string) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return finalizePublicText(config, runner.RunSecurityTests(context.Background(), config, securityInfo, output, tempOutput, &outputMutex))
}

// RunEmailTests 运行邮件端口测试
func RunEmailTests(config *Config, emailInfo string) string {
	var (
		wg2                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return finalizePublicText(config, runner.RunEmailTests(context.Background(), config, &wg2, &emailInfo, output, tempOutput, &outputMutex, &infoMutex))
}

// RunNetworkTests 运行网络测试（中文模式）
func RunNetworkTests(config *Config, ptInfo string) string {
	var (
		wg3                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return finalizePublicText(config, runner.RunNetworkTests(context.Background(), config, &wg3, &ptInfo, output, tempOutput, &outputMutex, &infoMutex))
}

// RunSpeedTests 运行测速测试（中文模式）
func RunSpeedTests(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return finalizePublicText(config, runner.RunSpeedTests(context.Background(), config, output, tempOutput, &outputMutex))
}

// RunEnglishNetworkTests 运行网络测试（英文模式）
func RunEnglishNetworkTests(config *Config, ptInfo string) string {
	var (
		wg3                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return finalizePublicText(config, runner.RunEnglishNetworkTests(context.Background(), config, &wg3, &ptInfo, output, tempOutput, &outputMutex, &infoMutex))
}

// RunEnglishSpeedTests 运行测速测试（英文模式）
func RunEnglishSpeedTests(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return finalizePublicText(config, runner.RunEnglishSpeedTests(context.Background(), config, output, tempOutput, &outputMutex))
}

// AppendTimeInfo 添加时间信息
func AppendTimeInfo(config *Config, output string, startTime time.Time) string {
	var (
		tempOutput  string
		outputMutex sync.Mutex
	)
	return finalizePublicText(config, runner.AppendTimeInfo(config, output, tempOutput, startTime, &outputMutex))
}

// HandleUploadResults 处理上传结果
func HandleUploadResults(config *Config, output string) {
	runner.HandleUploadResults(config, sanitizePublicOutput(output))
	sanitizeConfiguredLog(config)
}
