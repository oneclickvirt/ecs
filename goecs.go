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
	backtrace "github.com/oneclickvirt/backtrace/bk"
	backtracemodel "github.com/oneclickvirt/backtrace/model"
	basicmodel "github.com/oneclickvirt/basics/model"
	cputestmodel "github.com/oneclickvirt/cputest/model"
	disktestmodel "github.com/oneclickvirt/disktest/disk"
	"github.com/oneclickvirt/ecs/cputest"
	"github.com/oneclickvirt/ecs/disktest"
	"github.com/oneclickvirt/ecs/memorytest"
	"github.com/oneclickvirt/ecs/speedtest"
	"github.com/oneclickvirt/ecs/unlocktest"
	"github.com/oneclickvirt/ecs/utils"
	gostunmodel "github.com/oneclickvirt/gostun/model"
	memorytestmodel "github.com/oneclickvirt/memorytest/memory"
	nt3model "github.com/oneclickvirt/nt3/model"
	"github.com/oneclickvirt/nt3/nt"
	ptmodel "github.com/oneclickvirt/pingtest/model"
	"github.com/oneclickvirt/pingtest/pt"
	"github.com/oneclickvirt/portchecker/email"
	speedtestmodel "github.com/oneclickvirt/speedtest/model"
)

var (
	ecsVersion                                                        = "v0.1.38"
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
	autoChangeDiskTestMethod                                          = true
	filePath                                                          = "goecs.txt"
	enabelUpload                                                      = true
	help                                                              bool
	goecsFlag                                                         = flag.NewFlagSet("goecs", flag.ContinueOnError)
	finish                                                            bool
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
				case "1", "2", "3", "4", "5", "6", "7", "8", "9", "10":
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
	goecsFlag.StringVar(&cpuTestMethod, "cpum", "sysbench", "Set CPU test method (supported: sysbench, geekbench, winsat)")
	goecsFlag.StringVar(&cpuTestThreadMode, "cput", "multi", "Set CPU test thread mode (supported: single, multi)")
	goecsFlag.StringVar(&memoryTestMethod, "memorym", "sysbench", "Set memory test method (supported: sysbench, dd, winsat)")
	goecsFlag.StringVar(&diskTestMethod, "diskm", "fio", "Set disk test method (supported: fio, dd, winsat)")
	goecsFlag.StringVar(&diskTestPath, "diskp", "", "Set disk test path, e.g., -diskp /root")
	goecsFlag.BoolVar(&diskMultiCheck, "diskmc", false, "Enable/Disable multiple disk checks, e.g., -diskmc=false")
	goecsFlag.StringVar(&nt3Location, "nt3loc", "GZ", "Specify NT3 test location (supported: GZ, SH, BJ, CD for Guangzhou, Shanghai, Beijing, Chengdu)")
	goecsFlag.StringVar(&nt3CheckType, "nt3t", "ipv4", "Set NT3 test type (supported: both, ipv4, ipv6)")
	goecsFlag.IntVar(&spNum, "spnum", 2, "Set the number of servers per operator for speed test")
	goecsFlag.BoolVar(&enableLogger, "log", false, "Enable/Disable logging in the current path")
	goecsFlag.BoolVar(&enabelUpload, "upload", true, "Enable/Disable upload the result")
	goecsFlag.Parse(os.Args[1:])
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

func handleMenuMode(preCheck utils.NetCheckResult) {
	basicStatus, cpuTestStatus, memoryTestStatus, diskTestStatus = false, false, false, false
	commTestStatus, utTestStatus, securityTestStatus, emailTestStatus = false, false, false, false
	backtraceStatus, nt3Status, speedTestStatus = false, false, false
	autoChangeDiskTestMethod = true
	printMenuOptions()
Loop:
	for {
		choice = getMenuChoice(language)
		switch choice {
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
			setRouteTestStatus()
			break Loop
		default:
			printInvalidChoice()
		}
	}
}

func printMenuOptions() {
	switch language {
	case "zh":
		fmt.Println("VPS融合怪版本: ", ecsVersion)
		fmt.Println("1. 融合怪完全体")
		fmt.Println("2. 极简版(系统信息+CPU+内存+磁盘+测速节点5个)")
		fmt.Println("3. 精简版(系统信息+CPU+内存+磁盘+常用流媒体+路由+测速节点5个)")
		fmt.Println("4. 精简网络版(系统信息+CPU+内存+磁盘+回程+路由+测速节点5个)")
		fmt.Println("5. 精简解锁版(系统信息+CPU+内存+磁盘IO+御三家+常用流媒体+测速节点5个)")
		fmt.Println("6. 网络单项(IP质量检测+三网回程+三网路由与延迟+测速节点11个)")
		fmt.Println("7. 解锁单项(御三家解锁+常用流媒体解锁)")
		fmt.Println("8. 硬件单项(系统信息+CPU+内存+dd磁盘测试+fio磁盘测试)")
		fmt.Println("9. IP质量检测(15个数据库的IP检测+邮件端口检测)")
		fmt.Println("10. 三网回程线路+广州三网路由+全国三网延迟")
	case "en":
		fmt.Println("VPS Fusion Monster Test Version: ", ecsVersion)
		fmt.Println("1. VPS Fusion Monster Test Comprehensive Test Suite")
		fmt.Println("2. Minimal Test Suite (System Info + CPU + Memory + Disk + 5 Speed Test Nodes)")
		fmt.Println("3. Standard Test Suite (System Info + CPU + Memory + Disk + Basic Unlock Tests + 5 Speed Test Nodes)")
		fmt.Println("4. Network-Focused Test Suite (System Info + CPU + Memory + Disk + 5 Speed Test Nodes)")
		fmt.Println("5. Unlock-Focused Test Suite (System Info + CPU + Memory + Disk IO + Basic Unlock Tests + Common Streaming Services + 5 Speed Test Nodes)")
		fmt.Println("6. Network-Only Test (IP Quality Test + 5 Speed Test Nodes)")
		fmt.Println("7. Unlock-Only Test (Basic Unlock Tests + Common Streaming Services Unlock)")
		fmt.Println("8. Hardware-Only Test (System Info + CPU + Memory + dd Disk Test + fio Disk Test)")
		fmt.Println("9. IP Quality Test (IP Test with 15 Databases + Email Port Test)")
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
	securityTestStatus = true
	speedTestStatus = true
	backtraceStatus = true
	nt3Status = true
}

func setUnlockOnlyTestStatus() {
	commTestStatus = true
	utTestStatus = true
	enabelUpload = false
}

func setHardwareOnlyTestStatus(preCheck utils.NetCheckResult) {
	basicStatus = true
	cpuTestStatus = true
	memoryTestStatus = true
	diskTestStatus = true
	if preCheck.Connected {
		securityTestStatus = false
		autoChangeDiskTestMethod = false
	}
}

func setIPQualityTestStatus() {
	securityTestStatus = true
	emailTestStatus = true
}

func setRouteTestStatus() {
	backtraceStatus = true
	nt3Status = true
	pingTestStatus = true
	enabelUpload = false
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

func handleSignalInterrupt(sig chan os.Signal, startTime *time.Time, output *string, tempOutput string, uploadDone chan bool, outputMutex *sync.Mutex) {
	select {
	case <-sig:
		if !finish {
			endTime := time.Now()
			duration := endTime.Sub(*startTime)
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) % 60
			currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
			outputMutex.Lock()
			*output = utils.PrintAndCapture(func() {
				utils.PrintCenteredTitle("", width)
				fmt.Printf("Cost    Time          : %d min %d sec\n", minutes, seconds)
				fmt.Printf("Current Time          : %s\n", currentTime)
				utils.PrintCenteredTitle("", width)
			}, tempOutput, *output)
			outputMutex.Unlock()
			resultChan := make(chan struct {
				httpURL  string
				httpsURL string
			}, 1)
			go func() {
				outputMutex.Lock()
				finalOutput := *output
				outputMutex.Unlock()
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
		}
		os.Exit(0)
	}
}

func runChineseTests(preCheck utils.NetCheckResult, wg1, wg2, wg3 *sync.WaitGroup, basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo *string, output, tempOutput string, startTime time.Time, outputMutex *sync.Mutex) string {
	output = runBasicTests(preCheck, basicInfo, securityInfo, output, tempOutput, outputMutex)
	output = runCPUTest(output, tempOutput, outputMutex)
	output = runMemoryTest(output, tempOutput, outputMutex)
	output = runDiskTest(output, tempOutput, outputMutex)
	if (onlyChinaTest || pingTestStatus) && preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			*ptInfo = pt.PingTest()
		}()
	}
	if emailTestStatus && preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			*emailInfo = email.EmailCheck()
		}()
	}
	if utTestStatus && !onlyChinaTest && preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			*mediaInfo = unlocktest.MediaTest(language)
		}()
	}
	if preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		output = runStreamingTests(wg1, *mediaInfo, output, tempOutput, outputMutex)
		output = runSecurityTests(*securityInfo, output, tempOutput, outputMutex)
		output = runEmailTests(wg2, *emailInfo, output, tempOutput, outputMutex)
	}
	if runtime.GOOS != "windows" && preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		output = runNetworkTests(wg3, *ptInfo, output, tempOutput, outputMutex)
	}
	if preCheck.Connected && preCheck.StackType != "" && preCheck.StackType != "None" {
		output = runSpeedTests(output, tempOutput, outputMutex)
	}
	return appendTimeInfo(output, tempOutput, startTime, outputMutex)
}

