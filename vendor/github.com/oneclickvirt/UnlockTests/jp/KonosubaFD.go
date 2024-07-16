package jp

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
)

// KonosubaFD
// api.konosubafd.jp 仅 ipv4 且 post 请求
func KonosubaFD(c *http.Client) model.Result {
	name := "Konosuba Fantastic Days"
	hostname := "konosubafd.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api.konosubafd.jp/api/masterlist"
	headers := map[string]string{
		"User-Agent": model.UA_Pjsekai,
	}
	resp, _, err := utils.PostJson(c, url, "", headers)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get api.konosubafd.jp failed with code: %d", resp.StatusCode)}
}
