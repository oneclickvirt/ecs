package asia

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
)

// HotStar
// api.hotstar.com 双栈 get 请求
func HotStar(c *http.Client) model.Result {
	name := "HotStar"
	hostname := "api.hotstar.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api.hotstar.com/o/v1/page/1557?offset=0&size=20&tao=0&tas=20"
	client := utils.Req(c)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	//fmt.Println(body)
	if resp.StatusCode == 475 || resp.StatusCode == 403 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	resp1, err1 := client.R().Get("https://www.hotstar.com")
	if err1 != nil {
		return model.Result{Name: name, Status: model.StatusUnexpected,
			Err: fmt.Errorf("get api.hotstar.com failed with code1: %d", resp.StatusCode)}
	}
	defer resp1.Body.Close()
	//b, err := io.ReadAll(resp1.Body)
	//if err != nil {
	//	return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	//}
	//body := string(b)
	//fmt.Println(body)
	if resp1.StatusCode == 301 || resp.StatusCode == 475 || resp.StatusCode == 403 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	u := resp1.Header.Get("Location")
	if u == "" {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	// fmt.Println(u)
	t := strings.SplitN(u, "/", 4)
	if len(t) < 4 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if strings.ToLower(t[3]) == "us" {
		return model.Result{Name: name, Status: model.StatusNo, Region: t[3]}
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	return model.Result{Name: name, Status: model.StatusYes, Region: t[3], UnlockType: unlockType}
	// return model.Result{Name: name, Status: model.StatusUnexpected,
	// 	Err: fmt.Errorf("get api.hotstar.com failed with code: %d", resp.StatusCode)}
}
