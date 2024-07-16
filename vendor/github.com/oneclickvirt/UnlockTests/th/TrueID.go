package th

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

func getStringBetween(value string, a string, b string) string {
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	posLast := strings.Index(value[posFirstAdjusted:], b)
	if posLast == -1 {
		return ""
	}
	return value[posFirstAdjusted : posFirstAdjusted+posLast]
}

// TrueID
// tv.trueid.net 双栈 get 请求
func TrueID(c *http.Client) model.Result {
	name := "TrueID"
	hostname := "trueid.net"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://tv.trueid.net/th-en/live/thairathtv-hd"
	headers := map[string]string{
		"User-Agent":                "{UA_Browser}",
		"sec-ch-ua":                 "{UA_SecCHUA}",
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        "Windows",
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "same-origin",
		"sec-fetch-user":            "?1",
		"upgrade-insecure-requests": "1",
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
	channelId := getStringBetween(body, `"channelId":"`, `"`)
	authUser := getStringBetween(body, `"buildId":"`, `"`)
	if len(authUser) < 11 {
		return model.Result{Name: name, Status: model.StatusErr, Err: fmt.Errorf("authUser len < 11")}
	}
	authKey := authUser[10:]
	apiURL := fmt.Sprintf("https://tv.trueid.net/api/stream/checkedPlay?channelId=%s&lang=en&country=th", channelId)
	authHeader := fmt.Sprintf("%s:%s", authUser, authKey)
	headers2 := map[string]string{
		"Authorization": authHeader,
		"accept":        "application/json, text/plain, */*",
		"referer":       url,
	}
	client = utils.SetReqHeaders(client, headers2)
	resp2, err := client.R().Get(apiURL)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp2.Body.Close()
	b, err = io.ReadAll(resp2.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body = string(b)
	result := getStringBetween(body, `"billboardType":"`, `"`)
	if result == "GEO_BLOCK" || strings.Contains(body, "Access denied") || resp.StatusCode == 401 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if result == "LOADING" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get  failed with code: %d", resp.StatusCode)}
}
