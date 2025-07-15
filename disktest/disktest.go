package disktest

import (
	"runtime"
	"strings"

	"github.com/oneclickvirt/disktest/disk"
)

func DiskTest(language, testMethod, testPath string, isMultiCheck bool, autoChange bool) (realTestMethod, res string) {
	if runtime.GOOS == "windows" {
		if testMethod != "winsat" && testMethod != "" {
			// res = "Detected host is Windows, using Winsat for testing.\n"
			realTestMethod = "winsat"
		}
		res = disk.WinsatTest(language, isMultiCheck, testPath)
	} else {
		switch testMethod {
		case "fio":
			res = disk.FioTest(language, isMultiCheck, testPath)
			if res == "" && autoChange {
				// res = "Fio test failed, switching to DD for testing.\n"
				res += disk.DDTest(language, isMultiCheck, testPath)
				realTestMethod = "dd"
			} else {
				realTestMethod = "fio"
			}
		case "dd":
			res = disk.DDTest(language, isMultiCheck, testPath)
			if res == "" && autoChange {
				// res = "DD test failed, switching to Fio for testing.\n"
				res += disk.FioTest(language, isMultiCheck, testPath)
				realTestMethod = "fio"
			} else {
				realTestMethod = "dd"
			}
		default:
			// res = "Unsupported test method specified, switching to DD for testing.\n"
			res += disk.DDTest(language, isMultiCheck, testPath)
			realTestMethod = "dd"
		}
	}
	if !strings.Contains(res, "\n") && res != "" {
		res += "\n"
	}
	return
}
