package us

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// PlutoTV
// pluto.tv 仅 ipv4 且 get 请求
func PlutoTV(c *http.Client) model.Result {
	name := "Pluto TV"
	hostname := "pluto.tv"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://pluto.tv/"
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
	if strings.Contains(body, "thanks-for-watching") || strings.Contains(body, "plutotv-is-not-available") ||
		strings.Contains(resp.Request.URL.String(), "plutotv-is-not-available") ||
		resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if resp.StatusCode == 429 {
		return model.Result{Name: name, Status: model.StatusUnexpected, Info: "Rate Limit"}
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
}
