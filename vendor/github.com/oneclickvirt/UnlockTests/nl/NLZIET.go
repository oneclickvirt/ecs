package nl

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// ZIETCDN
// nlziet.nl 仅 ipv4 且 get 请求
// 直接通过CDN判断地区
func ZIETCDN(c *http.Client) model.Result {
	name := "NLZIET"
	hostname := "nlziet.nl"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://nlziet.nl/cdn-cgi/trace"
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
	tempList := strings.Split(body, "\n")
	var location string
	for _, line := range tempList {
		if strings.HasPrefix(line, "loc=") {
			location = strings.ReplaceAll(line, "loc=", "")
		}
	}
	loc := strings.ToLower(location)
	exit := utils.GetRegion(loc, model.NLZIETSupportCountry)
	if exit {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType, Region: loc}
	}
	return model.Result{Name: name, Status: model.StatusNo}
}

// NLZIET
// nlziet.nl 仅 ipv4 且 get 请求 cookie 有效期非常短
func NLZIET(c *http.Client) model.Result {
	name := "NLZIET"
	hostname := "nlziet.nl"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api.nlziet.nl/v7/stream/handshake/Widevine/Dash/VOD/rzIL9rb-TkSn-ek_wBmvaw?playerName=BitmovinWeb"
	headers := map[string]string{
		"User-Agent":         model.UA_Browser,
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"Authorization":      "Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IkM4M0YzQUFGOTRCOTM0ODA2NkQwRjZDRTNEODhGQkREIiwidHlwIjoiYXQrand0In0.eyJuYmYiOjE3MTIxMjY0NTMsImV4cCI6MTcxMjE1NTI0OCwiaXNzIjoiaHR0cHM6Ly9pZC5ubHppZXQubmwiLCJhdWQiOiJhcGkiLCJjbGllbnRfaWQiOiJ0cmlwbGUtd2ViIiwic3ViIjoiMDAzMTZiNGEtMDAwMC0wMDAwLWNhZmUtZjFkZTA1ZGVlZmVlIiwiYXV0aF90aW1lIjoxNzEyMTI2NDUzLCJpZHAiOiJsb2NhbCIsImVtYWlsIjoibXVsdGkuZG5zMUBvdXRsb29rLmNvbSIsInVzZXJJZCI6IjMyMzg3MzAiLCJjdXN0b21lcklkIjoiMCIsImRldmljZUlkZW50aWZpZXIiOiJJZGVudGl6aWV0LTI0NWJiNmYzLWM2ZjktNDNjZS05ODhmLTgxNDc2OTcwM2E5OCIsImV4dGVybmFsVXNlcklkIjoiZTM1ZjdkMzktMjQ0ZC00ZTkzLWFkOTItNGFjYzVjNGY0NGNlIiwicHJvZmlsZUlkIjoiMjdDMzM3RjktOTRDRS00NjBDLTlBNjktMTlDNjlCRTYwQUIzIiwicHJvZmlsZUNvbG9yIjoiRkY0MjdDIiwicHJvZmlsZVR5cGUiOiJBZHVsdCIsIm5hbWUiOiJTdHJlYW1pbmciLCJqdGkiOiI4Q0M1QzYzNkJGRjg3MEE2REJBOERBNUMwQTk0RUZDRiIsImlhdCI6MTcxMjEyNjQ1Mywic2NvcGUiOlsiYXBpIiwib3BlbmlkIl0sImFtciI6WyJwcm9maWxlIiwicHdkIl19.bk-ziFPJM00bpE7TcgPmIYFFx-2Q5N3BkUzEvQ_dDMK9O1F9f7DEe-Qzmnb5ym7ChlnXwrCV3QyOOA24hu_gCrlNlD7-vI3XGZR-54zFD-F7cRDOoL-1-iO_10tmgwb5Io-svY0bn0EDYKeRxYYBi0w_3bFVFDM2CxxA6tWeBYIfN5rCSzBHd3RPPjYtqX-sogyh_5W_7KJ83GK5kpsywT3mz8q7Cs1mtKs9QA1-o01N0RvTxZAcfzsHg3-qGgLnvaAuZ_XqRK9kLWqJWeJTWKWtUI6OlPex22sY3keKFpfZnUtFv-BvkCM6tvbIlMZAClk3lhI8rMFAWDpUcbcS3w",
		"nlziet-appname":     "WebApp",
		"nlziet-appversion":  "5.43.24",
		"Origin":             "https://app.nlziet.nl",
		"Referer":            "https//app.nlziet.nl/",
		"Sec-Ch-UA":          model.UA_SecCHUA,
		"Sec-Ch-UA-Mobile":   "?0",
		"Sec-Ch-UA-Platform": "\"Windows\"",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().Get(url)
	if err != nil {
		return ZIETCDN(c)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return ZIETCDN(c)
	}
	body := string(b)
	if resp.StatusCode == 200 {
		if strings.Contains(body, "CountryNotAllowed") {
			return model.Result{Name: name, Status: model.StatusNo}
		} else if strings.Contains(body, "streamSessionId") {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		} else {
			return ZIETCDN(c)
		}
	} else {
		return ZIETCDN(c)
	}
}
