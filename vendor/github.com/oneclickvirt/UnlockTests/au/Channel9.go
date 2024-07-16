package au

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Channel9
// login.nine.com.au 双栈 且 get 请求
func Channel9(c *http.Client) model.Result {
	name := "Channel 9"
	hostname := "nine.com.au"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://login.nine.com.au"
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
	if strings.Contains(body, "Geoblock") || resp.StatusCode == 403 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if strings.Contains(body, "Log in to") || resp.StatusCode == 302 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get login.nine.com.au failed with code: %d", resp.StatusCode)}
}
