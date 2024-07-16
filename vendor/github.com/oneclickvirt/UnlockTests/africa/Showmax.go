package africa

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// Showmax
// www.showmax.com 双栈 且 get 请求
func Showmax(c *http.Client) model.Result {
	name := "Showmax"
	hostname := "showmax.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.showmax.com/"
	headers := map[string]string{
		"Host":                      "www.showmax.com",
		"Connection":                "keep-alive",
		"Sec-Ch-UA":                 `"Chromium";v="124", "Microsoft Edge";v="124", "Not-A.Brand";v="99"`,
		"Sec-Ch-UA-Mobile":          "?0",
		"Sec-Ch-UA-Platform":        `"Windows"`,
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-User":            "?1",
		"Sec-Fetch-Dest":            "document",
		"Accept-Language":           "zh-CN,zh;q=0.9",
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body := string(b)
	//fmt.Println(body)
	regionStart := strings.Index(body, "activeTerritory")
	if regionStart == -1 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	regionEnd := strings.Index(body[regionStart:], "\n")
	region := strings.TrimSpace(body[regionStart+len("activeTerritory")+1 : regionStart+regionEnd])
	if region != "" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, Region: strings.ToLower(region), UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.showmax.com failed with code: %d", resp.StatusCode)}
}
