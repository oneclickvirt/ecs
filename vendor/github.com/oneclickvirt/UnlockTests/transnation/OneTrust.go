package transnation

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
)

// OneTrust
// geolocation.onetrust.com 双栈 get 请求
func OneTrust(c *http.Client) model.Result {
	name := "OneTrust"
	hostname := "onetrust.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://geolocation.onetrust.com/cookieconsentpub/v1/geo/location/dnsfeed"
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
	body := string(b)
	country := utils.ReParse(body, `"country"\s*:\s*"([^"]+)"`)
	stateName := utils.ReParse(body, `"stateName"\s*:\s*"([^"]+)"`)
	if country == "" {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	if stateName == "" {
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType, Region: country}
	} else {
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType,
			Region: country + " " + stateName}
	}
}
