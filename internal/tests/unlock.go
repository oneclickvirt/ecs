package tests

import (
	"fmt"
	"os"

	"github.com/oneclickvirt/UnlockTests/executor"
	"github.com/oneclickvirt/UnlockTests/utils"
	"github.com/oneclickvirt/defaultset"
)

// MediaTest runs streaming unlock tests.
// ipVersion controls which IP stacks to probe: "auto" (both), "ipv4", or "ipv6".
// showIP controls whether to prepend "IPV4:" / "IPV6:" section labels in the output.
// Unavailable IP versions are silently skipped regardless of the ipVersion setting.
func MediaTest(language, region, ipVersion string, showIP bool) string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[WARN] MediaTest panic: %v\n", r)
		}
	}()

	var res string
	readStatus := executor.ReadSelect(language, region)
	if !readStatus {
		return ""
	}
	testV4 := ipVersion == "auto" || ipVersion == "" || ipVersion == "ipv4"
	testV6 := ipVersion == "auto" || ipVersion == "" || ipVersion == "ipv6"
	if testV4 && IPV4 != "" {
		if showIP {
			res += defaultset.Blue("IPV4:") + "\n"
		}
		res += executor.RunTests(utils.Ipv4HttpClient, "ipv4", language, false)
	}
	if testV6 && IPV6 != "" {
		if showIP {
			res += defaultset.Blue("IPV6:") + "\n"
		}
		res += executor.RunTests(utils.Ipv6HttpClient, "ipv6", language, false)
	}
	return res
}
