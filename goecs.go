package main

import (
	"flag"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/uts"
	"github.com/oneclickvirt/ecs/backtrace"
	"github.com/oneclickvirt/ecs/basic"
	"github.com/oneclickvirt/ecs/commediatest"
	"github.com/oneclickvirt/ecs/cputest"
	"github.com/oneclickvirt/ecs/disktest"
	"github.com/oneclickvirt/ecs/memorytest"
	"github.com/oneclickvirt/ecs/network"
	"github.com/oneclickvirt/ecs/ntrace"
	"github.com/oneclickvirt/ecs/speedtest"
	"github.com/oneclickvirt/ecs/unlocktest"
	"github.com/oneclickvirt/portchecker/email"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	ecsVersion                   = "2024.06.30"
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

func printCenteredTitle(title string, width int) {
	titleLength := utf8.RuneCountInString(title) // 计算字符串的字符数
	totalPadding := width - titleLength
	padding := totalPadding / 2
	paddingStr := strings.Repeat("-", padding)
	fmt.Println(paddingStr + title + paddingStr + strings.Repeat("-", totalPadding%2))
}

func securityCheck() string {
	ipInfo, securityInfo, _ := network.NetworkCheck("both", true, language)
	if strings.Contains(ipInfo, "IPV4") && strings.Contains(ipInfo, "IPV6") {
		uts.IPV4 = true
		uts.IPV6 = true
		if nt3CheckType == "" {
			nt3CheckType = "ipv4"
		}
	} else if strings.Contains(ipInfo, "IPV4") {
		uts.IPV4 = true
		uts.IPV6 = false
		if nt3CheckType == "" {
			nt3CheckType = "ipv4"
		}
	} else if strings.Contains(ipInfo, "IPV6") {
		uts.IPV6 = true
		uts.IPV4 = false
		if nt3CheckType == "" {
			nt3CheckType = "ipv6"
		}
	}
	if nt3CheckType == "ipv4" && !strings.Contains(ipInfo, "IPV4") && strings.Contains(ipInfo, "IPV6") {
		nt3CheckType = "ipv6"
	} else if nt3CheckType == "ipv6" && !strings.Contains(ipInfo, "IPV6") && strings.Contains(ipInfo, "IPV4") {
		nt3CheckType = "ipv4"
	}
	return securityInfo
}

func mediatest(language string) string {
	return unlocktest.MediaTest(language)
}

func printHead() {
	if language == "zh" {
		printCenteredTitle("融合怪测试", width)
		fmt.Printf("版本：%s\n", ecsVersion)
		fmt.Println("测评频道: https://t.me/vps_reviews\n" +
			"Go项目地址：https://github.com/oneclickvirt/ecs\n" +
			"Shell项目地址：https://github.com/spiritLHLS/ecs")
	} else {
		printCenteredTitle("Fusion Monster Test", width)
		fmt.Printf("Version: %s\n", ecsVersion)
		fmt.Println("Review Channel: https://t.me/vps_reviews\n" +
			"Go Project URL: https://github.com/oneclickvirt/ecs\n" +
			"Shell Project URL: https://github.com/spiritLHLS/ecs")
	}
}

func main() {
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.StringVar(&language, "l", "zh", "Specify language (supported: en, zh)")
	flag.StringVar(&cpuTestMethod, "cpum", "sysbench", "Specify CPU test method (supported: sysbench, geekbench, winsat)")
	flag.StringVar(&cpuTestThread, "cput", "", "Specify CPU test thread count (supported: 1, 2, ...)")
	flag.StringVar(&memoryTestMethod, "memorym", "dd", "Specify Memory test method (supported: sysbench, dd, winsat)")
	flag.StringVar(&diskTestMethod, "diskm", "fio", "Specify Disk test method (supported: fio, dd, winsat)")
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
	var wg sync.WaitGroup
	var securityInfo, emailInfo, mediaInfo string
	if language == "zh" {
		printHead()
		printCenteredTitle("基础信息", width)
		wg.Add(2)
		go func() {
			defer wg.Done()
			basic.Basic(language)
		}()
		go func() {
			defer wg.Done()
			securityInfo = securityCheck()
		}()
		wg.Wait()
		printCenteredTitle(fmt.Sprintf("CPU测试-通过%s测试", cpuTestMethod), width)
		cputest.CpuTest(language, cpuTestMethod, cpuTestThread)
		printCenteredTitle(fmt.Sprintf("内存测试-通过%s测试", cpuTestMethod), width)
		memorytest.MemoryTest(language, memoryTestMethod)
		printCenteredTitle(fmt.Sprintf("硬盘测试-通过%s测试", diskTestMethod), width)
		disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck)
		wg.Add(1)
		go func() {
			defer wg.Done()
			emailInfo = email.EmailCheck()
		}()
		printCenteredTitle("御三家流媒体解锁", width)
		go func() {
			defer wg.Done()
			mediaInfo = mediatest(language)
		}()
		commediatest.ComMediaTest(language)
		printCenteredTitle("跨国流媒体解锁", width)
		wg.Wait()
		fmt.Printf(mediaInfo)
		printCenteredTitle("IP质量检测", width)
		fmt.Printf(securityInfo)
		printCenteredTitle("邮件端口检测", width)
		fmt.Println(emailInfo)
		printCenteredTitle("三网回程", width)
		backtrace.BackTrace()
		if runtime.GOOS != "windows" {
			// nexttrace 在win上不支持检测，报错 bind: An invalid argument was supplied.
			printCenteredTitle("路由检测", width)
			ntrace.TraceRoute3(language, nt3Location, nt3CheckType)
		}
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
		printHead()
		printCenteredTitle("Basic Information", width)
		wg.Add(2)
		go func() {
			defer wg.Done()
			basic.Basic(language)
		}()
		go func() {
			defer wg.Done()
			securityInfo = securityCheck()
		}()
		wg.Wait()
		printCenteredTitle(fmt.Sprintf("CPU Test - %s Method", cpuTestMethod), width)
		cputest.CpuTest(language, cpuTestMethod, cpuTestThread)
		printCenteredTitle(fmt.Sprintf("Memory Test - %s Method", memoryTestMethod), width)
		memorytest.MemoryTest(language, memoryTestMethod)
		printCenteredTitle(fmt.Sprintf("Disk Test - %s Method", diskTestMethod), width)
		disktest.DiskTest(language, diskTestMethod, diskTestPath, diskMultiCheck)
		wg.Add(1)
		go func() {
			defer wg.Done()
			emailInfo = email.EmailCheck()
		}()
		printCenteredTitle("The Three Families Streaming Media Unlock", width)
		commediatest.ComMediaTest(language)
		printCenteredTitle("Cross-Border Streaming Media Unlock", width)
		unlocktest.MediaTest(language)
		printCenteredTitle("IP Quality Check", width)
		fmt.Printf(securityInfo)
		printCenteredTitle("Email Port Check", width)
		wg.Wait()
		fmt.Println(emailInfo)
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
