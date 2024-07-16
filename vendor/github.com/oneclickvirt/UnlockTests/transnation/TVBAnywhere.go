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

// TVBAnywhere
// uapisfm.tvbanywhere.com.sg 仅 ipv4 且 get 请求
func TVBAnywhere(c *http.Client) model.Result {
	name := "TVBAnywhere+"
	hostname := "uapisfm.tvbanywhere.com.sg"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://uapisfm.tvbanywhere.com.sg/geoip/check/platform/android"
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
	//body := string(b)
	//fmt.Println(body)
	var res struct {
		AllowInThisCountry bool   `json:"allow_in_this_country"`
		Country            string `json:"country"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.AllowInThisCountry && res.Country != "" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType,
			Region: strings.ToLower(res.Country)}
	} else if !res.AllowInThisCountry && res.Country != "" {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get uapisfm.tvbanywhere.com.sg failed with code: %d", resp.StatusCode)}
}
