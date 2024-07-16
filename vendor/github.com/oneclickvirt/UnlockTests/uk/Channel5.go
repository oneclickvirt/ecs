package uk

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
	"time"
)

// Channel5
// cassie.channel5.com 仅 ipv4 且 get 请求
func Channel5(c *http.Client) model.Result {
	name := "Channel 5"
	hostname := "channel5.com"
	if c == nil {
		return model.Result{Name: name}
	}
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	url := fmt.Sprintf("https://cassie.channel5.com/api/v2/live_media/my5desktopng/C5.json?timestamp=%d&auth=0_rZDiY0hp_TNcDyk2uD-Kl40HqDbXs7hOawxyqPnbI", timestamp)
	client := utils.Req(c)
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
	// fmt.Println(body)
	var res struct {
		code string `json:"code"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.code == "3000" || strings.Contains(body, "this service is only available in restricted regions") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if res.code == "4003" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get cassie.channel5.com failed with code: %d", resp.StatusCode)}
}
