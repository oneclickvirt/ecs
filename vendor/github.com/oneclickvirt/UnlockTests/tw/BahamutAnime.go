package tw

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// BahamutAnime
// ani.gamer.com.tw 仅 ipv4 且 get 请求
func BahamutAnime(c *http.Client) model.Result {
	name := "Bahamut Anime"
	hostname := "gamer.com.tw"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://ani.gamer.com.tw/ajax/getdeviceid.php"
	client := utils.Req(c)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	body := string(b)
	var res struct {
		Deviceid string `json:"deviceid"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		if strings.Contains(body, "Just a moment") || strings.Contains(body, "系統異常回報") {
			return model.Result{Name: name, Status: model.StatusNo, Info: "Banned by cloudflare"}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	// fmt.Println(res.Deviceid)
	// 14667
	sn := "37783"
	resp2, err2 := client.R().Get("https://ani.gamer.com.tw/ajax/token.php?adID=89422&sn=" + sn + "&device=" + res.Deviceid)
	if err2 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
	}
	defer resp2.Body.Close()
	b2, err2 := io.ReadAll(resp2.Body)
	if err2 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
	}
	body2 := string(b2)

	resp3, err3 := client.R().Get("https://ani.gamer.com.tw/")
	if err3 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err3}
	}
	defer resp3.Body.Close()
	b3, err3 := io.ReadAll(resp3.Body)
	if err3 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err3}
	}
	body3 := string(b3)
	// fmt.Println(body3)
	if strings.Contains(body2, "\u5f88\u62b1\u6b49\uff01\u672c\u7bc0\u76ee\u56e0\u6388\u6b0a\u56e0\u7d20\u7121\u6cd5\u5728\u60a8\u7684\u6240\u5728\u5340\u57df\u64ad\u653e\u3002") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if (strings.Contains(body2, "animeSn") ||
		strings.Contains(body2, "\u88dd\u7f6e\u9a57\u8b49\u7570\u5e38\uff01")) && strings.Contains(body3, "data-geo") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	} else if resp2.StatusCode == 403 || resp2.StatusCode == 404 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get ani.gamer.com.tw failed with code: %d", resp.StatusCode)}
}
