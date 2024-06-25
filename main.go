package main

import (
	"flag"
	"fmt"
	"github.com/oneclickvirt/ecs/basic"
	"github.com/oneclickvirt/ecs/cputest"
	"github.com/oneclickvirt/ecs/disktest"
	"github.com/oneclickvirt/ecs/memorytest"
	"strings"
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
		width                        = 80
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
	if showVersion {
		fmt.Println(ecsVersion)
		return
	}
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
		printCenteredTitle("", width)
	}
}
