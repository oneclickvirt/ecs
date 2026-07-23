package runner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/oneclickvirt/cputest/cpu"
	"github.com/oneclickvirt/ecs/internal/analysis"
	"github.com/oneclickvirt/ecs/internal/params"
	"github.com/oneclickvirt/ecs/internal/tests"
	"github.com/oneclickvirt/ecs/utils"
	pingmodel "github.com/oneclickvirt/pingtest/model"
	"github.com/oneclickvirt/pingtest/pt"
)

type identityReadyContextKey struct{}

var runLegacyCPUBurn = cpu.RunBurn

// WithIdentityReady lets the orchestration layer wait until the legacy basic
// stage has finished publishing its IP identity. This prevents structured
// security/backtrace probes from racing the legacy global identity fields.
// The channel should be buffered with capacity one.
func WithIdentityReady(ctx context.Context, ready chan struct{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, identityReadyContextKey{}, ready)
}

func signalIdentityReady(ctx context.Context) {
	if ctx == nil {
		return
	}
	ready, _ := ctx.Value(identityReadyContextKey{}).(chan struct{})
	if ready == nil {
		return
	}
	select {
	case ready <- struct{}{}:
	default:
	}
}

// RunChineseTests runs all tests in Chinese mode
func RunChineseTests(ctx context.Context, preCheck utils.NetCheckResult, config *params.Config, wg1, wg2, wg3 *sync.WaitGroup, basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo *string, output *string, tempOutput string, startTime time.Time, outputMutex *sync.Mutex, infoMutex *sync.Mutex) {
	runLegacyTests(ctx, preCheck, config, wg1, wg2, wg3, basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo, output, tempOutput, startTime, outputMutex, infoMutex)
}

// RunEnglishTests runs all tests in English mode
func RunEnglishTests(ctx context.Context, preCheck utils.NetCheckResult, config *params.Config, wg1, wg2, wg3 *sync.WaitGroup, basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo *string, output *string, tempOutput string, startTime time.Time, outputMutex *sync.Mutex, infoMutex *sync.Mutex) {
	runLegacyTests(ctx, preCheck, config, wg1, wg2, wg3, basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo, output, tempOutput, startTime, outputMutex, infoMutex)
}

// RunTCPTests appends the new handshake diagnostics as an explicit standalone
// section. It is disabled by default so existing suites retain their original
// real-time output and duration.
func RunTCPTests(ctx context.Context, config *params.Config, output, tempOutput string, outputMutex *sync.Mutex) string {
	if ctx.Err() != nil || config == nil || !config.TCPProbeStatus {
		return output
	}
	_ = tempOutput
	section := bufferedTCPSection(ctx, config)
	if section == "" {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	fmt.Print(section)
	return output + section
}

func runConfiguredPing(config *params.Config) string {
	return pt.PingTestWithOptions(pt.PingOptions{
		Language: config.Language,
		Scope:    pingmodel.PingScope(config.PingScope),
		Sort:     pingmodel.PingSort(config.PingSortOrder),
	})
}

// RunIpInfoCheck performs IP info check
func RunIpInfoCheck(ctx context.Context, config *params.Config, output, tempOutput string, outputMutex *sync.Mutex) string {
	if ctx.Err() != nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		var ipinfo string
		tests.IPV4, tests.IPV6, ipinfo = utils.OnlyBasicsIpInfo(config.Language)
		if ipinfo != "" {
			if config.Language == "zh" {
				utils.PrintCenteredTitle("IP信息", config.Width)
			} else {
				utils.PrintCenteredTitle("IP-Information", config.Width)
			}
			fmt.Printf("%s", ipinfo)
		}
	}, tempOutput, output)
}

