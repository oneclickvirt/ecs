package asia

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// HBOGO
// api2.hbogoasia.com 仅 ipv4 且 get 请求
func HBOGO(c *http.Client) model.Result {
	name := "HBO GO Asia"
	hostname := "hbogoasia.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api2.hbogoasia.com/v1/geog?lang=undefined&version=0&bundleId=www.hbogoasia.com"
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
	//body := string(b)
	//fmt.Println(body)
	var hboRes struct {
		Country   string `json:"country"`
		Territory string `json:"territory"`
	}
	if err := json.Unmarshal(b, &hboRes); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if hboRes.Territory == "" {
		// 解析不到为空则识别为不解锁
		return model.Result{Name: name, Status: model.StatusNo}
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	return model.Result{Name: name, Status: model.StatusYes, Region: strings.ToLower(hboRes.Country), UnlockType: unlockType}
}
