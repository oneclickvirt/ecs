package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	unlocktestmodel "github.com/oneclickvirt/UnlockTests/model"
	backtracemodel "github.com/oneclickvirt/backtrace/model"
	basicmodel "github.com/oneclickvirt/basics/model"
	cputestmodel "github.com/oneclickvirt/cputest/model"
	disktestmodel "github.com/oneclickvirt/disktest/disk"
	ecsapi "github.com/oneclickvirt/ecs/api"
	menu "github.com/oneclickvirt/ecs/internal/menu"
	params "github.com/oneclickvirt/ecs/internal/params"
	"github.com/oneclickvirt/ecs/internal/runner"
	"github.com/oneclickvirt/ecs/internal/updater"
	"github.com/oneclickvirt/ecs/utils"
	gostunmodel "github.com/oneclickvirt/gostun/model"
	memorytestmodel "github.com/oneclickvirt/memorytest/memory"
	nt3model "github.com/oneclickvirt/nt3/model"
	ptmodel "github.com/oneclickvirt/pingtest/model"
	speedtestmodel "github.com/oneclickvirt/speedtest/model"
)

var (
	ecsVersion = ecsapi.DefaultVersion        // 融合怪版本号
	configs    = params.NewConfig(ecsVersion) // 全局配置实例
)

func initLogger() {
	if configs.EnableLogger {
		gostunmodel.EnableLoger = true
		basicmodel.EnableLoger = true
		cputestmodel.EnableLoger = true
		memorytestmodel.EnableLoger = true
		disktestmodel.EnableLoger = true
		unlocktestmodel.EnableLoger = true
		ptmodel.EnableLoger = true
		backtracemodel.EnableLoger = true
		nt3model.EnableLoger = true
		speedtestmodel.EnableLoger = true
	}
}

func sanitizeECSLog() {
	if configs.EnableLogger {
		_ = ecsapi.SanitizeLogFile("ecs.log")
	}
}

func handleLanguageSpecificSettings() {
	if configs.Language == "en" {
		configs.BacktraceStatus = false
		configs.Nt3Status = false
	}
}

func applyEnvironmentDefaults(config *params.Config) {
	// noninteractive env var only affects blocking prompts (Press Enter to exit, etc.)
	// Menu mode should only be disabled via explicit CLI flag -menu=false,
	// not by the noninteractive env var which may leak from install scripts.
}

func shouldWaitForExitInput() bool {
	return (runtime.GOOS == "windows" || runtime.GOOS == "darwin") && !utils.IsNonInteractive()
}

func shouldRunStructuredCLI(config *params.Config) bool {
	return config != nil && config.JSONPath != ""
}

func legacyDeadlineWindows(maxDuration time.Duration) (time.Duration, time.Duration) {
	if maxDuration <= 0 {
		return 0, 0
	}
	cleanupGrace := min(30*time.Second, maxDuration/5)
	softDeadline := maxDuration - cleanupGrace
	if softDeadline <= 0 {
		softDeadline = maxDuration
	}
	return softDeadline, maxDuration
}

