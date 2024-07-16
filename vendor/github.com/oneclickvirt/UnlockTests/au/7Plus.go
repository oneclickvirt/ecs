package au

import (
	"fmt"
	"net/http"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// Au7plus
// 7plus.com.au 仅 ipv4 且 get 请求
// 7plus-sevennetwork.akamaized.net 有问题 - 无论如何请求都失败
func Au7plus(c *http.Client) model.Result {
	name := "7plus"
	hostname := "7plus.com.au"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://7plus-sevennetwork.akamaized.net/media/v1/dash/live/cenc/5303576322001/68dca38b-85d7-4dae-b1c5-c88acc58d51c/f4ea4711-514e-4cad-824f-e0c87db0a614/225ec0a0-ef18-4b7c-8fd6-8dcdd16cf03a/1x/segment0.m4f?akamai_token=exp=1672500385~acl=/media/v1/dash/live/cenc/5303576322001/68dca38b-85d7-4dae-b1c5-c88acc58d51c/f4ea4711-514e-4cad-824f-e0c87db0a614/*~hmac=800e1e1d1943addf12b71339277c637c7211582fe12d148e486ae40d6549dbde"
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
	// fmt.Println(body)
	// fmt.Println(resp.StatusCode)
	if resp.StatusCode == 200 {
		return model.Result{Name: name, Status: model.StatusYes}
	} else {
		resp1, err1 := client.R().Get("https://7plus.com.au/")
		if err1 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err1}
		}
		defer resp1.Body.Close()
		// fmt.Println(body)
		if resp1.StatusCode == 403 || resp1.StatusCode == 451 {
			return model.Result{Name: name, Status: model.StatusNo}
		} else if resp1.StatusCode == 200 {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		} else {
			return model.Result{Name: name, Status: model.StatusUnexpected,
				Err: fmt.Errorf("get 7plus.com.au failed with code: %d %d", resp.StatusCode, resp1.StatusCode)}
		}
	}
}
