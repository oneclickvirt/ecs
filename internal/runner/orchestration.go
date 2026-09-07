package runner

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/oneclickvirt/cputest/cpu"
	"github.com/oneclickvirt/ecs/internal/params"
	"github.com/oneclickvirt/ecs/internal/tests"
	"github.com/oneclickvirt/ecs/utils"
	pingmodel "github.com/oneclickvirt/pingtest/model"
	"github.com/oneclickvirt/pingtest/pt"
	"github.com/oneclickvirt/portchecker/email"
)

type bufferedTask struct {
	name string
	run  func(context.Context) string
}

type legacyWorkflowPlan struct {
	basics           func(context.Context)
	concurrentBasics *bufferedTask
	hardware         []bufferedTask
	afterHardware    func(context.Context)
	identityReady    func(context.Context)
	preload          func(context.Context)
	independent      []bufferedTask
	speed            func(context.Context)
	concurrentSpeed  *bufferedTask
	fullConcurrent   bool
	emit             func(string)
}

var (
	runLegacyCPU          = tests.CpuTest
	runLegacyMemory       = tests.MemoryTest
	runLegacyDisk         = tests.DiskTest
	runLegacyMedia        = tests.MediaTest
	runLegacySecurity     = utils.SecurityInfoCheck
	runLegacyEmail        = email.EmailCheck
	runLegacyUpstream     = tests.UpstreamsCheckText
	runLegacyRoute        = tests.NextTrace3CheckText
	runLegacyTelegram     = pt.TelegramDCTest
	runLegacyWebsite      = pt.WebsiteTest
	runLoadedTCPRegistry  = pt.RunLoadedTCPRegistry
	runBuiltinTCPRegistry = pt.RunTCPRegistry
)