func runEnglishTests(preCheck utils.NetCheckResult, wg1, wg2 *sync.WaitGroup, basicInfo, securityInfo, emailInfo, mediaInfo *string, output, tempOutput string, startTime time.Time, outputMutex *sync.Mutex) string {
	output = runBasicTests(preCheck, basicInfo, securityInfo, output, tempOutput, outputMutex)
	output = runCPUTest(output, tempOutput, outputMutex)
	output = runMemoryTest(output, tempOutput, outputMutex)
	output = runDiskTest(output, tempOutput, outputMutex)
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
		output = runStreamingTests(wg1, *mediaInfo, output, tempOutput, outputMutex)
		output = runSecurityTests(*securityInfo, output, tempOutput, outputMutex)
		output = runEmailTests(wg2, *emailInfo, output, tempOutput, outputMutex)
		output = runEnglishSpeedTests(output, tempOutput, outputMutex)
	}
	return appendTimeInfo(output, tempOutput, startTime, outputMutex)
}

func runBasicTests(preCheck utils.NetCheckResult, basicInfo, securityInfo *string, output, tempOutput string, outputMutex *sync.Mutex) string {
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
				*basicInfo, *securityInfo, nt3CheckType = utils.BasicsAndSecurityCheck(language, nt3CheckType, securityTestStatus)
			} else if preCheck.Connected && preCheck.StackType == "IPv4" {
				*basicInfo, *securityInfo, nt3CheckType = utils.BasicsAndSecurityCheck(language, "ipv4", securityTestStatus)
			} else if preCheck.Connected && preCheck.StackType == "IPv6" {
				*basicInfo, *securityInfo, nt3CheckType = utils.BasicsAndSecurityCheck(language, "ipv6", securityTestStatus)
			} else {
				*basicInfo, *securityInfo, nt3CheckType = utils.BasicsAndSecurityCheck(language, "", false)
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
	return utils.PrintAndCapture(func() {
		if cpuTestStatus {
			if language == "zh" {
				utils.PrintCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", cpuTestMethod), width)
			} else {
				utils.PrintCenteredTitle(fmt.Sprintf("CPU-Test--%s-Method", cpuTestMethod), width)
			}
			cputest.CpuTest(language, cpuTestMethod, cpuTestThreadMode)
		}
	}, tempOutput, output)
}

