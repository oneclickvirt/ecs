package us

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// NFLPlus
// www.dazn.com 仅 ipv4 且 get 请求
// https://www.nfl.com/plus/ 重定向至于 https://nfl.com/dazn-watch-gp-row 约等于仅使用 dazn 进行观看
func NFLPlus(c *http.Client) model.Result {
	name := "NFL+"
	hostname := "nfl.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://nfl.com/dazn-watch-gp-row"
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
	lowBody := strings.ToLower(body)
	if strings.Contains(lowBody, "nflgamepass") || strings.Contains(lowBody, "nfl-game-pass") ||
		strings.Contains(lowBody, "gpi.nfl.com") {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get https://nfl.com/dazn-watch-gp-row failed with code: %d", resp.StatusCode)}
}
