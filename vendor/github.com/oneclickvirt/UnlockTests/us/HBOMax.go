package us

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// HBOMax
// www.hbomax.com 仅 ipv4 且 get 请求 (重定向至于 www.max.com 了)
// www.hbonow.com 仅 ipv4 且 get 请求 (重定向至于 www.max.com 了)
func HBOMax(c *http.Client) model.Result {
	name := "HBO Max"
	hostname := "max.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.max.com/"
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
	if strings.Contains(body, "geo-availability") || strings.Contains(resp.Header.Get("location"), "geo-availability") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	t := strings.Split(resp.Header.Get("location"), "/")
	region := ""
	if len(t) >= 4 {
		region = strings.Split(resp.Header.Get("location"), "/")[3]
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	return model.Result{Name: name, Status: model.StatusYes, Region: strings.ToUpper(region), UnlockType: unlockType}
}
