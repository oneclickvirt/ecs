package eu

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid/v5"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Eurosport
// www.eurosport.com 双栈
func Eurosport(c *http.Client) model.Result {
	name := "Eurosport RO"
	hostname := "eurosport.com"
	if c == nil {
		return model.Result{Name: name}
	}
	fakeUuid, _ := uuid.NewV4()
	url := "https://eu3-prod-direct.eurosport.ro/token?realm=eurosport"
	headers := map[string]string{
		"User-Agent":         model.UA_Browser,
		"accept":             "*/*",
		"accept-language":    "en-US,en;q=0.9",
		"origin":             "https://www.eurosport.ro",
		"referer":            "https://www.eurosport.ro/",
		"sec-ch-ua":          model.UA_SecCHUA,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"x-device-info":      fmt.Sprintf("escom/0.295.1 (unknown/unknown; Windows/10; %s)", fakeUuid),
		"x-disco-client":     "WEB:UNKNOWN:escom:0.295.1",
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp1, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp1.Body.Close()
	b, err := io.ReadAll(resp1.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	//body1 := string(b)
	//fmt.Println(body1)
	var res1 struct {
		Data struct {
			Attributes struct {
				Realm string `json:"realm"`
				Token string `json:"token"`
			} `json:"attributes"`
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &res1); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res1.Data.Attributes.Token != "" {
		//fmt.Println(res1.Data.Attributes.Token)
		sourceSystemId := "eurosport-vid2133403"
		playbackUrl := fmt.Sprintf("https://eu3-prod-direct.eurosport.ro/playback/v2/videoPlaybackInfo/sourceSystemId/%s?usePreAuth=true", sourceSystemId)
		headers2 := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", res1.Data.Attributes.Token),
		}
		client = utils.SetReqHeaders(client, headers2)
		resp2, err2 := client.R().Get(playbackUrl)
		if err2 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
		}
		defer resp2.Body.Close()
		b, err = io.ReadAll(resp2.Body)
		if err != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
		}
		body2 := string(b)
		//fmt.Println(body2)
		isBlocked := strings.Contains(body2, "access.denied.geoblocked")
		isOK := strings.Contains(body2, "eurosport-vod")
		if (!isBlocked && !isOK) || isBlocked {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		if isOK {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get eu3-prod-direct.eurosport.ro failed with code: %d", resp1.StatusCode)}
}
