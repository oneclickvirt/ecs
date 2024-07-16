package disk

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	. "github.com/oneclickvirt/defaultset"
)

// 获取硬盘性能数据
func getDiskPerformance(device string) string {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	cmd := exec.Command("winsat", "disk", "-drive", device)
	output, err := cmd.Output()
	if err != nil {
		if EnableLoger {
			Logger.Info("cannot match winsat command: " + err.Error())
		}
		return ""
	}
	var result string
	tempList := strings.Split(string(output), "\n")
	for _, l := range tempList {
		if strings.Contains(l, "> Disk  Random 16.0 Read") {
			// 随机读取速度
			tempText := strings.TrimSpace(strings.ReplaceAll(l, "> Disk  Random 16.0 Read", ""))
			if tempText != "" {
				tpList := strings.Split(tempText, "MB/s")
				result += fmt.Sprintf("%-20s", strings.TrimSpace(tpList[0]+"MB/s["+strings.TrimSpace(tpList[len(tpList)-1])+"]")) + "    "
			}
		} else if strings.Contains(l, "> Disk  Sequential 64.0 Read") {
			// 顺序读取速度
			tempText := strings.TrimSpace(strings.ReplaceAll(l, "> Disk  Sequential 64.0 Read", ""))
			if tempText != "" {
				tpList := strings.Split(tempText, "MB/s")
				result += fmt.Sprintf("%-20s", strings.TrimSpace(tpList[0]+"MB/s["+strings.TrimSpace(tpList[len(tpList)-1])+"]")) + "    "
			}
		} else if strings.Contains(l, "> Disk  Sequential 64.0 Write") {
			// 顺序写入速度
			tempText := strings.TrimSpace(strings.ReplaceAll(l, "> Disk  Sequential 64.0 Write", ""))
			if tempText != "" {
				tpList := strings.Split(tempText, "MB/s")
				result += fmt.Sprintf("%-20s", strings.TrimSpace(tpList[0]+"MB/s["+strings.TrimSpace(tpList[len(tpList)-1])+"]")) + "    "
			}
		}
	}
	result += "\n"
	return result
}

// isWritableMountpoint 检测挂载点是否为文件夹且可写入文件
func isWritableMountpoint(path string) bool {
	if EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	// 检测 mountpoint 是否是一个文件夹
	info, err := os.Stat(path)
	if err != nil {
		if EnableLoger {
			Logger.Info("cannot stat path: " + err.Error())
		}
		return false
	}
	if !info.IsDir() {
		if EnableLoger {
			Logger.Info("path is not a directory: " + path)
		}
		return false
	}
	// 尝试打开文件进行写入
	file, err := os.OpenFile(path+"/.temp_write_check", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		if EnableLoger {
			Logger.Info("cannot open file for writing: " + err.Error())
		}
		return false
	}
	defer file.Close()
	// 删除临时文件
	err = os.Remove(path + "/.temp_write_check")
	if err != nil {
		if EnableLoger {
			Logger.Info("cannot remove temporary file: " + err.Error())
		}
	}
	return true
}

// parseResultDD 提取dd测试的结果
func parseResultDD(tempText, blockCount string) string {
	var result string
	tp1 := strings.Split(tempText, "\n")
	var records, usageTime float64
	records, _ = strconv.ParseFloat(blockCount, 64)
	for _, t := range tp1 {
		if strings.Contains(t, "bytes") {
			// t 为 104857600 bytes (105 MB, 100 MiB) copied, 4.67162 s, 22.4 MB/s
			tp2 := strings.Split(t, ",")
			if len(tp2) == 4 {
				usageTime, _ = strconv.ParseFloat(strings.Split(strings.TrimSpace(tp2[2]), " ")[0], 64)
				ioSpeed := strings.Split(strings.TrimSpace(tp2[3]), " ")[0]
				ioSpeedFlat := strings.Split(strings.TrimSpace(tp2[3]), " ")[1]
				iops := records / usageTime
				var iopsText string
				if iops >= 1000 {
					iopsText = strconv.FormatFloat(iops/1000, 'f', 2, 64) + "K IOPS, " + strconv.FormatFloat(usageTime, 'f', 2, 64) + "s"
				} else {
					iopsText = strconv.FormatFloat(iops, 'f', 2, 64) + " IOPS, " + strconv.FormatFloat(usageTime, 'f', 2, 64) + "s"
				}
				result += fmt.Sprintf("%-30s", strings.TrimSpace(ioSpeed)+" "+ioSpeedFlat+"("+iopsText+")") + "    "
			}
		}
	}
	return result
}

// formatIOPS 转换fio的测试中的IOPS的值
// rawType 支持 string 或 int
func formatIOPS(raw interface{}, rawType string) string {
	// 确保 raw 值不为空，如果为空则返回空字符串
	var iops int
	var err error
	switch v := raw.(type) {
	case string:
		if v == "" {
			return ""
		}
		// 将 raw 字符串转换为整数
		iops, err = strconv.Atoi(v)
		if err != nil {
			return ""
		}
	case int:
		iops = v
	default:
		return ""
	}
	// 检查 IOPS 速度是否大于等于 10k
	if iops >= 10000 {
		// 将原始结果除以 1k
		result := float64(iops) / 1000.0
		// 将格式化后的结果保留一位小数（例如 x.x）
		resultStr := fmt.Sprintf("%.1fk", result)
		return resultStr
	}
	// 如果 IOPS 速度小于等于 1k，则返回原始值
	if rawType == "string" {
		return raw.(string)
	} else {
		return fmt.Sprintf("%d", iops)
	}
}

// formatSpeed 转换fio的测试中的TEST的值
// rawType 支持 string 或 float64
func formatSpeed(raw interface{}, rawType string) string {
	var rawFloat float64
	var err error
	// 根据 rawType 确定如何处理 raw 的类型
	switch v := raw.(type) {
	case string:
		if v == "" {
			return ""
		}
		// 将 raw 字符串转换为 float64
		rawFloat, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return ""
		}
	case float64:
		rawFloat = v
	default:
		return ""
	}
	// 初始化结果相关变量
	var resultFloat float64 = rawFloat
	var denom float64 = 1
	unit := "KB/s"
	// 根据速度大小确定单位
	if rawFloat >= 1000000 {
		denom = 1000000
		unit = "GB/s"
	} else if rawFloat >= 1000 {
		denom = 1000
		unit = "MB/s"
	}
	// 根据单位除以相应的分母以得到格式化后的结果
	resultFloat /= denom
	// 将格式化结果保留两位小数
	result := fmt.Sprintf("%.2f", resultFloat)
	// 将格式化结果值与单位拼接并返回结果
	return strings.Join([]string{result, unit}, " ")
}
