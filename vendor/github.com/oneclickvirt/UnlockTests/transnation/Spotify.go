package transnation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// Spotify
// spclient.wg.spotify.com 双栈 且 post 请求
func Spotify(c *http.Client) model.Result {
	name := "Spotify Registration"
	hostname := "spclient.wg.spotify.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://spclient.wg.spotify.com/signup/public/v1/account"
	headers := map[string]string{
		"User-Agent":      model.UA_Browser,
		"Accept-Language": "en",
		"content-type":    "application/json",
		"cache-control":   "no-cache",
	}
	payload := "birth_day=11&birth_month=11&birth_year=2000&collect_personal_info=undefined&creation_flow=&creation_point=https%3A%2F%2Fwww.spotify.com%2Fhk-en%2F&displayname=Gay%20Lord&gender=male&iagree=1&key=a1e486e2729f46d6bb368d6b2bcda326&platform=www&referrer=&send-email=0&thirdpartyemail=0&identifier_token=AgE6YTvEzkReHNfJpO114514"
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().SetQueryString(payload).Post(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	var res struct {
		Status            int    `json:"status"`
		Country           string `json:"country"`
		IsCountryLaunched bool   `json:"is_country_launched"`
	}
	// body := string(b)
	// fmt.Println(body, resp.StatusCode)
	if err := json.Unmarshal(b, &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.Status == 320 || res.Status == 120 || resp.StatusCode == 401 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if res.Status == 311 && res.IsCountryLaunched {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType,
			Region: strings.ToLower(res.Country)}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get spclient.wg.spotify.com failed with code: %d", resp.StatusCode)}
}
