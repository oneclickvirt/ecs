package eu

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Docplay
// - AU 、 New Zealand 、UK
// www.docplay.com 仅 ipv4 且 get 请求
func Docplay(c *http.Client) model.Result {
	name := "Docplay"
	hostname := "docplay.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.docplay.com/subscribe"
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
	if strings.Contains(body, "DocPlay hasn't launched in your part of the world yet.") ||
		resp.Request.URL.String() == "https://www.docplay.com/geoblocked" ||
		strings.Contains(resp.Header.Get("Set-Cookie"), "geoblocked=true") {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.docplay.com failed with code: %d", resp.StatusCode)}
}
