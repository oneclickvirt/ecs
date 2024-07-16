package system

import (
	"os/exec"
	"strings"
)

func checkSysctlVersion() bool {
	out, err := exec.Command("sysctl", "--version").Output()
	if err != nil {
		return false
	}
	if strings.Contains(string(out), "error") {
		return false
	}
	return true
}

func getSysctlValue(key string) (string, error) {
	out, err := exec.Command("sysctl", "-n", key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
