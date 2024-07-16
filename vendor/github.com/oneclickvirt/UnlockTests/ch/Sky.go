package ch

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// SkyCh
// sky.ch 双栈 且 get 请求
func SkyCh(c *http.Client) model.Result {
	name := "SKY CH"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://sky.ch/"
	hostname := "sky.ch"
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
	if strings.Contains(body, "out-of-country") || strings.Contains(body, "Are you using a VPN") ||
		strings.Contains(body, "Are you using a Proxy or similar Anonymizer technics") {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get sky.ch failed with code: %d", resp.StatusCode)}
}
