package tests

import (
	"fmt"
	"sync"
	"time"

	"github.com/oneclickvirt/UnlockTests/executor"
	bgptools "github.com/oneclickvirt/backtrace/bgptools"
	backtrace "github.com/oneclickvirt/backtrace/bk"
	. "github.com/oneclickvirt/defaultset"
)

type IpInfo struct {
	Ip      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

type ConcurrentResults struct {
	bgpResult       string
	backtraceResult string
	bgpError        error
	// backtraceError  error
}

var IPV4, IPV6 string

func UpstreamsCheck() {
	results := ConcurrentResults{}
	var wg sync.WaitGroup
	if IPV4 != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 2; i++ {
				result, err := bgptools.GetPoPInfo(IPV4)
				results.bgpError = err
				if err == nil && result.Result != "" {
					results.bgpResult = result.Result
					return
				}
				if i == 0 {
					time.Sleep(3 * time.Second)
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		result := backtrace.BackTrace(executor.IPV6)
		results.backtraceResult = result
	}()
	wg.Wait()
	if results.bgpResult != "" {
		fmt.Print(results.bgpResult)
	}
	if results.backtraceResult != "" {
		fmt.Printf("%s\n", results.backtraceResult)
	}
	fmt.Println(Yellow("准确线路自行查看详细路由，本测试结果仅作参考"))
	fmt.Println(Yellow("同一目标地址多个线路时，检测可能已越过汇聚层，除第一个线路外，后续信息可能无效"))
}
