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
	"github.com/oneclickvirt/ecs/utils"
	gostunmodel "github.com/oneclickvirt/gostun/model"
	memorytestmodel "github.com/oneclickvirt/memorytest/memory"
	nt3model "github.com/oneclickvirt/nt3/model"
	ptmodel "github.com/oneclickvirt/pingtest/model"
	speedtestmodel "github.com/oneclickvirt/speedtest/model"
)

var (
	ecsVersion = "v0.1.147"                   // 融合怪版本号
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

func runStructuredCLI(preCheck utils.NetCheckResult, config *params.Config) {
	if config == nil {
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	result := ecsapi.RunAllTestsContext(ctx, preCheck, config)
	if result == nil {
		fmt.Fprintln(os.Stderr, "failed to run structured ECS tests")
		return
	}
	if config.JSONPath != "-" && result.StructuredOutput != "" {
		fmt.Print(result.StructuredOutput)
	}
	finalized, err := ecsapi.FinalizeRunResultContext(ctx, preCheck, (*ecsapi.Config)(config), result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to finalize result: %v\n", err)
	}
	if finalized.HTTPURL != "" || finalized.HTTPSURL != "" {
		fmt.Printf("Http URL:  %s\nHttps URL: %s\n", finalized.HTTPURL, finalized.HTTPSURL)
	}
	if config.JSONPath == "-" {
		fmt.Println(string(result.JSON))
	}
}

func main() {
	configs.ParseFlags(os.Args[1:])
	applyEnvironmentDefaults(configs)
	if configs.HandleHelpAndVersion("goecs") {
		return
	}
	initLogger()
	utils.CheckAndFixAndroidDNS(configs.Language)
	preCheck := utils.CheckPublicAccess(3 * time.Second)
	go func() {
		if preCheck.Connected && !configs.PrivacyMode {
			resp, err := http.Get("https://hits.spiritlhl.net/goecs.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
			if err == nil && resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		}
	}()
	if configs.MenuMode {
		menu.HandleMenuMode(preCheck, configs)
	} else {
		configs.OnlyIpInfoCheck = true
	}
	handleLanguageSpecificSettings()
	if !preCheck.Connected {
		configs.EnableUpload = false
	}
	if configs.JSONPath != "" || ecsapi.UsesStructuredComponents() {
		runStructuredCLI(preCheck, configs)
		configs.Finish = true
		if shouldWaitForExitInput() {
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cleanupGrace := 30 * time.Second
	softDeadline := configs.MaxDuration - cleanupGrace
	if softDeadline <= 0 {
		softDeadline = configs.MaxDuration / 2
	}
	deadlineTimer := time.AfterFunc(softDeadline, func() {
		select {
		case sig <- runner.DeadlineSignal{}:
		default:
		}
	})
	defer deadlineTimer.Stop()
	// Keep a stable copy for the deadline callback. The legacy runner mutates
	// a few menu fields while tests are running, so reading configs directly
	// from the signal goroutine would introduce a race.
	deadlineConfig := *configs
	writeDeadlineReport := func() {
		if deadlineConfig.JSONPath == "" {
			return
		}
		reportConfig := deadlineConfig
		reportConfig.BasicStatus = false
		reportConfig.CpuTestStatus = false
		reportConfig.MemoryTestStatus = false
		reportConfig.DiskTestStatus = false
		reportConfig.TCPProbeStatus = false
		reportConfig.DeepMode = false
		reportCtx, reportCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer reportCancel()
		report := ecsapi.CollectStructuredReport(reportCtx, preCheck, &reportConfig, "", startTime, time.Now())
		report.Status = ecsapi.ReportStatusTimeout
		for index := range report.Sections {
			if report.Sections[index].Enabled {
				report.Sections[index].Status = ecsapi.ReportStatusTimeout
				report.Sections[index].Reason = "global deadline exceeded"
			}
		}
		encoded, err := report.JSON()
		if err != nil {
			return
		}
		if reportConfig.JSONPath == "-" {
			fmt.Println(string(encoded))
			return
		}
		_ = os.WriteFile(reportConfig.JSONPath, append(encoded, '\n'), 0o600)
	}
	go runner.HandleSignalInterrupt(ctx, cancel, sig, configs, &startTime, &output, tempOutput, uploadDone, &outputMutex, writeDeadlineReport)
	switch configs.Language {
	case "zh":
		runner.RunChineseTests(ctx, preCheck, configs, &wg1, &wg2, &wg3, &basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo, &output, tempOutput, startTime, &outputMutex, &infoMutex)
	case "en":
		runner.RunEnglishTests(ctx, preCheck, configs, &wg1, &wg2, &wg3, &basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo, &output, tempOutput, startTime, &outputMutex, &infoMutex)
	default:
		fmt.Println("Unsupported language")
	}
	if ctx.Err() == nil {
		if configs.AnalyzeResult {
			output = runner.AppendAnalysisSummary(configs, output, tempOutput, &outputMutex)
		}
		if preCheck.Connected {
			runner.HandleUploadResults(configs, output)
		}
	}
	configs.Finish = true
	if shouldWaitForExitInput() {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
