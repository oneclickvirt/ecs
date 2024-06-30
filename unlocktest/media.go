package unlocktest

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/utils"
	"github.com/oneclickvirt/UnlockTests/uts"
	"github.com/oneclickvirt/defaultset"
)

func MediaTest(language string) string {
	readStatus := uts.ReadSelect(language, "0")
	if !readStatus {
		return ""
	}
	if uts.IPV4 {
		fmt.Println(defaultset.Blue("IPV4:"))
		return uts.RunTests(utils.Ipv4HttpClient, "ipv4", language, false)
	}
	if uts.IPV6 {
		fmt.Println(defaultset.Blue("IPV6:"))
		return uts.RunTests(utils.Ipv6HttpClient, "ipv6", language, false)
	}
	return ""
}
