package memorytest

import (
	"fmt"
	"github.com/oneclickvirt/memorytest/memory"
	"runtime"
	"strings"
)

func MemoryTest(language, testMethod string) {
	var res string
	if runtime.GOOS == "windows" {
		if testMethod != "winsat" && testMethod != "" {
			res = "Detected host is Windows, using Winsat for testing.\n"
		}
		res += memory.WinsatTest(language)
	} else {
		switch testMethod {
		case "sysbench":
			res = memory.SysBenchTest(language)
			if res == "" {
				res = "sysbench test failed, switch to use dd test.\n"
				res += memory.DDTest(language)
			}
		case "dd":
			res = memory.DDTest(language)
		default:
			res = "Unsupported test method"
		}
	}
	if !strings.Contains(res, "\n") && res != "" {
		res += "\n"
	}
	fmt.Printf(res)
}
