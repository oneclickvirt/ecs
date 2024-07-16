package sp

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/imroc/req/v3"
	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/speedtest/model"
	"github.com/showwin/speedtest-go/speedtest"
)

var speedtestClient = speedtest.New(speedtest.WithUserConfig(
	&speedtest.UserConfig{
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.74 Safari/537.36",
		PingMode:       speedtest.TCP,
		MaxConnections: 8,
	}))

func getData(endpoint string) string {
	client := req.C()
	client.SetTimeout(10 * time.Second)
	client.R().
		SetRetryCount(2).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetRetryFixedInterval(2 * time.Second)
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	for _, baseUrl := range model.CdnList {
		url := baseUrl + endpoint
		resp, err := client.R().Get(url)
		if err == nil {
			defer resp.Body.Close()
			b, err := io.ReadAll(resp.Body)
			if err == nil {
				return string(b)
			}
		}
		if model.EnableLoger {
			Logger.Info(err.Error())
		}
	}
	return ""
}

func parseDataFromURL(data, url string) speedtest.Servers {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	var targets speedtest.Servers
	reader := csv.NewReader(strings.NewReader(data))
	reader.Comma = ','
	records, err := reader.ReadAll()
	if err == nil {
		if len(records) > 0 && (records[0][6] == "country_code" || records[0][1] == "country_code") {
			records = records[1:]
		}
		for _, record := range records {
			customURL := record[5]
			target, errFetch := speedtestClient.CustomServer(customURL)
			if errFetch != nil {
				if model.EnableLoger {
					Logger.Info(err.Error())
				}
				continue
			}
			target.Name = record[10] + record[7] + record[8]
			targets = append(targets, target)
		}
	}
	return targets
}

func parseDataFromID(data, url string) speedtest.Servers {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	var targets speedtest.Servers
	reader := csv.NewReader(strings.NewReader(data))
	reader.Comma = ','
	records, err := reader.ReadAll()
	if err == nil {
		if len(records) > 0 && (records[0][6] == "country_code" || records[0][1] == "country_code") {
			records = records[1:]
		}
		for _, record := range records {
			id := record[0]
			serverPtr, errFetch := speedtestClient.FetchServerByID(id)
			if errFetch != nil {
				if model.EnableLoger {
					Logger.Info(err.Error())
				}
				continue
			}
			if strings.Contains(url, "Mobile") {
				serverPtr.Name = "移动" + record[3]
			} else if strings.Contains(url, "Telecom") {
				serverPtr.Name = "电信" + record[3]
			} else if strings.Contains(url, "Unicom") {
				serverPtr.Name = "联通" + record[3]
			} else {
				serverPtr.Name = record[3]
			}
			targets = append(targets, serverPtr)
		}
	}
	return targets
}

// 计算字符串的显示宽度（考虑中文字符）
func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		if utf8.RuneLen(r) == 3 {
			// 假设每个中文字符宽度为2
			width += 2
		} else {
			width += 1
		}
	}
	return width
}

// 格式化字符串以确保左对齐
func formatString(s string, width int) string {
	displayW := displayWidth(s)
	if displayW < width {
		// 计算需要填充的空格数
		padding := width - displayW
		return s + fmt.Sprintf("%*s", padding, "")
	}
	return s
}

func ShowHead(language string) {
	headers1 := []string{"位置", "上传速度", "下载速度", "延迟", "丢包率"}
	headers2 := []string{"Location", "Upload Speed", "Download Speed", "Latency", "PacketLoss"}
	if language == "zh" {
		for _, header := range headers1 {
			fmt.Print(formatString(header, 16))
		}
		fmt.Println()
	} else if language == "en" {
		for _, header := range headers2 {
			fmt.Print(formatString(header, 16))
		}
		fmt.Println()
	}
}