func runLegacyTests(ctx context.Context, preCheck utils.NetCheckResult, config *params.Config,
	wg1, wg2, wg3 *sync.WaitGroup, basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo *string,
	output *string, tempOutput string, startTime time.Time, outputMutex, infoMutex *sync.Mutex,
) {
	_ = wg1
	_ = wg2
	_ = wg3
	if config == nil {
		return
	}
	network := preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None"
	speedNetwork := speedNetworkForStack(preCheck.StackType)
	var speedPreloads *tests.PrivateSpeedPreloads
	emit := func(value string) {
		if value == "" {
			return
		}
		outputMutex.Lock()
		fmt.Print(value)
		*output += value
		outputMutex.Unlock()
	}
	plan := legacyWorkflowPlan{
		fullConcurrent: config.Choice == "2",
		basics: func(context.Context) {
			*output = RunBasicTests(ctx, preCheck, config, basicInfo, securityInfo, *output, tempOutput, outputMutex)
		},
		// The concurrent suite cannot use RunBasicTests directly: that helper
		// temporarily redirects process-wide stdout. Keep the same visible text
		// in a buffered section so every test can start before any one completes.
		concurrentBasics: &bufferedTask{name: "basics", run: func(taskCtx context.Context) string {
			return bufferedBasicSection(taskCtx, preCheck, config, basicInfo, securityInfo, infoMutex)
		}},
		afterHardware: func(context.Context) {
			if config.OnlyIpInfoCheck && !config.BasicStatus && network {
				*output = RunIpInfoCheck(ctx, config, *output, tempOutput, outputMutex)
			}
		},
		identityReady: signalIdentityReady,
		emit:          emit,
	}
	if network && config.SpeedTestStatus && config.Language == "zh" {
		// Candidate probes are lightweight but still use the network. Start them
		// only after the hardware stage, then hide selection latency behind the
		// independent network diagnostics. The speed stage waits before opening
		// any transfer stream.
		plan.preload = func(taskCtx context.Context) {
			speedPreloads = tests.StartPrivateSpeedPreloads(taskCtx, []string{"ct", "cu", "cmcc"}, speedNetwork)
		}
	}
	if config.CpuTestStatus || (config.DeepMode && config.DeepBurnDuration > 0) {
		plan.hardware = append(plan.hardware, bufferedTask{name: "cpu", run: func(taskCtx context.Context) string {
			return bufferedCPUSection(taskCtx, config)
		}})
	}
	if config.MemoryTestStatus {
		plan.hardware = append(plan.hardware, bufferedTask{name: "memory", run: func(taskCtx context.Context) string {
			return bufferedMemorySection(taskCtx, config)
		}})
	}
	if config.DiskTestStatus {
		plan.hardware = append(plan.hardware, bufferedTask{name: "disk", run: func(taskCtx context.Context) string {
			return bufferedDiskSection(taskCtx, config)
		}})
	}
	if network {
		if config.UtTestStatus && (config.Language != "zh" || !config.OnlyChinaTest) {
			plan.independent = append(plan.independent, bufferedTask{name: "media", run: func(taskCtx context.Context) string {
				return bufferedMediaSection(taskCtx, config, mediaInfo, infoMutex)
			}})
		}
		if config.SecurityTestStatus {
			plan.independent = append(plan.independent, bufferedTask{name: "security", run: func(taskCtx context.Context) string {
				return bufferedSecuritySection(taskCtx, config, securityInfo, infoMutex)
			}})
		}
		if config.EmailTestStatus {
			plan.independent = append(plan.independent, bufferedTask{name: "email", run: func(taskCtx context.Context) string {
				return bufferedEmailSection(taskCtx, config, emailInfo, infoMutex)
			}})
		}
		if config.Language == "zh" && runtime.GOOS != "windows" {
			if config.BacktraceStatus && !config.OnlyChinaTest {
				plan.independent = append(plan.independent, bufferedTask{name: "backtrace", run: func(taskCtx context.Context) string {
					return bufferedUpstreamSection(taskCtx, config)
				}})
			}
			if config.Nt3Status && !config.OnlyChinaTest {
				plan.independent = append(plan.independent, bufferedTask{name: "routes", run: func(taskCtx context.Context) string {
					return bufferedRouteSection(taskCtx, config)
				}})
			}
		}
		allowPing := config.Language != "zh" || runtime.GOOS != "windows"
		pingTitlePending := true
		if allowPing && (config.OnlyChinaTest || config.PingTestStatus) {
			includeTitle := pingTitlePending
			pingTitlePending = false
			plan.independent = append(plan.independent, bufferedTask{name: "ping", run: func(taskCtx context.Context) string {
				return bufferedConfiguredPing(taskCtx, config, ptInfo, infoMutex, includeTitle)
			}})
		}
		if allowPing && config.TgdcTestStatus {
			includeTitle := pingTitlePending
			pingTitlePending = false
			plan.independent = append(plan.independent, bufferedTask{name: "tgdc", run: func(taskCtx context.Context) string {
				return bufferedTelegramSection(taskCtx, config, includeTitle)
			}})
		}
		if allowPing && config.WebTestStatus {
			includeTitle := pingTitlePending
			plan.independent = append(plan.independent, bufferedTask{name: "web", run: func(taskCtx context.Context) string {
				return bufferedWebsiteSection(taskCtx, config, includeTitle)
			}})
		}
		if config.TCPProbeStatus {
			plan.independent = append(plan.independent, bufferedTask{name: "tcp", run: func(taskCtx context.Context) string {
				return bufferedTCPSection(taskCtx, config)
			}})
		}
		if config.SpeedTestStatus {
			if config.Choice == "2" {
				plan.fullConcurrent = true
				plan.concurrentSpeed = &bufferedTask{name: "speed", run: func(taskCtx context.Context) string {
					if config.Language == "zh" {
						return captureChineseSpeedTests(taskCtx, config, speedNetwork, speedPreloads, false)
					}
					return captureEnglishSpeedTests(taskCtx, config, speedNetwork, false)
				}}
			} else {
				plan.speed = func(context.Context) {
					if config.Language == "zh" {
						*output = RunSpeedTestsWithNetwork(ctx, config, *output, tempOutput, outputMutex, speedNetwork, speedPreloads)
					} else {
						*output = RunEnglishSpeedTestsWithNetwork(ctx, config, *output, tempOutput, outputMutex, speedNetwork)
					}
				}
			}
		}
	}
	runLegacyWorkflowPlan(ctx, plan)
	*output = AppendTimeInfo(config, *output, tempOutput, startTime, outputMutex)
}

// runLegacyWorkflowPlan enforces the workflow barriers. Every concurrent task
// returns a complete chapter string; only the coordinator is allowed to emit
// terminal output.
func runLegacyWorkflowPlan(ctx context.Context, plan legacyWorkflowPlan) {
	if plan.fullConcurrent {
		runFullyConcurrentLegacyWorkflow(ctx, plan)
		return
	}
	if plan.basics != nil {
		plan.basics(ctx)
	}
	runSequentialBufferedTasks(ctx, plan.hardware, plan.emit)
	if plan.afterHardware != nil {
		plan.afterHardware(ctx)
	}
	if plan.identityReady != nil {
		plan.identityReady(ctx)
	}
	// Candidate selection is front-loaded relative to the transfer stage, but
	// starts only after CPU, memory, and disk benchmarks have finished. It then
	// overlaps independent network diagnostics; the speed stage joins it before
	// opening its first throughput stream.
	if plan.preload != nil && ctx.Err() == nil {
		plan.preload(ctx)
	}
	runOrderedBufferedTasks(ctx, plan.independent, plan.emit)
	if plan.speed != nil && ctx.Err() == nil {
		plan.speed(ctx)
	}
}

