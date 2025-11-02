package tests

import (
	"runtime"
	"strings"

	"github.com/oneclickvirt/cputest/cpu"
)

func CpuTest(language, testMethod, testThread string) (realTestMethod, res string) {
	if runtime.GOOS == "windows" {
		if testMethod != "winsat" && testMethod != "" {
			// res = "Detected host is Windows, using Winsat for testing.\n"
			realTestMethod = "winsat"
		}
		res += cpu.WinsatTest(language, testThread)
	} else {
		switch testMethod {
		case "sysbench":
			res = cpu.SysBenchTest(language, testThread)
			if res == "" {
				// res = "Sysbench test failed, switching to Geekbench for testing.\n"
				realTestMethod = "geekbench"
				res += cpu.GeekBenchTest(language, testThread)
			} else {
				realTestMethod = "sysbench"
			}
		case "geekbench":
			res = cpu.GeekBenchTest(language, testThread)
			if res == "" {
				// res = "Geekbench test failed, switching to Sysbench for testing.\n"
				realTestMethod = "sysbench"
				res += cpu.SysBenchTest(language, testThread)
			} else {
				realTestMethod = "geekbench"
			}
		default:
			res = "Invalid test method specified.\n"
			realTestMethod = "null"
		}
	}
	if !strings.Contains(res, "\n") && res != "" {
		res += "\n"
	}
	return
}
