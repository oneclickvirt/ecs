package uk

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// SkyGo
// skyid.sky.com 仅 ipv4 且 get 请求
func SkyGo(c *http.Client) model.Result {
	name := "Sky Go"
	hostname := "sky.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://skyid.sky.com/authorise/skygo?response_type=token&client_id=sky&appearance=compact&redirect_uri=skygo://auth"
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
	if strings.Contains(body, "You don't have permission to access") || resp.StatusCode == 403 || resp.StatusCode == 200 ||
		strings.Contains(body, "Access Denied") { // || resp.StatusCode == 451
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 302 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get skyid.sky.com failed with code: %d", resp.StatusCode)}
}
