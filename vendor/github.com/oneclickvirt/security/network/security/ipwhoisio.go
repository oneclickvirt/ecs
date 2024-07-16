package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// ipwhoisIoHttpRequest 发起 HTTP 请求
func ipwhoisIoHttpRequest(url, userAgent, accept, referer string) (*req.Response, error) {
	client := req.C()
	client.Headers = make(http.Header)
	client.SetTimeout(6 * time.Second)
	client.R().
		SetRetryCount(2).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetRetryFixedInterval(2 * time.Second)
	client.SetCommonHeader("User-Agent", userAgent)
	client.SetCommonHeader("Accept", accept)
	client.SetCommonHeader("Referer", referer)
	client.SetCommonHeader("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	client.SetCommonHeader("Connection", "keep-alive")
	client.SetCommonHeader("Sec-Fetch-Dest", "empty")
	client.SetCommonHeader("Sec-Fetch-Mode", "cors")
	client.SetCommonHeader("Sec-Fetch-Site", "same-origin")
	client.SetCommonHeader("TE", "trailers")
	resp, err := client.R().Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching info: %v", err)
	}
	return resp, nil
}

// IpwhoisIo 获取 ipwhois.io 的信息
func IpwhoisIo(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityInfo := &model.SecurityInfo{}
	url := fmt.Sprintf("https://ipwhois.io/widget?ip=%s&lang=en", ip)
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:123.0) Gecko/20100101 Firefox/123.0"
	accept := "*/*"
	referer := "https://ipwhois.io/"
	resp, err := ipwhoisIoHttpRequest(url, userAgent, accept, referer)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding %s info: %v", url, err)
	}
	if securitys, ok := data["security"].(map[string]interface{}); ok {
		setSecurityInfo := func(key string, field *string) {
			if val, ok := securitys[key].(bool); ok {
				*field = utils.BoolToString(val)
			}
		}
		setSecurityInfo("anonymous", &securityInfo.IsAnonymous)
		setSecurityInfo("proxy", &securityInfo.IsProxy)
		setSecurityInfo("vpn", &securityInfo.IsVpn)
		setSecurityInfo("tor", &securityInfo.IsTor)
		setSecurityInfo("hosting", &securityInfo.IsDatacenter)
	}
	securityInfo.Tag = "6"
	return nil, securityInfo, nil
}
