package uk

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// ITVX
// simulcast.itv.com 仅 ipv4 且 get 请求
func ITVX(c *http.Client) model.Result {
	name := "ITV Hub"
	hostname := "itv.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://simulcast.itv.com/playlist/itvonline/ITV"
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
	if resp.StatusCode == 403 || strings.Contains(body, "Outside Of Allowed Geographic Region") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if strings.Contains(body, "Playlist") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get simulcast.itv.com failed with code: %d", resp.StatusCode)}
}
