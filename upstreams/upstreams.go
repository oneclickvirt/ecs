package upstreams

import (
	"fmt"

	"github.com/oneclickvirt/UnlockTests/uts"
	bgptools "github.com/oneclickvirt/backtrace/bgptools"
	backtrace "github.com/oneclickvirt/backtrace/bk"
)

func UpstreamsCheck(ip string) {
	if ip != "" {
		if result, err := bgptools.GetPoPInfo(ip); err == nil {
			fmt.Print(result.Result)
		}
	}
	backtrace.BackTrace(uts.IPV6)
}
