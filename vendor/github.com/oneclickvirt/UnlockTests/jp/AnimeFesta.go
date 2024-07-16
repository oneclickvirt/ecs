package jp

import (
	"fmt"
	"net/http"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// AnimeFesta
// api-animefesta.iowl.jp 仅 ipv4 且 get 请求
func AnimeFesta(c *http.Client) model.Result {
	name := "AnimeFesta"
	hostname := "animefesta.iowl.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api-animefesta.iowl.jp/v1/titles/1305"
	client := utils.Req(c)
	headers := map[string]string{
		"Origin":  "https://animefesta.iowl.jp",
		"Referer": "https://animefesta.iowl.jp/",
	}
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	if resp.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	} else if resp.StatusCode == 403 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get api-animefesta.iowl.jp failed with code: %d", resp.StatusCode)}
}
