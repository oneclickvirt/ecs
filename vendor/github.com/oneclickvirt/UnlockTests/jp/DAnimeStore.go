package jp

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// DAnimeStore
// animestore.docomo.ne.jp 仅 ipv4 且 get 请求
func DAnimeStore(c *http.Client) model.Result {
	name := "D Anime Store"
	hostname := "docomo.ne.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://animestore.docomo.ne.jp/animestore/reg_pc"
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
	if resp.StatusCode == 403 || resp.StatusCode == 451 || strings.Contains(body, "海外") {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 && body != "" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get animestore.docomo.ne.jp failed with code: %d", resp.StatusCode)}
}
