package au

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
)

// SBSonDemand
// www.sbs.com.au 仅 ipv4 且 get 请求
func SBSonDemand(c *http.Client) model.Result {
	name := "SBS on Demand"
	hostname := "sbs.com.au"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.sbs.com.au/api/v3/network?context=odwebsite"
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
	var res struct {
		Get struct {
			Response struct {
				CountryCode string `json:"country_code"`
			} `json:"response"`
		} `json:"get"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.Get.Response.CountryCode == "AU" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusNo}
}
