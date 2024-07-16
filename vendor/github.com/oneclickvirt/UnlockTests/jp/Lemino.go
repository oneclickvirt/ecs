package jp

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
)

// Lemino
// if.lemino.docomo.ne.jp 双栈 且 get 请求
func Lemino(c *http.Client) model.Result {
	name := "Lemino"
	hostname := "docomo.ne.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://if.lemino.docomo.ne.jp/v1/user/delivery/watch/ready"
	headers := map[string]string{
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"Content-Type":       "application/json",
		"Origin":             "https://lemino.docomo.ne.jp",
		"Referer":            "https://lemino.docomo.ne.jp/",
		"Sec-CH-UA-Mobile":   "?0",
		"Sec-CH-UA-Platform": "\"Windows\"",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
		"X-Service-Token":    "f365771afd91452fa279863f240c233d",
		"X-Trace-ID":         "556db33f-d739-4a82-84df-dd509a8aa179",
		"sec-ch-ua":          model.UA_SecCHUA,
	}
	playload := "{\"inflow_flows\":[null,\"crid://plala.iptvf.jp/group/b100ce3\"],\"play_type\":1,\"key_download_only\":null,\"quality\":null,\"groupcast\":null,\"avail_status\":\"1\",\"terminal_type\":3,\"test_account\":0,\"content_list\":[{\"kind\":\"main\",\"service_id\":null,\"cid\":\"00lm78dz30\",\"lid\":\"a0lsa6kum1\",\"crid\":\"crid://plala.iptvf.jp/vod/0000000000_00lm78dymn\",\"preview\":0,\"trailer\":0,\"auto_play\":0,\"stop_position\":0}]}"
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().SetBodyJsonString(playload).Get(url)
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
		Err: fmt.Errorf("get if.lemino.docomo.ne.jp failed with code: %d", resp.StatusCode)}
}
