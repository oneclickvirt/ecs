package basic1

import (
	"fmt"
	"github.com/oneclickvirt/basics/network"
	"github.com/oneclickvirt/basics/system"
	"strings"
)

// 使用gopsutil查询可能会特别慢，执行命令查询反而更快
// TODO
// 迁移Shell的完整检测逻辑使用执行命令的方式查询，最后都失败才使用gopsutil查询
// 本包不在main中使用
func Basic(language string) {
	ipInfo, _, _ := network.NetworkCheck("both", false, language)
	systemInfo := system.CheckSystemInfo(language)
	basicInfo := strings.ReplaceAll(systemInfo+ipInfo, "\n\n", "\n")
	fmt.Printf(basicInfo)
}
