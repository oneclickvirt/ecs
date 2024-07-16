package us

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// FXNOW
// fxnow.fxnetworks.com 仅 ipv4 且 get 请求
func FXNOW(c *http.Client) model.Result {
	name := "FXNOW"
	hostname := "fxnetworks.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://fxnow.fxnetworks.com/"
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
	if strings.Contains(body, "is not accessible") {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if strings.Contains(body, "FX Movies") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get fxnow.fxnetworks.com with code: %d", resp.StatusCode)}
}
