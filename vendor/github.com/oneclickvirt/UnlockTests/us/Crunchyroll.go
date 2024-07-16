package us

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Crunchyroll
// c.evidon.com 仅 ipv4 且 get 请求
func Crunchyroll(c *http.Client) model.Result {
	name := "Crunchyroll"
	hostname := "evidon.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://c.evidon.com/geo/country.js"
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
	if strings.Contains(body, "'code':'us'") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	} else {
		return model.Result{Name: name, Status: model.StatusNo}
	}
}