// RunBasicTests runs basic system tests
func RunBasicTests(ctx context.Context, preCheck utils.NetCheckResult, config *params.Config, basicInfo, securityInfo *string, output, tempOutput string, outputMutex *sync.Mutex) string {
	if ctx.Err() != nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		utils.PrintHead(config.Language, config.Width, config.EcsVersion)
		if config.BasicStatus || config.SecurityTestStatus {
			if config.BasicStatus {
				if config.Language == "zh" {
					utils.PrintCenteredTitle("系统基础信息", config.Width)
				} else {
					utils.PrintCenteredTitle("System-Basic-Information", config.Width)
				}
			}
			if preCheck.Connected && preCheck.StackType == "DualStack" {
				tests.IPV4, tests.IPV6, *basicInfo, *securityInfo, config.Nt3CheckType = utils.BasicsAndSecurityCheck(config.Language, config.Nt3CheckType, false)
			} else if preCheck.Connected && preCheck.StackType == "IPv4" {
				tests.IPV4, tests.IPV6, *basicInfo, *securityInfo, config.Nt3CheckType = utils.BasicsAndSecurityCheck(config.Language, "ipv4", false)
			} else if preCheck.Connected && preCheck.StackType == "IPv6" {
				tests.IPV4, tests.IPV6, *basicInfo, *securityInfo, config.Nt3CheckType = utils.BasicsAndSecurityCheck(config.Language, "ipv6", false)
			} else {
				tests.IPV4, tests.IPV6, *basicInfo, *securityInfo, config.Nt3CheckType = utils.BasicsAndSecurityCheck(config.Language, "", false)
				config.SecurityTestStatus = false
			}
			if config.BasicStatus {
				fmt.Printf("%s", *basicInfo)
			} else if shouldPrintBriefIPLinesInBasicStage(config) {
				scanner := bufio.NewScanner(strings.NewReader(*basicInfo))
				for scanner.Scan() {
					line := scanner.Text()
					if strings.Contains(line, "IPV") {
						fmt.Println(line)
					}
				}
			}
		}
	}, tempOutput, output)
}

// shouldPrintBriefIPLinesInBasicStage decides whether to print brief IPV lines
// from basic info when basic system info is disabled.
func shouldPrintBriefIPLinesInBasicStage(config *params.Config) bool {
	if config == nil {
		return false
	}
	if !config.SecurityTestStatus || config.BasicStatus {
		return false
	}
	// IP-only profiles print a dedicated "IP信息" section later.
	if config.OnlyIpInfoCheck {
		return false
	}
	return config.Choice == "6" || config.Choice == "9"
}

func shouldPrintPingInfoSection(config *params.Config, info string) bool {
	if config == nil {
		return false
	}
	return info != "" && (config.OnlyChinaTest || config.PingTestStatus)
}

func shouldPrintPingExtraSectionWithoutInfo(config *params.Config) bool {
	if config == nil {
		return false
	}
	return !config.OnlyChinaTest && !config.PingTestStatus && (config.TgdcTestStatus || config.WebTestStatus)
}

// RunCPUTest runs CPU test
func RunCPUTest(ctx context.Context, config *params.Config, output, tempOutput string, outputMutex *sync.Mutex) string {
	if ctx.Err() != nil || config == nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		printedSection := false
		if config.CpuTestStatus {
			realTestMethod, res := tests.CpuTest(config.Language, config.CpuTestMethod, config.CpuTestThreadMode)
			if config.Language == "zh" {
				utils.PrintCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", realTestMethod), config.Width)
			} else {
				utils.PrintCenteredTitle(fmt.Sprintf("CPU-Test--%s-Method", realTestMethod), config.Width)
			}
			fmt.Print(res)
			printedSection = true
		}
		if config.DeepMode && config.DeepBurnDuration > 0 && ctx.Err() == nil {
			if !printedSection {
				if config.Language == "zh" {
					utils.PrintCenteredTitle("CPU测试", config.Width)
				} else {
					utils.PrintCenteredTitle("CPU-Test", config.Width)
				}
			}
			result := runLegacyCPUBurn(ctx, cpu.BurnConfig{
				Threads: runtime.NumCPU(), Duration: config.DeepBurnDuration, MaxPrime: 50000,
			})
			printCPUBurnResult(config.Language, result)
		}
	}, tempOutput, output)
}

// RunCPUBurnTest is retained for API compatibility. The full workflow calls
// RunCPUTest, which renders this result inside the existing CPU section.
func RunCPUBurnTest(ctx context.Context, config *params.Config, output, tempOutput string, outputMutex *sync.Mutex) string {
	if ctx.Err() != nil || config == nil || !config.DeepMode || config.DeepBurnDuration <= 0 {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		result := runLegacyCPUBurn(ctx, cpu.BurnConfig{
			Threads: runtime.NumCPU(), Duration: config.DeepBurnDuration, MaxPrime: 50000,
		})
		printCPUBurnResult(config.Language, result)
	}, tempOutput, output)
}

