package jp

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
)

// Kancolle
// Kancolle 仅 ipv4 且 get 请求
func Kancolle(c *http.Client) model.Result {
	name := "Kancolle Japan"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "http://203.104.209.7/kcscontents/news/"
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
	if resp.StatusCode == 403 || resp.StatusCode == 451 || resp.StatusCode == 302 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		return model.Result{Name: name, Status: model.StatusYes}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get Kancolle failed with code: %d", resp.StatusCode)}
}