func runStructuredCLI(preCheck utils.NetCheckResult, config *params.Config) {
	if config == nil {
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	resultCh := make(chan *ecsapi.RunResult, 1)
	go func() {
		// JSON on stdout remains a machine-only mode. JSON written to a file,
		// however, must not replace the classic real-time terminal workflow or
		// rerun benchmarks merely to populate a structured envelope.
		if config.JSONPath == "-" {
			resultCh <- ecsapi.RunAllTestsContext(runCtx, preCheck, config)
			return
		}
		var (
			wg1, wg2, wg3                                         sync.WaitGroup
			basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo string
			output, tempOutput                                    string
			outputMutex, infoMutex                                sync.Mutex
		)
		startedAt := time.Now()
		switch config.Language {
		case "en":
			runner.RunEnglishTests(runCtx, preCheck, config, &wg1, &wg2, &wg3,
				&basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo,
				&output, tempOutput, startedAt, &outputMutex, &infoMutex)
		default:
			runner.RunChineseTests(runCtx, preCheck, config, &wg1, &wg2, &wg3,
				&basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo,
				&output, tempOutput, startedAt, &outputMutex, &infoMutex)
		}
		if runCtx.Err() == nil && config.AnalyzeResult {
			output = runner.AppendAnalysisSummary(config, output, tempOutput, &outputMutex)
		}
		resultCh <- ecsapi.NewTextRunResult(runCtx, preCheck, config, output, startedAt, time.Now())
	}()
	var softTimer, hardTimer *time.Timer
	if softDeadline, hardDeadline := legacyDeadlineWindows(config.MaxDuration); hardDeadline > 0 {
		softTimer = time.AfterFunc(softDeadline, cancel)
		hardTimer = time.NewTimer(hardDeadline)
		defer softTimer.Stop()
		defer hardTimer.Stop()
	}
	var result *ecsapi.RunResult
	if hardTimer == nil {
		result = <-resultCh
	} else {
		select {
		case result = <-resultCh:
		case <-hardTimer.C:
			fmt.Fprintln(os.Stderr, "global structured deadline exceeded; terminating benchmark process group")
			sanitizeECSLog()
			runner.ForceExit(1)
			return
		}
	}
	if result == nil {
		fmt.Fprintln(os.Stderr, "failed to run structured ECS tests")
		return
	}
	if config.JSONPath != "-" && result.StructuredOutput != "" {
		fmt.Print(result.StructuredOutput)
	}
	finalized, err := ecsapi.FinalizeRunResultContext(ctx, preCheck, (*ecsapi.Config)(config), result)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to finalize result")
	}
	if finalized.HTTPSURL != "" {
		fmt.Printf("Share URL: %s\n", finalized.HTTPSURL)
	}
	if config.JSONPath == "-" {
		fmt.Println(string(result.JSON))
	}
}