// runFullyConcurrentLegacyWorkflow launches every enabled test before waiting
// for any result. It is intentionally allowed to distort performance values;
// its only output guarantee is the same chapter order as option 1.
func runFullyConcurrentLegacyWorkflow(ctx context.Context, plan legacyWorkflowPlan) {
	if plan.preload != nil && ctx.Err() == nil {
		plan.preload(ctx)
	}
	if ctx.Err() != nil {
		return
	}

	allTasks := make([]bufferedTask, 0, len(plan.hardware)+len(plan.independent)+2)
	basicOffset := 0
	if plan.concurrentBasics != nil {
		basicTask := *plan.concurrentBasics
		if plan.identityReady != nil {
			originalRun := basicTask.run
			basicTask.run = func(taskCtx context.Context) (value string) {
				defer plan.identityReady(taskCtx)
				return originalRun(taskCtx)
			}
		}
		allTasks = append(allTasks, basicTask)
		basicOffset = 1
	} else if plan.basics != nil {
		// Keep manually assembled legacy plans source-compatible. Production
		// option 2 always supplies concurrentBasics above.
		plan.basics(ctx)
		if plan.identityReady != nil {
			plan.identityReady(ctx)
		}
	}
	allTasks = append(allTasks, plan.hardware...)
	allTasks = append(allTasks, plan.independent...)
	if plan.concurrentSpeed != nil {
		allTasks = append(allTasks, *plan.concurrentSpeed)
	}
	channels := startBufferedTasks(ctx, allTasks)
	values, completed := collectBufferedTaskResults(ctx, channels)
	if !completed {
		return
	}

	hardwareStart := basicOffset
	hardwareEnd := hardwareStart + len(plan.hardware)
	independentEnd := hardwareEnd + len(plan.independent)
	if basicOffset == 1 && values[0] != "" && plan.emit != nil {
		plan.emit(values[0])
	}
	for _, value := range values[hardwareStart:hardwareEnd] {
		if value != "" && plan.emit != nil {
			plan.emit(value)
		}
	}
	// This path can capture its own output. It deliberately runs after every
	// concurrently-started task has completed so no process-wide stdout capture
	// can consume another section's result.
	if plan.afterHardware != nil && ctx.Err() == nil {
		plan.afterHardware(ctx)
	}
	for _, value := range values[hardwareEnd:independentEnd] {
		if value != "" && plan.emit != nil {
			plan.emit(value)
		}
	}
	for _, value := range values[independentEnd:] {
		if value != "" && plan.emit != nil {
			plan.emit(value)
		}
	}
}

// Hardware benchmarks are deliberately serialized so CPU, memory, and disk
// measurements do not distort one another.
func runSequentialBufferedTasks(ctx context.Context, tasks []bufferedTask, emit func(string)) {
	for _, task := range tasks {
		if ctx.Err() != nil {
			return
		}
		value := runBufferedTask(ctx, task)
		if value != "" && emit != nil && ctx.Err() == nil {
			emit(value)
		}
	}
}

// runOrderedBufferedTasks starts every task immediately, then consumes each
// private result channel in display order. A later chapter may finish early,
// but it cannot overtake the next chapter expected by the renderer.
func runOrderedBufferedTasks(ctx context.Context, tasks []bufferedTask, emit func(string)) {
	values, completed := collectBufferedTaskResults(ctx, startBufferedTasks(ctx, tasks))
	if !completed {
		return
	}
	for _, value := range values {
		if value != "" && emit != nil && ctx.Err() == nil {
			emit(value)
		}
	}
}

func startBufferedTasks(ctx context.Context, tasks []bufferedTask) []<-chan string {
	if len(tasks) == 0 {
		return nil
	}
	results := make([]<-chan string, len(tasks))
	for index := range tasks {
		result := make(chan string, 1)
		results[index] = result
		task := tasks[index]
		go func() {
			value := ""
			defer func() {
				_ = recover()
				result <- value
				close(result)
			}()
			value = runBufferedTask(ctx, task)
		}()
	}
	return results
}

