package unlocktest

import (
	"github.com/oneclickvirt/UnlockTests/utils"
	"github.com/oneclickvirt/UnlockTests/uts"
	"github.com/oneclickvirt/defaultset"
)

func MediaTest(language string) string {
	var res string
	readStatus := uts.ReadSelect(language, "0")
	if !readStatus {
		return ""
	}
	if uts.IPV4 {
		res += defaultset.Blue("IPV4:") + "\n"
		res += uts.RunTests(utils.Ipv4HttpClient, "ipv4", language, false)
		return res
	}
	if uts.IPV6 {
		res += defaultset.Blue("IPV6:") + "\n"
		res += uts.RunTests(utils.Ipv6HttpClient, "ipv6", language, false)
		return res
	}
	return ""
}
