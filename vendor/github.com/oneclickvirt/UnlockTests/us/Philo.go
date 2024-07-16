package us

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Philo
// content-us-east-2-fastly-b.www.philo.com 仅 ipv4 且 get 请求
func Philo(c *http.Client) model.Result {
	name := "Philo"
	hostname := "philo.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://content-us-east-2-fastly-b.www.philo.com/geo"
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
	var res struct {
		Status  string `json:"status"`
		Country string `json:"country"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		if resp.StatusCode == 403 || resp.StatusCode == 451 {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.Status == "FAIL" {
		return model.Result{Name: name, Status: model.StatusNo, Region: strings.ToLower(res.Country)}
	} else if res.Status == "SUCCESS" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, Region: strings.ToLower(res.Country), UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get content-us-east-2-fastly-b.www.philo.com failed with code: %d", resp.StatusCode)}
}
