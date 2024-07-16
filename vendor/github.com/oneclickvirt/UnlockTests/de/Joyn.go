package de

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
	"strings"
)

// Joyn
// auth.joyn.de 仅 ipv4 且 post 请求
func Joyn(c *http.Client) model.Result {
	name := "Joyn"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://auth.joyn.de/auth/anonymous"
	hostname := "joyn.de"
	payload := `{"client_id":"b74b9f27-a994-4c45-b7eb-5b81b1c856e7","client_name":"web","anon_device_id":"b74b9f27-a994-4c45-b7eb-5b81b1c856e7"}`
	resp, body, err := utils.PostJson(c, url, payload, nil)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	//fmt.Println(body)
	if strings.Contains(body, "Unauthorized") {
		return model.Result{Name: name, Status: model.StatusUnexpected, Err: fmt.Errorf("Unauthorized")}
	}
	var res struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		return model.Result{Name: name, Status: model.StatusUnexpected, Err: err}
	}

	url2 := "https://api.joyn.de/content/entitlement-token"
	headers := map[string]string{
		"authorization": "Bearer " + res.AccessToken,
		"x-api-key":     "36lp1t4wto5uu2i2nk57ywy9on1ns5yg",
	}
	payload2 := `{"content_id":"daserste-de-hd","content_type":"LIVE"}`
	resp2, body2, err2 := utils.PostJson(c, url2, payload2, headers)
	if err2 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
	}
	defer resp2.Body.Close()
	var res2a []struct {
		Code string `json:"code"`
	}
	var res2b struct {
		Token string `json:"entitlement_token"`
	}
	if err := json.Unmarshal([]byte(body2), &res2a); err != nil {
		if err := json.Unmarshal([]byte(body2), &res2b); err != nil {
			return model.Result{Name: name, Status: model.StatusUnexpected, Err: err}
		}
		if res2b.Token != "" {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		}
	}
	if len(res2a) > 0 && res2a[0].Code == "ENT_AssetNotAvailableInCountry" {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get joyn.de with code: %d %d", resp.StatusCode, resp2.StatusCode)}
}
