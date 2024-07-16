package kr

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// CoupangPlay
// www.coupangplay.com 仅 ipv4 且 get 请求
func CoupangPlay(c *http.Client) model.Result {
	name := "Coupang Play"
	hostname := "coupangplay.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.coupangplay.com/"
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
	//fmt.Println(resp.Request.URL.String())
	if strings.Contains(body, "is not available in your region") ||
		strings.Contains(resp.Request.URL.String(), "not-available") ||
		resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.coupangplay.com failed with code: %d", resp.StatusCode)}
}
