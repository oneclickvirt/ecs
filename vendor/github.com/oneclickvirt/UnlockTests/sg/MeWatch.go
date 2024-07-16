package sg

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// MeWatch
// cdn.mewatch.sg 仅 ipv4 且 get 请求
func MeWatch(c *http.Client) model.Result {
	name := "MeWatch"
	hostname := "mewatch.sg"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://cdn.mewatch.sg/api/items/97098/videos?delivery=stream%2Cprogressive&ff=idp%2Cldp%2Crpt%2Ccd&lang=en&resolution=External&segments=all"
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
		Code int `json:"code"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if strings.Contains(body, "You are accessing this item from a location that is not permitted by the license") ||
		res.Code == 8002 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 || strings.Contains(body, "deliveryType") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get cdn.mewatch.sg failed with code: %d", resp.StatusCode)}
}
