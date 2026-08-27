//go:build !ecs_public

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
			fmt.Fprintln(os.Stderr, "[WARN] speedtest header unavailable")
		}
	}()
	sp.ShowHead(language)
}

func NearbySP() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "[WARN] nearby speedtest unavailable")
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
	if len(location) > 15 {
		location = location[:15]
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
	fmt.Print(formatString(location, 15))
	fmt.Print(formatString(upload, 16))
	fmt.Print(formatString(download, 16))
	fmt.Print(formatString(latency, 16))
	fmt.Print(formatString(packetLoss, 16))
	fmt.Println()
}

func privateSpeedTest(num int, operator string) (testedCount int, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "[WARN] speedtest registry unavailable")
			testedCount = 0
			err = fmt.Errorf("私有测速运行失败")
		}
	}()
	*pst.NoProgress = true
	*pst.Quiet = true
	*pst.NoHeader = true
	*pst.NoProjectURL = true
	serverList, err := privateSpeedServerList()
	if err != nil {
		return 0, fmt.Errorf("加载自定义服务器列表失败")
	}
	serversPerISP := num
	if serversPerISP <= 0 || serversPerISP > 5 {
		serversPerISP = 2
	}
	var carrierType string
	switch strings.ToLower(operator) {
	case "cmcc":
		carrierType = "Mobile"
	case "cu":
		carrierType = "Unicom"
	case "ct":
		carrierType = "Telecom"
	case "other":
		carrierType = "Other"
	default:
		return 0, fmt.Errorf("不支持的运营商类型: %s", operator)
	}
	filteredServers := pst.FilterServersByISP(serverList.Servers, carrierType)
	candidateServers, err := pst.FindBestServers(
		filteredServers,
		len(filteredServers),
		5*time.Second, // ping 超时
		true,          // 显示进度条
		true,          // 静默
	)
	if err != nil {
		return 0, fmt.Errorf("分组查找失败")
	}
	bestServers := selectPrivateSpeedCandidates(candidateServers, serversPerISP)
	if len(bestServers) == 0 {
		return 0, fmt.Errorf("去重后没有可用的服务器")
	}
	for i, serverInfo := range bestServers {
		if testedCount >= serversPerISP {
			break
		}
		result := pst.RunSpeedTest(
			serverInfo.Server,
			false,          // 不禁用下载测试
			false,          // 不禁用上传测试
			6,              // 并发线程数
			12*time.Second, // 超时时间
			&serverInfo,
			false, // 不显示进度条
		)
		if result.UploadMbps > 0 || result.DownloadMbps > 0 {
			printTableRow(result)
			testedCount++
		}
		if testedCount < serversPerISP && i < len(bestServers)-1 {
			time.Sleep(1 * time.Second)
		}
	}
	// 返回实际成功输出的节点数量
	return testedCount, nil
}

func selectPrivateSpeedCandidates(candidates []pst.ServerWithLatencyInfo, serversPerISP int) []pst.ServerWithLatencyInfo {
	if serversPerISP <= 0 {
		return nil
	}
	return pst.SelectDistinctCityServers(candidates, serversPerISP*2)
}

func privateSpeedTestWithFallback(num int, operator, language string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "[WARN] preferred speedtest unavailable; using fallback")
		}
	}()
	testedCount, err := privateSpeedTest(num, operator)
	if err != nil || testedCount == 0 {
		var url, parseType string
		url = model.NetGlobal
		parseType = "id"
		if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
			sp.CustomSpeedTest(url, parseType, num, language)
		} else {
			sp.OfficialCustomSpeedTest(url, parseType, num, language)
		}
	}
}

func CustomSP(platform, operator string, num int, language string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "[WARN] custom speedtest unavailable")
		}
	}()
	opLower := strings.ToLower(operator)
	if opLower == "cmcc" || opLower == "cu" || opLower == "ct" || opLower == "other" {
		testedCount, err := privateSpeedTest(num, opLower)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[WARN] preferred speedtest unavailable; using fallback")
		} else if testedCount >= num {
			return
		} else if testedCount > 0 {
			fmt.Fprintf(os.Stderr, "[INFO] 私有节点仅测试了 %d 个，补充 %d 个公共节点\n", testedCount, num-testedCount)
			num = num - testedCount
		} else {
			// testedCount == 0，继续使用公共节点
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
		} else if strings.ToLower(operator) == "global" || strings.ToLower(operator) == "other" {
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
