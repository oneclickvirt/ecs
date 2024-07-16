package jp

import (
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
)

// Hulu
// www.hulu.jp 或 id.hulu.jp 仅 ipv4 且 get 请求
// https://www.hulu.jp/login
func Hulu(c *http.Client) model.Result {
	name := "Hulu Japan"
	hostname := "hulu.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	headers := map[string]string{
		"User-Agent":                model.UA_Browser,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Encoding":           "gzip, deflate, br",
		"Cache-Control":             "no-cache",
		"DNT":                       "1",
		"Pragma":                    "no-cache",
		"Sec-CH-UA":                 `"Chromium";v="106", "Google Chrome";v="106", "Not;A=Brand";v="99"`,
		"Sec-CH-UA-Mobile":          "?0",
		"Sec-CH-UA-Platform":        "Windows",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().Get("https://id.hulu.jp")
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	if resp.Request.URL.Path == "/restrict.html" || resp.StatusCode == 403 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
}
