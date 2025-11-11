package tests

import (
	"fmt"
	"os"
	"strings"

	"github.com/oneclickvirt/nt3/nt"
)

func NextTrace3Check(language, nt3Location, nt3CheckType string) {
	// 添加panic恢复机制，防止因权限问题导致程序崩溃
	defer func() {
		if r := recover(); r != nil {
			if language == "zh" {
				fmt.Println("\n路由追踪测试出现错误（可能因为权限不足），已跳过")
				fmt.Fprintf(os.Stderr, "[WARN] 路由追踪panic: %v\n", r)
			} else {
				fmt.Println("\nRoute tracing test failed (possibly due to insufficient permissions), skipped")
				fmt.Fprintf(os.Stderr, "[WARN] Route tracing panic: %v\n", r)
			}
		}
	}()
	
	resultChan := make(chan nt.TraceResult, 100)
	go func() {
		// 在goroutine中也添加错误恢复
		defer func() {
			if r := recover(); r != nil {
				// 发送错误结果并关闭channel
				resultChan <- nt.TraceResult{
					Index:   -1,
					ISPName: "Error",
					Output:  []string{fmt.Sprintf("Route tracing error: %v", r)},
				}
				close(resultChan)
			}
		}()
		nt.TraceRoute(language, nt3Location, nt3CheckType, resultChan)
	}()
	
	for result := range resultChan {
		if result.Index == -1 {
			for index, res := range result.Output {
				res = strings.TrimSpace(res)
				if res != "" && index == 0 {
					fmt.Println(res)
				}
			}
			continue
		}
		if result.ISPName == "Error" {
			for _, res := range result.Output {
				res = strings.TrimSpace(res)
				if res != "" {
					fmt.Println(res)
				}
			}
			continue
		}
		for _, res := range result.Output {
			res = strings.TrimSpace(res)
			if res == "" {
				continue
			}
			if strings.Contains(res, "ICMP") {
				fmt.Print(res)
			} else {
				fmt.Println(res)
			}
		}
	}
}
