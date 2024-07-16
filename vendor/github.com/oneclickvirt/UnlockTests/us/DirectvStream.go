package us

import (
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
)

// ATTNOW - DirectvStream
// www.atttvnow.com 双栈 且 get 请求
func DirectvStream(c *http.Client) model.Result {
	name := "Directv Stream"
	hostname := "atttvnow.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.atttvnow.com/"
	client := utils.Req(c)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	//b, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	//}
	//body := string(b)
	if resp.StatusCode == 403 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
}
