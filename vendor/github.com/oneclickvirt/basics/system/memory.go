package system

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/oneclickvirt/basics/model"
	"github.com/shirou/gopsutil/v4/mem"
)

func getMemoryInfo() (string, string, string, string, string, string) {
	var memoryTotalStr, memoryUsageStr, swapTotalStr, swapUsageStr, virtioBalloonStatus, KernelSamepageMerging string
	mv, err := mem.VirtualMemory()
	if err != nil {
		println("mem.VirtualMemory error:", err)
	} else {
		memoryTotal := float64(mv.Total)
		memoryUsage := float64(mv.Total - mv.Available)
		if memoryTotal < 1024*1024*1024 {
			memoryTotalStr = strconv.FormatFloat(memoryTotal/(1024*1024), 'f', 2, 64) + " MB"
		} else {
			memoryTotalStr = strconv.FormatFloat(memoryTotal/(1024*1024*1024), 'f', 2, 64) + " GB"
		}
		if memoryUsage < 1024*1024*1024 {
			memoryUsageStr = strconv.FormatFloat(memoryUsage/(1024*1024), 'f', 2, 64) + " MB"
		} else {
			memoryUsageStr = strconv.FormatFloat(memoryUsage/(1024*1024*1024), 'f', 2, 64) + " GB"
		}
		if runtime.GOOS != "windows" {
			swapTotal := float64(mv.SwapTotal)
			swapUsage := float64(mv.SwapTotal - mv.SwapFree)
			if swapTotal != 0 {
				if swapTotal < 1024*1024*1024 {
					swapTotalStr = strconv.FormatFloat(swapTotal/(1024*1024), 'f', 2, 64) + " MB"
				} else {
					swapTotalStr = strconv.FormatFloat(swapTotal/(1024*1024*1024), 'f', 2, 64) + " GB"
				}
				if swapUsage < 1024*1024*1024 {
					swapUsageStr = strconv.FormatFloat(swapUsage/(1024*1024), 'f', 2, 64) + " MB"
				} else {
					swapUsageStr = strconv.FormatFloat(swapUsage/(1024*1024*1024), 'f', 2, 64) + " GB"
				}
			}
		}
	}
	// MAC需要额外获取信息进行判断
	if runtime.GOOS == "darwin" {
		if len(model.MacOSInfo) > 0 {
			for _, line := range model.MacOSInfo {
				if strings.Contains(line, "Memory") {
					memoryTotalStr = strings.TrimSpace(strings.Split(line, ":")[1])
				}
			}
		}
	}
	if runtime.GOOS == "windows" {
		// gopsutil 在 Windows 下不能正确取 swap
		ms, err := mem.SwapMemory()
		if err != nil {
			println("mem.SwapMemory error:", err)
		} else {
			swapTotal := float64(ms.Total)
			swapUsage := float64(ms.Used)
			if swapTotal != 0 {
				if swapTotal < 1024*1024*1024 {
					swapTotalStr = strconv.FormatFloat(swapTotal/(1024*1024), 'f', 2, 64) + " MB"
				} else {
					swapTotalStr = strconv.FormatFloat(swapTotal/(1024*1024*1024), 'f', 2, 64) + " GB"
				}
				if swapUsage < 1024*1024*1024 {
					swapUsageStr = strconv.FormatFloat(swapUsage/(1024*1024), 'f', 2, 64) + " MB"
				} else {
					swapUsageStr = strconv.FormatFloat(swapUsage/(1024*1024*1024), 'f', 2, 64) + " GB"
				}
			}
		}
	}
	virtioBalloon, err := os.ReadFile("/proc/modules")
	if err == nil {
		if strings.Contains(string(virtioBalloon), "virtio_balloon") {
			virtioBalloonStatus = "✔️ Enabled"
		} else {
			virtioBalloonStatus = ""
		}
	} else {
		virtioBalloonStatus = ""
	}
	ksmStatus, err := os.ReadFile("/sys/kernel/mm/ksm/run")
	if err == nil {
		if strings.Contains(string(ksmStatus), "1") {
			KernelSamepageMerging = "✔️ Enabled"
		} else {
			KernelSamepageMerging = ""
		}
	} else {
		KernelSamepageMerging = ""
	}
	return memoryTotalStr, memoryUsageStr, swapTotalStr, swapUsageStr, virtioBalloonStatus, KernelSamepageMerging
}
