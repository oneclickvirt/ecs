package basic1

import (
	"fmt"
	"github.com/oneclickvirt/basics/network"
	"github.com/oneclickvirt/basics/system"
	"strings"
)

// 本包不在main中使用，仅做测试使用
func Basic(language string) {
	ipInfo, _, _ := network.NetworkCheck("both", false, language)
	systemInfo := system.CheckSystemInfo(language)
	basicInfo := strings.ReplaceAll(systemInfo+ipInfo, "\n\n", "\n")
	fmt.Printf(basicInfo)
}
