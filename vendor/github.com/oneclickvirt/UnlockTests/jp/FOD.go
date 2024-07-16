package jp

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// FOD
// geocontrol1.stream.ne.jp 仅 ipv4 且 get 请求
func FOD(c *http.Client) model.Result {
	name := "FOD(Fuji TV)"
	hostname := "stream.ne.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://geocontrol1.stream.ne.jp/fod-geo/check.xml?time=1624504256"
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
	//fmt.Println(body)
	if strings.Contains(body, "FLAG TYPE=\"false\"") || resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 || strings.Contains(body, "true") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get geocontrol1.stream.ne.jp failed with code: %d", resp.StatusCode)}
}
