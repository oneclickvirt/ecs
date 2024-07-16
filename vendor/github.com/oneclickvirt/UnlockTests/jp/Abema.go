package jp

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Abema
// api.abema.io 仅 ipv4 且 get 请求
func Abema(c *http.Client) model.Result {
	name := "Abema.TV"
	hostname := "abema.io"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api.abema.io/v1/ip/check?device=android"
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
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body := string(b)
	// fmt.Println(body)
	var abemaRes struct {
		Message        string `json:"message"`
		IsoCountryCode string `json:"isoCountryCode"`
	}
	if err := json.Unmarshal(b, &abemaRes); err != nil {
		if strings.Contains(body, "blocked_location") || strings.Contains(body, "anonymous_ip") {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if abemaRes.IsoCountryCode == "JP" || strings.Contains(body, "JP") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType, Region: "JP"}
	}
	if abemaRes.Message == "blocked_location" || abemaRes.Message == "anonymous_ip" {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType, Info: "Oversea Only"}
}
