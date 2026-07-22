package tests

import (
	"fmt"
	"net"
	"strings"

	"github.com/oneclickvirt/nt3/nt"
)

func NextTrace3Check(language, nt3Location, nt3CheckType string) {
	fmt.Print(NextTrace3CheckText(language, nt3Location, nt3CheckType))
}

// NextTrace3CheckText returns the complete route section without writing to
// stdout/stderr. This allows the caller to execute routes concurrently while
// preserving deterministic section output.
func NextTrace3CheckText(language, nt3Location, nt3CheckType string) (output string) {
	var builder strings.Builder
	// 先检查 ICMP 权限
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		// 没有权限，显示友好提示并跳过
		if language == "zh" {
			return "路由追踪测试需要 root 权限或 CAP_NET_RAW 能力，已跳过\n"
		} else {
			return "Route tracing test requires root privileges or CAP_NET_RAW capability, skipped\n"
		}
	}
	conn.Close()
	defer func() {
		if r := recover(); r != nil {
			if language == "zh" {
				builder.WriteString("路由追踪测试出现错误，已跳过\n")
			} else {
				builder.WriteString("Route tracing test failed, skipped\n")
			}
			output = builder.String()
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
					builder.WriteString(res + "\n")
				}
			}
			continue
		}
		if result.ISPName == "Error" {
			if language == "zh" {
				builder.WriteString("路由追踪测试失败（可能因为权限不足），已跳过\n")
			} else {
				builder.WriteString("Route tracing test failed (possibly due to insufficient permissions), skipped\n")
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
				builder.WriteString(res)
			} else {
				builder.WriteString(res + "\n")
			}
		}
	}
	if errorOccurred {
		if language == "zh" {
			builder.WriteString("提示: 路由追踪需要 root 权限或 CAP_NET_RAW 能力\n")
		} else {
			builder.WriteString("Hint: Route tracing requires root privileges or CAP_NET_RAW capability\n")
		}
	}
	return builder.String()
}
