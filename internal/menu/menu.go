package menu

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"github.com/oneclickvirt/ecs/internal/params"
	"github.com/oneclickvirt/ecs/utils"
)

// GetMenuChoice prompts user for menu choice
func GetMenuChoice(language string) string {
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

// PrintMenuOptions displays menu options
func PrintMenuOptions(preCheck utils.NetCheckResult, config *params.Config) {
	var stats *utils.StatsResponse
	var statsErr error
	var githubInfo *utils.GitHubRelease
	var githubErr error
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
		if statsErr != nil {
			statsInfo = "NULL"
		} else {
			switch config.Language {
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
			cmp = utils.CompareVersions(config.EcsVersion, githubInfo.TagName)
		} else {
			cmp = 0
		}
	}
	switch config.Language {
	case "zh":
		fmt.Printf("VPS融合怪版本: %s\n", config.EcsVersion)
		if preCheck.Connected {
			switch cmp {
			case -1:
				fmt.Printf("检测到新版本 %s 如有必要请更新！\n", githubInfo.TagName)
			}
			fmt.Printf("使用统计: %s\n", statsInfo)
		}
	fmt.Println("1. 融合怪完全体(能测全测)")
	fmt.Println("2. 极简版(系统信息+CPU+内存+磁盘+测速节点5个)")
	fmt.Println("3. 精简版(系统信息+CPU+内存+磁盘+跨国平台解锁+路由+测速节点5个)")
	fmt.Println("4. 精简网络版(系统信息+CPU+内存+磁盘+回程+路由+测速节点5个)")
	fmt.Println("5. 精简解锁版(系统信息+CPU+内存+磁盘IO+跨国平台解锁+测速节点5个)")
	fmt.Println("6. 网络单项(IP质量检测+上游及三网回程+广州三网回程详细路由+全国延迟+TGDC+网站延迟+测速节点11个)")
	fmt.Println("7. 解锁单项(跨国平台解锁)")
		fmt.Println("8. 硬件单项(系统信息+CPU+dd磁盘测试+fio磁盘测试)")
		fmt.Println("9. IP质量检测(15个数据库的IP质量检测+邮件端口检测)")
		fmt.Println("10. 三网回程线路检测+三网回程详细路由(北京上海广州成都)+全国延迟+TGDC+网站延迟")
		fmt.Println("0. 退出程序")
	case "en":
		fmt.Printf("VPS Fusion Monster Test Version: %s\n", config.EcsVersion)
		if preCheck.Connected {
			switch cmp {
			case -1:
				fmt.Printf("New version detected %s update if necessary!\n", githubInfo.TagName)
			}
			fmt.Printf("%s\n", statsInfo)
		}
	fmt.Println("1. VPS Fusion Monster Test (Full Test)")
	fmt.Println("2. Minimal Test Suite (System Info + CPU + Memory + Disk + 5 Speed Test Nodes)")
	fmt.Println("3. Standard Test Suite (System Info + CPU + Memory + Disk + International Platform Unlock + Routing + 5 Speed Test Nodes)")
	fmt.Println("4. Network-Focused Test Suite (System Info + CPU + Memory + Disk + Backtrace + Routing + 5 Speed Test Nodes)")
	fmt.Println("5. Unlock-Focused Test Suite (System Info + CPU + Memory + Disk IO + International Platform Unlock + 5 Speed Test Nodes)")
	fmt.Println("6. Network-Only Test (IP Quality Test + Upstream & 3-Network Backtrace + Guangzhou 3-Network Detailed Routing + National Latency + TGDC + Websites + 11 Speed Test Nodes)")
	fmt.Println("7. Unlock-Only Test (International Platform Unlock)")
	fmt.Println("8. Hardware-Only Test (System Info + CPU + Memory + dd Disk Test + fio Disk Test)")
	fmt.Println("9. IP Quality Test (IP Test with 15 Databases + Email Port Test)")
	fmt.Println("0. Exit Program")
	}
}

// HandleMenuMode handles menu selection
func HandleMenuMode(preCheck utils.NetCheckResult, config *params.Config) {
	savedParams := config.SaveUserSetParams()
	config.BasicStatus = false
	config.CpuTestStatus = false
	config.MemoryTestStatus = false
	config.DiskTestStatus = false
	config.UtTestStatus = false
	config.SecurityTestStatus = false
	config.EmailTestStatus = false
	config.BacktraceStatus = false
	config.Nt3Status = false
	config.SpeedTestStatus = false
	config.TgdcTestStatus = false
	config.WebTestStatus = false
	config.AutoChangeDiskMethod = true
	PrintMenuOptions(preCheck, config)
Loop:
	for {
		config.Choice = GetMenuChoice(config.Language)
		switch config.Choice {
		case "0":
			os.Exit(0)
		case "1":
			SetFullTestStatus(preCheck, config)
			config.OnlyChinaTest = utils.CheckChina(config.EnableLogger)
			break Loop
		case "2":
			SetMinimalTestStatus(preCheck, config)
			break Loop
		case "3":
			SetStandardTestStatus(preCheck, config)
			break Loop
		case "4":
			SetNetworkFocusedTestStatus(preCheck, config)
			break Loop
		case "5":
			SetUnlockFocusedTestStatus(preCheck, config)
			break Loop
		case "6":
			if !preCheck.Connected {
				fmt.Println("Can not test without network connection!")
				return
			}
			SetNetworkOnlyTestStatus(config)
			break Loop
		case "7":
			if !preCheck.Connected {
				fmt.Println("Can not test without network connection!")
				return
			}
			SetUnlockOnlyTestStatus(config)
			break Loop
		case "8":
			SetHardwareOnlyTestStatus(preCheck, config)
			break Loop
		case "9":
			if !preCheck.Connected {
				fmt.Println("Can not test without network connection!")
				return
			}
			SetIPQualityTestStatus(config)
			break Loop
		case "10":
			if !preCheck.Connected {
				fmt.Println("Can not test without network connection!")
				return
			}
			config.Nt3Location = "ALL"
			SetRouteTestStatus(config)
			break Loop
		default:
			PrintInvalidChoice(config.Language)
		}
	}
	config.RestoreUserSetParams(savedParams)
}

// SetFullTestStatus enables all tests
func SetFullTestStatus(preCheck utils.NetCheckResult, config *params.Config) {
	config.BasicStatus = true
	config.CpuTestStatus = true
	config.MemoryTestStatus = true
	config.DiskTestStatus = true
	if preCheck.Connected {
		config.UtTestStatus = true
		config.SecurityTestStatus = true
		config.EmailTestStatus = true
		config.BacktraceStatus = true
		config.Nt3Status = true
		config.SpeedTestStatus = true
		config.TgdcTestStatus = true
		config.WebTestStatus = true
	}
}

// SetMinimalTestStatus sets minimal test configuration
func SetMinimalTestStatus(preCheck utils.NetCheckResult, config *params.Config) {
	config.BasicStatus = true
	config.CpuTestStatus = true
	config.MemoryTestStatus = true
	config.DiskTestStatus = true
	if preCheck.Connected {
		config.SpeedTestStatus = true
	}
}

// SetStandardTestStatus sets standard test configuration
func SetStandardTestStatus(preCheck utils.NetCheckResult, config *params.Config) {
	config.BasicStatus = true
	config.CpuTestStatus = true
	config.MemoryTestStatus = true
	config.DiskTestStatus = true
	if preCheck.Connected {
		config.UtTestStatus = true
		config.Nt3Status = true
		config.SpeedTestStatus = true
	}
}

// SetNetworkFocusedTestStatus sets network-focused test configuration
func SetNetworkFocusedTestStatus(preCheck utils.NetCheckResult, config *params.Config) {
	config.BasicStatus = true
	config.CpuTestStatus = true
	config.MemoryTestStatus = true
	config.DiskTestStatus = true
	if preCheck.Connected {
		config.BacktraceStatus = true
		config.Nt3Status = true
		config.SpeedTestStatus = true
	}
}

// SetUnlockFocusedTestStatus sets unlock-focused test configuration
func SetUnlockFocusedTestStatus(preCheck utils.NetCheckResult, config *params.Config) {
	config.BasicStatus = true
	config.CpuTestStatus = true
	config.MemoryTestStatus = true
	config.DiskTestStatus = true
	if preCheck.Connected {
		config.UtTestStatus = true
		config.SpeedTestStatus = true
	}
}

// SetNetworkOnlyTestStatus sets network-only test configuration
func SetNetworkOnlyTestStatus(config *params.Config) {
	config.OnlyIpInfoCheck = true
	config.SecurityTestStatus = true
	config.SpeedTestStatus = true
	config.BacktraceStatus = true
	config.Nt3Status = true
	config.PingTestStatus = true
	config.TgdcTestStatus = true
	config.WebTestStatus = true
}

// SetUnlockOnlyTestStatus sets unlock-only test configuration
func SetUnlockOnlyTestStatus(config *params.Config) {
	config.OnlyIpInfoCheck = true
	config.UtTestStatus = true
}

// SetHardwareOnlyTestStatus sets hardware-only test configuration
func SetHardwareOnlyTestStatus(preCheck utils.NetCheckResult, config *params.Config) {
	_ = preCheck
	config.BasicStatus = true
	config.CpuTestStatus = true
	config.MemoryTestStatus = true
	config.DiskTestStatus = true
	config.SecurityTestStatus = false
	config.AutoChangeDiskMethod = false
}

// SetIPQualityTestStatus sets IP quality test configuration
func SetIPQualityTestStatus(config *params.Config) {
	config.OnlyIpInfoCheck = true
	config.SecurityTestStatus = true
	config.EmailTestStatus = true
}

// SetRouteTestStatus sets route test configuration
func SetRouteTestStatus(config *params.Config) {
	config.OnlyIpInfoCheck = true
	config.BacktraceStatus = true
	config.Nt3Status = true
	config.PingTestStatus = true
	config.TgdcTestStatus = true
	config.WebTestStatus = true
}

// PrintInvalidChoice prints invalid choice message
func PrintInvalidChoice(language string) {
	if language == "zh" {
		fmt.Println("无效的选项")
	} else {
		fmt.Println("Invalid choice")
	}
}