func collectBufferedTaskResults(ctx context.Context, results []<-chan string) ([]string, bool) {
	values := make([]string, len(results))
	for index, result := range results {
		select {
		case values[index] = <-result:
		case <-ctx.Done():
			return values, false
		}
	}
	return values, true
}

func runBufferedTask(ctx context.Context, task bufferedTask) (value string) {
	defer func() {
		if recover() != nil {
			value = ""
		}
	}()
	if task.run == nil {
		return ""
	}
	return task.run(ctx)
}

func bufferedCPUSection(ctx context.Context, config *params.Config) string {
	if config == nil || ctx.Err() != nil {
		return ""
	}
	var body strings.Builder
	title := ""
	if config.CpuTestStatus {
		method, result := runLegacyCPU(config.Language, config.CpuTestMethod, config.CpuTestThreadMode)
		if config.Language == "zh" {
			title = fmt.Sprintf("CPU测试-通过%s测试", method)
		} else {
			title = fmt.Sprintf("CPU-Test--%s-Method", method)
		}
		body.WriteString(result)
	}
	if config.DeepMode && config.DeepBurnDuration > 0 && ctx.Err() == nil {
		if title == "" {
			if config.Language == "zh" {
				title = "CPU测试"
			} else {
				title = "CPU-Test"
			}
		}
		body.WriteString(formatCPUBurnResult(config.Language, runLegacyCPUBurn(ctx, cpu.BurnConfig{
			Threads: runtime.NumCPU(), Duration: config.DeepBurnDuration, MaxPrime: 50000,
		})))
	}
	if title == "" {
		return ""
	}
	return legacySectionText(title, config.Width, body.String())
}

func bufferedMemorySection(ctx context.Context, config *params.Config) string {
	if config == nil || !config.MemoryTestStatus || ctx.Err() != nil {
		return ""
	}
	method, result := runLegacyMemory(config.Language, config.MemoryTestMethod)
	title := fmt.Sprintf("Memory-Test--%s-Method", method)
	if config.Language == "zh" {
		title = fmt.Sprintf("内存测试-通过%s测试", method)
	}
	return legacySectionText(title, config.Width, result)
}

func bufferedDiskSection(ctx context.Context, config *params.Config) string {
	if config == nil || !config.DiskTestStatus || ctx.Err() != nil {
		return ""
	}
	section := func(method, result string) string {
		title := fmt.Sprintf("Disk-Test--%s-Method", method)
		if config.Language == "zh" {
			title = fmt.Sprintf("硬盘测试-通过%s测试", method)
		}
		if strings.TrimSpace(result) == "" {
			if config.Language == "en" {
				result = "Disk benchmark returned no usable data.\n"
			} else {
				result = "磁盘测试未返回可用的性能数据。\n"
			}
		}
		return legacySectionText(title, config.Width, result)
	}
	if config.AutoChangeDiskMethod {
		method, result := runLegacyDisk(config.Language, config.DiskTestMethod, config.DiskTestPath, config.DiskMultiCheck, config.AutoChangeDiskMethod)
		return section(method, result)
	}
	_, ddResult := runLegacyDisk(config.Language, "dd", config.DiskTestPath, config.DiskMultiCheck, config.AutoChangeDiskMethod)
	_, fioResult := runLegacyDisk(config.Language, "fio", config.DiskTestPath, config.DiskMultiCheck, config.AutoChangeDiskMethod)
	return section("dd", ddResult) + section("fio", fioResult)
}

func bufferedMediaSection(ctx context.Context, config *params.Config, mediaInfo *string, infoMutex *sync.Mutex) string {
	if config == nil || ctx.Err() != nil || !waitIdentityReady(ctx) {
		return ""
	}
	result := runLegacyMedia(config.Language, config.UnlockTestRegion, config.UnlockTestIPVersion, config.UnlockTestShowIP)
	setBufferedInfo(mediaInfo, result, infoMutex)
	title := "Cross-Border-Platform-Unlock"
	if config.Language == "zh" {
		title = "跨国平台解锁"
	}
	return legacySectionText(title, config.Width, result)
}

func bufferedSecuritySection(ctx context.Context, config *params.Config, securityInfo *string, infoMutex *sync.Mutex) string {
	if config == nil || ctx.Err() != nil || !waitIdentityReady(ctx) {
		return ""
	}
	result := runLegacySecurity(config.Language)
	setBufferedInfo(securityInfo, result, infoMutex)
	title := "IP-Quality-Check"
	if config.Language == "zh" {
		title = "IP质量检测"
	}
	return legacySectionText(title, config.Width, result)
}

