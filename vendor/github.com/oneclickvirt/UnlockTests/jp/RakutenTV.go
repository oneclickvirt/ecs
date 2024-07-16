package jp

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// RakutenTV
// www.rakuten.tv 仅 ipv4 且 get 请求 带 cloudflare 的 5秒盾 无法使用 "is not available in your country"
// api.tv.rakuten.co.jp 仅 ipv4 且 get 请求 无盾可使用
func RakutenTV(c *http.Client) model.Result {
	name := "Rakuten TV"
	hostname := "rakuten.co.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api.tv.rakuten.co.jp/content/playinfo.json?content_id=476611&device_id=14&trailer=1&auth=0&log=0&serial_code=&tmp_eng_flag=1&multi_audio_support=1&_=1716694365356"
	headers := map[string]string{
		"connection": "keep-alive",
		"Cookie":     "alt_id=kdPG3ErDszsWchi~f3P7Y3Mk; _ra=1716693934724|fbf06bf6-0e63-49bc-b5ae-ea8e785126ba; sec_token=6d518581124ba17c1b9968dca83aba7d441dcf88s%3A40%3A%220f817994db4925695da3375e3248a7552d981647%22%3B",
		"origin":     "https://tv.rakuten.co.jp",
		"referer":    "https://tv.rakuten.co.jp/",
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
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
	if resp.StatusCode == 403 || strings.Contains(body, "海外からのアクセスのため、動画を再生できません。") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if resp.StatusCode == 200 {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.rakuten.tv failed with code: %d", resp.StatusCode)}
}
