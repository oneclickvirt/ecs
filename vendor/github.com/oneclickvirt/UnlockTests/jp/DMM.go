package jp

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// DMM
// bitcoin.dmm.com 仅 ipv4 且 get 请求
func DMM(c *http.Client) model.Result {
	name := "DMM"
	hostname := "dmm.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://bitcoin.dmm.com"
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
	if strings.Contains(body, "This page is not available in your area") {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if strings.Contains(body, "暗号資産") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get bitcoin.dmm.com failed with code: %d", resp.StatusCode)}
}
