package tw

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
)

// LineTV
// www.linetv.tw 仅 ipv4 且 get 请求
func LineTV(c *http.Client) model.Result {
	name := "LineTV.TW"
	hostname := "linetv.tw"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.linetv.tw/api/part/11829/eps/1/part?chocomemberId="
	client := utils.ReqDefault(c)
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
		CountryCode int `json:"countryCode"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.CountryCode == 228 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	} else if resp.StatusCode == 400 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.linetv.tw failed with code: %d", resp.StatusCode)}
}