func printCPUBurnResult(language string, result cpu.BurnResult) {
	fmt.Print(formatCPUBurnResult(language, result))
}

func formatCPUBurnResult(language string, result cpu.BurnResult) string {
	duration := time.Duration(result.DurationMS) * time.Millisecond
	if result.Status == "ok" {
		if language == "zh" {
			return fmt.Sprintf("压力测试            : %s / %d 线程 / %.2f 次/秒 / %d 次\n",
				formatCompactDuration(duration), result.EffectiveThreads, result.EventsPerSecond, result.Events)
		}
		return fmt.Sprintf("Pressure Test        : %s / %d threads / %.2f events/s / %d events\n",
			formatCompactDuration(duration), result.EffectiveThreads, result.EventsPerSecond, result.Events)
	}
	reason := strings.TrimSpace(result.Error)
	if reason == "" {
		reason = strings.TrimSpace(result.Status)
	}
	if language == "zh" {
		return fmt.Sprintf("压力测试            : %s\n", reason)
	}
	return fmt.Sprintf("Pressure Test        : %s\n", reason)
}

func formatCompactDuration(duration time.Duration) string {
	if duration <= 0 {
		return "0s"
	}
	if duration%time.Second == 0 {
		return fmt.Sprintf("%ds", int64(duration/time.Second))
	}
	return duration.Round(time.Millisecond).String()
}

// RunMemoryTest runs memory test
func RunMemoryTest(ctx context.Context, config *params.Config, output, tempOutput string, outputMutex *sync.Mutex) string {
	if ctx.Err() != nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if config.MemoryTestStatus {
			realTestMethod, res := tests.MemoryTest(config.Language, config.MemoryTestMethod)
			if config.Language == "zh" {
				utils.PrintCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", realTestMethod), config.Width)
			} else {
				utils.PrintCenteredTitle(fmt.Sprintf("Memory-Test--%s-Method", realTestMethod), config.Width)
			}
			fmt.Print(res)
		}
	}, tempOutput, output)
}

// RunDiskTest runs disk test
func RunDiskTest(ctx context.Context, config *params.Config, output, tempOutput string, outputMutex *sync.Mutex) string {
	if ctx.Err() != nil || config == nil || !config.DiskTestStatus {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		fmt.Print(bufferedDiskSection(ctx, config))
	}, tempOutput, output)
}

// RunStreamingTests runs platform unlock tests
func RunStreamingTests(ctx context.Context, config *params.Config, wg1 *sync.WaitGroup, mediaInfo *string, output, tempOutput string, outputMutex *sync.Mutex, infoMutex *sync.Mutex) string {
	if ctx.Err() != nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if config.UtTestStatus && (config.Language == "zh" && !config.OnlyChinaTest || config.Language == "en") {
			wg1.Wait()
			if config.Language == "zh" {
				utils.PrintCenteredTitle("跨国平台解锁", config.Width)
			} else {
				utils.PrintCenteredTitle("Cross-Border-Platform-Unlock", config.Width)
			}
			infoMutex.Lock()
			info := *mediaInfo
			infoMutex.Unlock()
			fmt.Printf("%s", info)
		}
	}, tempOutput, output)
}

// RunSecurityTests runs security tests
func RunSecurityTests(ctx context.Context, config *params.Config, securityInfo, output, tempOutput string, outputMutex *sync.Mutex) string {
	if ctx.Err() != nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if config.SecurityTestStatus {
			if config.Language == "zh" {
				utils.PrintCenteredTitle("IP质量检测", config.Width)
			} else {
				utils.PrintCenteredTitle("IP-Quality-Check", config.Width)
			}
			fmt.Printf("%s", securityInfo)
		}
	}, tempOutput, output)
}

// RunEmailTests runs email port tests
func RunEmailTests(ctx context.Context, config *params.Config, wg2 *sync.WaitGroup, emailInfo *string, output, tempOutput string, outputMutex *sync.Mutex, infoMutex *sync.Mutex) string {
	if ctx.Err() != nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if config.EmailTestStatus {
			wg2.Wait()
			if config.Language == "zh" {
				utils.PrintCenteredTitle("邮件端口检测", config.Width)
			} else {
				utils.PrintCenteredTitle("Email-Port-Check", config.Width)
			}
			infoMutex.Lock()
			info := *emailInfo
			infoMutex.Unlock()
			fmt.Println(info)
		}
	}, tempOutput, output)
}

