package transnation

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// MetaAI
// www.meta.ai 双栈 且 get 请求 有问题
func MetaAI(c *http.Client) model.Result {
	name := "MetaAI"
	hostname := "www.meta.ai"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.meta.ai/"
	headers := map[string]string{
		"User-Agent":                model.UA_Browser,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Accept-Language":           "en-US,en;q=0.9",
		"sec-ch-ua":                 "${UA_SEC_CH_UA}",
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        "Windows",
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "none",
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
	// 检查是否被阻止
	if strings.Contains(body, "AbraGeoBlockedErrorRoot") {
		return model.Result{Name: name, Status: model.StatusNo, Info: "AbraGeoBlocked"}
	}
	// 检查是否成功
	if strings.Contains(body, "AbraHomeRootConversationQuery") {
		start := strings.Index(body, `"code"`)
		if start != -1 {
			start = strings.Index(body[start:], `"`) + start + 1
			end := strings.Index(body[start:], `"`) + start
			code := body[start:end]
			region := strings.Split(code, "_")[1]
			if region != "" {
				result1, result2, result3 := utils.CheckDNS(hostname)
				unlockType := utils.GetUnlockType(result1, result2, result3)
				return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType, Region: region}
			}
		}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.meta.ai failed with code: %d", resp.StatusCode)}
}
