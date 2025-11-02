package tests

import (
	"runtime"
	"strings"

	"github.com/oneclickvirt/memorytest/memory"
)

func MemoryTest(language, testMethod string) (realTestMethod, res string) {
	testMethod = strings.ToLower(testMethod)
	if testMethod == "" {
		testMethod = "auto"
	}
	if runtime.GOOS == "windows" {
		switch testMethod {
		case "stream":
			res = memory.WinsatTest(language)
			realTestMethod = "winsat"
		case "dd":
			res = memory.WindowsDDTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				res += memory.WinsatTest(language)
				realTestMethod = "winsat"
			} else {
				realTestMethod = "dd"
			}
		case "sysbench":
			res = memory.WinsatTest(language)
			realTestMethod = "winsat"
		case "auto", "winsat":
			res = memory.WinsatTest(language)
			realTestMethod = "winsat"
		default:
			res = memory.WinsatTest(language)
			realTestMethod = "winsat"
		}
	} else {
		switch testMethod {
		case "stream":
			res = memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				res += memory.DDTest(language)
				realTestMethod = "dd"
			} else {
				realTestMethod = "stream"
			}
		case "dd":
			res = memory.DDTest(language)
			realTestMethod = "dd"
		case "sysbench":
			res = memory.SysBenchTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				res += memory.DDTest(language)
				realTestMethod = "dd"
			} else {
				realTestMethod = "sysbench"
			}
		case "auto":
			res = memory.StreamTest(language)
			if res == "" || strings.TrimSpace(res) == "" {
				res = memory.DDTest(language)
				if res == "" || strings.TrimSpace(res) == "" {
					res = memory.SysBenchTest(language)
					if res == "" || strings.TrimSpace(res) == "" {
						realTestMethod = ""
					} else {
						realTestMethod = "sysbench"
					}
				} else {
					realTestMethod = "dd"
				}
			} else {
				realTestMethod = "stream"
			}
		case "winsat":
			// winsat 仅 Windows 支持，非 Windows fallback 到 dd
			res = memory.DDTest(language)
			realTestMethod = "dd"
		default:
			res = "Unsupported test method"
			realTestMethod = ""
		}
	}
	if !strings.Contains(res, "\n") && res != "" {
		res += "\n"
	}
	return
}