// RunNetworkTests runs network tests (Chinese mode)
func RunNetworkTests(ctx context.Context, config *params.Config, wg3 *sync.WaitGroup, ptInfo *string, output, tempOutput string, outputMutex *sync.Mutex, infoMutex *sync.Mutex) string {
	if ctx.Err() != nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if config.BacktraceStatus && !config.OnlyChinaTest {
			utils.PrintCenteredTitle("上游及回程线路检测", config.Width)
			tests.UpstreamsCheck(config.Language)
		}
		if config.Nt3Status && !config.OnlyChinaTest {
			utils.PrintCenteredTitle("三网回程路由检测", config.Width)
			tests.NextTrace3Check(config.Language, config.Nt3Location, config.Nt3CheckType)
		}
		printedPingTitle := false
		if config.OnlyChinaTest || config.PingTestStatus {
			wg3.Wait()
			infoMutex.Lock()
			info := *ptInfo
			infoMutex.Unlock()
			if shouldPrintPingInfoSection(config, info) {
				utils.PrintCenteredTitle("PING值检测", config.Width)
				printedPingTitle = true
				fmt.Println(info)
			}
			if config.TgdcTestStatus {
				if !printedPingTitle {
					utils.PrintCenteredTitle("PING值检测", config.Width)
					printedPingTitle = true
				}
				fmt.Println(pt.TelegramDCTest())
			}
			if config.WebTestStatus {
				if !printedPingTitle {
					utils.PrintCenteredTitle("PING值检测", config.Width)
					printedPingTitle = true
				}
				fmt.Println(pt.WebsiteTest())
			}
		}
		if shouldPrintPingExtraSectionWithoutInfo(config) {
			utils.PrintCenteredTitle("PING值检测", config.Width)
			if config.TgdcTestStatus {
				fmt.Println(pt.TelegramDCTest())
			}
			if config.WebTestStatus {
				fmt.Println(pt.WebsiteTest())
			}
		}
		// 等待第三方库的输出完全刷新到标准输出
		time.Sleep(300 * time.Millisecond)
	}, tempOutput, output)
}

// RunSpeedTests runs speed tests (Chinese mode)
func RunSpeedTests(ctx context.Context, config *params.Config, output, tempOutput string, outputMutex *sync.Mutex) string {
	if ctx.Err() != nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if config.SpeedTestStatus {
			utils.PrintCenteredTitle("就近节点测速", config.Width)
			tests.ShowHead(config.Language)
			if config.Choice == "1" || !config.MenuMode {
				tests.NearbySP()
				tests.CustomSP("net", "global", 2, config.Language)
				tests.CustomSP("net", "cu", config.SpNum, config.Language)
				tests.CustomSP("net", "ct", config.SpNum, config.Language)
				tests.CustomSP("net", "cmcc", config.SpNum, config.Language)
			} else if config.Choice == "2" || config.Choice == "3" || config.Choice == "4" || config.Choice == "5" {
				// 中文模式：就近测速 + 三网各1个 + Other 1个（带回退）
				if config.Language == "zh" {
					tests.NearbySP()
					tests.CustomSP("net", "other", 1, config.Language)
					tests.CustomSP("net", "cu", 1, config.Language)
					tests.CustomSP("net", "ct", 1, config.Language)
					tests.CustomSP("net", "cmcc", 1, config.Language)
				} else {
					// 英文模式：保持原有逻辑，测4个global节点
					tests.CustomSP("net", "global", 4, config.Language)
				}
			} else if config.Choice == "6" {
				tests.CustomSP("net", "global", 11, config.Language)
			} else {
				// Custom menu mode and any other fallback choices.
				tests.NearbySP()
				tests.CustomSP("net", "cu", config.SpNum, config.Language)
				tests.CustomSP("net", "ct", config.SpNum, config.Language)
				tests.CustomSP("net", "cmcc", config.SpNum, config.Language)
			}
			// 等待第三方库的输出完全刷新到标准输出
			time.Sleep(500 * time.Millisecond)
		}
	}, tempOutput, output)
}

