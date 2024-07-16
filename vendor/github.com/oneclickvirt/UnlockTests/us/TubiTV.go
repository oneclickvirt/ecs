package us

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// TubiTV
// tubitv.com 双栈 get 请求
func TubiTV(c *http.Client) model.Result {
	name := "Tubi TV"
	hostname := "tubitv.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://tubitv.com/home"
	client := utils.Req(c)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	//b, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	//}
	//body := string(b)
	if resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		return model.Result{Name: name, Status: model.StatusYes}
	}
	if resp.StatusCode == 302 {
		url2 := "https://gdpr.tubi.tv"
		resp2, err2 := client.R().Get(url2)
		if err2 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
		}
		defer resp2.Body.Close()
		b2, err2 := io.ReadAll(resp2.Body)
		if err2 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
		}
		body2 := string(b2)
		if strings.Contains(body2, "Unfortunately") {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get tubitv.com failed with code: %d", resp.StatusCode)}
}
