package tests

import (
	"github.com/oneclickvirt/UnlockTests/executor"
	"github.com/oneclickvirt/UnlockTests/utils"
	"github.com/oneclickvirt/defaultset"
)

func MediaTest(language string) string {
	var res string
	readStatus := executor.ReadSelect(language, "0")
	if !readStatus {
		return ""
	}
	if executor.IPV4 {
		res += defaultset.Blue("IPV4:") + "\n"
		res += executor.RunTests(utils.Ipv4HttpClient, "ipv4", language, false)
		return res
	}
	if executor.IPV6 {
		res += defaultset.Blue("IPV6:") + "\n"
		res += executor.RunTests(utils.Ipv6HttpClient, "ipv6", language, false)
		return res
	}
	return ""
}