// RunEnglishNetworkTests runs network tests (English mode)
func RunEnglishNetworkTests(ctx context.Context, config *params.Config, wg3 *sync.WaitGroup, ptInfo *string, output, tempOutput string, outputMutex *sync.Mutex, infoMutex *sync.Mutex) string {
	if ctx.Err() != nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		printedPingTitle := false
		if config.PingTestStatus {
			wg3.Wait()
			infoMutex.Lock()
			info := *ptInfo
			infoMutex.Unlock()
			if strings.TrimSpace(info) != "" {
				utils.PrintCenteredTitle("PING-Test", config.Width)
				printedPingTitle = true
				fmt.Println(info)
			}
		}
		if config.TgdcTestStatus || config.WebTestStatus {
			if !printedPingTitle {
				utils.PrintCenteredTitle("PING-Test", config.Width)
			}
			if config.TgdcTestStatus {
				fmt.Println(pt.TelegramDCTest())
			}
			if config.WebTestStatus {
				fmt.Println(pt.WebsiteTest())
			}
		}
		// 等待第三方库的输出完全刷新到标准输出
		time.Sleep(300 * time.Millisecond)
	}, tempOutput, output)
}

// RunEnglishSpeedTests runs speed tests (English mode)
func RunEnglishSpeedTests(ctx context.Context, config *params.Config, output, tempOutput string, outputMutex *sync.Mutex) string {
	if ctx.Err() != nil {
		return output
	}
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if config.SpeedTestStatus {
			utils.PrintCenteredTitle("Speed-Test", config.Width)
			tests.ShowHead(config.Language)
			// English mode deliberately excludes nearby/CN operator discovery;
			// the global registry supplies representative international nodes.
			tests.CustomSP("net", "global", max(4, config.SpNum), config.Language)
			// 等待第三方库的输出完全刷新到标准输出
			time.Sleep(500 * time.Millisecond)
		}
	}, tempOutput, output)
}

// AppendTimeInfo appends timing information
func AppendTimeInfo(config *params.Config, output, tempOutput string, startTime time.Time, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
	return utils.PrintAndCapture(func() {
		utils.PrintCenteredTitle("", config.Width)
		if config.Language == "zh" {
			fmt.Printf("花费          : %d 分 %d 秒\n", minutes, seconds)
			fmt.Printf("时间          : %s\n", currentTime)
		} else {
			fmt.Printf("Cost    Time          : %d min %d sec\n", minutes, seconds)
			fmt.Printf("Current Time          : %s\n", currentTime)
		}
		utils.PrintCenteredTitle("", config.Width)
	}, tempOutput, output)
}

// AppendAnalysisSummary appends a concise bilingual summary for easier interpretation.
func AppendAnalysisSummary(config *params.Config, output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	finalOutput := output
	return utils.PrintAndCapture(func() {
		summary := analysis.GenerateSummary(config, finalOutput)
		if strings.TrimSpace(summary) == "" {
			return
		}
		fmt.Println(summary)
	}, tempOutput, output)
}

// printTimeInfo prints elapsed-time / current-time to stdout directly (no mutex).
func printTimeInfo(config *params.Config, minutes, seconds int, currentTime string) {
	utils.PrintCenteredTitle("", config.Width)
	if config.Language == "zh" {
		fmt.Printf("花费          : %d 分 %d 秒\n", minutes, seconds)
		fmt.Printf("时间          : %s\n", currentTime)
	} else {
		fmt.Printf("Cost    Time          : %d min %d sec\n", minutes, seconds)
		fmt.Printf("Current Time          : %s\n", currentTime)
	}
	utils.PrintCenteredTitle("", config.Width)
}

