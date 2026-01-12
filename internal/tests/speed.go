package tests

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/oneclickvirt/privatespeedtest/pst"
	"github.com/oneclickvirt/speedtest/model"
	"github.com/oneclickvirt/speedtest/sp"
)

func ShowHead(language string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[WARN] ShowHead panic: %v\n", r)
		}
	}()
	sp.ShowHead(language)
}

func NearbySP() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[WARN] NearbySP panic: %v\n", r)
		}
	}()
	if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
		sp.NearbySpeedTest()
	} else {
		sp.OfficialNearbySpeedTest()
	}
}

// formatString 格式化字符串到指定宽度
func formatString(s string, width int) string {
	return fmt.Sprintf("%-*s", width, s)
}

// printTableRow 打印表格行
func printTableRow(result pst.SpeedTestResult) {
	location := result.City
	if result.CarrierType != "" {
		carrier := result.CarrierType
		switch carrier {
		case "Telecom":
			carrier = "电信"
		case "Unicom":
			carrier = "联通"
		case "Mobile":
			carrier = "移动"
		case "Other":
			carrier = "其他"
		}
		location = fmt.Sprintf("%s%s", carrier, result.City)
	}
	if len(location) > 14 {
		location = location[:14] + "..."
	}
	upload := "N/A"
	if result.UploadMbps > 0 {
		upload = fmt.Sprintf("%.2f Mbps", result.UploadMbps)
	}
	download := "N/A"
	if result.DownloadMbps > 0 {
		download = fmt.Sprintf("%.2f Mbps", result.DownloadMbps)
	}
	latency := fmt.Sprintf("%.2f ms", result.PingLatency.Seconds()*1000)
	packetLoss := "N/A"
	fmt.Print(formatString(location, 16))
	fmt.Print(formatString(upload, 16))
	fmt.Print(formatString(download, 16))
	fmt.Print(formatString(latency, 16))
	fmt.Print(formatString(packetLoss, 16))
	fmt.Println()
}

// privateSpeedTest 使用 privatespeedtest 进行单个运营商测速
// operator 参数：只支持 "cmcc"、"cu"、"ct"
func privateSpeedTest(num int, operator string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[WARN] privateSpeedTest panic: %v\n", r)
		}
	}()
	*pst.NoProgress = true
    *pst.Quiet = true
    *pst.NoHeader = true
    *pst.NoProjectURL = true
	// 加载服务器列表
	serverList, err := pst.LoadServerList()
	if err != nil {
		return fmt.Errorf("加载自定义服务器列表失败")
	}
	// 使用三网测速模式（每个运营商选择指定数量的最低延迟节点）
	serversPerISP := num
	if serversPerISP <= 0 || serversPerISP > 5{
		serversPerISP = 2
	}
	// 单个运营商测速：先过滤服务器列表
	var carrierType string
	switch strings.ToLower(operator) {
	case "cmcc":
		carrierType = "Mobile"
	case "cu":
		carrierType = "Unicom"
	case "ct":
		carrierType = "Telecom"
	default:
		return fmt.Errorf("不支持的运营商类型: %s", operator)
	}
	// 过滤出指定运营商的服务器
	filteredServers := pst.FilterServersByISP(serverList.Servers, carrierType)
	// 先找足够多的候选服务器用于去重（找 serversPerISP * 3 个，确保去重后还能剩下足够的服务器）
	candidateCount := serversPerISP * 3
	if candidateCount > len(filteredServers) {
		candidateCount = len(filteredServers)
	}
	// 使用 FindBestServers 选择最佳服务器
	candidateServers, err := pst.FindBestServers(
		filteredServers,
		candidateCount,   // 选择更多候选节点用于去重
		5*time.Second,    // ping 超时
		true,             // 显示进度条
		true,             // 静默
	)
	if err != nil {
		return fmt.Errorf("分组查找失败")
	}
	// 去重：确保同一运营商内城市不重复
	seenCities := make(map[string]bool)
	var bestServers []pst.ServerWithLatencyInfo
	for _, serverInfo := range candidateServers {
		city := serverInfo.Server.City
		if city == "" {
			city = "Unknown"
		}
		if !seenCities[city] {
			seenCities[city] = true
			bestServers = append(bestServers, serverInfo)
			// 去重后取前 serversPerISP 个
			if len(bestServers) >= serversPerISP {
				break
			}
		}
	}
	if len(bestServers) == 0 {
		return fmt.Errorf("去重后没有可用的服务器")
	}
	// 执行测速并逐个打印结果（不打印表头）
	for i, serverInfo := range bestServers {
		result := pst.RunSpeedTest(
			serverInfo.Server,
			false,            // 不禁用下载测试
			false,            // 不禁用上传测试
			6,                // 并发线程数
			12*time.Second,   // 超时时间
			&serverInfo,
			false,             // 不显示进度条
		)
		if result.Success {
			printTableRow(result)
		}
		// 在测试之间暂停（除了最后一个）
		if i < len(bestServers)-1 {
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

func CustomSP(platform, operator string, num int, language string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[WARN] CustomSP panic: %v\n", r)
		}
	}()
	// 对于三网测速（cmcc、cu、ct），优先使用 privatespeedtest 进行私有测速
	opLower := strings.ToLower(operator)
	if opLower == "cmcc" || opLower == "cu" || opLower == "ct" {
		err := privateSpeedTest(num, opLower)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[WARN] privatespeedtest failed\n")
			// 继续使用原有的兜底方案
		} else {
			// 测速成功，直接返回
			return
		}
	}
	var url, parseType string
	if strings.ToLower(platform) == "cn" {
		if strings.ToLower(operator) == "cmcc" {
			url = model.CnCMCC
		} else if strings.ToLower(operator) == "cu" {
			url = model.CnCU
		} else if strings.ToLower(operator) == "ct" {
			url = model.CnCT
		} else if strings.ToLower(operator) == "hk" {
			url = model.CnHK
		} else if strings.ToLower(operator) == "tw" {
			url = model.CnTW
		} else if strings.ToLower(operator) == "jp" {
			url = model.CnJP
		} else if strings.ToLower(operator) == "sg" {
			url = model.CnSG
		}
		parseType = "url"
	} else if strings.ToLower(platform) == "net" {
		if strings.ToLower(operator) == "cmcc" {
			url = model.NetCMCC
		} else if strings.ToLower(operator) == "cu" {
			url = model.NetCU
		} else if strings.ToLower(operator) == "ct" {
			url = model.NetCT
		} else if strings.ToLower(operator) == "hk" {
			url = model.NetHK
		} else if strings.ToLower(operator) == "tw" {
			url = model.NetTW
		} else if strings.ToLower(operator) == "jp" {
			url = model.NetJP
		} else if strings.ToLower(operator) == "sg" {
			url = model.NetSG
		} else if strings.ToLower(operator) == "global" {
			url = model.NetGlobal
		}
		parseType = "id"
	}
	if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
		sp.CustomSpeedTest(url, parseType, num, language)
	} else {
		sp.OfficialCustomSpeedTest(url, parseType, num, language)
	}
}
