package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/oneclickvirt/CommonMediaTests/commediatests"
	unlocktestmodel "github.com/oneclickvirt/UnlockTests/model"
	backtracemodel "github.com/oneclickvirt/backtrace/model"
	basicmodel "github.com/oneclickvirt/basics/model"
	cputestmodel "github.com/oneclickvirt/cputest/model"
	disktestmodel "github.com/oneclickvirt/disktest/disk"
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
	ecsVersion   = "v0.1.96"
	configs      = params.NewConfig(ecsVersion) // 全局配置实例
	userSetFlags = make(map[string]bool)        // 用于跟踪哪些参数是用户显式设置的
)

func initLogger() {
	if configs.EnableLogger {
		gostunmodel.EnableLoger = true
		basicmodel.EnableLoger = true
		cputestmodel.EnableLoger = true
		memorytestmodel.EnableLoger = true
		disktestmodel.EnableLoger = true
		commediatests.EnableLoger = true
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
	if !configs.EnableUpload {
		configs.SecurityTestStatus = false
	}
}

func main() {
	configs.ParseFlags(os.Args[1:])
	if configs.HandleHelpAndVersion("goecs") {
		return
	}
	initLogger()
	preCheck := utils.CheckPublicAccess(3 * time.Second)
	go func() {
		if preCheck.Connected {
			http.Get("https://hits.spiritlhl.net/goecs.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
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
	var (
		wg1, wg2, wg3                                         sync.WaitGroup
		basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo string
		output, tempOutput                                    string
		outputMutex                                           sync.Mutex
	)
	startTime := time.Now()
	uploadDone := make(chan bool, 1)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go runner.HandleSignalInterrupt(sig, configs, &startTime, &output, tempOutput, uploadDone, &outputMutex)
	switch configs.Language {
	case "zh":
		runner.RunChineseTests(preCheck, configs, &wg1, &wg2, &wg3, &basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo, &output, tempOutput, startTime, &outputMutex)
	case "en":
		runner.RunEnglishTests(preCheck, configs, &wg1, &wg2, &wg3, &basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo, &output, tempOutput, startTime, &outputMutex)
	default:
		fmt.Println("Unsupported language")
	}
	if preCheck.Connected {
		runner.HandleUploadResults(configs, output)
	}
	configs.Finish = true
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