// HandleSignalInterrupt handles interrupt signals
//
// First Ctrl+C  → cancel the context (no new tests start) and wait for the
// currently-running test to finish, then upload & exit gracefully. Once the
// soft deadline has fired, the next signal is treated as a hard stop.
//
// Second Ctrl+C (or if cleanup cannot make progress) → kill the process group
// so that any child subprocess (stream, fio, dd, sysbench, geekbench …) is
// also terminated immediately, then os.Exit(1).
func HandleSignalInterrupt(ctx context.Context, cancel context.CancelFunc, sig chan os.Signal, config *params.Config, startTime *time.Time, output *string, tempOutput string, uploadDone chan bool, outputMutex *sync.Mutex) {
	select {
	case <-sig:
		if ctx.Err() != nil {
			forceExit(1)
			return
		}

		// ── First Ctrl+C ────────────────────────────────────────────────────────
		// Cancel context so that tests that have not yet started are skipped.
		cancel()

		// Arm a goroutine that watches for a second Ctrl+C or a 30-second
		// timeout, whichever comes first, and then force-terminates everything.
		go func() {
			select {
			case <-sig:
				// Second Ctrl+C → hard kill
				forceExit(1)
			case <-time.After(30 * time.Second):
				// Cleanup stuck for 30 s → hard kill
				forceExit(1)
			}
		}()

		if config.Finish {
			os.Exit(0)
		}
		// ── Snapshot timing information ──────────────────────────────────────────
		endTime := time.Now()
		duration := endTime.Sub(*startTime)
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")

		// ── Acquire outputMutex without blocking forever ─────────────────────────
		// The currently running test (memory/disk/cpu…) may hold the lock for a
		// long time. We try to get it in a background goroutine and give it up to
		// 10 seconds. If we timeout we still print the time info directly to the
		// terminal and exit – the output won't be captured for upload but the
		// important thing is that the user sees the summary and the program exits.
		lockCh := make(chan struct{}, 1)
		go func() {
			outputMutex.Lock()
			select {
			case lockCh <- struct{}{}:
			default:
				// Timed out while we were waiting; release immediately.
				outputMutex.Unlock()
			}
		}()

		var finalOutput string
		select {
		case <-lockCh:
			// Got the lock – capture time info so it goes into the upload too.
			timeInfo := utils.PrintAndCapture(func() {
				printTimeInfo(config, minutes, seconds, currentTime)
			}, "", "")
			*output += timeInfo
			finalOutput = *output
			outputMutex.Unlock()

		case <-time.After(10 * time.Second):
			// A plain os.Exit could orphan a benchmark that still holds the SSH
			// stdout pipe open, so an exhausted cleanup window is a hard stop.
			fmt.Println()
			printTimeInfo(config, minutes, seconds, currentTime)
			forceExit(1)
			return
		}
		// ── Upload and exit ──────────────────────────────────────────────────────
		if config.EnableUpload {
			uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer uploadCancel()
			resultChan := make(chan struct {
				httpURL  string
				httpsURL string
			}, 1)
			go func() {
				httpURL, httpsURL := utils.ProcessAndUpload(finalOutput, config.FilePath, config.EnableUpload, config.Language)
				select {
				case resultChan <- struct {
					httpURL  string
					httpsURL string
				}{httpURL, httpsURL}:
				case <-uploadCtx.Done():
				}
			}()
			select {
			case result := <-resultChan:
				uploadCancel()
				if result.httpURL != "" || result.httpsURL != "" {
					if config.Language == "en" {
						fmt.Printf("Upload successfully!\nHttp URL:  %s\nHttps URL: %s\n", result.httpURL, result.httpsURL)
					} else {
						fmt.Printf("上传成功!\nHttp URL:  %s\nHttps URL: %s\n", result.httpURL, result.httpsURL)
					}
				}
				time.Sleep(100 * time.Millisecond)
			case <-uploadCtx.Done():
				if config.Language == "en" {
					fmt.Println("Upload timeout, program exit")
				} else {
					fmt.Println("上传超时，程序退出")
				}
			}
		}
		if (runtime.GOOS == "windows" || runtime.GOOS == "darwin") && !utils.IsNonInteractive() {
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
		}
		os.Exit(0)
	case <-ctx.Done():
		// Keep signal delivery armed after the soft deadline. Returning while
		// signal.Notify is active would swallow a later timeout SIGTERM.
		<-sig
		forceExit(1)
	}
}

// HandleUploadResults handles uploading results
func HandleUploadResults(config *params.Config, output string) {
	httpURL, httpsURL := utils.ProcessAndUpload(output, config.FilePath, config.EnableUpload, config.Language)
	if httpURL != "" || httpsURL != "" {
		if config.Language == "en" {
			fmt.Printf("Upload successfully!\nHttp URL:  %s\nHttps URL: %s\n", httpURL, httpsURL)
			fmt.Println("Each Test Benchmark: https://bash.spiritlhl.net/ecsguide")
		} else {
			fmt.Printf("上传成功!\nHttp URL:  %s\nHttps URL: %s\n", httpURL, httpsURL)
			fmt.Println("每项测试基准见: https://bash.spiritlhl.net/ecsguide")
		}
	}
}
