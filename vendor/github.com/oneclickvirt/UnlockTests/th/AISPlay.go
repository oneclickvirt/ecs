package th

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gofrs/uuid/v5"
	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
	"time"
)

func genUUID() string {
	fakeUuid, _ := uuid.NewV4()
	return fakeUuid.String()
}

func generateMD5(input string) string {
	hash := md5.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

func extractValue(body, start, end string) string {
	startIndex := strings.Index(body, start)
	if startIndex == -1 {
		return ""
	}
	startIndex += len(start)
	endIndex := strings.Index(body[startIndex:], end)
	if endIndex == -1 {
		return ""
	}
	return body[startIndex : startIndex+endIndex]
}

func extractHeaderValue(resp *req.Response, headerName string) string {
	if resp.Header.Get(headerName) != "" {
		return resp.Header.Get(headerName)
	}
	return ""
}

// AISPlay
func AISPlay(c *http.Client) model.Result {
	name := "AIS Play"
	hostname := "ais-vidnt.com"
	if c == nil {
		return model.Result{Name: name}
	}
	userId := "09e8b25510"
	userPasswd := "e49e9f9e7f"
	fakeApiKey := generateMD5(genUUID())
	fakeUdid := generateMD5(genUUID())
	timestamp := fmt.Sprint(time.Now().Unix())
	url := fmt.Sprintf("https://web-tls.ais-vidnt.com/device/login/?d=gstweb&gst=1&user=%s&pass=%s", userId, userPasswd)
	headers := map[string]string{
		"accept-language":    "th",
		"api-version":        "2.8.2",
		"api_key":            fakeApiKey,
		"content-type":       "multipart/form-data; boundary=----WebKitFormBoundaryBj2RhUIW7BtRvfK0",
		"device-info":        "com.vimmi.ais.portal, Windows + Chrome, AppVersion: 4.9.97, 10, language: tha",
		"origin":             "https://aisplay.ais.co.th",
		"privateid":          userId,
		"referer":            "https://aisplay.ais.co.th/",
		"sec-ch-ua":          `\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"98\", \"Google Chrome\";v=\"98\"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "cross-site",
		"time":               timestamp,
		"udid":               fakeUdid,
	}
	playload := "------WebKitFormBoundaryBj2RhUIW7BtRvfK0--"
	resp, body, err := utils.PostJson(c, url, playload, headers)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	sId := extractValue(body, `"sid" : "`, `"`)
	datAuth := extractValue(body, `"dat" : "`, `"`)
	if sId == "" || datAuth == "" {
		// fmt.Println(body)
		// return model.Result{Name: name, Status: model.StatusErr, Err: fmt.Errorf("sid or datauth is null")}
		return AnotherAISPlay(c)
	}

	timestamp = fmt.Sprint(time.Now().Unix())
	url = "https://web-sila.ais-vidnt.com/playtemplate/?d=gstweb"
	headers["dat"] = datAuth
	headers["sid"] = sId
	headers["time"] = timestamp
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err = client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body = string(b)

	tmpLiveUrl := extractValue(body, `"live" : "`, `"`)
	if tmpLiveUrl == "" {
		// return model.Result{Name: name, Status: model.StatusErr, Err: fmt.Errorf("tmpLiveUrl is null")}
		return AnotherAISPlay(c)
	}

	mediaId := "B0006"
	realLiveUrl := strings.ReplaceAll(tmpLiveUrl, "{MID}", mediaId)
	realLiveUrl = strings.ReplaceAll(realLiveUrl, "metadata.xml", "metadata.json")
	realLiveUrl = fmt.Sprintf("%s-https&tuid=%s&tdid=%s&chunkHttps=true&origin=anevia", realLiveUrl, userId, fakeUdid)

	headers2 := map[string]string{
		"Accept-Language":    "en-US,en;q=0.9",
		"Origin":             "https://web-player.ais-vidnt.com",
		"Referer":            "https://web-player.ais-vidnt.com/",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
		"sec-ch-ua":          `\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"98\", \"Google Chrome\";v=\"98\"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
	}
	client2 := utils.Req(c)
	client2 = utils.SetReqHeaders(client2, headers2)
	resp2, err := client2.R().Get(realLiveUrl)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp2.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body = string(b)
	playUrl := extractValue(body, `"url" : "`, `"`)
	if playUrl == "" {
		return model.Result{Name: name, Status: model.StatusErr, Err: fmt.Errorf("playUrl is null")}
	}

	headers3 := map[string]string{
		"Accept-Language":    "en-US,en;q=0.9",
		"Origin":             "https://web-player.ais-vidnt.com",
		"Referer":            "https://web-player.ais-vidnt.com/",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
		"sec-ch-ua":          `\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"98\", \"Google Chrome\";v=\"98\"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
	}
	client3 := utils.Req(c)
	client3 = utils.SetReqHeaders(client3, headers3)
	resp3, err := client3.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp3.Body.Close()

	baseRequstCheckStatus := extractHeaderValue(resp3, "X-Base-Request-Check-Status")
	if baseRequstCheckStatus == "INCORRECT" {
		// return model.Result{Name: name, Status: model.StatusErr,
		// 	Err: fmt.Errorf("X-Base-Request-Check-Status is INCORRECT")}
		return AnotherAISPlay(c)
	}

	result := extractHeaderValue(resp, "X-Geo-Protection-System-Status")
	fmt.Println(result)
	switch result {
	case "BLOCK":
		return model.Result{Name: name, Status: model.StatusNo}
	case "SUCCESS":
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	default:
		return AnotherAISPlay(c)
	}
}

// AnotherAISPlay
// 49-231-37-237-rewriter.ais-vidnt.com 双栈 get 请求
func AnotherAISPlay(c *http.Client) model.Result {
	name := "AIS Play"
	hostname := "ais-vidnt.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://49-231-37-237-rewriter.ais-vidnt.com/ais/play/origin/VOD/playlist/ais-yMzNH1-bGUxc/index.m3u8"
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
	if resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		if strings.Contains(body, "X-Geo-Protection-System-Status") {
			if strings.Contains(body, "ALLOW") {
				result1, result2, result3 := utils.CheckDNS(hostname)
				unlockType := utils.GetUnlockType(result1, result2, result3)
				return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
			} else if strings.Contains(body, "BLOCK") {
				return model.Result{Name: name, Status: model.StatusNo}
			}
		}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get 49-231-37-237-rewriter.ais-vidnt.com failed with code: %d", resp.StatusCode)}
}
