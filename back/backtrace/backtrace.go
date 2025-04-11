package backtrace

import (
	"github.com/oneclickvirt/backtrace/bk"
)

func BackTrace(enableIpv6 bool) {
	backtrace.BackTrace(enableIpv6)
}
