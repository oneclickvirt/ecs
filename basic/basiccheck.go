package basic

import (
	"fmt"
	"github.com/oneclickvirt/basics/network"
	"github.com/oneclickvirt/basics/system"
	"strings"
)

// 使用gopsutil查询可能会特别慢，执行命令查询反而更快
// TODO
// 迁移Shell的完整检测逻辑使用执行命令的方式查询，最后都失败才使用gopsutil查询

func basic() {
	language := "zh"
	ipInfo, _, _ := network.NetworkCheck("both", false, language)
	res := system.CheckSystemInfo(language)
	fmt.Println("--------------------------------------------------")
	fmt.Printf(strings.ReplaceAll(res+ipInfo, "\n\n", "\n"))
	fmt.Println("--------------------------------------------------")
}
