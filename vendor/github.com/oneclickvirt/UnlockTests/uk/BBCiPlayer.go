package uk

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// BBCiPlayer
// open.live.bbc.co.uk 仅 ipv4 且 get 请求
func BBCiPlayer(c *http.Client) model.Result {
	name := "BBC iPLAYER"
	hostname := "bbc.co.uk"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://open.live.bbc.co.uk/mediaselector/6/select/version/2.0/mediaset/pc/vpid/bbc_one_london/format/json/jsfunc/JS_callbacks0"
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
	if resp.StatusCode == 200 {
		if strings.Contains(body, "geolocation") {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		if strings.Contains(body, "vs-hls-push-uk") {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		}
	} else if resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get open.live.bbc.co.uk failed with code: %d", resp.StatusCode)}
}
