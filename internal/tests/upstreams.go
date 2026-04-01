package tests

import (
	"fmt"
	"os"
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

func UpstreamsCheck(language string) {
	// 添加panic恢复机制
	defer func() {
		if r := recover(); r != nil {
			if language == "zh" {
				fmt.Println("\n上游检测出现错误，已跳过")
			} else {
				fmt.Println("\nUpstream check failed, skipped")
			}
			fmt.Fprintf(os.Stderr, "[WARN] Upstream check panic: %v\n", r)
		}
	}()
	
	results := ConcurrentResults{}
	var wg sync.WaitGroup
	if IPV4 != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr, "[WARN] BGP info panic: %v\n", r)
				}
			}()
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
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "[WARN] Backtrace panic: %v\n", r)
			}
		}()
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
	if language == "zh" {
		fmt.Println(Yellow("准确线路自行查看详细路由，本测试结果仅作参考"))
		fmt.Println(Yellow("同一目标地址多个线路时，检测可能已越过汇聚层，除第一个线路外，后续信息可能无效"))
	} else {
		fmt.Println(Yellow("For accurate routing, check the detailed routes yourself. This result is for reference only."))
		fmt.Println(Yellow("When multiple routes share the same destination, detection may have passed the aggregation layer; only the first route is reliable."))
	}
}