func bufferedEmailSection(ctx context.Context, config *params.Config, emailInfo *string, infoMutex *sync.Mutex) string {
	if config == nil || ctx.Err() != nil {
		return ""
	}
	result := runLegacyEmail()
	setBufferedInfo(emailInfo, result, infoMutex)
	title := "Email-Port-Check"
	if config.Language == "zh" {
		title = "邮件端口检测"
	}
	return legacySectionText(title, config.Width, result)
}

func bufferedUpstreamSection(ctx context.Context, config *params.Config) string {
	if config == nil || ctx.Err() != nil || !waitIdentityReady(ctx) {
		return ""
	}
	return legacySectionText("上游及回程线路检测", config.Width, runLegacyUpstream(config.Language))
}

func bufferedRouteSection(ctx context.Context, config *params.Config) string {
	if config == nil || ctx.Err() != nil || !waitIdentityReady(ctx) {
		return ""
	}
	return legacySectionText("三网回程路由检测", config.Width, runLegacyRoute(config.Language, config.Nt3Location, config.Nt3CheckType))
}

func bufferedConfiguredPing(ctx context.Context, config *params.Config, ptInfo *string, infoMutex *sync.Mutex, includeTitle bool) string {
	if config == nil || ctx.Err() != nil {
		return ""
	}
	result := runConfiguredPing(config)
	setBufferedInfo(ptInfo, result, infoMutex)
	if strings.TrimSpace(result) == "" && !includeTitle {
		return ""
	}
	return bufferedPingPart(config, result, includeTitle)
}

func bufferedTelegramSection(ctx context.Context, config *params.Config, includeTitle bool) string {
	if config == nil || ctx.Err() != nil {
		return ""
	}
	return bufferedPingPart(config, runLegacyTelegram(), includeTitle)
}

func bufferedWebsiteSection(ctx context.Context, config *params.Config, includeTitle bool) string {
	if config == nil || ctx.Err() != nil {
		return ""
	}
	return bufferedPingPart(config, runLegacyWebsite(), includeTitle)
}

func bufferedPingPart(config *params.Config, body string, includeTitle bool) string {
	body = reserveLeadingCell(body)
	if !includeTitle {
		return body
	}
	title := "PING-Test"
	if config.Language == "zh" {
		title = "PING值检测"
	}
	return legacySectionText(title, config.Width, body)
}

func bufferedTCPSection(ctx context.Context, config *params.Config) string {
	if config == nil || !config.TCPProbeStatus || ctx.Err() != nil {
		return ""
	}
	probeConfig := pt.DefaultTCPProbeConfig()
	var results []pt.TCPResult
	if config.DataOffline {
		results = runBuiltinTCPRegistry(ctx, probeConfig)
	} else {
		loaded, _, err := runLoadedTCPRegistry(ctx, probeConfig)
		if err == nil {
			results = loaded
		} else {
			results = runBuiltinTCPRegistry(ctx, probeConfig)
		}
	}
	formatted := pt.FormatTCPResultsWithOptions(results, pt.TCPFormatOptions{
		Format: pt.TCPTextFormat(config.TCPTextFormat), Sort: pingmodel.TCPSort(config.TCPSortOrder), Language: config.Language,
	})
	title := "TCP-Handshake-Latency"
	if config.Language == "zh" {
		title = "TCP握手延迟"
	}
	return legacySectionText(title, config.Width, formatted+"\n")
}

func setBufferedInfo(target *string, value string, mutex *sync.Mutex) {
	if target == nil {
		return
	}
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}
	*target = value
}

func legacySectionText(title string, width int, body string) string {
	if width <= 0 {
		width = 80
	}
	titleWidth := runewidth.StringWidth(title)
	padding := width - titleWidth
	if padding < 0 {
		padding = 0
	}
	left := padding / 2
	var builder strings.Builder
	builder.WriteString(strings.Repeat("-", left))
	builder.WriteString(title)
	builder.WriteString(strings.Repeat("-", padding-left))
	builder.WriteByte('\n')
	builder.WriteString(reserveLeadingCell(body))
	return builder.String()
}

func reserveLeadingCell(value string) string {
	if value == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(value) + 8)
	lineStart := true
	for _, current := range value {
		if lineStart && current != '\n' && current != '\r' && current != ' ' && current != '\t' && current != '-' && current != '=' {
			builder.WriteByte(' ')
		}
		builder.WriteRune(current)
		lineStart = current == '\n' || current == '\r'
	}
	if !strings.HasSuffix(value, "\n") {
		builder.WriteByte('\n')
	}
	return builder.String()
}
