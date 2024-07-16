package security

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// Scamalytics 查询 scamalytics.com 的信息 传入IPV4或者IPV6都可
func Scamalytics(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	if ip == "" {
		return nil, nil, fmt.Errorf("IP address is null")
	}
	const (
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36"
		accept    = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"
		referer   = "https://scamalytics.com"
		urlFormat = "https://scamalytics.com/ip/%s"
	)
	url := fmt.Sprintf(urlFormat, ip)
	client := req.C()
	client.Headers = make(http.Header)
	client.SetTimeout(6 * time.Second)
	client.SetCommonHeader("User-Agent", userAgent)
	client.SetCommonHeader("Accept", accept)
	client.SetCommonHeader("Referer", referer)
	resp, err := client.R().Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("can not load response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("HTTP request failed: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("can not parse body")
	}
	body := string(b)
	doc, readErr := goquery.NewDocumentFromReader(strings.NewReader(body))
	if readErr != nil {
		return nil, nil, fmt.Errorf("can not parse page: %v", readErr.Error())
	}
	score := &model.SecurityScore{
		FraudScore: new(int),
	}
	info := &model.SecurityInfo{}
	// 解析欺诈分数
	fraudScoreText := doc.Find("div.score").First().Text()
	if fraudScoreText != "" {
		fraudScore, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(fraudScoreText, "Fraud Score: ")))
		if err != nil {
			return nil, nil, fmt.Errorf("can not find fraud score: %v", err)
		}
		*score.FraudScore = fraudScore
	}
	// 解析安全信息
	doc.Find("th.title").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Proxies") {
			tr := s.Parent()
			for j := 0; j < 6 && tr != nil && tr.Is("tr"); j++ {
				text := tr.Text()
				switch {
				case strings.Contains(text, "Anonymizing VPN"):
					info.IsAnonymous = utils.ParseYesNo(text)
					info.IsVpn = info.IsAnonymous
				case strings.Contains(text, "Tor Exit Node"):
					info.IsTorExit = utils.ParseYesNo(text)
					info.IsTor = info.IsTorExit
				case strings.Contains(text, "Server"):
					info.IsDatacenter = utils.ParseYesNo(text)
				case strings.Contains(text, "Search Engine Robot"):
					info.IsBot = utils.ParseYesNo(text)
				case strings.Contains(text, "Public Proxy"), strings.Contains(text, "Web Proxy"):
					if info.IsProxy == "" || info.IsProxy == "No" {
						info.IsProxy = utils.ParseYesNo(text)
					}
				}
				tr = tr.Next()
			}
		}
	})
	score.Tag = "1"
	info.Tag = "1"
	return score, info, nil
}
