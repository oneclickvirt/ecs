package jp

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Telasa
// api-videopass-anon.kddi-video.com 双栈 get 请求
func Telasa(c *http.Client) model.Result {
	name := "Telasa"
	hostname := "kddi-video.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api-videopass-anon.kddi-video.com/v1/playback/system_status"
	headers := map[string]string{
		"X-Device-ID": "d36f8e6b-e344-4f5e-9a55-90aeb3403799",
	}
	client := utils.ReqDefault(c)
	client = utils.SetReqHeaders(client, headers)
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
		Status struct {
			Type    string `json:"type"`
			Subtype string `json:"subtype"`
		} `json:"status"`
	}
	//fmt.Println(body)
	if err := json.Unmarshal(b, &res); err != nil {
		if strings.Contains(body, "RequestForbidden") {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.Status.Subtype == "IPLocationNotAllowed" {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if res.Status.Type != "" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get api-videopass-anon.kddi-video.com failed with code: %d", resp.StatusCode)}
}
