package jp

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Niconico
// www.nicovideo.jp 仅 ipv4 且 get 请求
func Niconico(c *http.Client) model.Result {
	name := "Niconico"
	hostname := "nicovideo.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	url1 := "https://www.nicovideo.jp/watch/so40278367" // 进击的巨人
	//url2 := "https://www.nicovideo.jp/watch/so23017073" // 假面骑士
	headers := map[string]string{
		"User-Agent": model.UA_Browser,
	}
	client1 := utils.Req(c)
	resp1, err1 := client1.R().Get(url1)
	if err1 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err1}
	}
	defer resp1.Body.Close()
	b1, err := io.ReadAll(resp1.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body1 := string(b1)
	if strings.Contains(body1, "同じ地域") || resp1.StatusCode == 403 || resp1.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	headers = map[string]string{
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language":           "en-US,en;q=0.9",
		"sec-ch-ua":                 `"(Not(A:Brand";v="8", "Chromium";v="114", "Google Chrome";v="114")"`,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        `"Windows"`,
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "none",
		"sec-fetch-user":            "?1",
		"upgrade-insecure-requests": "1",
		"user-agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	}
	client1 = utils.SetReqHeaders(client1, headers)
	url11 := "https://live.nicovideo.jp/?cmnhd_ref=device=pc&site=nicolive&pos=header_servicelink&ref=WatchPage-Anchor"
	resp, err := client1.R().Get(url11)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body := string(b)
	// 查找第一个官方直播剧的ID
	splitted := strings.Split(body, "&quot;isOfficialChannelMemberFree&quot;:false")
	var liveID string
	for _, part := range splitted {
		if strings.Contains(part, "話") && !strings.Contains(part, "&quot;isOfficialChannelMemberFree&quot;:true") && !strings.Contains(part, "playerProgram") && !strings.Contains(part, "&quot;ON_AIR&quot;") {
			startIdx := strings.Index(part, "&quot;id&quot;:&quot;")
			if startIdx != -1 {
				startIdx += len("&quot;id&quot;:&quot;")
				endIdx := strings.Index(part[startIdx:], "&quot;")
				if endIdx != -1 {
					liveID = part[startIdx : startIdx+endIdx]
					break
				}
			}
		}
	}
	if liveID != "" {
		resp, err = client1.R().Get("https://live.nicovideo.jp/watch/" + liveID)
		if err != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
		}
		defer resp.Body.Close()
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
		}
		body = string(b)
		if strings.Contains(body, "notAllowedCountry") && resp1.StatusCode == 200 {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType,
				Info: fmt.Sprintf("But Official Live is Unavailable. LiveID: %s", liveID)}
		}
		if resp1.StatusCode == 200 {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType,
				Info: fmt.Sprintf("LiveID: %s", liveID)}
		}
	} else if resp1.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType,
			Info: "But Official Live is Unavailable"}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.nicovideo.jp failed with code: %d", resp.StatusCode)}
}
