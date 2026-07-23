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
			fmt.Fprintln(os.Stderr, "[WARN] DiskTest failed")
			res = diskUnavailableMessage(language)
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
	if strings.TrimSpace(res) == "" {
		res = diskUnavailableMessage(language)
	}
	if !strings.Contains(res, "\n") && res != "" {
		res += "\n"
	}
	return
}

func diskUnavailableMessage(language string) string {
	if strings.EqualFold(strings.TrimSpace(language), "en") {
		return "Disk test unavailable\n"
	}
	return "硬盘测试不可用\n"
}
