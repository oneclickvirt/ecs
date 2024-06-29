package main

import (
	"flag"
	"fmt"
	"github.com/oneclickvirt/ecs/backtrace"
	"github.com/oneclickvirt/ecs/basic"
	"github.com/oneclickvirt/ecs/commediatest"
	"github.com/oneclickvirt/ecs/cputest"
	"github.com/oneclickvirt/ecs/disktest"
	"github.com/oneclickvirt/ecs/memorytest"
	"github.com/oneclickvirt/ecs/network"
	"github.com/oneclickvirt/ecs/ntrace"
	"github.com/oneclickvirt/ecs/port"
	"github.com/oneclickvirt/ecs/speedtest"
	"github.com/oneclickvirt/ecs/unlocktest"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"
)

func printCenteredTitle(title string, width int) {
	titleLength := utf8.RuneCountInString(title) // 计算字符串的字符数
	totalPadding := width - titleLength
	padding := totalPadding / 2
	paddingStr := strings.Repeat("-", padding)
	fmt.Println(paddingStr + title + paddingStr + strings.Repeat("-", totalPadding%2))
}

func main() {
	var (
		ecsVersion                   = "2024.06.25"
		showVersion                  bool
		language                     string
		cpuTestMethod, cpuTestThread string
		memoryTestMethod             string
		diskTestMethod, diskTestPath string
		diskMultiCheck               bool
		nt3CheckType, nt3Location    string
		spNum                        int
		width                        = 84
	)
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.StringVar(&language, "l", "zh", "Specify language (supported: en, zh)")
	flag.StringVar(&cpuTestMethod, "cpum", "sysbench", "Specify CPU test method (supported: sysbench, geekbench, winsat)")
	flag.StringVar(&cpuTestThread, "cput", "", "Specify CPU test thread count (supported: 1, 2, ...)")
	flag.StringVar(&memoryTestMethod, "memorym", "", "Specify Memory test method (supported: sysbench, dd, winsat)")
	flag.StringVar(&diskTestMethod, "diskm", "", "Specify Disk test method (supported: sysbench, dd, winsat)")
	flag.StringVar(&diskTestPath, "diskp", "", "Specify Disk test path, example: -diskp /root")
	flag.BoolVar(&diskMultiCheck, "diskmc", false, "Enable multiple disk checks, example: -diskmc=false")
	flag.Parse()
	if language == "zh" {
		flag.StringVar(&nt3Location, "nt3loc", "GZ", "指定三网回程路由检测的地址，支持 GZ, SH, BJ, CD 对应 广州，上海，北京，成都")
		flag.StringVar(&nt3CheckType, "nt3t", "ipv4", "指定三网回程路由检测的类型，支持 both, ipv4, ipv6")
	}
	flag.IntVar(&spNum, "spnum", 2, "Specify speedtest each operator servers num")
	flag.Parse()
	if showVersion {
		fmt.Println(ecsVersion)
		return
	}
	startTime := time.Now()
	if language == "zh" {
		printCenteredTitle("融合怪测试", width)
		fmt.Printf("版本：%s\n", ecsVersion)
		fmt.Println("测评频道: https://t.me/vps_reviews\nGo项目地址：https://github.com/oneclickvirt/ecs\nShell项目地址：https://github.com/spiritLHLS/ecs")
		printCenteredTitle("基础信息", width)
		basic.Basic(language)
		printCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", cpuTestMethod), width)
		cputest.CpuTest(language, cpuTestMethod, cpuTestThread)
		printCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", cpuTestMethod), width)
		memorytest.MemoryTest(language, memoryTestMethod)
		printCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", diskTestMethod), width)
		disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck)
		printCenteredTitle("御三家流媒体解锁", width)
		commediatest.ComMediaTest(language)
		printCenteredTitle("跨国流媒体解锁", width)
		unlocktest.MediaTest(language)
		printCenteredTitle("IP质量检测", width)
		ipInfo, securityInfo, _ := network.NetworkCheck("both", true, language)
		fmt.Printf(securityInfo)
		printCenteredTitle("邮件端口检测", width)
		port.EmailCheck()
		if runtime.GOOS != "windows" {
			// nexttrace 在win上不支持检测，报错 bind: An invalid argument was supplied.
			printCenteredTitle("三网回程", width)
			backtrace.BackTrace()
		}
		printCenteredTitle("三网回程路由", width)
		if nt3CheckType == "" && strings.Contains(ipInfo, "IPV4") {
			nt3CheckType = "ipv4"
		} else if nt3CheckType == "" && strings.Contains(ipInfo, "IPV6") {
			nt3CheckType = "ipv6"
		} else if nt3CheckType == "ipv4" && !strings.Contains(ipInfo, "IPV4") && strings.Contains(ipInfo, "IPV6") {
			nt3CheckType = "ipv6"
		} else if nt3CheckType == "ipv6" && !strings.Contains(ipInfo, "IPV6") && strings.Contains(ipInfo, "IPV4") {
			nt3CheckType = "ipv4"
		}
		ntrace.TraceRoute3(language, nt3Location, nt3CheckType)
		printCenteredTitle("就近节点测速", width)
		speedtest.ShowHead(language)
		speedtest.NearbySP()
		speedtest.CustomSP("net", "global", 4)
		speedtest.CustomSP("net", "cu", spNum)
		speedtest.CustomSP("net", "ct", spNum)
		speedtest.CustomSP("net", "cmcc", spNum)
		printCenteredTitle("", width)
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		fmt.Printf("花费          : %d 分 %d 秒\n", minutes, seconds)
		currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
		fmt.Printf("时间          : %s\n", currentTime)
		printCenteredTitle("", width)
	} else if language == "en" {
		printCenteredTitle("Fusion Monster Test", width)
		fmt.Printf("Version: %s\n", ecsVersion)
		fmt.Println("Review Channel: https://t.me/vps_reviews\nGo Project URL: https://github.com/oneclickvirt/ecs\nShell Project URL: https://github.com/spiritLHLS/ecs")
		printCenteredTitle("Basic Information", width)
		basic.Basic(language)
		printCenteredTitle(fmt.Sprintf("CPU Test - %s Method", cpuTestMethod), width)
		cputest.CpuTest(language, cpuTestMethod, cpuTestThread)
		printCenteredTitle(fmt.Sprintf("Memory Test - %s Method", memoryTestMethod), width)
		memorytest.MemoryTest(language, memoryTestMethod)
		printCenteredTitle(fmt.Sprintf("Disk Test - %s Method", diskTestMethod), width)
		disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck)
		printCenteredTitle("The Three Families Streaming Media Unlock", width)
		commediatest.ComMediaTest(language)
		printCenteredTitle("Cross-Border Streaming Media Unlock", width)
		unlocktest.MediaTest(language)
		printCenteredTitle("IP Quality Check", width)
		_, securityInfo, _ := network.NetworkCheck("both", true, language)
		fmt.Printf(securityInfo)
		printCenteredTitle("Email Port Check", width)
		port.EmailCheck()
		//printCenteredTitle("Return Path Routing", width)
		printCenteredTitle("Nearby Node Speed Test", width)
		speedtest.ShowHead(language)
		speedtest.NearbySP()
		speedtest.CustomSP("net", "global", -1)
		printCenteredTitle("", width)
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		fmt.Printf("Cost    Time          : %d 分 %d 秒\n", minutes, seconds)
		currentTime := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
		fmt.Printf("Current Time          : %s\n", currentTime)
		printCenteredTitle("", width)
	}
}
