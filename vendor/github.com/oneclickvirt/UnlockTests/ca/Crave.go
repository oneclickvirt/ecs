package ca

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Crave
// capi.9c9media.com 仅 ipv4 且 get 请求
func Crave(c *http.Client) model.Result {
	name := "Crave"
	hostname := "9c9media.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://capi.9c9media.com/destinations/se_atexace/platforms/desktop/bond/contents/2205173/contentpackages/4279732/manifest.mpd"
	client := utils.ReqDefault(c)
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
	if strings.Contains(body, "Geo Constraint Restrictions") || resp.StatusCode == 404 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if strings.Contains(body, "video.9c9media.com") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get capi.9c9media.com with code: %d", resp.StatusCode)}
}
