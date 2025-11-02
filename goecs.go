package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/oneclickvirt/CommonMediaTests/commediatests"
	unlocktestmodel "github.com/oneclickvirt/UnlockTests/model"
	backtracemodel "github.com/oneclickvirt/backtrace/model"
	basicmodel "github.com/oneclickvirt/basics/model"
	cputestmodel "github.com/oneclickvirt/cputest/model"
	disktestmodel "github.com/oneclickvirt/disktest/disk"
	"github.com/oneclickvirt/ecs/cputest"
	"github.com/oneclickvirt/ecs/disktest"
	"github.com/oneclickvirt/ecs/memorytest"
	"github.com/oneclickvirt/ecs/nexttrace"
	"github.com/oneclickvirt/ecs/speedtest"
	"github.com/oneclickvirt/ecs/unlocktest"
	"github.com/oneclickvirt/ecs/upstreams"
	"github.com/oneclickvirt/ecs/utils"
	gostunmodel "github.com/oneclickvirt/gostun/model"
	memorytestmodel "github.com/oneclickvirt/memorytest/memory"
	nt3model "github.com/oneclickvirt/nt3/model"
	ptmodel "github.com/oneclickvirt/pingtest/model"
	"github.com/oneclickvirt/pingtest/pt"
	"github.com/oneclickvirt/portchecker/email"
	speedtestmodel "github.com/oneclickvirt/speedtest/model"
)

var (
	ecsVersion                                                        = "v0.1.95"
	menuMode                                                          bool
	onlyChinaTest                                                     bool
	input, choice                                                     string
	showVersion                                                       bool
	enableLogger                                                      bool
	language                                                          string
	cpuTestMethod, cpuTestThreadMode                                  string
	memoryTestMethod                                                  string
	diskTestMethod, diskTestPath                                      string
	diskMultiCheck                                                    bool
	nt3CheckType, nt3Location                                         string
	spNum                                                             int
	width                                                             = 82
	basicStatus, cpuTestStatus, memoryTestStatus, diskTestStatus      bool
	commTestStatus, utTestStatus, securityTestStatus, emailTestStatus bool
	backtraceStatus, nt3Status, speedTestStatus, pingTestStatus       bool
	tgdcTestStatus, webTestStatus                                     bool
	autoChangeDiskTestMethod                                          = true
	filePath                                                          = "goecs.txt"
	enabelUpload                                                      = true
	onlyIpInfoCheckStatus, help                                       bool
	goecsFlag                                                         = flag.NewFlagSet("goecs", flag.ContinueOnError)
	finish                                                            bool
	// 用于跟踪哪些参数是用户显式设置的
	userSetFlags = make(map[string]bool)
)

func getMenuChoice(language string) string {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	inputChan := make(chan string, 1)
	go func() {
		select {
		case <-sigChan:
			fmt.Println("\n程序在选择过程中被用户中断")
			os.Exit(0)
		case <-ctx.Done():
			return
		}
	}()
	for {
		go func() {
			var input string
			fmt.Print("请输入选项 / Please enter your choice: ")
			fmt.Scanln(&input)
			input = strings.TrimSpace(input)
			input = strings.TrimRight(input, "\n")
			select {
			case inputChan <- input:
			case <-ctx.Done():
				return
			}
		}()
		select {
		case input := <-inputChan:
			re := regexp.MustCompile(`^\d+$`)
			if re.MatchString(input) {
				inChoice := input
				switch inChoice {
				case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10":
					return inChoice
				default:
					if language == "zh" {
						fmt.Println("无效的选项")
					} else {
						fmt.Println("Invalid choice")
					}
				}
			} else {
				if language == "zh" {
					fmt.Println("输入错误，请输入一个纯数字")
				} else {
					fmt.Println("Invalid input, please enter a number")
				}
			}
		case <-ctx.Done():
			return ""
		}
	}
}

