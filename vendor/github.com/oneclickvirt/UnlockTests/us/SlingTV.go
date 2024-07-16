package us

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
)

// SlingTV
// www.sling.com 双栈 且 get 请求
func SlingTV(c *http.Client) model.Result {
	name := "Sling TV"
	hostname := "sling.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.sling.com/"
	headers := map[string]string{
		"User-Agent": model.UA_Dalvik,
	}
	client := utils.ReqDefault(c)
	client = utils.SetReqHeaders(client, headers)
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
	if resp.StatusCode == 403 || resp.StatusCode == 451 || resp.StatusCode == 302 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.sling.com failed with code: %d", resp.StatusCode)}
}
