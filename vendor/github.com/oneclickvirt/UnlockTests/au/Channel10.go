package au

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Channel10
// 10play.com.au 仅 ipv4 且 get 请求
// https://e410fasadvz.global.ssl.fastly.net/geo 仅 ipv4 且 get 请求
// https://10play.com.au/geo-web 仅 ipv4 且 get 请求
func Channel10(c *http.Client) model.Result {
	name := "Channel 10"
	hostname := "10play.com.au"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://10play.com.au/geo-web"
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
	//fmt.Println(body)
	if strings.Contains(body, "Sorry, 10 play is not available in your region.") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	url = "https://e410fasadvz.global.ssl.fastly.net/geo"
	resp, err = client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body = string(b)
	//fmt.Println(body)
	var res struct {
		State string `json:"state"`
		Allow bool   `json:"allow"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		if strings.Contains(body, "not available") {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if !res.Allow {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if res.Allow && res.State != "" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, Region: res.State, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get 10play.com.au failed with code: %d", resp.StatusCode)}
}