func parseFlags() {
	goecsFlag.BoolVar(&help, "h", false, "Show help information")
	goecsFlag.BoolVar(&showVersion, "v", false, "Display version information")
	goecsFlag.BoolVar(&menuMode, "menu", true, "Enable/Disable menu mode, disable example: -menu=false") // true 默认启用菜单栏模式
	goecsFlag.StringVar(&language, "l", "zh", "Set language (supported: en, zh)")
	goecsFlag.BoolVar(&basicStatus, "basic", true, "Enable/Disable basic test")
	goecsFlag.BoolVar(&cpuTestStatus, "cpu", true, "Enable/Disable CPU test")
	goecsFlag.BoolVar(&memoryTestStatus, "memory", true, "Enable/Disable memory test")
	goecsFlag.BoolVar(&diskTestStatus, "disk", true, "Enable/Disable disk test")
	goecsFlag.BoolVar(&commTestStatus, "comm", true, "Enable/Disable common media test")
	goecsFlag.BoolVar(&utTestStatus, "ut", true, "Enable/Disable unlock media test")
	goecsFlag.BoolVar(&securityTestStatus, "security", true, "Enable/Disable security test")
	goecsFlag.BoolVar(&emailTestStatus, "email", true, "Enable/Disable email port test")
	goecsFlag.BoolVar(&backtraceStatus, "backtrace", true, "Enable/Disable backtrace test (in 'en' language or on windows it always false)")
	goecsFlag.BoolVar(&nt3Status, "nt3", true, "Enable/Disable NT3 test (in 'en' language or on windows it always false)")
	goecsFlag.BoolVar(&speedTestStatus, "speed", true, "Enable/Disable speed test")
	goecsFlag.BoolVar(&pingTestStatus, "ping", false, "Enable/Disable ping test")
	goecsFlag.BoolVar(&tgdcTestStatus, "tgdc", false, "Enable/Disable Telegram DC test")
	goecsFlag.BoolVar(&webTestStatus, "web", false, "Enable/Disable popular websites test")
	goecsFlag.StringVar(&cpuTestMethod, "cpum", "sysbench", "Set CPU test method (supported: sysbench, geekbench, winsat)")
	goecsFlag.StringVar(&cpuTestThreadMode, "cput", "multi", "Set CPU test thread mode (supported: single, multi)")
	goecsFlag.StringVar(&memoryTestMethod, "memorym", "stream", "Set memory test method (supported: stream, sysbench, dd, winsat, auto)")
	goecsFlag.StringVar(&diskTestMethod, "diskm", "fio", "Set disk test method (supported: fio, dd, winsat)")
	goecsFlag.StringVar(&diskTestPath, "diskp", "", "Set disk test path, e.g., -diskp /root")
	goecsFlag.BoolVar(&diskMultiCheck, "diskmc", false, "Enable/Disable multiple disk checks, e.g., -diskmc=false")
	goecsFlag.StringVar(&nt3Location, "nt3loc", "GZ", "Specify NT3 test location (supported: GZ, SH, BJ, CD, ALL for Guangzhou, Shanghai, Beijing, Chengdu and all)")
	goecsFlag.StringVar(&nt3CheckType, "nt3t", "ipv4", "Set NT3 test type (supported: both, ipv4, ipv6)")
	goecsFlag.IntVar(&spNum, "spnum", 2, "Set the number of servers per operator for speed test")
	goecsFlag.BoolVar(&enableLogger, "log", false, "Enable/Disable logging in the current path")
	goecsFlag.BoolVar(&enabelUpload, "upload", true, "Enable/Disable upload the result")
	goecsFlag.Parse(os.Args[1:])

	// 记录用户显式设置的参数
	goecsFlag.Visit(func(f *flag.Flag) {
		userSetFlags[f.Name] = true
	})
}

func handleHelpAndVersion() bool {
	if help {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		goecsFlag.PrintDefaults()
		return true
	}
	if showVersion {
		fmt.Println(ecsVersion)
		return true
	}
	return false
}

