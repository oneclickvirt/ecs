package tests

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/oneclickvirt/disktest/disk"
)

type legacyDiskMethod func(string, bool, string) string

func DiskTest(language, testMethod, testPath string, isMultiCheck bool, autoChange bool) (realTestMethod, res string) {
	return diskTestWithMethods(language, testMethod, testPath, isMultiCheck, autoChange, disk.FioTest, disk.DDTest)
}

func diskTestWithMethods(language, testMethod, testPath string, isMultiCheck bool, autoChange bool, fioTest, ddTest legacyDiskMethod) (realTestMethod, res string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "[WARN] DiskTest failed")
			res = selectedDiskFailureText(language, "")
			realTestMethod = "error"
		}
	}()

	switch testMethod {
	case "fio":
		realTestMethod = "fio"
		res = fioTest(language, isMultiCheck, testPath)
		if !legacyDiskResultUsable(res) && autoChange {
			primary := res
			fallback := ddTest(language, isMultiCheck, testPath)
			if legacyDiskResultUsable(fallback) {
				res, realTestMethod = fallback, "dd"
			} else {
				res = legacyDiskFailureText(language, true, primary, fallback)
			}
		}
	case "dd":
		realTestMethod = "dd"
		res = ddTest(language, isMultiCheck, testPath)
		if !legacyDiskResultUsable(res) && autoChange {
			primary := res
			fallback := fioTest(language, isMultiCheck, testPath)
			if legacyDiskResultUsable(fallback) {
				res, realTestMethod = fallback, "fio"
			} else {
				res = legacyDiskFailureText(language, true, primary, fallback)
			}
		}
	default:
		if runtime.GOOS == "windows" {
			realTestMethod = "winsat"
			res = disk.WinsatTest(language, isMultiCheck, testPath)
		} else {
			res = ddTest(language, isMultiCheck, testPath)
			realTestMethod = "dd"
		}
	}
	if !legacyDiskResultUsable(res) && !containsDiskFailureMessage(language, res) {
		res = selectedDiskFailureText(language, res)
	}
	if !strings.Contains(res, "\n") && res != "" {
		res += "\n"
	}
	return
}

func diskUnavailableMessage(language string) string {
	return selectedDiskFailureMessage(language) + "\n"
}

func selectedDiskFailureMessage(language string) string {
	if strings.EqualFold(strings.TrimSpace(language), "en") {
		return "Disk benchmark returned no usable data."
	}
	return "磁盘测试未返回可用的性能数据。"
}

func bothDiskMethodsFailureMessage(language string) string {
	if strings.EqualFold(strings.TrimSpace(language), "en") {
		return "DD and FIO returned no usable benchmark data."
	}
	return "DD和FIO均未返回可用的性能数据。"
}

func selectedDiskFailureText(language, output string) string {
	return legacyDiskFailureText(language, false, output)
}

func legacyDiskFailureText(language string, bothAttempted bool, outputs ...string) string {
	var retained string
	for _, output := range outputs {
		if strings.TrimSpace(output) != "" {
			retained = output
			break
		}
	}
	if retained != "" && !strings.HasSuffix(retained, "\n") {
		retained += "\n"
	}
	message := selectedDiskFailureMessage(language)
	if bothAttempted {
		message = bothDiskMethodsFailureMessage(language)
	}
	return retained + message + "\n"
}

func containsDiskFailureMessage(language, output string) bool {
	return strings.Contains(output, selectedDiskFailureMessage(language)) || strings.Contains(output, bothDiskMethodsFailureMessage(language))
}

func legacyDiskResultUsable(output string) bool {
	normalized := strings.ToLower(output)
	for _, unit := range []string{"kb/s", "mb/s", "gb/s", "tb/s", "kib/s", "mib/s", "gib/s", "tib/s"} {
		if strings.Contains(normalized, unit) {
			return true
		}
	}
	return false
}