func runSelfUpdate(language string) {
	if language == "en" {
		fmt.Println("Checking the official GoECS release for an update...")
	} else {
		fmt.Println("正在检查官方 GoECS Release 更新...")
	}
	result, err := updater.Update(context.Background(), updater.Options{CurrentVersion: configs.EcsVersion})
	if err != nil {
		if language == "en" {
			fmt.Fprintf(os.Stderr, "GoECS update failed: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "GoECS 自动升级失败: %v\n", err)
		}
		return
	}
	switch result.Status {
	case updater.StatusUpToDate:
		if language == "en" {
			fmt.Printf("GoECS is already up to date (%s).\n", result.CurrentVersion)
		} else {
			fmt.Printf("GoECS 已是最新版本（%s）。\n", result.CurrentVersion)
		}
	case updater.StatusUpdated:
		if result.Deferred {
			if language == "en" {
				fmt.Printf("GoECS %s has been verified and will replace the current executable after it exits.\n", result.LatestVersion)
			} else {
				fmt.Printf("GoECS %s 已校验完成，程序退出后将替换当前可执行文件。\n", result.LatestVersion)
			}
			return
		}
		if language == "en" {
			fmt.Printf("GoECS updated to %s. Run it again to use the new version.\n", result.LatestVersion)
		} else {
			fmt.Printf("GoECS 已更新至 %s，请重新运行程序。\n", result.LatestVersion)
		}
	}
}

func main() {
	if handled, err := updater.HandleHelper(os.Args[1:]); handled {
		if err != nil {
			fmt.Fprintln(os.Stderr, "GoECS update helper failed:", err)
		}
		return
	}
	runner.IsolateProcessGroup()
	configs.ParseFlags(os.Args[1:])
	applyEnvironmentDefaults(configs)
	if configs.HandleHelpAndVersion("goecs") {
		return
	}
	initLogger()
	runner.SetExitCleanup(sanitizeECSLog)
	defer runner.SetExitCleanup(nil)
	runner.SetOutputSanitizer(ecsapi.SanitizeOutput)
	defer runner.SetOutputSanitizer(nil)
	defer sanitizeECSLog()
	// The legacy Android resolver-file repair is retained only for callers that
	// explicitly require the system resolver. Auto/DoH/DoT runs stay process-local.
	if configs.DNSMode == "system" {
		utils.CheckAndFixAndroidDNS(configs.Language)
	}
	preCheck := utils.CheckPublicAccess(3 * time.Second)
	utils.ConfigureDNS(context.Background(), configs.DNSMode, &preCheck)
	defer utils.ShutdownDNS()
	if configs.MenuMode {
		configuredDNSMode := configs.DNSMode
		menu.HandleMenuMode(preCheck, configs)
		if configs.DNSMode != configuredDNSMode {
			utils.ConfigureDNS(context.Background(), configs.DNSMode, &preCheck)
			if status := utils.DNSStatusText(preCheck, configs.Language); status != "" {
				fmt.Println(status)
			}
		}
	} else {
		if status := utils.DNSStatusText(preCheck, configs.Language); status != "" {
			fmt.Println(status)
		}
		configs.OnlyIpInfoCheck = true
	}
	if configs.Choice == "12" {
		runSelfUpdate(configs.Language)
		return
	}
	handleLanguageSpecificSettings()
	if !preCheck.Connected {
		configs.EnableUpload = false
	}
	if preCheck.Connected && !configs.PrivacyMode {
		go func() {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get("https://hits.spiritlhl.net/goecs.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
			if err == nil && resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		}()
	}
	// Keep the established interactive/text runner as the default user-facing
	// path. Structured orchestration is an explicit JSON/API mode; it must not
	// replace the legacy streaming sections merely because structured adapters
	// are compiled into this build.
	if shouldRunStructuredCLI(configs) {
		runStructuredCLI(preCheck, configs)
		configs.Finish = true
		if shouldWaitForExitInput() && configs.JSONPath != "-" {
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
		}
		return
	}
	var (
		wg1, wg2, wg3                                         sync.WaitGroup
		basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo string
		output, tempOutput                                    string
		outputMutex                                           sync.Mutex
		infoMutex                                             sync.Mutex // 保护并发字符串写入
	)
	startTime := time.Now()
	uploadDone := make(chan bool, 1)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if softDeadline, hardDeadline := legacyDeadlineWindows(configs.MaxDuration); hardDeadline > 0 {
		softDeadlineTimer := time.AfterFunc(softDeadline, cancel)
		hardDeadlineTimer := time.AfterFunc(hardDeadline, func() {
			sanitizeECSLog()
			runner.ForceExit(1)
		})
		defer softDeadlineTimer.Stop()
		defer hardDeadlineTimer.Stop()
	}
	go runner.HandleSignalInterrupt(ctx, cancel, sig, configs, &startTime, &output, tempOutput, uploadDone, &outputMutex)
	switch configs.Language {
	case "zh":
		runner.RunChineseTests(ctx, preCheck, configs, &wg1, &wg2, &wg3, &basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo, &output, tempOutput, startTime, &outputMutex, &infoMutex)
	case "en":
		runner.RunEnglishTests(ctx, preCheck, configs, &wg1, &wg2, &wg3, &basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo, &output, tempOutput, startTime, &outputMutex, &infoMutex)
	default:
		fmt.Println("Unsupported language")
	}
	if ctx.Err() == nil && configs.AnalyzeResult {
		output = runner.AppendAnalysisSummary(configs, output, tempOutput, &outputMutex)
	}
	output = ecsapi.SanitizeOutput(output)
	// HandleUploadResults always writes the local result file. Keep that
	// behavior after a deadline/cancellation; EnableUpload alone controls the
	// optional remote share.
	if preCheck.Connected || output != "" {
		runner.HandleUploadResults(configs, output)
	}
	sanitizeECSLog()
	configs.Finish = true
	if shouldWaitForExitInput() {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