func initLogger() {
	if enableLogger {
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

// saveUserSetParams 保存用户通过命令行显式设置的参数值
func saveUserSetParams() map[string]interface{} {
	saved := make(map[string]interface{})

	if userSetFlags["basic"] {
		saved["basic"] = basicStatus
	}
	if userSetFlags["cpu"] {
		saved["cpu"] = cpuTestStatus
	}
	if userSetFlags["memory"] {
		saved["memory"] = memoryTestStatus
	}
	if userSetFlags["disk"] {
		saved["disk"] = diskTestStatus
	}
	if userSetFlags["comm"] {
		saved["comm"] = commTestStatus
	}
	if userSetFlags["ut"] {
		saved["ut"] = utTestStatus
	}
	if userSetFlags["security"] {
		saved["security"] = securityTestStatus
	}
	if userSetFlags["email"] {
		saved["email"] = emailTestStatus
	}
	if userSetFlags["backtrace"] {
		saved["backtrace"] = backtraceStatus
	}
	if userSetFlags["nt3"] {
		saved["nt3"] = nt3Status
	}
	if userSetFlags["speed"] {
		saved["speed"] = speedTestStatus
	}
	if userSetFlags["ping"] {
		saved["ping"] = pingTestStatus
	}
	if userSetFlags["tgdc"] {
		saved["tgdc"] = tgdcTestStatus
	}
	if userSetFlags["web"] {
		saved["web"] = webTestStatus
	}
	if userSetFlags["cpum"] {
		saved["cpum"] = cpuTestMethod
	}
	if userSetFlags["cput"] {
		saved["cput"] = cpuTestThreadMode
	}
	if userSetFlags["memorym"] {
		saved["memorym"] = memoryTestMethod
	}
	if userSetFlags["diskm"] {
		saved["diskm"] = diskTestMethod
	}
	if userSetFlags["diskp"] {
		saved["diskp"] = diskTestPath
	}
	if userSetFlags["diskmc"] {
		saved["diskmc"] = diskMultiCheck
	}
	if userSetFlags["nt3loc"] {
		saved["nt3loc"] = nt3Location
	}
	if userSetFlags["nt3t"] {
		saved["nt3t"] = nt3CheckType
	}
	if userSetFlags["spnum"] {
		saved["spnum"] = spNum
	}

	return saved
}

// restoreUserSetParams 恢复用户通过命令行显式设置的参数值，覆盖菜单的默认值
func restoreUserSetParams(saved map[string]interface{}) {
	if val, ok := saved["basic"]; ok {
		basicStatus = val.(bool)
	}
	if val, ok := saved["cpu"]; ok {
		cpuTestStatus = val.(bool)
	}
	if val, ok := saved["memory"]; ok {
		memoryTestStatus = val.(bool)
	}
	if val, ok := saved["disk"]; ok {
		diskTestStatus = val.(bool)
	}
	if val, ok := saved["comm"]; ok {
		commTestStatus = val.(bool)
	}
	if val, ok := saved["ut"]; ok {
		utTestStatus = val.(bool)
	}
	if val, ok := saved["security"]; ok {
		securityTestStatus = val.(bool)
	}
	if val, ok := saved["email"]; ok {
		emailTestStatus = val.(bool)
	}
	if val, ok := saved["backtrace"]; ok {
		backtraceStatus = val.(bool)
	}
	if val, ok := saved["nt3"]; ok {
		nt3Status = val.(bool)
	}
	if val, ok := saved["speed"]; ok {
		speedTestStatus = val.(bool)
	}
	if val, ok := saved["ping"]; ok {
		pingTestStatus = val.(bool)
	}
	if val, ok := saved["tgdc"]; ok {
		tgdcTestStatus = val.(bool)
	}
	if val, ok := saved["web"]; ok {
		webTestStatus = val.(bool)
	}
	if val, ok := saved["cpum"]; ok {
		cpuTestMethod = val.(string)
	}
	if val, ok := saved["cput"]; ok {
		cpuTestThreadMode = val.(string)
	}
	if val, ok := saved["memorym"]; ok {
		memoryTestMethod = val.(string)
	}
	if val, ok := saved["diskm"]; ok {
		diskTestMethod = val.(string)
	}
	if val, ok := saved["diskp"]; ok {
		diskTestPath = val.(string)
	}
	if val, ok := saved["diskmc"]; ok {
		diskMultiCheck = val.(bool)
	}
	if val, ok := saved["nt3loc"]; ok {
		// 如果用户没有在菜单中选择选项10，才恢复用户设置的nt3Location
		// 选项10会强制设置 nt3Location = "ALL"
		if choice != "10" {
			nt3Location = val.(string)
		}
	}
	if val, ok := saved["nt3t"]; ok {
		nt3CheckType = val.(string)
	}
	if val, ok := saved["spnum"]; ok {
		spNum = val.(int)
	}

	// 验证参数的有效性
	validateParams()
}

// validateParams 验证参数的有效性，如果无效则使用默认值
func validateParams() {
	// 验证 cpuTestMethod
	validCpuMethods := map[string]bool{"sysbench": true, "geekbench": true, "winsat": true}
	if !validCpuMethods[cpuTestMethod] {
		if language == "zh" {
			fmt.Printf("警告: CPU测试方法 '%s' 无效，使用默认值 'sysbench'\n", cpuTestMethod)
		} else {
			fmt.Printf("Warning: Invalid CPU test method '%s', using default 'sysbench'\n", cpuTestMethod)
		}
		cpuTestMethod = "sysbench"
	}

	// 验证 cpuTestThreadMode
	validThreadModes := map[string]bool{"single": true, "multi": true}
	if !validThreadModes[cpuTestThreadMode] {
		if language == "zh" {
			fmt.Printf("警告: CPU线程模式 '%s' 无效，使用默认值 'multi'\n", cpuTestThreadMode)
		} else {
			fmt.Printf("Warning: Invalid CPU thread mode '%s', using default 'multi'\n", cpuTestThreadMode)
		}
		cpuTestThreadMode = "multi"
	}

	// 验证 memoryTestMethod
	validMemoryMethods := map[string]bool{"stream": true, "sysbench": true, "dd": true, "winsat": true, "auto": true}
	if !validMemoryMethods[memoryTestMethod] {
		if language == "zh" {
			fmt.Printf("警告: 内存测试方法 '%s' 无效，使用默认值 'stream'\n", memoryTestMethod)
		} else {
			fmt.Printf("Warning: Invalid memory test method '%s', using default 'stream'\n", memoryTestMethod)
		}
		memoryTestMethod = "stream"
	}

	// 验证 diskTestMethod
	validDiskMethods := map[string]bool{"fio": true, "dd": true, "winsat": true}
	if !validDiskMethods[diskTestMethod] {
		if language == "zh" {
			fmt.Printf("警告: 磁盘测试方法 '%s' 无效，使用默认值 'fio'\n", diskTestMethod)
		} else {
			fmt.Printf("Warning: Invalid disk test method '%s', using default 'fio'\n", diskTestMethod)
		}
		diskTestMethod = "fio"
	}

	// 验证 nt3Location
	validNt3Locations := map[string]bool{"GZ": true, "SH": true, "BJ": true, "CD": true, "ALL": true}
	if !validNt3Locations[nt3Location] {
		if language == "zh" {
			fmt.Printf("警告: NT3测试位置 '%s' 无效，使用默认值 'GZ'\n", nt3Location)
		} else {
			fmt.Printf("Warning: Invalid NT3 location '%s', using default 'GZ'\n", nt3Location)
		}
		nt3Location = "GZ"
	}

	// 验证 nt3CheckType
	validNt3Types := map[string]bool{"both": true, "ipv4": true, "ipv6": true}
	if !validNt3Types[nt3CheckType] {
		if language == "zh" {
			fmt.Printf("警告: NT3测试类型 '%s' 无效，使用默认值 'ipv4'\n", nt3CheckType)
		} else {
			fmt.Printf("Warning: Invalid NT3 check type '%s', using default 'ipv4'\n", nt3CheckType)
		}
		nt3CheckType = "ipv4"
	}

	// 验证 spNum (应该是正数)
	if spNum < 0 {
		if language == "zh" {
			fmt.Printf("警告: 测速节点数量 '%d' 无效，使用默认值 2\n", spNum)
		} else {
			fmt.Printf("Warning: Invalid speed test node count '%d', using default 2\n", spNum)
		}
		spNum = 2
	}

	// 验证 language
	validLanguages := map[string]bool{"zh": true, "en": true}
	if !validLanguages[language] {
		fmt.Printf("Warning: Invalid language '%s', using default 'zh'\n", language)
		language = "zh"
	}
}

func handleMenuMode(preCheck utils.NetCheckResult) {
	// 保存用户显式设置的参数值
	savedParams := saveUserSetParams()

	basicStatus, cpuTestStatus, memoryTestStatus, diskTestStatus = false, false, false, false
	commTestStatus, utTestStatus, securityTestStatus, emailTestStatus = false, false, false, false
	backtraceStatus, nt3Status, speedTestStatus = false, false, false
	tgdcTestStatus, webTestStatus = false, false
	autoChangeDiskTestMethod = true
	printMenuOptions(preCheck)
Loop:
	for {
		choice = getMenuChoice(language)
		switch choice {
		case "0":
			os.Exit(0)
		case "1":
			setFullTestStatus(preCheck)
			onlyChinaTest = utils.CheckChina(enableLogger)
			break Loop
		case "2":
			setMinimalTestStatus(preCheck)
			break Loop
		case "3":
			setStandardTestStatus(preCheck)
			break Loop
		case "4":
			setNetworkFocusedTestStatus(preCheck)
			break Loop
		case "5":
			setUnlockFocusedTestStatus(preCheck)
			break Loop
		case "6":
			if !preCheck.Connected {
				fmt.Println("Can not test without network connection!")
				return
			}
			setNetworkOnlyTestStatus()
			break Loop
		case "7":
			if !preCheck.Connected {
				fmt.Println("Can not test without network connection!")
				return
			}
			setUnlockOnlyTestStatus()
			break Loop
		case "8":
			setHardwareOnlyTestStatus(preCheck)
			break Loop
		case "9":
			if !preCheck.Connected {
				fmt.Println("Can not test without network connection!")
				return
			}
			setIPQualityTestStatus()
			break Loop
		case "10":
			if !preCheck.Connected {
				fmt.Println("Can not test without network connection!")
				return
			}
			nt3Location = "ALL"
			setRouteTestStatus()
			break Loop
		default:
			printInvalidChoice()
		}
	}

	// 恢复用户显式设置的参数，覆盖菜单的默认值
	restoreUserSetParams(savedParams)
}

func printMenuOptions(preCheck utils.NetCheckResult) {
	var stats *utils.StatsResponse
	var statsErr error
	var githubInfo *utils.GitHubRelease
	var githubErr error
	// 只有在网络连接正常时才获取统计信息和版本信息
	if preCheck.Connected {
		var pwg sync.WaitGroup
		pwg.Add(2)
		go func() {
			defer pwg.Done()
			stats, statsErr = utils.GetGoescStats()
		}()
		go func() {
			defer pwg.Done()
			githubInfo, githubErr = utils.GetLatestEcsRelease()
		}()
		pwg.Wait()
	} else {
		statsErr = fmt.Errorf("network not connected")
		githubErr = fmt.Errorf("network not connected")
	}
	var statsInfo string
	var cmp int
	if preCheck.Connected {
		// 网络连接正常时处理统计信息和版本比较
		if statsErr != nil {
			statsInfo = "NULL"
		} else {
			switch language {
			case "zh":
				statsInfo = fmt.Sprintf("总使用量: %s | 今日使用: %s",
					utils.FormatGoecsNumber(stats.Total),
					utils.FormatGoecsNumber(stats.Daily))
			case "en":
				statsInfo = fmt.Sprintf("Total Usage: %s | Daily Usage: %s",
					utils.FormatGoecsNumber(stats.Total),
					utils.FormatGoecsNumber(stats.Daily))
			}
		}
		if githubErr == nil {
			cmp = utils.CompareVersions(ecsVersion, githubInfo.TagName)
		} else {
			cmp = 0
		}
	}
	switch language {
	case "zh":
		fmt.Printf("VPS融合怪版本: %s\n", ecsVersion)
		if preCheck.Connected {
			switch cmp {
			case -1:
				fmt.Printf("检测到新版本 %s 如有必要请更新！\n", githubInfo.TagName)
			}
			fmt.Printf("使用统计: %s\n", statsInfo)
		}
		fmt.Println("1. 融合怪完全体(能测全测)")
		fmt.Println("2. 极简版(系统信息+CPU+内存+磁盘+测速节点5个)")
		fmt.Println("3. 精简版(系统信息+CPU+内存+磁盘+常用流媒体+路由+测速节点5个)")
		fmt.Println("4. 精简网络版(系统信息+CPU+内存+磁盘+回程+路由+测速节点5个)")
		fmt.Println("5. 精简解锁版(系统信息+CPU+内存+磁盘IO+御三家+常用流媒体+测速节点5个)")
		fmt.Println("6. 网络单项(IP质量检测+上游及三网回程+广州三网回程详细路由+全国延迟+TGDC+网站延迟+测速节点11个)")
		fmt.Println("7. 解锁单项(御三家解锁+常用流媒体解锁)")
		fmt.Println("8. 硬件单项(系统信息+CPU+dd磁盘测试+fio磁盘测试)")
		fmt.Println("9. IP质量检测(15个数据库的IP质量检测+邮件端口检测)")
		fmt.Println("10. 三网回程线路检测+三网回程详细路由(北京上海广州成都)+全国延迟+TGDC+网站延迟")
		fmt.Println("0. 退出程序")
	case "en":
		fmt.Printf("VPS Fusion Monster Test Version: %s\n", ecsVersion)
		if preCheck.Connected {
			switch cmp {
			case -1:
				fmt.Printf("New version detected %s update if necessary!\n", githubInfo.TagName)
			}
			fmt.Printf("%s\n", statsInfo)
		}
		fmt.Println("1. VPS Fusion Monster Test (Full Test)")
		fmt.Println("2. Minimal Test Suite (System Info + CPU + Memory + Disk + 5 Speed Test Nodes)")
		fmt.Println("3. Standard Test Suite (System Info + CPU + Memory + Disk + Basic Unlock Tests + 5 Speed Test Nodes)")
		fmt.Println("4. Network-Focused Test Suite (System Info + CPU + Memory + Disk + 5 Speed Test Nodes)")
		fmt.Println("5. Unlock-Focused Test Suite (System Info + CPU + Memory + Disk IO + Basic Unlock Tests + Common Streaming Services + 5 Speed Test Nodes)")
		fmt.Println("6. Network-Only Test (IP Quality Test + TGDC + Websites + 11 Speed Test Nodes)")
		fmt.Println("7. Unlock-Only Test (Basic Unlock Tests + Common Streaming Services Unlock)")
		fmt.Println("8. Hardware-Only Test (System Info + CPU + Memory + dd Disk Test + fio Disk Test)")
		fmt.Println("9. IP Quality Test (IP Test with 15 Databases + Email Port Test)")
		fmt.Println("0. Exit Program")
	}
}

func setFullTestStatus(preCheck utils.NetCheckResult) {
	basicStatus = true
	cpuTestStatus = true
	memoryTestStatus = true
	diskTestStatus = true
	if preCheck.Connected {
		commTestStatus = true
		utTestStatus = true
		securityTestStatus = true
		emailTestStatus = true
		backtraceStatus = true
		nt3Status = true
		speedTestStatus = true
		tgdcTestStatus = true
		webTestStatus = true
	}
}

func setMinimalTestStatus(preCheck utils.NetCheckResult) {
	basicStatus = true
	cpuTestStatus = true
	memoryTestStatus = true
	diskTestStatus = true
	if preCheck.Connected {
		speedTestStatus = true
	}
}

func setStandardTestStatus(preCheck utils.NetCheckResult) {
	basicStatus = true
	cpuTestStatus = true
	memoryTestStatus = true
	diskTestStatus = true
	if preCheck.Connected {
		utTestStatus = true
		nt3Status = true
		speedTestStatus = true
	}
}

func setNetworkFocusedTestStatus(preCheck utils.NetCheckResult) {
	basicStatus = true
	cpuTestStatus = true
	memoryTestStatus = true
	diskTestStatus = true
	if preCheck.Connected {
		backtraceStatus = true
		nt3Status = true
		speedTestStatus = true
	}
}

func setUnlockFocusedTestStatus(preCheck utils.NetCheckResult) {
	basicStatus = true
	cpuTestStatus = true
	memoryTestStatus = true
	diskTestStatus = true
	if preCheck.Connected {
		commTestStatus = true
		utTestStatus = true
		speedTestStatus = true
	}
}

func setNetworkOnlyTestStatus() {
	onlyIpInfoCheckStatus = true
	securityTestStatus = true
	speedTestStatus = true
	backtraceStatus = true
	nt3Status = true
	pingTestStatus = true
	tgdcTestStatus = true
	webTestStatus = true
}

func setUnlockOnlyTestStatus() {
	onlyIpInfoCheckStatus = true
	commTestStatus = true
	utTestStatus = true
}

func setHardwareOnlyTestStatus(preCheck utils.NetCheckResult) {
	_ = preCheck
	basicStatus = true
	cpuTestStatus = true
	memoryTestStatus = true
	diskTestStatus = true
	securityTestStatus = false
	autoChangeDiskTestMethod = false
}

func setIPQualityTestStatus() {
	onlyIpInfoCheckStatus = true
	securityTestStatus = true
	emailTestStatus = true
}

func setRouteTestStatus() {
	onlyIpInfoCheckStatus = true
	backtraceStatus = true
	nt3Status = true
	pingTestStatus = true
	tgdcTestStatus = true
	webTestStatus = true
}

func printInvalidChoice() {
	if language == "zh" {
		fmt.Println("无效的选项")
	} else {
		fmt.Println("Invalid choice")
	}
}

func handleLanguageSpecificSettings() {
	if language == "en" {
		backtraceStatus = false
		nt3Status = false
	}
	if !enabelUpload {
		securityTestStatus = false
	}
}

func handleSignalInterrupt(sig chan os.Signal, startTime *time.Time, output *string, _ string, uploadDone chan bool, outputMutex *sync.Mutex) {
	select {
	case <-sig:
		if !finish {
			endTime := time.Now()
			duration := endTime.Sub(*startTime)
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) % 60
			currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
			outputMutex.Lock()
			timeInfo := utils.PrintAndCapture(func() {
				utils.PrintCenteredTitle("", width)
				if language == "zh" {
					fmt.Printf("花费          : %d 分 %d 秒\n", minutes, seconds)
					fmt.Printf("时间          : %s\n", currentTime)
				} else {
					fmt.Printf("Cost    Time          : %d min %d sec\n", minutes, seconds)
					fmt.Printf("Current Time          : %s\n", currentTime)
				}
				utils.PrintCenteredTitle("", width)
			}, "", "")
			*output += timeInfo
			finalOutput := *output
			outputMutex.Unlock()
			resultChan := make(chan struct {
				httpURL  string
				httpsURL string
			}, 1)
			if enabelUpload {
				go func() {
					httpURL, httpsURL := utils.ProcessAndUpload(finalOutput, filePath, enabelUpload)
					resultChan <- struct {
						httpURL  string
						httpsURL string
					}{httpURL, httpsURL}
					uploadDone <- true
				}()
				select {
				case result := <-resultChan:
					if result.httpURL != "" || result.httpsURL != "" {
						if language == "en" {
							fmt.Printf("Upload successfully!\nHttp URL:  %s\nHttps URL: %s\n", result.httpURL, result.httpsURL)
						} else {
							fmt.Printf("上传成功!\nHttp URL:  %s\nHttps URL: %s\n", result.httpURL, result.httpsURL)
						}
					}
					time.Sleep(100 * time.Millisecond)
					if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
						fmt.Println("Press Enter to exit...")
						fmt.Scanln()
					}
					os.Exit(0)
				case <-time.After(30 * time.Second):
					if language == "en" {
						fmt.Println("Upload timeout, program exit")
					} else {
						fmt.Println("上传超时，程序退出")
					}
					if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
						fmt.Println("Press Enter to exit...")
						fmt.Scanln()
					}
					os.Exit(1)
				}
			} else {
				if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
					fmt.Println("Press Enter to exit...")
					fmt.Scanln()
				}
				os.Exit(0)
			}
		}
		os.Exit(0)
	}
}

