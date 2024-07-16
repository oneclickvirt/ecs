package transnation

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// ViuCom
// www.viu.com 仅 ipv4 且 get 请求
func ViuCom(c *http.Client) model.Result {
	name := "Viu.com"
	hostname := "www.viu.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.viu.com"
	client := utils.Req(c)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	// b, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	// }
	// body := string(b)
	// fmt.Println(body)
	location := fmt.Sprintf("%s", resp.Request.URL)
	// fmt.Println(location)
	if strings.Contains(location, "no-service") {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if location != "" {
		regions := strings.Split(location, "/")
		if regions[len(regions)-1] == "no-service" || strings.Contains(location, "no-service") {
			return model.Result{Name: name, Status: model.StatusNo}
		}
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		if len(regions) >= 4 {
			return model.Result{Name: name, Status: model.StatusYes,
				Region: strings.ToLower(regions[len(regions)-1]), UnlockType: unlockType}
		} else {
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.viu.com failed with code: %d", resp.StatusCode)}
}
