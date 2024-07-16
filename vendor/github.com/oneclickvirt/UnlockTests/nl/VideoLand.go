package nl

import (
	"encoding/json"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// VideoLand
// api.videoland.com 双栈 且 post 请求
func VideoLand(c *http.Client) model.Result {
	name := "Videoland"
	hostname := "videoland.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api.videoland.com/subscribe/videoland-account/graphql"
	payload := `{"operationName":"IsOnboardingGeoBlocked","variables":{},"query":"query IsOnboardingGeoBlocked {\n  isOnboardingGeoBlocked\n}\n"}`
	headers := map[string]string{
		"connection":                "keep-alive",
		"apollographql-client-name": "apollo_accounts_base",
		"traceparent":               "00-cab2dbd109bf1e003903ec43eb4c067d-623ef8e56174b85a-01",
		"origin":                    "https://www.videoland.com",
		"referer":                   "https://www.videoland.com/",
		"accept":                    "application/json, text/plain, */*",
	}
	client := utils.ReqDefault(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().SetBodyString(payload).Post(url)
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
	var res struct {
		Data struct {
			Blocked bool `json:"isOnboardingGeoBlocked"`
		} `json:"data"`
	}
	if err = json.Unmarshal(b, &res); err != nil {
		if strings.Contains(body, "\"isOnboardingGeoBlocked\":true") || body == "" {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.Data.Blocked {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
}