func runChineseTests(preCheck utils.NetCheckResult, wg1, wg2, wg3 *sync.WaitGroup, basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo *string, output *string, tempOutput string, startTime time.Time, outputMutex *sync.Mutex) {
	*output = runBasicTests(preCheck, basicInfo, securityInfo, *output, tempOutput, outputMutex)
	*output = runCPUTest(*output, tempOutput, outputMutex)
	*output = runMemoryTest(*output, tempOutput, outputMutex)
	*output = runDiskTest(*output, tempOutput, outputMutex)
	if onlyIpInfoCheckStatus && !basicStatus && preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		*output = runIpInfoCheck(*output, tempOutput, outputMutex)
	}
	if utTestStatus && preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" && !onlyChinaTest {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			*mediaInfo = unlocktest.MediaTest(language)
		}()
	}
	if emailTestStatus && preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			*emailInfo = email.EmailCheck()
		}()
	}
	if (onlyChinaTest || pingTestStatus) && preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			*ptInfo = pt.PingTest()
		}()
	}
	if preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		*output = runStreamingTests(wg1, mediaInfo, *output, tempOutput, outputMutex)
		*output = runSecurityTests(*securityInfo, *output, tempOutput, outputMutex)
		*output = runEmailTests(wg2, emailInfo, *output, tempOutput, outputMutex)
	}
	if runtime.GOOS != "windows" && preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		*output = runNetworkTests(wg3, ptInfo, *output, tempOutput, outputMutex)
	}
	if preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		*output = runSpeedTests(*output, tempOutput, outputMutex)
	}
	*output = appendTimeInfo(*output, tempOutput, startTime, outputMutex)
}

