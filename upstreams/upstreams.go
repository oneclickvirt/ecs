package upstreams

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/oneclickvirt/UnlockTests/uts"
	bgptools "github.com/oneclickvirt/backtrace/bgptools"
	backtrace "github.com/oneclickvirt/backtrace/bk"
)

type IpInfo struct {
	Ip      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

func UpstreamsCheck() {
	info := IpInfo{}
	rsp, err := http.Get("http://ipinfo.io")
	if err == nil {
		err = json.NewDecoder(rsp.Body).Decode(&info)
		if err == nil {
			result, err := bgptools.GetPoPInfo(info.Ip)
			if err == nil {
				fmt.Print(result.Result)
			}
		}
	}
	if uts.IPV6 {
		backtrace.BackTrace(true)
	} else {
		backtrace.BackTrace(false)
	}
}
