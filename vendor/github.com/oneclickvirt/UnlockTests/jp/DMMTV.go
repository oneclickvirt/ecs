package jp

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
	"strings"
)

// DMMTV
// api.beacon.dmm.com 双栈 且 post 请求
func DMMTV(c *http.Client) model.Result {
	name := "DMM TV"
	hostname := "dmm.com"
	if c == nil {
		return model.Result{Name: name}
	}
	resp, body, err := utils.PostJson(c, "https://api.beacon.dmm.com/v1/streaming/start",
		`{"player_name":"dmmtv_browser","player_version":"0.0.0","content_type_detail":"VOD_SVOD","content_id":"11uvjcm4fw2wdu7drtd1epnvz","purchase_product_id":null}`,
		map[string]string{"User-Agent": model.UA_Browser},
	)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	var res struct {
		IsBkocked   bool   `json:"is_blocked"`
		BlockStatus string `json:"block_status"`
	}
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		if strings.Contains(body, "UNAUTHORIZED") {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		}
		if strings.Contains(body, "FOREIGN") {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.IsBkocked {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if !res.IsBkocked {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get api.beacon.dmm.com failed with code: %d", resp.StatusCode)}
}