func runEnglishTests(preCheck utils.NetCheckResult, wg1, wg2, wg3 *sync.WaitGroup, basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo *string, output *string, tempOutput string, startTime time.Time, outputMutex *sync.Mutex) {
	*output = runBasicTests(preCheck, basicInfo, securityInfo, *output, tempOutput, outputMutex)
	*output = runCPUTest(*output, tempOutput, outputMutex)
	*output = runMemoryTest(*output, tempOutput, outputMutex)
	*output = runDiskTest(*output, tempOutput, outputMutex)
	if onlyIpInfoCheckStatus && !basicStatus && preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		*output = runIpInfoCheck(*output, tempOutput, outputMutex)
	}
	if preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		if utTestStatus {
			wg1.Add(1)
			go func() {
				defer wg1.Done()
				*mediaInfo = unlocktest.MediaTest(language)
			}()
		}
		if emailTestStatus {
			wg2.Add(1)
			go func() {
				defer wg2.Done()
				*emailInfo = email.EmailCheck()
			}()
		}
		// 英文模式不进行三网PING测试,所以不启动 pt.PingTest()
		*output = runStreamingTests(wg1, mediaInfo, *output, tempOutput, outputMutex)
		*output = runSecurityTests(*securityInfo, *output, tempOutput, outputMutex)
		*output = runEmailTests(wg2, emailInfo, *output, tempOutput, outputMutex)
		*output = runEnglishNetworkTests(wg3, ptInfo, *output, tempOutput, outputMutex)
		*output = runEnglishSpeedTests(*output, tempOutput, outputMutex)
	}
	*output = appendTimeInfo(*output, tempOutput, startTime, outputMutex)
}

