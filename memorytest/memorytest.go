package memorytest

import (
	"fmt"
	"github.com/oneclickvirt/memorytest/memory"
	"runtime"
)

func memorytest() {
	var res string
	language := "zh"
	testMethod := ""
	if runtime.GOOS == "windows" {
		res = memory.WinsatTest(language)
	} else {
		if testMethod == "sysbench" {
			res = memory.SysBenchTest(language)
			if res == "" {
				res = "sysbench test failed, switch to use dd test.\n"
				res += memory.DDTest(language)
			}
		} else if testMethod == "dd" {
			res = memory.DDTest(language)
		}
	}
	fmt.Println("--------------------------------------------------")
	fmt.Printf(res)
	fmt.Println("--------------------------------------------------")
}
