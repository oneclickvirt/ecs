package hk

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
)

// NowE
// webtvapi.nowe.com 仅 ipv4 且 post 请求
func NowE(c *http.Client) model.Result {
	name := "Now E"
	hostname := "nowe.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url1 := "https://webtvapi.nowe.com/16/1/getVodURL"
	data1 := `{"contentId":"202403181904703","contentType":"Vod","pin":"","deviceName":"Browser","deviceId":"w-663bcc51-913c-913c-913c-913c913c","deviceType":"WEB","secureCookie":null,"callerReferenceNo":"W17151951620081575","profileId":null,"mupId":null}`
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, body, err := utils.PostJson(c, url1, data1, headers)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	var res struct {
		ResponseCode string `json:"responseCode"`
	}
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.ResponseCode == "GEO_CHECK_FAIL" {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if res.ResponseCode == "SUCCESS" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{
		Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("webtvapi.nowe.com get responseCode: %s", res.ResponseCode),
	}
}