// runIpInfoCheck 系统和网络基础信息检测不进行测试的时候，该函数检测取得本机IP信息并显示(单项测试中输出)
func runIpInfoCheck(output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		var ipinfo string
		upstreams.IPV4, upstreams.IPV6, ipinfo = utils.OnlyBasicsIpInfo(language)
		if ipinfo != "" {
			if language == "zh" {
				utils.PrintCenteredTitle("IP信息", width)
			} else {
				utils.PrintCenteredTitle("IP-Information", width)
			}
			fmt.Printf("%s", ipinfo)
		}
	}, tempOutput, output)
}

func runBasicTests(preCheck utils.NetCheckResult, basicInfo, securityInfo *string, output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		utils.PrintHead(language, width, ecsVersion)
		if basicStatus || securityTestStatus {
			if basicStatus {
				if language == "zh" {
					utils.PrintCenteredTitle("系统基础信息", width)
				} else {
					utils.PrintCenteredTitle("System-Basic-Information", width)
				}
			}
			if preCheck.Connected && preCheck.StackType == "DualStack" {
				upstreams.IPV4, upstreams.IPV6, *basicInfo, *securityInfo, nt3CheckType = utils.BasicsAndSecurityCheck(language, nt3CheckType, securityTestStatus)
			} else if preCheck.Connected && preCheck.StackType == "IPv4" {
				upstreams.IPV4, upstreams.IPV6, *basicInfo, *securityInfo, nt3CheckType = utils.BasicsAndSecurityCheck(language, "ipv4", securityTestStatus)
			} else if preCheck.Connected && preCheck.StackType == "IPv6" {
				upstreams.IPV4, upstreams.IPV6, *basicInfo, *securityInfo, nt3CheckType = utils.BasicsAndSecurityCheck(language, "ipv6", securityTestStatus)
			} else {
				upstreams.IPV4, upstreams.IPV6, *basicInfo, *securityInfo, nt3CheckType = utils.BasicsAndSecurityCheck(language, "", false)
				securityTestStatus = false
			}
			if basicStatus {
				fmt.Printf("%s", *basicInfo)
			} else if (input == "6" || input == "9") && securityTestStatus {
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

func runCPUTest(output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if cpuTestStatus {
			realTestMethod, res := cputest.CpuTest(language, cpuTestMethod, cpuTestThreadMode)
			if language == "zh" {
				utils.PrintCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", realTestMethod), width)
			} else {
				utils.PrintCenteredTitle(fmt.Sprintf("CPU-Test--%s-Method", realTestMethod), width)
			}
			fmt.Print(res)
		}
	}, tempOutput, output)
}

func runMemoryTest(output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if memoryTestStatus {
			realTestMethod, res := memorytest.MemoryTest(language, memoryTestMethod)
			if language == "zh" {
				utils.PrintCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", realTestMethod), width)
			} else {
				utils.PrintCenteredTitle(fmt.Sprintf("Memory-Test--%s-Method", realTestMethod), width)
			}
			fmt.Print(res)
		}
	}, tempOutput, output)
}