func runMemoryTest(output, tempOutput string, outputMutex *sync.Mutex) string {
	return utils.PrintAndCapture(func() {
		if memoryTestStatus {
			if language == "zh" {
				utils.PrintCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", memoryTestMethod), width)
			} else {
				utils.PrintCenteredTitle(fmt.Sprintf("Memory-Test--%s-Method", memoryTestMethod), width)
			}
			memorytest.MemoryTest(language, memoryTestMethod)
		}
	}, tempOutput, output)
}

func runDiskTest(output, tempOutput string, outputMutex *sync.Mutex) string {
	return utils.PrintAndCapture(func() {
		if diskTestStatus && autoChangeDiskTestMethod {
			if language == "zh" {
				utils.PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", diskTestMethod), width)
			} else {
				utils.PrintCenteredTitle(fmt.Sprintf("Disk-Test--%s-Method", diskTestMethod), width)
			}
			disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck, autoChangeDiskTestMethod)
		} else if diskTestStatus && !autoChangeDiskTestMethod {
			if language == "zh" {
				utils.PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", "dd"), width)
				disktest.DiskTest(language, "dd", diskTestPath, diskMultiCheck, autoChangeDiskTestMethod)
				utils.PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", "fio"), width)
				disktest.DiskTest(language, "fio", diskTestPath, diskMultiCheck, autoChangeDiskTestMethod)
			} else {
				utils.PrintCenteredTitle(fmt.Sprintf("Disk-Test--%s-Method", "dd"), width)
				disktest.DiskTest(language, "dd", diskTestPath, diskMultiCheck, autoChangeDiskTestMethod)
				utils.PrintCenteredTitle(fmt.Sprintf("Disk-Test--%s-Method", "fio"), width)
				disktest.DiskTest(language, "fio", diskTestPath, diskMultiCheck, autoChangeDiskTestMethod)
			}
		}
	}, tempOutput, output)
}

