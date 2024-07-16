package asia

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// Bilibili
// B站主体请求逻辑
func Bilibili(c *http.Client, name, url string) model.Result {
	hostname := "bilibili.com"
	if c == nil {
		return model.Result{Name: name}
	}
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
	var res struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		if strings.Contains(body, "抱歉您所在地区不可观看") {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if strings.Contains(body, "抱歉您所在地区不可观看") || strings.Contains(body, "The area is inaccessible") ||
		res.Code == 10004001 || res.Code == 10003003 || res.Code == -10403 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if res.Code == 0 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		//fmt.Println(result1, result2, result3)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get api.bilibili.com failed with code: %d", resp.StatusCode)}
}