func runDiskTest(output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if diskTestStatus && autoChangeDiskTestMethod {
			realTestMethod, res := disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck, autoChangeDiskTestMethod)
			if language == "zh" {
				utils.PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", realTestMethod), width)
			} else {
				utils.PrintCenteredTitle(fmt.Sprintf("Disk-Test--%s-Method", realTestMethod), width)
			}
			fmt.Print(res)
		} else if diskTestStatus && !autoChangeDiskTestMethod {
			if language == "zh" {
				utils.PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", "dd"), width)
				_, res := disktest.DiskTest(language, "dd", diskTestPath, diskMultiCheck, autoChangeDiskTestMethod)
				fmt.Print(res)
				utils.PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", "fio"), width)
				_, res = disktest.DiskTest(language, "fio", diskTestPath, diskMultiCheck, autoChangeDiskTestMethod)
				fmt.Print(res)
			} else {
				utils.PrintCenteredTitle(fmt.Sprintf("Disk-Test--%s-Method", "dd"), width)
				_, res := disktest.DiskTest(language, "dd", diskTestPath, diskMultiCheck, autoChangeDiskTestMethod)
				fmt.Print(res)
				utils.PrintCenteredTitle(fmt.Sprintf("Disk-Test--%s-Method", "fio"), width)
				_, res = disktest.DiskTest(language, "fio", diskTestPath, diskMultiCheck, autoChangeDiskTestMethod)
				fmt.Print(res)
			}
		}
	}, tempOutput, output)
}

func runStreamingTests(wg1 *sync.WaitGroup, mediaInfo *string, output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if language == "zh" {
			if commTestStatus && !onlyChinaTest {
				utils.PrintCenteredTitle("御三家流媒体解锁", width)
				fmt.Printf("%s", commediatests.MediaTests(language))
			}
		}
		if utTestStatus && (language == "zh" && !onlyChinaTest || language == "en") {
			wg1.Wait()
			if language == "zh" {
				utils.PrintCenteredTitle("跨国流媒体解锁", width)
			} else {
				utils.PrintCenteredTitle("Cross-Border-Streaming-Media-Unlock", width)
			}
			fmt.Printf("%s", *mediaInfo)
		}
	}, tempOutput, output)
}

func runSecurityTests(securityInfo, output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if securityTestStatus {
			if language == "zh" {
				utils.PrintCenteredTitle("IP质量检测", width)
			} else {
				utils.PrintCenteredTitle("IP-Quality-Check", width)
			}
			fmt.Printf("%s", securityInfo)
		}
	}, tempOutput, output)
}

func runEmailTests(wg2 *sync.WaitGroup, emailInfo *string, output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if emailTestStatus {
			wg2.Wait()
			if language == "zh" {
				utils.PrintCenteredTitle("邮件端口检测", width)
			} else {
				utils.PrintCenteredTitle("Email-Port-Check", width)
			}
			fmt.Println(*emailInfo)
		}
	}, tempOutput, output)
}

