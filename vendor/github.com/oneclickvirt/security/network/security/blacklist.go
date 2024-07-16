package security

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/basics/model"
	. "github.com/oneclickvirt/defaultset"
)

// buildAllBlackList 获取 https://multirbl.valli.org/list/ 所有黑名单查询地址的列表
func buildAllBlackList() []string {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	url := "https://multirbl.valli.org/list/"
	client := req.C()
	client.SetTimeout(6 * time.Second)
	resp, err := client.R().Get(url)
	if err != nil {
		if model.EnableLoger {
			Logger.Info(err.Error())
		}
		return nil
	}
	if resp.StatusCode != 200 {
		if model.EnableLoger {
			Logger.Info(fmt.Sprintf("failed to fetch the URL: %s", resp.Status))
		}
		return nil
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		if model.EnableLoger {
			Logger.Info(fmt.Sprintf("failed to fetch the URL body: %s", resp.Status))
		}
		return nil
	}
	body := string(b)
	lines := strings.Split(body, "\n")
	var AllBlackList []string
	var aliveCount string
	for _, line := range lines {
		if strings.Contains(line, "<td>") {
			tdContent := strings.ReplaceAll(strings.TrimSpace(strings.Split(line, "<td>")[3]), "</td>", "")
			AllBlackList = append(AllBlackList, tdContent)
		} else if strings.Contains(line, "alive (") {
			aliveCount = strings.ReplaceAll(line, "<h2>alive (", "")
			aliveCount = strings.ReplaceAll(aliveCount, ")</h2>", "")
			aliveCount = strings.TrimSpace(aliveCount)
		}
	}
	var filteredList []string
	for _, item := range AllBlackList {
		if !strings.Contains(item, "(hidden)") {
			filteredList = append(filteredList, item)
		}
	}
	//fmt.Println(aliveCount)
	//fmt.Println("Filtered List:")
	//for _, item := range filteredList {
	//	fmt.Println(item)
	//}
	aliveCountInt, err := strconv.Atoi(aliveCount)
	if err == nil {
		return filteredList[:aliveCountInt]
	} else {
		return nil
	}
}

func checkDNSBL(ipToCheck string, AllBlackList []string, ipType string) (total, clean, blacklisted, other int) {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	var mu sync.Mutex
	reversedIP := reverseIP(ipToCheck)
	var wg sync.WaitGroup
	wg.Add(len(AllBlackList))
	for _, domain := range AllBlackList {
		go func(domain string) {
			defer wg.Done()
			var ips []net.IP
			var err error
			if ipType == "ipv4" {
				ips, err = net.LookupIP(fmt.Sprintf("%s.%s", reversedIP, domain))
			} else if ipType == "ipv6" {
				ip := net.ParseIP(ipToCheck)
				ips = append(ips, ip)
			} else {
				if model.EnableLoger {
					Logger.Info(fmt.Sprint("Invalid IpType specified"))
				}
				return
			}
			if err != nil {
				//fmt.Println("Error performing DNS lookup:", err)
				return
			}
			result := "Other"
			for _, ip := range ips {
				if ip.To4() != nil && ip.String() == "127.0.0.2" {
					result = "Blacklisted"
					break
				} else if ip.To4() == nil && ip.String() == "" {
					result = "Clean"
					break
				}
			}
			mu.Lock()
			defer mu.Unlock()
			switch result {
			case "Clean":
				total++
				clean++
			case "Blacklisted":
				total++
				blacklisted++
			default:
				total++
				other++
			}
		}(domain)
	}
	wg.Wait()

	return total, clean, blacklisted, other
}

func reverseIP(ip string) string {
	parts := strings.Split(ip, ".")
	reversedParts := make([]string, len(parts))
	for i := range parts {
		reversedParts[len(parts)-1-i] = parts[i]
	}
	return strings.Join(reversedParts, ".")
}

func BlackList(ipInfo *model.IpInfo, ipType string, language string) string {
	var result string
	AllBlackList := buildAllBlackList()
	if AllBlackList != nil && len(AllBlackList) > 0 && ipInfo != nil && ipInfo.Ip != "" {
		_, clean, blacklisted, other := checkDNSBL(ipInfo.Ip, AllBlackList, ipType)
		//fmt.Printf("Total: %d\n", len(AllBlackList))
		//fmt.Printf("Clean: %d\n", clean)
		//fmt.Printf("Blacklisted: %d\n", blacklisted)
		//fmt.Printf("Other: %d\n", other)
		var head string
		if language == "zh" {
			head = "DNS-黑名单: "
		} else {
			head = "DNS-BlackList: "
		}
		result = head + fmt.Sprintf("%d(Total_Check) ", len(AllBlackList)) +
			fmt.Sprintf("%d(Clean) ", clean) + fmt.Sprintf("%d(Blacklisted) ", blacklisted) +
			fmt.Sprintf("%d(Other) ", other) + "\n"
	}
	return result
}
