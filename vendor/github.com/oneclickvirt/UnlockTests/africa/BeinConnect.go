package africa

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// BeinConnect
// proxies.bein-mena-production.eu-west-2.tuc.red 仅 ipv4 且 get 请求
func BeinConnect(c *http.Client) model.Result {
	name := "Bein Sports Connect"
	hostname := "beinconnect.com.tr"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://proxies.bein-mena-production.eu-west-2.tuc.red/proxy/availableOffers"
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
	if strings.Contains(body, "Unavailable For Legal Reasons") ||
		resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 || resp.StatusCode == 500 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get proxies.bein-mena-production.eu-west-2.tuc.red failed with code: %d", resp.StatusCode)}
}
