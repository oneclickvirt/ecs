package tests

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/oneclickvirt/nt3/nt"
)

func NextTrace3Check(language, nt3Location, nt3CheckType string) {
	// 先检查 ICMP 权限
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		// 没有权限，显示友好提示并跳过
		if language == "zh" {
			fmt.Println("路由追踪测试需要 root 权限或 CAP_NET_RAW 能力，已跳过")
			fmt.Fprintf(os.Stderr, "[WARN] ICMP权限不足: %v\n", err)
		} else {
			fmt.Println("Route tracing test requires root privileges or CAP_NET_RAW capability, skipped")
			fmt.Fprintf(os.Stderr, "[WARN] Insufficient ICMP permission: %v\n", err)
		}
		return
	}
	conn.Close()
	defer func() {
		if r := recover(); r != nil {
			if language == "zh" {
				fmt.Println("路由追踪测试出现错误，已跳过")
				fmt.Fprintf(os.Stderr, "[WARN] 路由追踪panic: %v\n", r)
			} else {
				fmt.Println("Route tracing test failed, skipped")
				fmt.Fprintf(os.Stderr, "[WARN] Route tracing panic: %v\n", r)
			}
		}
	}()
	resultChan := make(chan nt.TraceResult, 100)
	errorOccurred := false
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorOccurred = true
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
			if language == "zh" {
				fmt.Println("路由追踪测试失败（可能因为权限不足），已跳过")
			} else {
				fmt.Println("Route tracing test failed (possibly due to insufficient permissions), skipped")
			}
			for _, res := range result.Output {
				res = strings.TrimSpace(res)
				if res != "" {
					fmt.Fprintf(os.Stderr, "[WARN] %s\n", res)
				}
			}
			errorOccurred = true
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
	if errorOccurred {
		if language == "zh" {
			fmt.Println("提示: 路由追踪需要 root 权限或 CAP_NET_RAW 能力")
		} else {
			fmt.Println("Hint: Route tracing requires root privileges or CAP_NET_RAW capability")
		}
	}
}
