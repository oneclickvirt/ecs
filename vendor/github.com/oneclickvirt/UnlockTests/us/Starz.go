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

// Starz
// www.starz.com 双栈 get 请求
func Starz(c *http.Client) model.Result {
	name := "Starz"
	hostname := "starz.com"
	if c == nil {
		return model.Result{Name: name}
	}
	client := utils.Req(c)
	client.Headers.Set("Referer", "https://www.starz.com/us/en/")
	// client.Headers.Set("Authtokenauthorization", "")
	url := "https://www.starz.com/sapi/header/v1/starz/us/09b397fc9eb64d5080687fc8a218775b" // 请求有tls校验
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	authorization := string(b)
	// fmt.Printf(authorization)
	if authorization != "" && !strings.Contains(authorization, "AccessDenied") {
		url2 := "https://auth.starz.com/api/v4/User/geolocation"
		headers2 := map[string]string{
			"AuthTokenAuthorization": authorization,
		}
		client2 := utils.Req(c)
		client2 = utils.SetReqHeaders(client2, headers2)
		resp2, err2 := client2.R().Get(url2)
		if err2 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
		}
		defer resp2.Body.Close()
		b2, err2 := io.ReadAll(resp2.Body)
		if err2 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
		}
		//body2 := string(b2)
		var res struct {
			IsAllowedAccess  bool   `json:"isAllowedAccess"`
			IsAllowedCountry bool   `json:"isAllowedCountry"`
			IsKnownProxy     bool   `json:"isKnownProxy"`
			Country          string `json:"country"`
		}
		// fmt.Println(body2)
		if err := json.Unmarshal(b2, &res); err != nil {
			return model.Result{Name: name, Status: model.StatusErr, Err: err}
		}
		if res.IsAllowedAccess && res.IsAllowedCountry && !res.IsKnownProxy {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		}
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.starz.com failed with code: %d", resp.StatusCode)}
}
