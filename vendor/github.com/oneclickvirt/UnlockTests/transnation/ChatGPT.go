package transnation

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// OpenAI
// api.openai.com 仅 ipv4 且 get 请求
func OpenAI(c *http.Client) model.Result {
	name := "ChatGPT"
	if c == nil {
		return model.Result{Name: name}
	}
	var body1, body2, body3 string
	url1 := "https://api.openai.com/compliance/cookie_requirements"
	headers1 := map[string]string{
		"User-Agent":         model.UA_Browser,
		"authority":          "api.openai.com",
		"accept":             "*/*",
		"accept-language":    "zh-CN,zh;q=0.9",
		"authorization":      "Bearer null",
		"content-type":       "application/json",
		"origin":             "https://platform.openai.com",
		"referer":            "https://platform.openai.com/",
		"sec-ch-ua":          model.UA_SecCHUA,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "Windows",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
	}
	client1 := utils.Req(c)
	client1 = utils.SetReqHeaders(client1, headers1)
	resp1, err1 := client1.R().Get(url1)

	url2 := "https://ios.chat.openai.com/"
	headers2 := map[string]string{
		"User-Agent":                model.UA_Browser,
		"authority":                 "ios.chat.openai.com",
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language":           "zh-CN,zh;q=0.9",
		"sec-ch-ua":                 model.UA_SecCHUA,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        "Windows",
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "none",
		"sec-fetch-user":            "?1",
		"upgrade-insecure-requests": "1",
	}
	client2 := utils.Req(c)
	client2 = utils.SetReqHeaders(client2, headers2)
	resp2, err2 := client2.R().Get(url2)

	url3 := "https://chat.openai.com/cdn-cgi/trace"
	client3 := utils.Req(c)
	resp3, err3 := client3.R().Get(url3)

	var reqStatus1, reqStatus2, reqStatus3 bool
	if err1 != nil {
		//fmt.Println(err1)
		reqStatus1 = false
	} else {
		reqStatus1 = true
		defer resp1.Body.Close()
		b1, err11 := io.ReadAll(resp1.Body)
		if err11 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
		}
		body1 = string(b1)
	}
	if err2 != nil {
		//fmt.Println(err2)
		reqStatus2 = false
	} else {
		reqStatus2 = true
		defer resp2.Body.Close()
		b2, err22 := io.ReadAll(resp2.Body)
		if err22 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
		}
		body2 = string(b2)
	}
	if err3 != nil {
		//fmt.Println(err3)
		reqStatus3 = false
	} else {
		reqStatus3 = true
		defer resp3.Body.Close()
		b3, err33 := io.ReadAll(resp3.Body)
		if err33 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
		}
		body3 = string(b3)
	}
	unsupportedCountry := strings.Contains(body1, "unsupported_country")
	VPN := strings.Contains(body2, "VPN")
	tempList := strings.Split(body3, "\n")
	var location string
	if reqStatus3 {
		for _, line := range tempList {
			if strings.HasPrefix(line, "loc=") {
				location = strings.ReplaceAll(line, "loc=", "")
			}
		}
	}
	if (reqStatus1 && resp1 != nil && resp1.StatusCode == 429) || (reqStatus2 && resp2 != nil && resp2.StatusCode == 429) {
		if location != "" {
			loc := strings.ToLower(location)
			exit := utils.GetRegion(loc, model.GptSupportCountry)
			if exit {
				return model.Result{Name: name, Status: model.StatusNo, Info: "429 Rate limit", Region: loc}
			}
		}
		return model.Result{Name: name, Status: model.StatusNo, Info: "429 Rate limit"}
	}
	if !VPN && !unsupportedCountry && reqStatus1 && reqStatus2 && reqStatus3 {
		if location != "" {
			loc := strings.ToLower(location)
			exit := utils.GetRegion(loc, model.GptSupportCountry)
			result1, result2, result3 := utils.CheckDNS("api.openai.com")
			unlockType := utils.GetUnlockType(result1, result2, result3)
			if exit {
				return model.Result{Name: name, Status: model.StatusYes, Region: loc, UnlockType: unlockType}
			} else {
				return model.Result{Name: name, Status: model.StatusYes, Info: "but cdn-cgi not unsupported",
					Region: location, UnlockType: unlockType}
			}
		} else {
			return model.Result{Name: name, Status: model.StatusYes}
		}
	} else if !unsupportedCountry && VPN && reqStatus1 {
		result1, result2, result3 := utils.CheckDNS("chat.openai.com")
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, Info: "Only Available with Web Browser",
			UnlockType: unlockType}
	} else if unsupportedCountry && !VPN && reqStatus2 {
		result1, result2, result3 := utils.CheckDNS("ios.chat.openai.com")
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, Info: "Only Available with Mobile APP",
			UnlockType: unlockType}
	} else if !reqStatus1 && VPN {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if VPN && unsupportedCountry {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if !reqStatus1 && !reqStatus2 && !reqStatus3 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else {
		return model.Result{Name: name, Status: model.StatusUnexpected}
	}
}
