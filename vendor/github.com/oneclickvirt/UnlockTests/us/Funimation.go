package us

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
)

// Funimation
// www.crunchyroll.com 仅 ipv4 且 get 请求 ( www.funimation.com 重定向为 www.crunchyroll.com 了)
func Funimation(c *http.Client) model.Result {
	name := "Funimation"
	hostname := "crunchyroll.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.crunchyroll.com/"
	client := utils.Req(c)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	// b, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	// }
	// body := string(b)
	// fmt.Println(body)
	if resp.StatusCode == 403 || resp.StatusCode == 400 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	// fmt.Println(resp.Request.Cookies)
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	for _, ck := range resp.Request.Cookies {
		if ck.Name == "region" {
			return model.Result{Name: name, Status: model.StatusYes, Region: ck.Value, UnlockType: unlockType}
		}
	}
	if resp.StatusCode == 200 {
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.crunchyroll.com failed with code: %d", resp.StatusCode)}
}
