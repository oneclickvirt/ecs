package eu

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// SkyShowTime
// www.skyshowtime.com 双栈 get 请求
func SkyShowTime(c *http.Client) model.Result {
	name := "SkyShowTime"
	hostname := "skyshowtime.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.skyshowtime.com/"
	headers := map[string]string{
		"User-Agent": model.UA_Browser,
		"accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
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
	if strings.Contains(body, "Access Denied") ||
		strings.Contains(resp.Request.URL.String(), "where-can-i-stream") ||
		strings.Contains(body, "is not available in your location") ||
		resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		// fmt.Println(resp.Request.URL.String())
		var region string
		tempList := strings.Split(body, "\n")
		for _, line := range tempList {
			if strings.Contains(line, "location") && strings.Contains(line, ":") {
				if strings.Contains(line, "https://www.skyshowtime.com/watch/home") {
					continue
				}
				tpList := strings.Split(line, ":")
				if len(tpList) >= 2 && len(tpList[1]) <= 3 {
					region = strings.TrimSpace(strings.ReplaceAll(tpList[1], "?", ""))
					break
				} else {
					return model.Result{Name: name, Status: model.StatusNo}
				}
			}
		}
		if region != "" {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, Region: strings.ToLower(region), UnlockType: unlockType}
		} else {
			return model.Result{Name: name, Status: model.StatusNo}
		}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.skyshowtime.com failed with code: %d", resp.StatusCode)}
}
