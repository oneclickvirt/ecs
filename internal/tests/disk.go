package tests

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/oneclickvirt/disktest/disk"
)

func DiskTest(language, testMethod, testPath string, isMultiCheck bool, autoChange bool) (realTestMethod, res string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[WARN] DiskTest panic: %v\n", r)
			res = fmt.Sprintf("\nDisk test failed: %v\n", r)
			realTestMethod = "error"
		}
	}()
	
	switch testMethod {
	case "fio":
		res = disk.FioTest(language, isMultiCheck, testPath)
		if res == "" && autoChange {
			res += disk.DDTest(language, isMultiCheck, testPath)
			realTestMethod = "dd"
		} else {
			realTestMethod = "fio"
		}
	case "dd":
		res = disk.DDTest(language, isMultiCheck, testPath)
		if res == "" && autoChange {
			res += disk.FioTest(language, isMultiCheck, testPath)
			realTestMethod = "fio"
		} else {
			realTestMethod = "dd"
		}
	default:
		if runtime.GOOS == "windows" {
			realTestMethod = "winsat"
			res = disk.WinsatTest(language, isMultiCheck, testPath)
		} else {
			res = disk.DDTest(language, isMultiCheck, testPath)
			realTestMethod = "dd"
		}
	}
	if !strings.Contains(res, "\n") && res != "" {
		res += "\n"
	}
	return
}
