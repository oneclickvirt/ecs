package tests

import (
	"fmt"
	"strings"
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
	fmt.Print(UpstreamsCheckText(language))
}

// UpstreamsCheckText runs the upstream probes without writing to the process
// terminal. The orchestrator can therefore run it concurrently and publish
// the complete section later without interleaving another section's output.
func UpstreamsCheckText(language string) (output string) {
	var builder strings.Builder
	defer func() {
		if r := recover(); r != nil {
			if language == "zh" {
				builder.WriteString("上游检测出现错误，已跳过\n")
			} else {
				builder.WriteString("Upstream check failed, skipped\n")
			}
			output = builder.String()
		}
	}()

	results := ConcurrentResults{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	if IPV4 != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				_ = recover()
			}()
			for i := 0; i < 2; i++ {
				result, err := bgptools.GetPoPInfo(IPV4)
				mu.Lock()
				results.bgpError = err
				if err == nil && result.Result != "" {
					results.bgpResult = result.Result
					mu.Unlock()
					return
				}
				mu.Unlock()
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
			_ = recover()
		}()
		result := backtrace.BackTrace(executor.IPV6)
		mu.Lock()
		results.backtraceResult = result
		mu.Unlock()
	}()
	wg.Wait()
	mu.Lock()
	finalResults := results
	mu.Unlock()
	if finalResults.bgpResult != "" {
		builder.WriteString(finalResults.bgpResult)
	}
	if finalResults.backtraceResult != "" {
		builder.WriteString(finalResults.backtraceResult)
		if !strings.HasSuffix(finalResults.backtraceResult, "\n") {
			builder.WriteByte('\n')
		}
	}
	if language == "zh" {
		builder.WriteString(Yellow("准确线路自行查看详细路由，本测试结果仅作参考") + "\n")
		builder.WriteString(Yellow("同一目标地址多个线路时，检测可能已越过汇聚层，除第一个线路外，后续信息可能无效") + "\n")
	} else {
		builder.WriteString(Yellow("For accurate routing, check the detailed routes yourself. This result is for reference only.") + "\n")
		builder.WriteString(Yellow("When multiple routes share the same destination, detection may have passed the aggregation layer; only the first route is reliable.") + "\n")
	}
	return builder.String()
}
