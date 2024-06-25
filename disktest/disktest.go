package disktest

import (
	"fmt"
	"github.com/oneclickvirt/disktest/disk"
	"runtime"
)

func diskIoTest() {
	var language, res string
	language = "zh"
	isMultiCheck := false
	if runtime.GOOS == "windows" {
		res = disk.WinsatTest(language, isMultiCheck, "")
	} else {
		res = disk.FioTest(language, isMultiCheck, "")
		if res == "" {
			res = disk.DDTest(language, isMultiCheck, "")
		}
	}
	fmt.Println("--------------------------------------------------")
	fmt.Printf(res)
	fmt.Println("--------------------------------------------------")
}
