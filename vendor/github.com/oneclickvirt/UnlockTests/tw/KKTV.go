package tw

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// KKTV
// api.kktv.me 仅 ipv4 且 get 请求
func KKTV(c *http.Client) model.Result {
	name := "KKTV"
	hostname := "kktv.me"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api.kktv.me/v3/ipcheck"
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
	var res struct {
		Data struct {
			Country   string `json:"country"`
			IsAllowed bool   `json:"is_allowed"`
		} `json:"data"`
	}
	if err = json.Unmarshal(b, &res); err != nil {
		if strings.Contains(body, "\"is_allowed\":false") {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.Data.Country == "TW" && res.Data.IsAllowed {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	if !res.Data.IsAllowed {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get api.kktv.me failed with head: %d", resp.StatusCode)}
}
