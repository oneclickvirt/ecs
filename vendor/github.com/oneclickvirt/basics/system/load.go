package system

import (
	"bufio"
	"github.com/oneclickvirt/basics/system/utils"
	"github.com/shirou/gopsutil/v4/load"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// getLoadFromProc 从 /proc/loadavg 文件中获取负载信息
func getLoadFromProc() (float64, float64, float64, error) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		// 解析负载信息并转换为 float64 类型
		load1, _ := strconv.ParseFloat(fields[0], 64)
		load5, _ := strconv.ParseFloat(fields[1], 64)
		load15, _ := strconv.ParseFloat(fields[2], 64)
		return load1, load5, load15, nil
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, 0, err
	}
	return 0, 0, 0, nil
}

// 获取系统负载信息
func getSystemLoad() (float64, float64, float64, error) {
	var load1, load5, load15 float64
	var err error
	if runtime.GOOS == "linux" {
		// 尝试从 /proc/loadavg 文件获取负载信息
		load1, load5, load15, err = getLoadFromProc()
		if err != nil {
			load1, load5, load15 = 0, 0, 0
		}
	}
	// 使用 gopsutil 获取负载
	avg, err := load.Avg()
	if err != nil {
		load1, load5, load15 = 0, 0, 0
	} else {
		if avg.Load1 != 0 && avg.Load5 != 0 && avg.Load15 != 0 {
			load1, load5, load15 = avg.Load1, avg.Load5, avg.Load15
		}
	}
	// 有时候获取全为零时尝试使用 wmic 获取负载
	if runtime.GOOS == "windows" && load1 == 0 && load5 == 0 && load15 == 0 {
		load1 = utils.GetLoad1()
	}
	return load1, load5, load15, nil
}
