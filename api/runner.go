package api

import (
	"context"
	"errors"
	"strings"
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
// versioned structured report. The context is bounded by Config.MaxDuration.
func RunAllTestsContext(parent context.Context, preCheck utils.NetCheckResult, config *Config) *RunResult {
	if parent == nil {
		parent = context.Background()
	}
	if config == nil {
		config = NewDefaultConfig()
	}
	config.ValidateParams()
	ctx, cancel := context.WithTimeout(parent, config.MaxDuration)
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
	legacyConfig := legacyConfigForStructured(config)
	if UsesStructuredComponents() {
		configureStructuredLogging(config.EnableLogger)
		extras := collectStructuredExtras(ctx, preCheck, &structuredConfig)
		status, reason := structuredRunStatus(ctx, extras.err)
		output = renderStructuredRunText(config, extras.dataFiles, extras.components, extras.tcp)
		if ctx.Err() == nil && config.AnalyzeResult {
			progressStarted(ctx, "analysis")
			output = runner.AppendAnalysisSummary(config, output, "", &outputMutex)
			progressCompleted(ctx, "analysis", ReportStatusOK, "")
		}
		endTime := time.Now()
		sections := sectionReports(config, preCheck, extras, status, reason)
		status = aggregateReportStatus(status, sections)
		report := &StructuredReport{
			SchemaVersion: StructuredReportSchema, ECSVersion: config.EcsVersion,
			Status: status, StartedAt: startTime, FinishedAt: endTime,
			DurationMS: endTime.Sub(startTime).Milliseconds(), DeepMode: config.DeepMode,
			PrivacyMode: config.PrivacyMode, Data: extras.data, DataFiles: extras.dataFiles,
			Components: extras.components, TCP: extras.tcp, Sections: sections, Text: output,
		}
		if config.PrivacyMode {
			applyStructuredPrivacy(report)
			output = renderStructuredRunText(config, report.DataFiles, report.Components, report.TCP)
		}
		output = appendStructuredTimeText(output, config, startTime, endTime)
		if !config.PrivacyMode {
			report.Text = output
		}
		jsonData, _ := report.JSON()
		return &RunResult{
			Output: output, StructuredOutput: output, Duration: endTime.Sub(startTime),
			StartTime: startTime, EndTime: endTime, Report: report, JSON: jsonData,
		}
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
		status, reason = ReportStatusPartial, extras.err.Error()
	}
	structuredOutput := ""
	if workflowFinished && (structuredOwnsHardware() || structuredOwnsNetwork()) {
		legacyOutputLength := len(output)
		output = appendStructuredHardwareText(output, config, extras.components)
		output = appendStructuredTCPText(output, config, extras.tcp)
		structuredOutput = output[legacyOutputLength:]
	}
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
		return ReportStatusTimeout, ctx.Err().Error()
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return ReportStatusCanceled, ctx.Err().Error()
	}
	if runErr != nil {
		return ReportStatusPartial, runErr.Error()
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

// appendStructuredHardwareText keeps the text-facing RunAllTests API useful
// when local builds execute hardware through the context-aware structured
// adapters. It only renders already-collected payloads and never invokes a
// benchmark. Published builds retain their original legacy text output.
func appendStructuredHardwareText(output string, config *Config, components []ComponentReport) string {
	if config == nil || (!structuredOwnsHardware() && !structuredOwnsNetwork()) {
		return output
	}
	renderer := newStructuredTextRenderer(config)
	if strings.TrimSpace(output) == "" && len(components) > 0 {
		renderer.header(config)
	}
	for _, component := range components {
		renderer.component(component)
	}
	return output + renderer.builder.String()
}

func appendStructuredTCPText(output string, config *Config, reports []TCPReport) string {
	if config == nil || !structuredOwnsNetwork() || !config.TCPProbeStatus || len(reports) == 0 {
		return output
	}
	renderer := newStructuredTextRenderer(config)
	if strings.TrimSpace(output) == "" {
		renderer.header(config)
	}
	renderer.tcp(reports)
	return output + renderer.builder.String()
}

// RunBasicTests 运行基础信息测试
func RunBasicTests(preCheck utils.NetCheckResult, config *Config) string {
	var (
		basicInfo, securityInfo string
		output, tempOutput      string
		outputMutex             sync.Mutex
	)
	return runner.RunBasicTests(context.Background(), preCheck, config, &basicInfo, &securityInfo, output, tempOutput, &outputMutex)
}

// RunCPUTest 运行CPU测试
func RunCPUTest(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunCPUTest(context.Background(), config, output, tempOutput, &outputMutex)
}

// RunMemoryTest 运行内存测试
func RunMemoryTest(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunMemoryTest(context.Background(), config, output, tempOutput, &outputMutex)
}

// RunDiskTest 运行硬盘测试
func RunDiskTest(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunDiskTest(context.Background(), config, output, tempOutput, &outputMutex)
}

// RunIpInfoCheck 执行IP信息检测
func RunIpInfoCheck(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunIpInfoCheck(context.Background(), config, output, tempOutput, &outputMutex)
}

// RunStreamingTests 运行流媒体测试
func RunStreamingTests(config *Config, mediaInfo string) string {
	var (
		wg1                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return runner.RunStreamingTests(context.Background(), config, &wg1, &mediaInfo, output, tempOutput, &outputMutex, &infoMutex)
}

// RunSecurityTests 运行安全测试
func RunSecurityTests(config *Config, securityInfo string) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunSecurityTests(context.Background(), config, securityInfo, output, tempOutput, &outputMutex)
}

// RunEmailTests 运行邮件端口测试
func RunEmailTests(config *Config, emailInfo string) string {
	var (
		wg2                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return runner.RunEmailTests(context.Background(), config, &wg2, &emailInfo, output, tempOutput, &outputMutex, &infoMutex)
}

// RunNetworkTests 运行网络测试（中文模式）
func RunNetworkTests(config *Config, ptInfo string) string {
	var (
		wg3                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return runner.RunNetworkTests(context.Background(), config, &wg3, &ptInfo, output, tempOutput, &outputMutex, &infoMutex)
}

// RunSpeedTests 运行测速测试（中文模式）
func RunSpeedTests(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunSpeedTests(context.Background(), config, output, tempOutput, &outputMutex)
}

// RunEnglishNetworkTests 运行网络测试（英文模式）
func RunEnglishNetworkTests(config *Config, ptInfo string) string {
	var (
		wg3                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return runner.RunEnglishNetworkTests(context.Background(), config, &wg3, &ptInfo, output, tempOutput, &outputMutex, &infoMutex)
}

// RunEnglishSpeedTests 运行测速测试（英文模式）
func RunEnglishSpeedTests(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunEnglishSpeedTests(context.Background(), config, output, tempOutput, &outputMutex)
}

// AppendTimeInfo 添加时间信息
func AppendTimeInfo(config *Config, output string, startTime time.Time) string {
	var (
		tempOutput  string
		outputMutex sync.Mutex
	)
	return runner.AppendTimeInfo(config, output, tempOutput, startTime, &outputMutex)
}

// HandleUploadResults 处理上传结果
func HandleUploadResults(config *Config, output string) {
	runner.HandleUploadResults(config, output)
}
