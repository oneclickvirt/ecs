package transnation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// SonyLiv
// www.sonyliv.com 双栈 且 get 请求 - 有问题，获取不到地区
func SonyLiv(c *http.Client) model.Result {
	name := "SonyLiv"
	hostname := "www.sonyliv.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.sonyliv.com/"
	client := utils.Req(c)
	resp1, err1 := client.R().Get(url)
	if err1 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err1}
	}
	defer resp1.Body.Close()
	b, err := io.ReadAll(resp1.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	body1 := string(b)
	if strings.Contains(body1, "geolocation_notsupported") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	jwtToken := utils.ReParse(body1, `resultObj:"([^"]+)`)
	if jwtToken == "" {
		return model.Result{Name: name, Status: model.StatusErr, Err: fmt.Errorf("can not find jwtToken")}
	}
	// fmt.Println(jwtToken)
	// 获取不到region
	headers2 := map[string]string{
		"accept":         "application/json, text/plain, */*",
		"referer":        "https://www.sonyliv.com/",
		"device_id":      "25a417c3b5f246a393fadb022adc82d5-1715309762699",
		"app_version":    "3.5.59",
		"security_token": jwtToken,
	}
	url2 := "https://apiv2.sonyliv.com/AGL/1.4/A/ENG/WEB/ALL/USER/ULD"
	client = utils.SetReqHeaders(client, headers2)
	resp2, err2 := client.R().Get(url2)
	if err2 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
	}
	defer resp2.Body.Close()
	b, err = io.ReadAll(resp2.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	body2 := string(b)
	// fmt.Println(body2)
	var region string
	if body2 != "" && strings.Contains(body2, "country_code") {
		var res1 struct {
			ResultObj struct {
				CountryCode string `json:"country_code"`
			} `json:"resultObj"`
		}
		if err := json.Unmarshal([]byte(body2), &res1); err != nil {
			return model.Result{Name: name, Status: model.StatusErr, Err: err}
		}
		region = res1.ResultObj.CountryCode
		if region == "" {
			return model.Result{Name: name, Status: model.StatusErr, Err: fmt.Errorf("can not found region")}
		}
	} else {
		return model.Result{Name: name, Status: model.StatusErr, Err: fmt.Errorf("can not found region")}
	}

	headers3 := map[string]string{
		"upgrade-insecure-requests": "1",
		"accept":                    "application/json, text/plain, */*",
		"origin":                    "https://www.sonyliv.com",
		"referer":                   "https://www.sonyliv.com/",
		"device_id":                 "25a417c3b5f246a393fadb022adc82d5-1715309762699",
		"security_token":            jwtToken,
	}
	// 1000273613 1000045427
	url3 := "https://apiv2.sonyliv.com/AGL/3.8/A/ENG/WEB/" + region + "/ALL/CONTENT/VIDEOURL/VOD/1000045427/prefetch"
	client = utils.SetReqHeaders(client, headers3)
	resp3, err3 := client.R().Get(url3)
	if err3 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err3}
	}
	defer resp3.Body.Close()
	b, err = io.ReadAll(resp3.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	body3 := string(b)
	// fmt.Println(body3)
	var res2 struct {
		ResultCode string `json:"resultCode"`
	}
	if err := json.Unmarshal([]byte(body3), &res2); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res2.ResultCode == "OK" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType,
			Region: strings.ToLower(region)}
	}
	if res2.ResultCode == "KO" || strings.Contains(body3, "It seems you are trying to access SonyLIV via <b>VPN, Proxy</b> or a <b>Routed Service</b>.") {
		return model.Result{Name: name, Status: model.StatusNo, Info: "Proxy Detected", Region: strings.ToLower(region)}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get apiv2.sonyliv.com failed with code: %d", resp3.StatusCode)}
}
