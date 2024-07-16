package us

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

// ESPNPlus
// espn.api.edge.bamgrid.com 双栈 且 post 请求 可能 有 cloudflare 的5秒盾
func ESPNPlus(c *http.Client) model.Result {
	name := "ESPN+"
	if c == nil {
		return model.Result{Name: name}
	}
	url1 := "https://espn.api.edge.bamgrid.com/token"
	data1 := url.Values{
		"grant_type":         {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"latitude":           {"0"},
		"longitude":          {"0"},
		"platform":           {"browser"},
		"subject_token":      {"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJzdWIiOiJjYWJmMDNkMi0xMmEyLTQ0YjYtODJjOS1lOWJkZGNhMzYwNjkiLCJhdWQiOiJ1cm46YmFtdGVjaDpzZXJ2aWNlOnRva2VuIiwibmJmIjoxNjMyMjMwMTY4LCJpc3MiOiJ1cm46YmFtdGVjaDpzZXJ2aWNlOmRldmljZSIsImV4cCI6MjQ5NjIzMDE2OCwiaWF0IjoxNjMyMjMwMTY4LCJqdGkiOiJhYTI0ZWI5Yi1kNWM4LTQ5ODctYWI4ZS1jMDdhMWVhMDgxNzAifQ.8RQ-44KqmctKgdXdQ7E1DmmWYq0gIZsQw3vRL8RvCtrM_hSEHa-CkTGIFpSLpJw8sMlmTUp5ZGwvhghX-4HXfg"},
		"subject_token_type": {"urn:bamtech:params:oauth:token-type:device"},
	}
	headers1 := map[string]string{
		"authorization": "Bearer ZXNwbiZicm93c2VyJjEuMC4w.ptUt7QxsteaRruuPmGZFaJByOoqKvDP2a5YkInHrc7c",
	}
	client1 := utils.Req(c)
	client1 = utils.SetReqHeaders(client1, headers1)
	resp, err := client1.R().SetFormDataFromValues(data1).Post(url1)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body := string(b)
	// fmt.Println(body)
	if strings.Contains(body, "forbidden-location") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	var res struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}

	url2 := "https://espn.api.edge.bamgrid.com/graph/v1/device/graphql"
	data2 := `{"query":"mutation registerDevice($input: RegisterDeviceInput!) {\n            registerDevice(registerDevice: $input) {\n                grant {\n                    grantType\n                    assertion\n                }\n            }\n        }","variables":{"input":{"deviceFamily":"browser","applicationRuntime":"chrome","deviceProfile":"windows","deviceLanguage":"zh-CN","attributes":{"osDeviceIds":[],"manufacturer":"microsoft","model":null,"operatingSystem":"windows","operatingSystemVersion":"10.0","browserName":"chrome","browserVersion":"96.0.4664"}}}}`
	headers2 := map[string]string{
		"authorization": "ZXNwbiZicm93c2VyJjEuMC4w.ptUt7QxsteaRruuPmGZFaJByOoqKvDP2a5YkInHrc7c",
	}
	resp2, body2, err2 := utils.PostJson(c, url2, data2, headers2)
	if err2 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
	}
	defer resp2.Body.Close()
	var res2 struct {
		Extensions struct {
			Sdk struct {
				Session struct {
					Location struct {
						CountryCode string `json:"countryCode"`
					}
					InSupportedLocation bool `json:"inSupportedLocation"`
				}
			}
		}
	}
	if err := json.Unmarshal([]byte(body2), &res2); err != nil {
		return model.Result{Name: name, Status: model.StatusUnexpected, Err: err}
	}
	if res2.Extensions.Sdk.Session.Location.CountryCode == "US" && res2.Extensions.Sdk.Session.InSupportedLocation {
		return model.Result{Name: name, Status: model.StatusYes}
	}
	return model.Result{Name: name, Status: model.StatusNo}
	// return model.Result{Name: name, Status: model.StatusUnexpected,
	// 	Err: fmt.Errorf("get espn.api.edge.bamgrid.com failed with code: %d", resp.StatusCode)}
}
