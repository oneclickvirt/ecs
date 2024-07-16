package nz

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// MaoriTV
// www.maoriplus.co.nz 双栈 且 get 请求
func MaoriTV(c *http.Client) model.Result {
	name := "Maori TV"
	hostname := "maoriplus.co.nz"
	if c == nil {
		return model.Result{Name: name}
	}
	// https://www.maoriplus.co.nz/show/kapa-haka-regionals-2024-tamaki-makaurau/play/6352727601112
	url := "https://edge.api.brightcove.com/playback/v1/accounts/1614493167001/videos/6352727601112"
	client := utils.Req(c)
	headers := map[string]string{
		"User-Agent": model.UA_Browser,
		"Accept":     "application/json;pk=BCpkADawqM2E9yW4lLgKIEIV5majz5djzZCIqJiYMkP5yYaYdF6AQYq4isPId1ZLtQdGnK1ErLYG0-r1N-3DzAEdbfvw9SFdDWz_i09pLp8Njx1ybslyIXid-X_Dx31b7-PLdQhJCws-vk6Y",
		"Origin":     "https://www.maoritelevision.com",
	}
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	body := string(b)
	//fmt.Println(body)
	var res1 struct {
		ErrorSubcode string `json:"error_subcode"`
		AccountId    string `json:"account_id"`
	}
	var res2 []struct {
		ClientGeo    string `json:"client_geo"`
		ErrorSubcode string `json:"error_subcode"`
		ErrorCode    string `json:"error_code"`
		Message      string `json:"message"`
	}
	if err := json.Unmarshal(b, &res1); err != nil {
		if err := json.Unmarshal(b, &res2); err != nil {
			if strings.Contains(body, "CLIENT_GEO") || strings.Contains(body, "ACCESS_DENIED") {
				return model.Result{Name: name, Status: model.StatusNo}
			}
			return model.Result{Name: name, Status: model.StatusErr, Err: err}
		}
		if res2[0].ErrorSubcode == "CLIENT_GEO" {
			return model.Result{Name: name, Status: model.StatusNo, Region: res2[0].ClientGeo}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res1.AccountId != "0" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType, Region: "nz"}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get edge.api.brightcove.com failed with code: %d", resp.StatusCode)}
}
