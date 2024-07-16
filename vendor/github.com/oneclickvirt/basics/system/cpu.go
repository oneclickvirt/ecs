package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/system/utils"
	"github.com/shirou/gopsutil/v4/cpu"
)

func checkCPUFeatureLinux(filename string, feature string) (string, bool) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "Error reading file", false
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, feature) {
			return "✔️ Enabled", true
		}
	}
	return "❌ Disabled", false
}

func checkCPUFeature(filename string, feature string) (string, bool) {
	if runtime.GOOS == "windows" {
		return utils.CheckCPUFeatureWindows(filename, feature)
	} else if runtime.GOOS == "linux" {
		return checkCPUFeatureLinux(filename, feature)
	}
	return "Unsupported OS", false
}

// convertBytes 转换字节数
func convertBytes(bytes int64) (string, int64) {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return "GB", bytes / GB
	case bytes >= MB:
		return "MB", bytes / MB
	case bytes >= KB:
		return "KB", bytes / KB
	default:
		return "Bytes", bytes
	}
}

func getCpuInfo(ret *model.SystemInfo, cpuType string) (*model.SystemInfo, error) {
	var aesFeature, virtFeature, hypervFeature string
	var st bool
	if runtime.NumCPU() != 0 {
		ret.CpuCores = fmt.Sprintf("%d %s CPU(s)", runtime.NumCPU(), cpuType)
	}
	if runtime.GOOS == "windows" {
		ci, err := cpu.Info()
		if err != nil {
			return nil, fmt.Errorf("cpu.Info error: %v", err.Error())
		} else {
			for i := 0; i < len(ci); i++ {
				if len(ret.CpuModel) < len(ci[i].ModelName) {
					ret.CpuModel = strings.TrimSpace(ci[i].ModelName)
				}
			}
		}
		ret.CpuCache = utils.GetCpuCache()
	} else {
		// 使用 /proc/cpuinfo 检测信息
		cpuinfoFile, err := os.Open("/proc/cpuinfo")
		if err == nil {
			scanner := bufio.NewScanner(cpuinfoFile)
			for scanner.Scan() {
				line := scanner.Text()
				fields := strings.Split(line, ":")
				if len(fields) >= 2 {
					if strings.Contains(fields[0], "model name") {
						ret.CpuModel = strings.TrimSpace(strings.Join(fields[1:], " "))
					} else if strings.Contains(fields[0], "cache size") {
						ret.CpuCache = strings.TrimSpace(strings.Join(fields[1:], " "))
					} else if strings.Contains(fields[0], "cpu MHz") && !strings.Contains(ret.CpuModel, "@") {
						ret.CpuModel += " @ " + strings.TrimSpace(strings.Join(fields[1:], " ")) + " MHz"
					}
				}
			}
		}
		defer cpuinfoFile.Close()
		// 使用 lscpu -B 检测信息
		cmd := exec.Command("lscpu", "-B") // 以字节数为单位查询
		output, err := cmd.Output()
		if err == nil {
			var L1dcache, L1icache, L1cache, L2cache, L3cache string
			outputStr := string(output)
			lines := strings.Split(outputStr, "\n")
			for _, line := range lines {
				fields := strings.Split(line, ":")
				if len(fields) >= 2 {
					if strings.Contains(fields[0], "Model name") && !strings.Contains(fields[0], "BIOS Model name") && ret.CpuModel == "" {
						ret.CpuModel = strings.TrimSpace(strings.Join(fields[1:], " "))
					} else if strings.Contains(fields[0], "CPU MHz") && !strings.Contains(ret.CpuModel, "@") {
						ret.CpuModel += " @ " + strings.TrimSpace(strings.Join(fields[1:], " ")) + " MHz"
					} else if strings.Contains(fields[0], "L1d cache") {
						L1dcache = strings.TrimSpace(strings.Join(fields[1:], " "))
					} else if strings.Contains(fields[0], "L1i cache") {
						L1icache = strings.TrimSpace(strings.Join(fields[1:], " "))
					} else if strings.Contains(fields[0], "L2 cache") {
						L2cache = strings.TrimSpace(strings.Join(fields[1:], " "))
					} else if strings.Contains(fields[0], "L3 cache") {
						L3cache = strings.TrimSpace(strings.Join(fields[1:], " "))
					}
				}
			}
			if L1dcache != "" && L1icache != "" && L2cache != "" && L3cache != "" && !strings.Contains(ret.CpuCache, "/") {
				bytes1, err1 := strconv.ParseInt(L1dcache, 10, 64)
				bytes2, err2 := strconv.ParseInt(L1icache, 10, 64)
				if err1 == nil && err2 == nil {
					bytes3 := bytes1 + bytes2
					unit, size := convertBytes(bytes3)
					L1cache = fmt.Sprintf("L1: %d %s", size, unit)
				}
				bytes4, err4 := strconv.ParseInt(L2cache, 10, 64)
				if err4 == nil {
					unit, size := convertBytes(bytes4)
					L2cache = fmt.Sprintf("L2: %d %s", size, unit)
				}
				bytes5, err5 := strconv.ParseInt(L3cache, 10, 64)
				if err5 == nil {
					unit, size := convertBytes(bytes5)
					L3cache = fmt.Sprintf("L3: %d %s", size, unit)
				}
				if err1 == nil && err2 == nil && err4 == nil && err5 == nil {
					ret.CpuCache = L1cache + " / " + L2cache + " / " + L3cache
				}
			}
		}
	}
	// 使用 /proc/device-tree 获取信息 - 特化适配嵌入式系统
	deviceTreeContent, err := os.ReadFile("/proc/device-tree")
	if err == nil {
		ret.CpuModel = string(deviceTreeContent)
	}
	// 获取虚拟化架构
	if runtime.GOOS == "windows" {
		aesFeature = `HARDWARE\DESCRIPTION\System\CentralProcessor\0`
		virtFeature = `HARDWARE\DESCRIPTION\System\CentralProcessor\0`
		hypervFeature = `SYSTEM\CurrentControlSet\Control\Hypervisor\0`
	} else if runtime.GOOS == "linux" {
		aesFeature = "/proc/cpuinfo"
		virtFeature = "/proc/cpuinfo"
		hypervFeature = "/proc/cpuinfo"
	}
	ret.CpuAesNi, _ = checkCPUFeature(aesFeature, "aes")
	ret.CpuVAH, st = checkCPUFeature(virtFeature, "vmx")
	if !st {
		ret.CpuVAH, _ = checkCPUFeature(hypervFeature, "hypervisor")
	}
	// 使用 sysctl 获取信息 - 特化适配 freebsd openbsd 系统
	if checkSysctlVersion() {
		if ret.CpuModel == "" {
			cname, err := getSysctlValue("hw.model")
			if err == nil && !strings.Contains(cname, "cannot") {
				ret.CpuModel = cname
				// 获取CPU频率
				freq, err := getSysctlValue("dev.cpu.0.freq")
				if err == nil && !strings.Contains(freq, "cannot") {
					ret.CpuModel += " @" + freq + "MHz"
				}
			}
		}
		if ret.CpuCores == "" {
			cores, err := getSysctlValue("hw.ncpu")
			if err == nil && !strings.Contains(cores, "cannot") {
				ret.CpuCores = fmt.Sprintf("%s %s CPU(s)", cores, cpuType)
			}
		}
		if ret.CpuCache == "" {
			// 获取CPU缓存配置
			ccache, err := getSysctlValue("hw.cacheconfig")
			if err == nil && !strings.Contains(ccache, "cannot") {
				ret.CpuCache = strings.TrimSpace(strings.Split(ccache, ":")[1])
			}
		}
		aesOut, err := exec.Command("sysctl", "-a").Output()
		if ret.CpuAesNi == "Unsupported OS" || ret.CpuAesNi == "" {
			// 检查AES指令集支持
			var CPU_AES string
			if err == nil {
				aesReg := regexp.MustCompile(`crypto\.aesni\s*=\s*(\d)`)
				aesMatch := aesReg.FindStringSubmatch(string(aesOut))
				if len(aesMatch) > 1 {
					CPU_AES = aesMatch[1]
				}
				if CPU_AES != "" {
					ret.CpuAesNi = "✔️ Enabled"
				} else {
					ret.CpuAesNi = "❌ Disabled"
				}
			}
		}
		if ret.CpuVAH == "Unsupported OS" || ret.CpuVAH == "" {
			// 检查虚拟化支持
			var CPU_VIRT string
			if err == nil {
				virtReg := regexp.MustCompile(`(hw\.vmx|hw\.svm)\s*=\s*(\d)`)
				virtMatch := virtReg.FindStringSubmatch(string(aesOut))
				if len(virtMatch) > 2 {
					CPU_VIRT = virtMatch[2]
				}
				if CPU_VIRT != "" {
					ret.CpuVAH = "✔️ Enabled"
				} else {
					ret.CpuVAH = "❌ Disabled"
				}
			}
		}
		if ret.Uptime == "" {
			// 获取系统运行时间
			boottimeStr, err := getSysctlValue("kern.boottime")
			if err == nil {
				boottimeReg := regexp.MustCompile(`sec = (\d+), usec = (\d+)`)
				boottimeMatch := boottimeReg.FindStringSubmatch(boottimeStr)
				if len(boottimeMatch) > 1 {
					boottime, err := strconv.ParseInt(boottimeMatch[1], 10, 64)
					if err == nil {
						uptime := time.Now().Unix() - boottime
						days := uptime / 86400
						hours := (uptime % 86400) / 3600
						minutes := (uptime % 3600) / 60
						ret.Uptime = fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
					}
				}
			}
		}
		if ret.Load == "" {
			// 获取系统负载
			var load string
			out, err := exec.Command("w").Output()
			if err == nil {
				loadFields := strings.Fields(string(out))
				load = loadFields[len(loadFields)-3] + " " + loadFields[len(loadFields)-2] + " " + loadFields[len(loadFields)-1]
			} else {
				out, err = exec.Command("uptime").Output()
				if err == nil {
					fields := strings.Fields(string(out))
					load = fields[len(fields)-3] + " " + fields[len(fields)-2] + " " + fields[len(fields)-1]
				}
			}
			if load != "" {
				ret.Load = load
			}
		}
	}
	// MAC需要额外获取信息进行判断
	if runtime.GOOS == "darwin" {
		if len(model.MacOSInfo) > 0 {
			for _, line := range model.MacOSInfo {
				if strings.Contains(line, "Chip") {
					ret.CpuModel = strings.TrimSpace(strings.Split(line, ":")[1])
				}
				if strings.Contains(line, "Total Number of Cores") {
					ret.CpuCores = strings.TrimSpace(strings.Split(line, ":")[1])
				}
				if strings.Contains(line, "Memory") {
					ret.MemoryTotal = strings.TrimSpace(strings.Split(line, ":")[1])
				}
			}
		}
	}
	return ret, nil
}
