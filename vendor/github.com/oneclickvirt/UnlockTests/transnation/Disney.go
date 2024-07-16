package transnation

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DisneyPlus
// www.disneyplus.com 双栈 且 post 请求
func DisneyPlus(c *http.Client) model.Result {
	name := "Disney+"
	hostname := "disneyplus.com"
	if c == nil {
		return model.Result{Name: name}
	}
	// 首次请求，获取assertion
	url1 := "https://disney.api.edge.bamgrid.com/devices"
	playload := `{"deviceFamily":"browser","applicationRuntime":"chrome","deviceProfile":"windows","attributes":{}}`
	headers := map[string]string{
		"authorization": "Bearer ZGlzbmV5JmJyb3dzZXImMS4wLjA.Cu56AgSfBTDag5NiRA81oLHkDZfu5L3CKadnefEAY84",
		"Content-Type":  "application/json",
	}
	client := utils.Req(c)
	client = utils.SetReqHeaders(client, headers)
	resp1, err := client.R().SetBodyString(playload).Post(url1)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp1.Body.Close()
	body1, err := io.ReadAll(resp1.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	if strings.Contains(string(body1), "403 ERROR") {
		return model.Result{Name: name, Status: model.StatusNo, Info: "Can not get assertion"}
	}
	// fmt.Println(string(body1))
	var res1 struct {
		Assertion string `json:"assertion"`
	}
	if err := json.Unmarshal(body1, &res1); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}

	// 二次请求修改subject_token
	data := url.Values{
		"grant_type":         {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"latitude":           {"0"},
		"longitude":          {"0"},
		"platform":           {"browser"},
		"subject_token":      {res1.Assertion},
		"subject_token_type": {"urn:bamtech:params:oauth:token-type:device"},
	}
	url2 := "https://disney.api.edge.bamgrid.com/token"
	headers2 := map[string]string{
		"authorization": "ZGlzbmV5JmJyb3dzZXImMS4wLjA.Cu56AgSfBTDag5NiRA81oLHkDZfu5L3CKadnefEAY84",
	}
	client = utils.SetReqHeaders(client, headers2)
	resp2, err := client.R().SetFormDataFromValues(data).Post(url2)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp2.Body.Close()
	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	// fmt.Println(string(body2))
	if strings.Contains(string(body2), "forbidden-location") || resp2.StatusCode == 403 {
		return model.Result{Name: name, Status: model.StatusNo, Info: "forbidden-location"}
	}
	var res2 struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body2, &res2); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}

	// 三次请求获取地址
	resp3, err := client.R().Get("https://disneyplus.com")
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp3.Body.Close()
	if strings.Contains(resp3.Request.URL.String(), "preview") || strings.Contains(resp3.Request.URL.String(), "unavailable") {
		return model.Result{Name: name, Status: model.StatusNo, Info: "Can not visit page"}
	}

	// 四次请求刷新token并获取支持
	url4 := "https://disney.api.edge.bamgrid.com/graph/v1/device/graphql"
	playload4 := fmt.Sprintf(`{"query":"mutation refreshToken($input: RefreshTokenInput!) {\n refreshToken(refreshToken: $input) {\n activeSession {\n sessionId\n }\n }\n}","variables":{"input":{"refreshToken":"%s"}}}`, res2.RefreshToken)
	headers4 := map[string]string{
		"authorization": "ZGlzbmV5JmJyb3dzZXImMS4wLjA.Cu56AgSfBTDag5NiRA81oLHkDZfu5L3CKadnefEAY84",
	}
	resp4, body4, err := utils.PostJson(c, url4, playload4, headers4)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp4.Body.Close()
	//fmt.Println(body4)
	if utils.ReParse(body4, `"inSupportedLocation"\s*:\s*(false|true)`) != "true" {
		return model.Result{Name: name, Status: model.StatusNo, Info: "UnSupported"}
	}
	region := utils.ReParse(body4, `"countryCode"\s*:\s*"([^"]+)"`)
	if region == "" {
		return model.Result{Name: name, Status: model.StatusUnexpected}
	}
	result1, result2, result3 := utils.CheckDNS(hostname)
	unlockType := utils.GetUnlockType(result1, result2, result3)
	return model.Result{Name: name, Status: model.StatusYes, Region: strings.ToLower(region), UnlockType: unlockType}
}
