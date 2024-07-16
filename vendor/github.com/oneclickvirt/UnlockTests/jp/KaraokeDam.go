package jp

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
)

// KaraokeDam
// cds1.clubdam.com 仅 ipv4 且 get 请求
func KaraokeDam(c *http.Client) model.Result {
	name := "Karaoke@DAM"
	hostname := "clubdam.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "http://cds1.clubdam.com/vhls-cds1/site/xbox/sample_1.mp4.m3u8"
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
	if resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get Karaoke@Dam failed with code: %d", resp.StatusCode)}
}
