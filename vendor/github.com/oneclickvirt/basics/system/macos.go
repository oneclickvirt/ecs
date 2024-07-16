package system

import (
	"os/exec"
	"strings"

	"github.com/oneclickvirt/basics/model"
)

func getMacOSInfo() {
	out, err := exec.Command("system_profiler", "SPHardwareDataType").Output()
	if err == nil && !strings.Contains(string(out), "error") {
		model.MacOSInfo = strings.Split(string(out), "\n")
	}
}
