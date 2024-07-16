package in

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// MXPlayer
// www.mxplayer.in 仅 ipv4 且 get 请求
func MXPlayer(c *http.Client) model.Result {
	name := "MX Player"
	hostname := "mxplayer.in"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.mxplayer.in/"
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
	//fmt.Println(resp.Header.Get("set-cookie"))
	if strings.Contains(body, "We are currently not available in your region") ||
		strings.Contains(body, "403 ERROR") ||
		resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 || resp.Header.Get("set-cookie") != "" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.mxplayer.in failed with code: %d", resp.StatusCode)}
}
