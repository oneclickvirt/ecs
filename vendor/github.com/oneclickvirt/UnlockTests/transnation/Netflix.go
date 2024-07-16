package transnation

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// NetflixCDN
// api.fast.com 双栈 get 请求
func NetflixCDN(c *http.Client) model.Result {
	name := "Netflix CDN"
	hostname := "fast.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://api.fast.com/netflix/speedtest/v2?https=true&token=YXNkZmFzZGxmbnNkYWZoYXNkZmhrYWxm&urlCount=5"
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
	//body := string(b)
	//fmt.Println(body)
	if resp.StatusCode == 403 || resp.StatusCode == 451 {
		return model.Result{Name: name, Status: model.StatusNo, Info: "IP Banned By Netflix"}
	}
	type netflixCdnTarget struct {
		Name     string `json:"name"`
		Url      string `json:"url"`
		Location struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"location"`
	}
	var res struct {
		Targets []netflixCdnTarget `json:"targets"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.Targets[0].Location.Country != "" {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{
			Name: name, Status: model.StatusYes,
			Region:     res.Targets[0].Location.Country,
			UnlockType: unlockType,
		}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get api.fast.com failed with code: %d", resp.StatusCode)}
}

// Netflix
// www.netflix.com 双栈 且 get 请求
func Netflix(c *http.Client) model.Result {
	name := "Netflix"
	hostname := "netflix.com"
	if c == nil {
		return model.Result{Name: name}
	}
	url1 := "https://www.netflix.com/title/81280792" // 乐高
	url2 := "https://www.netflix.com/title/70143836" // 绝命毒师
	url3 := "https://www.netflix.com/title/80018499" // Test Patterns
	client1 := utils.Req(c)
	resp1, err1 := client1.R().Get(url1)
	if err1 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err1}
	}
	defer resp1.Body.Close()
	//b1, err1 := io.ReadAll(resp1.Body)
	//if err1 != nil {
	//	return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	//}
	//body1 := string(b1)
	//if body1 == "" {
	//	return model.Result{
	//		Name: name, Status: model.StatusNo,
	//	}
	//}
	client2 := utils.Req(c)
	resp2, err2 := client2.R().Get(url2)
	if err2 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
	}
	defer resp2.Body.Close()
	//b2, err2 := io.ReadAll(resp2.Body)
	//if err2 != nil {
	//	return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	//}
	//body2 := string(b2)
	//if body2 == "" {
	//	return model.Result{
	//		Name: name, Status: model.StatusNo,
	//	}
	//}
	if resp1.StatusCode == 404 && resp2.StatusCode == 404 {
		return model.Result{Name: name, Status: model.StatusRestricted, Info: "Originals Only"}
	}
	if resp1.StatusCode == 403 && resp2.StatusCode == 403 {
		return model.Result{Name: name, Status: model.StatusBanned}
	}
	if (resp1.StatusCode == 200 || resp1.StatusCode == 301) || (resp2.StatusCode == 200 || resp2.StatusCode == 301) {
		client3 := utils.Req(c)
		resp3, err3 := client3.R().Get(url3)
		if err3 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err3}
		}
		defer resp3.Body.Close()
		//b3, err3 := io.ReadAll(resp3.Body)
		//if err3 != nil {
		//	return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
		//}
		//body3 := string(b3)
		//if body3 == "" {
		//	return model.Result{
		//		Name: name, Status: model.StatusNo,
		//	}
		//}
		u := resp3.Header.Get("location")
		if u == "" {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType, Region: "us"}
		}
		//fmt.Println("nf", u)
		t := strings.SplitN(u, "/", 5)
		if len(t) < 5 {
			return model.Result{Name: name, Status: model.StatusUnexpected, Err: fmt.Errorf("can not find region")}
		}
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType,
			Region: strings.SplitN(t[3], "-", 2)[0]}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.netflix.com failed with code: %d %d", resp1.StatusCode, resp2.StatusCode)}
}
