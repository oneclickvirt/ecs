package transnation

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Bing
// www.bing.com 双栈 且 post 请求
func Bing(c *http.Client) model.Result {
	name := "Bing Region"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.bing.com/search?q=www.spiritysdx.top"
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
	if resp.StatusCode == 200 {
		region := utils.ReParse(body, `Region:"([^"]*)"`)
		if region == "CN" {
			if strings.Contains(body, "cn.bing.com") {
				return model.Result{Name: name, Status: model.StatusYes, Region: "cn", Info: "Only cn.bing.com"}
			}
			return model.Result{Name: name, Status: model.StatusNo, Region: "cn"}
		}
		if region != "" {
			return model.Result{Name: name, Status: model.StatusYes, Region: strings.ToLower(region)}
		}
	}
	if strings.Contains(body, "cn.bing.com") {
		return model.Result{Name: name, Status: model.StatusYes, Region: "cn", Info: "Only cn.bing.com"}
	}
	if resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusBanned}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.bing.com failed with code: %d", resp.StatusCode)}
}
