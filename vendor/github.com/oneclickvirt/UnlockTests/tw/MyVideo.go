package tw

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// MyVideo
// www.myvideo.net.tw 仅 ipv4 且 get 请求
func MyVideo(c *http.Client) model.Result {
	name := "MyVideo"
	hostname := "myvideo.net.tw"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.myvideo.net.tw/login.do"
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
	if strings.Contains(body, "serviceAreaBlock") {
		return model.Result{Name: name, Status: model.StatusNo}
	} else {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
}