func runNetworkTests(wg3 *sync.WaitGroup, ptInfo *string, output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if backtraceStatus && !onlyChinaTest {
			utils.PrintCenteredTitle("上游及回程线路检测", width)
			upstreams.UpstreamsCheck() // 不能在重定向的同时外部并发，此处仅可以顺序执行
		}
		if nt3Status && !onlyChinaTest {
			utils.PrintCenteredTitle("三网回程路由检测", width)
			nexttrace.NextTrace3Check(language, nt3Location, nt3CheckType) // 不能在重定向的同时外部并发，此处仅可以顺序执行
		}
		// 中国模式：显示三网 PING 测试
		if onlyChinaTest && *ptInfo != "" {
			wg3.Wait()
			utils.PrintCenteredTitle("PING值检测", width)
			fmt.Println(*ptInfo)
		}
		// 选项 6/10：显示三网 PING + TGDC + 网站
		if pingTestStatus && *ptInfo != "" {
			wg3.Wait()
			utils.PrintCenteredTitle("PING值检测", width)
			fmt.Println(*ptInfo)
			if tgdcTestStatus {
				fmt.Println(pt.TelegramDCTest())
			}
			if webTestStatus {
				fmt.Println(pt.WebsiteTest())
			}
		}
		// 非中国模式且非 pingTestStatus：只显示 TGDC + 网站（选项 1 的情况）
		if !onlyChinaTest && !pingTestStatus && (tgdcTestStatus || webTestStatus) {
			utils.PrintCenteredTitle("PING值检测", width)
			if tgdcTestStatus {
				fmt.Println(pt.TelegramDCTest())
			}
			if webTestStatus {
				fmt.Println(pt.WebsiteTest())
			}
		}
	}, tempOutput, output)
}

func runSpeedTests(output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if speedTestStatus {
			utils.PrintCenteredTitle("就近节点测速", width)
			speedtest.ShowHead(language)
			if choice == "1" || !menuMode {
				speedtest.NearbySP()
				speedtest.CustomSP("net", "global", 2, language)
				speedtest.CustomSP("net", "cu", spNum, language)
				speedtest.CustomSP("net", "ct", spNum, language)
				speedtest.CustomSP("net", "cmcc", spNum, language)
			} else if choice == "2" || choice == "3" || choice == "4" || choice == "5" {
				speedtest.CustomSP("net", "global", 4, language)
			} else if choice == "6" {
				speedtest.CustomSP("net", "global", 11, language)
			}
		}
	}, tempOutput, output)
}

func runEnglishNetworkTests(wg3 *sync.WaitGroup, ptInfo *string, output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		// 英文模式只测试 TGDC 和主流网站，不测试三网PING
		if tgdcTestStatus || webTestStatus {
			utils.PrintCenteredTitle("PING-Test", width)
			if tgdcTestStatus {
				fmt.Println(pt.TelegramDCTest())
			}
			if webTestStatus {
				fmt.Println(pt.WebsiteTest())
			}
		}
	}, tempOutput, output)
}

func runEnglishSpeedTests(output, tempOutput string, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	return utils.PrintAndCapture(func() {
		if speedTestStatus {
			utils.PrintCenteredTitle("Speed-Test", width)
			speedtest.ShowHead(language)
			speedtest.NearbySP()
			speedtest.CustomSP("net", "global", -1, language)
		}
	}, tempOutput, output)
}

func appendTimeInfo(output, tempOutput string, startTime time.Time, outputMutex *sync.Mutex) string {
	outputMutex.Lock()
	defer outputMutex.Unlock()
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
	return utils.PrintAndCapture(func() {
		utils.PrintCenteredTitle("", width)
		if language == "zh" {
			fmt.Printf("花费          : %d 分 %d 秒\n", minutes, seconds)
			fmt.Printf("时间          : %s\n", currentTime)
		} else {
			fmt.Printf("Cost    Time          : %d min %d sec\n", minutes, seconds)
			fmt.Printf("Current Time          : %s\n", currentTime)
		}
		utils.PrintCenteredTitle("", width)
	}, tempOutput, output)
}

func handleUploadResults(output string) {
	httpURL, httpsURL := utils.ProcessAndUpload(output, filePath, enabelUpload)
	if httpURL != "" || httpsURL != "" {
		if language == "en" {
			fmt.Printf("Upload successfully!\nHttp URL:  %s\nHttps URL: %s\n", httpURL, httpsURL)
			fmt.Println("Each Test Benchmark: https://bash.spiritlhl.net/ecsguide")
		} else {
			fmt.Printf("上传成功!\nHttp URL:  %s\nHttps URL: %s\n", httpURL, httpsURL)
			fmt.Println("每项测试基准见: https://bash.spiritlhl.net/ecsguide")
		}
	}
}

func main() {
	parseFlags()
	if handleHelpAndVersion() {
		return
	}
	initLogger()
	preCheck := utils.CheckPublicAccess(3 * time.Second)
	go func() {
		if preCheck.Connected {
			http.Get("https://hits.spiritlhl.net/goecs.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
		}
	}()
	if menuMode {
		handleMenuMode(preCheck)
	} else {
		onlyIpInfoCheckStatus = true
	}
	handleLanguageSpecificSettings()
	if !preCheck.Connected {
		enabelUpload = false
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
	go handleSignalInterrupt(sig, &startTime, &output, tempOutput, uploadDone, &outputMutex)
	switch language {
	case "zh":
		runChineseTests(preCheck, &wg1, &wg2, &wg3, &basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo, &output, tempOutput, startTime, &outputMutex)
	case "en":
		runEnglishTests(preCheck, &wg1, &wg2, &wg3, &basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo, &output, tempOutput, startTime, &outputMutex)
	default:
		fmt.Println("Unsupported language")
	}
	if preCheck.Connected {
		handleUploadResults(output)
	}
	finish = true
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
