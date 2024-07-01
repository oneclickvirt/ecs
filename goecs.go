package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/oneclickvirt/CommonMediaTests/commediatests"
	backtraceori "github.com/oneclickvirt/backtrace/bk"
	basicmodel "github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/ecs/backtrace"
	"github.com/oneclickvirt/ecs/commediatest"
	"github.com/oneclickvirt/ecs/cputest"
	"github.com/oneclickvirt/ecs/disktest"
	"github.com/oneclickvirt/ecs/memorytest"
	"github.com/oneclickvirt/ecs/ntrace"
	"github.com/oneclickvirt/ecs/speedtest"
	"github.com/oneclickvirt/ecs/unlocktest"
	"github.com/oneclickvirt/ecs/utils"
	gostunmodel "github.com/oneclickvirt/gostun/model"
	"github.com/oneclickvirt/portchecker/email"
	speedtestmodel "github.com/oneclickvirt/speedtest/model"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	ecsVersion                                                        = "2024.07.01.1"
	menuMode                                                          bool
	choice                                                            string
	showVersion                                                       bool
	enableLogger                                                      bool
	language                                                          string
	cpuTestMethod, cpuTestThreadMode                                  string
	memoryTestMethod                                                  string
	diskTestMethod, diskTestPath                                      string
	diskMultiCheck                                                    bool
	nt3CheckType, nt3Location                                         string
	spNum                                                             int
	width                                                             = 84
	basicStatus, cpuTestStatus, memoryTestStatus, diskTestStatus      bool
	commTestStatus, utTestStatus, securityTestStatus, emailTestStatus bool
	backtraceStatus, nt3Status, speedTestStatus                       bool
)

