package jp

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
	"time"
)

func getFirstLink(jsonStr string) string {
	type Drama struct {
		Link string `json:"link"`
	}
	var dramas []Drama
	err := json.Unmarshal([]byte(jsonStr), &dramas)
	if err != nil || len(dramas) == 0 {
		return ""
	}
	return dramas[0].Link
}

func getWodUrl(htmlStr string) string {
	lines := strings.Split(htmlStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "https://wod.wowow.co.jp/content/") {
			tempList := strings.Split(line, "https://wod.wowow.co.jp/content/")
			if len(tempList) >= 2 {
				tpList := strings.Split(tempList[1], "\"")
				if len(tpList) >= 2 {
					return "https://wod.wowow.co.jp/content/" + tpList[0]
				} else {
					return ""
				}
			} else {
				return ""
			}
		}
	}
	return ""
}

func getProgramUrl(htmlStr string) string {
	lines := strings.Split(htmlStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "https://wod.wowow.co.jp/program/") {
			tempList := strings.Split(line, ":")
			if len(tempList) >= 2 {
				tpList := strings.Split(tempList[len(tempList)-1], "\"")
				if len(tpList) >= 2 {
					for _, l := range tpList {
						if strings.Contains(l, "//wod.wowow.co.jp/program/") {
							return "https:" + l
						}
					}
				} else {
					return ""
				}
			} else {
				return ""
			}
		}
	}
	return ""
}

func getMetaId(htmlStr string) string {
	lines := strings.Split(htmlStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "https://wod.wowow.co.jp/watch/") {
			tempList := strings.Split(line, "https://wod.wowow.co.jp/watch/")
			if len(tempList) >= 2 {
				tpList := strings.Split(tempList[1], "\"")
				if len(tpList) >= 2 {
					return tpList[0]
				} else {
					return ""
				}
			} else {
				return ""
			}
		}
	}
	return ""
}

// Wowow
// www.wowow.co.jp 仅 ipv4 且 get 请求
func Wowow(c *http.Client) model.Result {
	name := "WOWOW"
	hostname := "wowow.co.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	// 获取当前时间戳
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	// 第一次请求：获取原创剧集列表
	url := fmt.Sprintf("https://www.wowow.co.jp/drama/original/json/lineup.json?_=%d", timestamp)
	headers := map[string]string{
		"Accept":             "application/json, text/javascript, */*; q=0.01",
		"Referer":            "https://www.wowow.co.jp/drama/original/",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-origin",
		"X-Requested-With":   "XMLHttpRequest",
		"sec-ch-ua":          model.UA_SecCHUA,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
		"User-Agent":         model.UA_Browser,
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Info: "1-1", Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Info: "1-1", Err: fmt.Errorf("can not parse body")}
	}
	body := string(b)
	// 获取第一个剧集的链接
	playUrl := getFirstLink(body)
	if playUrl == "" {
		return model.Result{Name: name, Status: model.StatusErr, Err: fmt.Errorf("failed to get play URL")}
	}

	// 第二次请求：获取真实链接
	resp2, err := client.R().Get(playUrl)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Info: "2-1", Err: err}
	}
	defer resp2.Body.Close()
	b2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Info: "2-1", Err: fmt.Errorf("can not parse body")}
	}
	body2 := string(b2)

	// 获取真实链接
	wodUrl := getWodUrl(body2)
	if wodUrl == "" {
		programUrl := getProgramUrl(body2)
		// 第二次请求的二次请求：获取真实链接
		resp22, err22 := client.R().Get(programUrl)
		if err22 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Info: "2-2", Err: err22}
		}
		defer resp22.Body.Close()
		b22, err22 := io.ReadAll(resp22.Body)
		if err22 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
		}
		body = string(b22)
		tempList := strings.Split(body, "\"refId\":\"")
		if len(tempList) >= 2 {
			for _, l := range tempList {
				if strings.Contains(l, "media_meta") {
					tpList := strings.Split(l, "\"")
					if len(tpList) >= 2 {
						wodUrl = "https://wod.wowow.co.jp/content/" + tpList[0]
						break
					}
				}
			}
		}
		if wodUrl == "" {
			return model.Result{Name: name, Status: model.StatusErr, Err: fmt.Errorf("failed to get WOD URL")}
		}
	}

	// 第三次请求：获取 meta_id
	resp3, err := client.R().Get(wodUrl)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Info: "3", Err: err}
	}
	defer resp3.Body.Close()
	b3, err := io.ReadAll(resp3.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Info: "3", Err: fmt.Errorf("can not parse body")}
	}
	body = string(b3)
	metaId := getMetaId(body)
	if metaId == "" {
		return model.Result{Name: name, Status: model.StatusErr, Err: fmt.Errorf("failed to get meta ID")}
	}

	// 生成 vUid
	hash := md5.Sum([]byte(fmt.Sprintf("%d", timestamp)))
	vUid := hex.EncodeToString(hash[:])
	// 最终测试请求
	authUrl := "https://mapi.wowow.co.jp/api/v1/playback/auth"
	data := fmt.Sprintf(`{"meta_id":"%s","vuid":"%s","device_code":1,"app_id":1,"ua":"%s"}`, metaId, vUid, model.UA_Browser)
	headers4 := map[string]string{
		"accept":             "application/json, text/plain, */*",
		"content-type":       "application/json;charset=UTF-8",
		"origin":             "https://wod.wowow.co.jp",
		"referer":            "https://wod.wowow.co.jp/",
		"sec-ch-ua":          model.UA_SecCHUA,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"x-requested-with":   "XMLHttpRequest",
		"User-Agent":         model.UA_Browser,
	}
	resp, body, err = utils.PostJson(c, authUrl, data, headers4)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Info: "4", Err: err}
	}
	//fmt.Println(body)
	// {"error":{"message":"サポート外ネットワークからの接続です。日本国外からの接続、VPN・プロキシ経由の接続等ではご利用いただけません。","code":2055,"type":"Forbidden",
	if strings.Contains(body, "VPN") || strings.Contains(body, "Forbidden") {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if strings.Contains(body, "playback_session_id") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{
		Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get mapi.wowow.co.jp failed with code: %d", resp.StatusCode)}
}
