package hk

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// MyTvSuper
// www.mytvsuper.com 仅 ipv4 且 get 请求
func MyTvSuper(c *http.Client) model.Result {
	name := "MyTVSuper"
	hostname := "mytvsuper.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.mytvsuper.com/api/auth/getSession/self/"
	headers := map[string]string{
		"User-Agent":   model.UA_Browser,
		"Content-Type": "application/json",
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
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
	var mytvsuperRes struct {
		Region      int    `json:"region"`
		CountryCode string `json:"country_code"`
	}
	if err := json.Unmarshal(b, &mytvsuperRes); err != nil {
		if strings.Contains(body, "HK") {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if mytvsuperRes.Region == 1 && mytvsuperRes.CountryCode == "HK" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	if mytvsuperRes.Region != 1 || mytvsuperRes.CountryCode != "HK" {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.mytvsuper.com failed with code: %d", resp.StatusCode)}
}
