package eu

import (
	"crypto/rand"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// MathsSpot
func MathsSpot(c *http.Client) model.Result {
	name := "Maths Spot"
	hostname := "mathsspot.com"
	if c == nil {
		return model.Result{Name: name}
	}
	headers := map[string]string{
		"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language": "en-US,en;q=0.9",
		"User-Agent":      model.UA_Browser,
	}
	url := "https://mathsspot.com/"
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
	//fmt.Println(body)
	if len(body) > 0 && strings.Contains(body, "FailureServiceNotInRegion") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	region := utils.ReParse(body, `"countryCode"\s{0,}:\s{0,}"([^"]+)"`)
	nggFeVersion := utils.ReParse(body, `"NEXT_PUBLIC_FE_VERSION"\s{0,}:\s{0,}"([^"]+)"`)
	if region == "" || nggFeVersion == "" {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	headers2 := map[string]string{
		"accept":                "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language":       "en-US,en;q=0.9",
		"User-Agent":            model.UA_Browser,
		"referer":               "https://mathsspot.com/",
		"sec-ch-ua":             model.UA_SecCHUA,
		"sec-fetch-dest":        "empty",
		"sec-fetch-mode":        "cors",
		"sec-fetch-site":        "same-origin",
		"x-ngg-skip-evar-check": "true",
		"x-ngg-fe-version":      nggFeVersion,
	}
	client2 := utils.Req(c)
	client2 = utils.SetReqHeaders(client2, headers2)
	url2 := fmt.Sprintf("https://mathsspot.com/3/api/play/v1/startSession?appId=5349&uaId=ua-%s&uaSessionId=uasess-%s&feSessionId=fesess-%s&visitId=visitid-%s&initialOrientation=landscape&utmSource=NA&utmMedium=NA&utmCampaign=NA&deepLinkUrl=&accessCode=&ngReferrer=NA&pageReferrer=NA&ngEntryPoint=https%%3A%%2F%%2Fmathsspot.com%%2F&ntmSource=NA&customData=&appLaunchExtraData=&feSessionTags=nowgg&sdpType=&eVar=&isIframe=false&feDeviceType=desktop&feOsName=window&userSource=direct&visitSource=direct&userCampaign=NA&visitCampaign=NA", generateRandomString(21), generateRandomString(21), generateRandomString(21), generateRandomString(21))
	resp2, err2 := client.R().Get(url2)
	if err2 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
	}
	defer resp2.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body2 := string(b)
	//fmt.Println(body2)
	status := utils.ReParse(body2, `"status"\s{0,}:\s{0,}"([^"]+)"`)
	switch status {
	case "FailureServiceNotInRegion":
		return model.Result{Name: name, Status: model.StatusNo}
	case "FailureProxyUserLimitExceeded":
		return model.Result{Name: name, Status: model.StatusNo, Info: "Proxy/VPN Detected"}
	case "Success":
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, Region: region, UnlockType: unlockType}
	default:
		return model.Result{Name: name, Status: model.StatusUnexpected, Info: status}
	}
}
