package kr

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NaverTV
// apis.naver.com 仅 ipv4 且 get 请求
func NaverTV(c *http.Client) model.Result {
	name := "Naver TV"
	hostname := "naver.com"
	if c == nil {
		return model.Result{Name: name}
	}
	ts := time.Now().UnixNano() / int64(time.Millisecond)
	baseURL := "https://apis.naver.com/"
	key := "nbxvs5nwNG9QKEWK0ADjYA4JZoujF4gHcIwvoCxFTPAeamq5eemvt5IWAYXxrbYM"
	signText := fmt.Sprintf("https://apis.naver.com/now_web2/now_web_api/v1/clips/31030608/play-info%d", ts)
	// 生成 HMAC-SHA1 签名并进行 base64 编码
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(signText))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	// URL 对签名进行编码
	signatureEncoded := url.QueryEscape(signature)
	reqURL := fmt.Sprintf("%snow_web2/now_web_api/v1/clips/31030608/play-info?msgpad=%d&md=%s", baseURL, ts, signatureEncoded)
	// 进行请求
	headers := map[string]string{
		"User-Agent": model.UA_Browser,
		"Host":       "apis.naver.com",
		"Connection": "keep-alive",
		"Accept":     "application/json, text/plain, */*",
		"Origin":     "https://tv.naver.com",
		"Referer":    "https://tv.naver.com/v/31030608",
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().Get(reqURL)
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
		var res struct {
			Result struct {
				Play struct {
					Playable string `json:"playable"`
				} `json:"play"`
			} `json:"result"`
		}
		if err := json.Unmarshal(b, &res); err != nil {
			if strings.Contains(body, "NOT_COUNTRY_AVAILABLE") {
				return model.Result{Name: name, Status: model.StatusNo}
			}
			return model.Result{Name: name, Status: model.StatusErr, Err: err}
		}
		if res.Result.Play.Playable == "NOT_COUNTRY_AVAILABLE" {
			return model.Result{Name: name, Status: model.StatusNo}
		} else if res.Result.Play.Playable != "NOT_COUNTRY_AVAILABLE" {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get apis.naver.com failed with code: %d", resp.StatusCode)}
}
