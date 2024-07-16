package jp

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

type PlatformResponse struct {
	PlatformUID   string `json:"platform_uid"`
	PlatformToken string `json:"platform_token"`
}

type Episode struct {
	AccountID  string `json:"accountID"`
	PlayerID   string `json:"playerID"`
	VideoID    string `json:"videoID"`
	VideoRefID string `json:"videoRefID"`
}

func getEpisodeID(body string) string {
	if idx := strings.Index(body, `"newer-drama"`); idx != -1 {
		body = body[idx:]
		if idx = strings.Index(body, `"id"`); idx != -1 {
			body = body[idx:]
			parts := strings.Split(body, `"`)
			if len(parts) > 3 {
				return parts[3]
			}
		}
	}
	return ""
}

func getPolicyKey(body string) string {
	if idx := strings.Index(body, `policyKey:"`); idx != -1 {
		body = body[idx+len(`policyKey:"`):]
		parts := strings.Split(body, `"`)
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return ""
}

func getDeliveryConfigID(body string) string {
	if idx := strings.Index(body, `deliveryConfigId:"`); idx != -1 {
		body = body[idx+len(`deliveryConfigId:"`):]
		parts := strings.Split(body, `"`)
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return ""
}

// TVer
// edge.api.brightcove.com 仅 ipv4 且 get 请求
// 双重检测逻辑
func TVer(c *http.Client) model.Result {
	firstCheck := FirstTVer(c)
	if firstCheck.Status == model.StatusNetworkErr || firstCheck.Status == model.StatusErr {
		secondCheck := AnotherTVer(c)
		if secondCheck.Status == model.StatusNetworkErr || secondCheck.Status == model.StatusErr {
			return firstCheck
		} else {
			return secondCheck
		}
	}
	return firstCheck
}

// FirstTVer
// 主要的检测逻辑
func FirstTVer(c *http.Client) model.Result {
	name := "TVer"
	hostname := "tver.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	// 创建平台用户
	headers := map[string]string{
		"content-type":       "application/x-www-form-urlencoded",
		"origin":             "https://s.tver.jp",
		"referer":            "https://s.tver.jp/",
		"sec-ch-ua":          model.UA_SecCHUA,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "Windows",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"user-agent":         model.UA_Browser,
	}
	res, body, err := utils.PostJson(c, "https://platform-api.tver.jp/v2/api/platform_users/browser/create",
		"device_type=pc", headers)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return model.Result{Name: name, Status: model.StatusNetworkErr,
			Err: fmt.Errorf("1. get platform-api.tver.jp failed with code: %d", res.StatusCode)}
	}
	var platformResp PlatformResponse
	if err := json.Unmarshal([]byte(body), &platformResp); err != nil {
		return model.Result{Name: name, Status: model.StatusUnexpected}
	}

	// 获取当前播放的剧集
	url := fmt.Sprintf("https://platform-api.tver.jp/service/api/v1/callHome?"+
		"platform_uid=%s&platform_token=%s&require_data=mylist%%2Cresume%%2Clater", platformResp.PlatformUID,
		platformResp.PlatformToken)
	headers2 := map[string]string{
		"origin":               "https://tver.jp",
		"referer":              "https://tver.jp/",
		"sec-ch-ua":            model.UA_SecCHUA,
		"sec-ch-ua-mobile":     "?0",
		"sec-ch-ua-platform":   "Windows",
		"sec-fetch-dest":       "empty",
		"sec-fetch-mode":       "cors",
		"sec-fetch-site":       "same-site",
		"x-tver-platform-type": "web",
		"user-agent":           model.UA_Browser,
	}
	client2 := utils.Req(c)
	client2 = utils.SetReqHeaders(client2, headers2)
	resp2, err := client2.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp2.Body.Close()
	b, err := io.ReadAll(resp2.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body2 := string(b)
	if res.StatusCode != 200 {
		return model.Result{Name: name, Status: model.StatusNetworkErr,
			Err: fmt.Errorf("2. get platform-api.tver.jp failed with code: %d", res.StatusCode)}
	}
	episodeId := getEpisodeID(body2)
	if episodeId == "" {
		return model.Result{Name: name, Status: model.StatusUnexpected, Err: fmt.Errorf("failed (No Episode ID)")}
	}

	// 获取剧集的信息
	url = fmt.Sprintf("https://statics.tver.jp/content/episode/%s.json", episodeId)
	headers3 := map[string]string{
		"origin":             "https://tver.jp",
		"referer":            "https://tver.jp/",
		"sec-ch-ua":          model.UA_SecCHUA,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "Windows",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"user-agent":         model.UA_Browser,
	}
	client3 := utils.Req(c)
	client3 = utils.SetReqHeaders(client3, headers3)
	resp3, err := client3.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp3.Body.Close()
	b3, err := io.ReadAll(resp3.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	//body = string(b3)
	if res.StatusCode != 200 {
		return model.Result{Name: name, Status: model.StatusNetworkErr,
			Err: fmt.Errorf("get platform-api.tver.jp failed with code: %d", res.StatusCode)}
	}
	var episode Episode
	if err := json.Unmarshal(b3, &episode); err != nil {
		return model.Result{Name: name, Status: model.StatusUnexpected, Err: fmt.Errorf("failed (Parsing JSON)")}
	}

	// 获取 Brightcove 播放器信息
	url = fmt.Sprintf("https://players.brightcove.net/%s/%s_default/index.min.js", episode.AccountID,
		episode.PlayerID)
	headers4 := map[string]string{
		"Referer":            "https://tver.jp/",
		"Sec-Fetch-Dest":     "script",
		"Sec-Fetch-Mode":     "no-cors",
		"Sec-Fetch-Site":     "cross-site",
		"sec-ch-ua":          model.UA_SecCHUA,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "Windows",
		"user-agent":         model.UA_Browser,
	}
	client4 := utils.Req(c)
	client4 = utils.SetReqHeaders(client4, headers4)
	resp4, err := client4.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp4.Body.Close()
	//b4, err := io.ReadAll(resp4.Body)
	//if err != nil {
	//	return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	//}
	//body4 := string(b4)
	if res.StatusCode != 200 {
		return model.Result{Name: name, Status: model.StatusNetworkErr,
			Err: fmt.Errorf("get platform-api.tver.jp failed with code: %d", res.StatusCode)}
	}
	policyKey := getPolicyKey(body)
	if policyKey == "" {
		return model.Result{Name: name, Status: model.StatusUnexpected, Err: fmt.Errorf("failed (No policyKey)")}
	}

	// 最终测试
	var finalURL string
	if episode.VideoRefID == "" {
		deliveryConfigId := getDeliveryConfigID(body)
		if deliveryConfigId != "" {
			finalURL = fmt.Sprintf("https://edge.api.brightcove.com/playback/v1/accounts/%s/videos/%s?config_id=%s",
				episode.AccountID, episode.VideoID, deliveryConfigId)
		} else {
			finalURL = fmt.Sprintf("https://edge.api.brightcove.com/playback/v1/accounts/%s/videos/ref%%3A%s",
				episode.AccountID, episode.VideoRefID)
		}
	} else {
		finalURL = fmt.Sprintf("https://edge.api.brightcove.com/playback/v1/accounts/%s/videos/ref%%3A%s",
			episode.AccountID, episode.VideoRefID)
	}

	// 构建请求
	headers5 := map[string]string{
		"accept":             fmt.Sprintf("application/json;pk=%s", policyKey),
		"origin":             "https://tver.jp",
		"referer":            "https://tver.jp/",
		"sec-ch-ua":          model.UA_SecCHUA,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "Windows",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "cross-site",
		"user-agent":         model.UA_Browser,
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers5)
	resp, err := client.R().Get(finalURL)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	body = string(b)
	var res1 struct {
		ErrorSubcode string `json:"error_subcode"`
		AccountId    string `json:"account_id"`
	}
	var res2 []struct {
		ClientGeo    string `json:"client_geo"`
		ErrorSubcode string `json:"error_subcode"`
		ErrorCode    string `json:"error_code"`
		Message      string `json:"message"`
	}
	if err := json.Unmarshal(b, &res1); err != nil {
		if err := json.Unmarshal(b, &res2); err != nil {
			if strings.Contains(body, "CLIENT_GEO") || strings.Contains(body, "ACCESS_DENIED") {
				return model.Result{Name: name, Status: model.StatusNo}
			}
			return model.Result{Name: name, Status: model.StatusErr, Err: err}
		}
		if res2[0].ErrorSubcode == "CLIENT_GEO" {
			return model.Result{Name: name, Status: model.StatusNo, Region: res2[0].ClientGeo}
		}
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res1.AccountId != "0" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType, Region: "jp"}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get edge.api.brightcove.com failed with code: %d", resp.StatusCode)}
}

// AnotherTVer
func AnotherTVer(c *http.Client) model.Result {
	name := "TVer"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://edge.api.brightcove.com/playback/v1/accounts/5102072605001/videos/ref%3Akaguyasama_01"
	headers := map[string]string{
		"User-Agent": model.UA_Browser,
		"Accept":     "application/json;pk=BCpkADawqM0_rzsjsYbC1k1wlJLU4HiAtfzjxdUmfvvLUQB-Ax6VA-p-9wOEZbCEm3u95qq2Y1CQQW1K9tPaMma9iAqUqhpISCmyXrgnlpx9soEmoVNuQpiyGsTpePGumWxSs1YoKziYB6Wz",
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Status: model.StatusNetworkErr, Err: err}
	}
	body := string(b)
	var res1 struct {
		ErrorSubcode string `json:"error_subcode"`
		AccountId    string `json:"account_id"`
	}
	var res2 []struct {
		ClientGeo    string `json:"client_geo"`
		ErrorSubcode string `json:"error_subcode"`
		ErrorCode    string `json:"error_code"`
		Message      string `json:"message"`
	}
	//fmt.Println(body)
	if err := json.Unmarshal(b, &res1); err != nil {
		if err := json.Unmarshal(b, &res2); err != nil {
			if strings.Contains(body, "CLIENT_GEO") || strings.Contains(body, "ACCESS_DENIED") {
				return model.Result{
					Status: model.StatusNo,
				}
			}
			return model.Result{Status: model.StatusErr, Err: err}
		}
		if res2[0].ErrorSubcode == "CLIENT_GEO" {
			return model.Result{Status: model.StatusNo, Region: res2[0].ClientGeo}
		}
		return model.Result{Status: model.StatusErr, Err: err}
	}
	if res1.AccountId != "0" {
		return model.Result{Status: model.StatusYes, Region: "jp"}
	}
	return model.Result{Status: model.StatusUnexpected,
		Err: fmt.Errorf("get edge.api.brightcove.com failed with code: %d", resp.StatusCode)}
}