func runStreamingTests(wg1 *sync.WaitGroup, mediaInfo string, output, tempOutput string, outputMutex *sync.Mutex) string {
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
			fmt.Printf("%s", mediaInfo)
		} else {
			wg1.Wait()
		}
	}, tempOutput, output)
}

func runSecurityTests(securityInfo, output, tempOutput string, outputMutex *sync.Mutex) string {
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

func runEmailTests(wg2 *sync.WaitGroup, emailInfo string, output, tempOutput string, outputMutex *sync.Mutex) string {
	return utils.PrintAndCapture(func() {
		if emailTestStatus {
			wg2.Wait()
			if language == "zh" {
				utils.PrintCenteredTitle("邮件端口检测", width)
			} else {
				utils.PrintCenteredTitle("Email-Port-Check", width)
			}
			fmt.Println(emailInfo)
		} else {
			wg2.Wait()
		}
	}, tempOutput, output)
}

func runNetworkTests(wg3 *sync.WaitGroup, ptInfo string, output, tempOutput string, outputMutex *sync.Mutex) string {
	output = utils.PrintAndCapture(func() {
		if backtraceStatus && !onlyChinaTest {
			utils.PrintCenteredTitle("三网回程线路检测", width)
			if strings.Contains(output, "IPV6") {
				backtrace.BackTrace(true)
			} else {
				backtrace.BackTrace(false)
			}
		}
	}, tempOutput, output)
	output = utils.PrintAndCapture(func() {
		if nt3Status && !onlyChinaTest {
			utils.PrintCenteredTitle("三网回程路由检测", width)
			nt.TraceRoute(language, nt3Location, nt3CheckType)
		}
	}, tempOutput, output)
	return utils.PrintAndCapture(func() {
		if onlyChinaTest || pingTestStatus {
			wg3.Wait()
			utils.PrintCenteredTitle("三网ICMP的PING值检测", width)
			fmt.Println(ptInfo)
		} else {
			wg3.Wait()
		}
	}, tempOutput, output)
}

func runSpeedTests(output, tempOutput string, outputMutex *sync.Mutex) string {
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

func runEnglishSpeedTests(output, tempOutput string, outputMutex *sync.Mutex) string {
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
		} else {
			fmt.Printf("上传成功!\nHttp URL:  %s\nHttps URL: %s\n", httpURL, httpsURL)
		}
	}
}

func main() {
	parseFlags()
	if handleHelpAndVersion() {
		return
	}
	initLogger()
	go func() {
		http.Get("https://hits.spiritlhl.net/goecs.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false")
	}()
	preCheck := utils.CheckPublicAccess(3 * time.Second)
	if menuMode {
		handleMenuMode(preCheck)
	}
	handleLanguageSpecificSettings()
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
		output = runChineseTests(preCheck, &wg1, &wg2, &wg3, &basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo, output, tempOutput, startTime, &outputMutex)
	case "en":
		output = runEnglishTests(preCheck, &wg1, &wg2, &basicInfo, &securityInfo, &emailInfo, &mediaInfo, output, tempOutput, startTime, &outputMutex)
	default:
		fmt.Println("Unsupported language")
	}
	handleUploadResults(output)
	finish = true
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
