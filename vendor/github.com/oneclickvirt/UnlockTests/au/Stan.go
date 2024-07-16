package au

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
	"strings"
)

// Stan
// api.stan.com.au 仅 ipv4 且 post 请求
func Stan(c *http.Client) model.Result {
	name := "Stan"
	hostname := "stan.com.au"
	if c == nil {
		return model.Result{Name: name}
	}
	resp, body, err := utils.PostJson(c, "https://api.stan.com.au/login/v1/sessions/web/account", "{}", nil)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	if strings.Contains(string(body), "Access Denied") || resp.StatusCode == 404 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if strings.Contains(string(body), "VPNDetected") {
		return model.Result{Name: name, Status: model.StatusNo, Info: "VPN Detected"}
	}
	if resp.StatusCode == 400 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get api.stan.com.au failed with code: %d", resp.StatusCode)}
}
