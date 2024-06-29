package cputest

import (
	"fmt"
	"github.com/oneclickvirt/cputest/cpu"
	"runtime"
	"strings"
)

func CpuTest(language, testMethod, testThread string) {
	var res string
	if runtime.GOOS == "windows" {
		if testMethod != "winsat" && testMethod != "" {
			res = "Detected host is Windows, using Winsat for testing.\n"
		}
		res += cpu.WinsatTest(language, testThread)
	} else {
		switch testMethod {
		case "sysbench":
			res = cpu.SysBenchTest(language, testThread)
			if res == "" {
				res = "Sysbench test failed, switching to Geekbench for testing.\n"
				res += cpu.GeekBenchTest(language, testThread)
			}
		case "geekbench":
			res = cpu.GeekBenchTest(language, testThread)
			if res == "" {
				res = "Geekbench test failed, switching to Sysbench for testing.\n"
				res += cpu.SysBenchTest(language, testThread)
			}
		default:
			res = "Invalid test method specified.\n"
		}
	}
	if !strings.Contains(res, "\n") && res != "" {
		res += "\n"
	}
	fmt.Print(res)
}
