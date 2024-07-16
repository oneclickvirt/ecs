package us

import (
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
	"strings"
)

// PeacockTV
// www.peacocktv.com 双栈 get 请求
func PeacockTV(c *http.Client) model.Result {
	name := "Peacock TV"
	hostname := "peacocktv.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.peacocktv.com/"
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
	if strings.Contains(resp.Header.Get("location"), "unavailable") ||
		resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
}
