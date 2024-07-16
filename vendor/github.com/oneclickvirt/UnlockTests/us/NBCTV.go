package us

import (
	"fmt"
	"github.com/gofrs/uuid/v5"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// NBCTV
// geolocation.digitalsvc.apps.nbcuni.com 双栈 get 请求
func NBCTV(c *http.Client) model.Result {
	name := "NBC TV"
	hostname := "nbcuni.com"
	if c == nil {
		return model.Result{Name: name}
	}
	fakeUuid, _ := uuid.NewV4()
	url := "https://geolocation.digitalsvc.apps.nbcuni.com/geolocation/live/usa"
	client := utils.Req(c)
	headers := map[string]string{
		"accept-language":    "en-US,en;q=0.9",
		"app-session-id":     fakeUuid.String(),
		"authorization":      "NBC-Basic key=\"usa_live\", version=\"3.0\", type=\"cpc\"",
		"client":             "oneapp",
		"content-type":       "application/json",
		"origin":             "https://www.nbc.com",
		"referer":            "https://www.nbc.com/",
		"sec-ch-ua":          model.UA_SecCHUA,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "cross-site",
	}
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().SetBodyJsonString(`{"adobeMvpdId":null,"serviceZip":null,"device":"web"}`).Post(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	body := string(b)
	// fmt.Println(body)
	if strings.Contains(body, `"restricted":false`) {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	} else if strings.Contains(body, `"restricted":true`) || body == "" {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("getgeolocation.digitalsvc.apps.nbcuni.com failed with code: %d", resp.StatusCode)}
}
