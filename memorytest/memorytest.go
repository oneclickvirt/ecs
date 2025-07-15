package memorytest

import (
	"runtime"
	"strings"

	"github.com/oneclickvirt/memorytest/memory"
)

func MemoryTest(language, testMethod string) (realTestMethod, res string) {
	if runtime.GOOS == "windows" {
		if testMethod != "winsat" && testMethod != "" {
			// res = "Detected host is Windows, using Winsat for testing.\n"
			realTestMethod = "winsat"
		}
		res += memory.WinsatTest(language)
	} else {
		switch testMethod {
		case "sysbench":
			res = memory.SysBenchTest(language)
			if res == "" {
				// res = "sysbench test failed, switch to use dd test.\n"
				res += memory.DDTest(language)
				realTestMethod = "dd"
			} else {
				realTestMethod = "sysbench"
			}
		case "dd":
			res = memory.DDTest(language)
			realTestMethod = "dd"
		default:
			// res = "Unsupported test method, switch to use dd test.\n"
			res += memory.DDTest(language)
			realTestMethod = "dd"
		}
	}
	if !strings.Contains(res, "\n") && res != "" {
		res += "\n"
	}
	return
}
