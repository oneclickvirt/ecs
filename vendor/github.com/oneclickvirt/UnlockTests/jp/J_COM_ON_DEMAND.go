package jp

import (
	"fmt"
	"net/http"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// J_COM_ON_DEMAND
// linkvod.myjcom.jp 仅 ipv4 且 get 请求
func J_COM_ON_DEMAND(c *http.Client) model.Result {
	name := "J:com On Demand"
	hostname := "id2.zaq.ne.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://auth.id2.zaq.ne.jp/login"
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
	//fmt.Println(body)
	if resp.StatusCode == 400 || resp.StatusCode == 403 || resp.StatusCode == 404 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get linkvod.myjcom.jp failed with code: %d", resp.StatusCode)}
}
