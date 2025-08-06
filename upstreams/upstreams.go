package upstreams

import (
	"fmt"

	"github.com/oneclickvirt/UnlockTests/uts"
	bgptools "github.com/oneclickvirt/backtrace/bgptools"
	backtrace "github.com/oneclickvirt/backtrace/bk"
)

var IPV4, IPV6 string

func UpstreamsCheck() {
	if IPV4 != "" {
		if result, err := bgptools.GetPoPInfo(IPV4); err == nil {
			fmt.Print(result.Result)
		}
	}
	backtrace.BackTrace(uts.IPV6)
}