func main() {
	flag.BoolVar(&showVersion, "v", false, "Display version information")
	flag.BoolVar(&menuMode, "menu", true, "Enable/Disable menu mode, disable example: -menu=false") // true 默认启用菜单栏模式
	flag.StringVar(&language, "l", "zh", "Set language (supported: en, zh)")
	flag.BoolVar(&basicStatus, "basic", true, "Enable/Disable basic test")
	flag.BoolVar(&cpuTestStatus, "cpu", true, "Enable/Disable CPU test")
	flag.BoolVar(&memoryTestStatus, "memory", true, "Enable/Disable memory test")
	flag.BoolVar(&diskTestStatus, "disk", true, "Enable/Disable disk test")
	flag.BoolVar(&commTestStatus, "comm", true, "Enable/Disable common media test")
	flag.BoolVar(&utTestStatus, "ut", true, "Enable/Disable unlock media test")
	flag.BoolVar(&securityTestStatus, "security", true, "Enable/Disable security test")
	flag.BoolVar(&emailTestStatus, "email", true, "Enable/Disable email port test")
	flag.BoolVar(&backtraceStatus, "backtrace", true, "Enable/Disable backtrace test (in 'en' language or on `windows` it always false)")
	flag.BoolVar(&nt3Status, "nt3", true, "Enable/Disable NT3 test (in 'en' language or on `windows` it always false)")
	flag.BoolVar(&speedTestStatus, "speed", true, "Enable/Disable speed test")
	flag.StringVar(&cpuTestMethod, "cpum", "sysbench", "Set CPU test method (supported: sysbench, geekbench, winsat)")
	flag.StringVar(&cpuTestThreadMode, "cput", "multi", "Set CPU test thread mode (supported: single, multi)")
	flag.StringVar(&memoryTestMethod, "memorym", "dd", "Set memory test method (supported: sysbench, dd, winsat)")
	flag.StringVar(&diskTestMethod, "diskm", "fio", "Set disk test method (supported: fio, dd, winsat)")
	flag.StringVar(&diskTestPath, "diskp", "", "Set disk test path, e.g., -diskp /root")
	flag.BoolVar(&diskMultiCheck, "diskmc", false, "Enable/Disable multiple disk checks, e.g., -diskmc=false")
	flag.StringVar(&nt3Location, "nt3loc", "GZ", "Specify NT3 test location (supported: GZ, SH, BJ, CD for Guangzhou, Shanghai, Beijing, Chengdu)")
	flag.StringVar(&nt3CheckType, "nt3t", "ipv4", "Set NT3 test type (supported: both, ipv4, ipv6)")
	flag.IntVar(&spNum, "spnum", 2, "Set the number of servers per operator for speed test")
	flag.BoolVar(&enableLogger, "log", false, "Enable/Disable logging in the current path")
	flag.Parse()
	if showVersion {
		fmt.Println(ecsVersion)
		return
	}
	if enableLogger {
		basicmodel.EnableLoger = true
		speedtestmodel.EnableLoger = true
		gostunmodel.EnableLoger = true
		commediatests.EnableLoger = true
		backtraceori.EnableLoger = true
	}
	if menuMode {
		basicStatus, cpuTestStatus, memoryTestStatus, diskTestStatus = false, false, false, false
		commTestStatus, utTestStatus, securityTestStatus, emailTestStatus = false, false, false, false
		backtraceStatus, nt3Status, speedTestStatus = false, false, false
		// 正则表达式匹配纯数字
		re := regexp.MustCompile(`^\d+$`)
		reader := bufio.NewReader(os.Stdin)
		switch language {
		case "zh":
			fmt.Println("1. 融合怪完全体")
			fmt.Println("2. 极简版(系统信息+CPU+内存+磁盘+测速节点5个)")
			fmt.Println("3. 精简版(系统信息+CPU+内存+磁盘+御三家+常用流媒体+回程+路由+测速节点5个)")
			fmt.Println("4. 精简网络版(系统信息+CPU+内存+磁盘+回程+路由+测速节点5个)")
			fmt.Println("5. 精简解锁版(系统信息+CPU+内存+磁盘IO+御三家+常用流媒体+测速节点5个)")
			fmt.Println("6. 网络单项(IP质量检测+三网回程+三网路由与延迟+测速节点11个)")
			fmt.Println("7. 解锁单项(御三家解锁+常用流媒体解锁)")
			fmt.Println("8. 硬件单项(基础系统信息+CPU+内存+dd磁盘测试+fio磁盘测试)")
			fmt.Println("9. IP质量检测(15个数据库的IP检测+邮件端口检测)")
			fmt.Println("10. 三网回程线路+广州三网路由+全国三网延迟")
		case "en":
			fmt.Println("1. VPS Fusion Monster Test Comprehensive Test Suite")
			fmt.Println("2. Minimal Test Suite (System Info + CPU + Memory + Disk + 5 Speed Test Nodes)")
			fmt.Println("3. Standard Test Suite (System Info + CPU + Memory + Disk + Basic Unlock Tests + Common Streaming Services + 5 Speed Test Nodes)")
			fmt.Println("4. Network-Focused Test Suite (System Info + CPU + Memory + Disk + 5 Speed Test Nodes)")
			fmt.Println("5. Unlock-Focused Test Suite (System Info + CPU + Memory + Disk IO + Basic Unlock Tests + Common Streaming Services + 5 Speed Test Nodes)")
			fmt.Println("6. Network-Only Test (IP Quality Test + 5 Speed Test Nodes)")
			fmt.Println("7. Unlock-Only Test (Basic Unlock Tests + Common Streaming Services Unlock)")
			fmt.Println("8. Hardware-Only Test (Basic System Info + CPU + Memory + dd Disk Test + fio Disk Test)")
			fmt.Println("9. IP Quality Test (IP Test with 15 Databases + Email Port Test)")
		}
		for {
			fmt.Print("请输入选项 / Please enter your choice: ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if re.MatchString(input) {
				choice = input
				switch choice {
				case "1":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					commTestStatus = true
					utTestStatus = true
					securityTestStatus = true
					emailTestStatus = true
					backtraceStatus = true
					nt3Status = true
					speedTestStatus = true
					break
				case "2":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					speedTestStatus = true
					break
				case "3":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					commTestStatus = true
					utTestStatus = true
					securityTestStatus = true
					backtraceStatus = true
					nt3Status = true
					speedTestStatus = true
					break
				case "4":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					backtraceStatus = true
					nt3Status = true
					speedTestStatus = true
					break
				case "5":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					securityTestStatus = true
					speedTestStatus = true
					break
				case "6":
					speedTestStatus = true
					backtraceStatus = true
					nt3Status = true
					break
				case "7":
					securityTestStatus = true
					commTestStatus = true
					break
				case "8":
					basicStatus = true
					cpuTestStatus = true
					memoryTestStatus = true
					diskTestStatus = true
					break
				case "9":
					emailTestStatus = true
					break
				case "10":
					backtraceStatus = true
					nt3Status = true
					speedTestStatus = true
					break
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
		}
	}
	if language == "en" {
		backtraceStatus = false
		nt3Status = false
	}
	startTime := time.Now()
	var (
		wg1, wg2, wg3                                                sync.WaitGroup
		basicInfo, securityInfo, emailInfo, mediaInfo, backtraceInfo string
		output, tempOutput                                           string
	)
	output = utils.PrintAndCapture(func() {
		switch language {
		case "zh":
			utils.PrintHead(language, width, ecsVersion)
			if basicStatus || securityTestStatus {
				if basicStatus {
					utils.PrintCenteredTitle("基础信息", width)
				}
				basicInfo, securityInfo, nt3CheckType = utils.SecurityCheck(language, nt3CheckType, securityTestStatus)
				if basicStatus {
					fmt.Printf(basicInfo)
				}
			}
			if cpuTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", cpuTestMethod), width)
				cputest.CpuTest(language, cpuTestMethod, cpuTestThreadMode)
			}
			if memoryTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", cpuTestMethod), width)
				memorytest.MemoryTest(language, memoryTestMethod)
			}
			if diskTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", diskTestMethod), width)
				disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck)
			}
			if emailTestStatus {
				wg2.Add(1)
				go func() {
					defer wg1.Done()
					emailInfo = email.EmailCheck()
				}()
			}
			if utTestStatus {
				wg1.Add(1)
				go func() {
					defer wg1.Done()
					mediaInfo = unlocktest.MediaTest(language)
				}()
			}
			if runtime.GOOS == "windows" {
				if backtraceStatus {
					wg3.Add(1)
					go func() {
						defer wg3.Done()
						backtraceInfo = utils.GetCaptureOutput(func() {
							backtrace.BackTrace()
						})
					}()
				}
			}
			if commTestStatus {
				utils.PrintCenteredTitle("御三家流媒体解锁", width)
				commediatest.ComMediaTest(language)
			}
			if utTestStatus {
				utils.PrintCenteredTitle("跨国流媒体解锁", width)
				wg1.Wait()
				fmt.Printf(mediaInfo)
			}
			if securityTestStatus {
				utils.PrintCenteredTitle("IP质量检测", width)
				fmt.Printf(securityInfo)
			}
			if emailTestStatus {
				utils.PrintCenteredTitle("邮件端口检测", width)
				wg2.Wait()
				fmt.Println(emailInfo)
			}
			if runtime.GOOS != "windows" {
				if backtraceStatus {
					utils.PrintCenteredTitle("三网回程", width)
					wg3.Wait()
					fmt.Printf(backtraceInfo)
				}
				// nexttrace 在win上不支持检测，报错 bind: An invalid argument was supplied.
				if nt3Status {
					utils.PrintCenteredTitle("路由检测", width)
					ntrace.TraceRoute3(language, nt3Location, nt3CheckType)
				}
			}
			if speedTestStatus {
				utils.PrintCenteredTitle("就近节点测速", width)
				speedtest.ShowHead(language)
				if (menuMode && choice == "1") || !menuMode {
					speedtest.NearbySP()
					speedtest.CustomSP("net", "global", 2)
					speedtest.CustomSP("net", "cu", spNum)
					speedtest.CustomSP("net", "ct", spNum)
					speedtest.CustomSP("net", "cmcc", spNum)
				} else if menuMode && choice == "2" || choice == "3" || choice == "4" || choice == "5" {
					speedtest.CustomSP("net", "global", 4)
				}
			}
			utils.PrintCenteredTitle("", width)
			endTime := time.Now()
			duration := endTime.Sub(startTime)
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) % 60
			currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
			fmt.Printf("花费          : %d 分 %d 秒\n", minutes, seconds)
			fmt.Printf("时间          : %s\n", currentTime)
			utils.PrintCenteredTitle("", width)
		case "en":
			utils.PrintHead(language, width, ecsVersion)
			if basicStatus || securityTestStatus {
				if basicStatus {
					utils.PrintCenteredTitle("Basic Information", width)
				}
				basicInfo, securityInfo, nt3CheckType = utils.SecurityCheck(language, nt3CheckType, securityTestStatus)
				if basicStatus {
					fmt.Printf(basicInfo)
				}
			}
			if cpuTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("CPU Test - %s Method", cpuTestMethod), width)
				cputest.CpuTest(language, cpuTestMethod, cpuTestThreadMode)
			}
			if memoryTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("Memory Test - %s Method", memoryTestMethod), width)
				memorytest.MemoryTest(language, memoryTestMethod)
			}
			if diskTestStatus {
				utils.PrintCenteredTitle(fmt.Sprintf("Disk Test - %s Method", diskTestMethod), width)
				disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck)
			}
			if emailTestStatus {
				wg1.Add(1)
				go func() {
					defer wg1.Done()
					emailInfo = email.EmailCheck()
				}()
			}
			if commTestStatus {
				utils.PrintCenteredTitle("The Three Families Streaming Media Unlock", width)
				commediatest.ComMediaTest(language)
			}
			if utTestStatus {
				utils.PrintCenteredTitle("Cross-Border Streaming Media Unlock", width)
				unlocktest.MediaTest(language)
			}
			if securityTestStatus {
				utils.PrintCenteredTitle("IP Quality Check", width)
				fmt.Printf(securityInfo)
			}
			if emailTestStatus {
				utils.PrintCenteredTitle("Email Port Check", width)
				wg1.Wait()
				fmt.Println(emailInfo)
			}
			if speedTestStatus {
				utils.PrintCenteredTitle("Nearby Node Speed Test", width)
				speedtest.ShowHead(language)
				speedtest.NearbySP()
				speedtest.CustomSP("net", "global", -1)
			}
			utils.PrintCenteredTitle("", width)
			endTime := time.Now()
			duration := endTime.Sub(startTime)
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) % 60
			currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
			fmt.Printf("Cost    Time          : %d 分 %d 秒\n", minutes, seconds)
			fmt.Printf("Current Time          : %s\n", currentTime)
			utils.PrintCenteredTitle("", width)
		default:
			fmt.Println("Unsupported language")
		}
	}, tempOutput, output)
	filePath := "goecs.txt"
	utils.ProcessAndUpload(output, filePath, true)
}
