package us

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// AETV
// ccpa-service.sp-prod.net 仅 ipv4 且 post 请求
func AETV(c *http.Client) model.Result {
	name := "A&E TV"
	hostname := "aetv.com"
	if c == nil {
		return model.Result{Name: name}
	}

	url1 := "https://link.theplatform.com/s/xc6n8B/UR27JDU0bu2s/"
	client1 := utils.Req(c)
	resp1, err1 := client1.R().Post(url1)
	if err1 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err1}
	}
	defer resp1.Body.Close()
	b1, err1 := io.ReadAll(resp1.Body)
	if err1 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body1 := string(b1)
	if strings.Contains(body1, "GeoLocationBlocked") {
		return model.Result{Name: name, Status: model.StatusNo}
	}

	url2 := "https://play.aetv.com/"
	client2 := utils.Req(c)
	resp2, err2 := client2.R().Post(url2)
	if err2 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
	}
	defer resp2.Body.Close()
	b2, err2 := io.ReadAll(resp2.Body)
	if err2 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body2 := string(b2)
	if body2 != "" {
		tp := utils.ReParse(body2, `AETN-Country-Code=([A-Z]+)`)
		if tp != "" {
			region := strings.ToLower(tp)
			if region == "ca" || region == "us" {
				return model.Result{Name: name, Status: model.StatusYes, Region: region}
			} else {
				return model.Result{Name: name, Status: model.StatusNo}
			}
		}
	}

	url3 := "https://ccpa-service.sp-prod.net/ccpa/consent/10265/display-dns"
	client3 := utils.Req(c)
	resp3, err3 := client3.R().Post(url3)
	if err3 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err3}
	}
	defer resp3.Body.Close()
	b3, err3 := io.ReadAll(resp3.Body)
	if err3 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body3 := string(b3)
	//fmt.Println(body)
	var res struct {
		CcpaApplies bool `json:"ccpaApplies"`
	}
	if err := json.Unmarshal([]byte(body3), &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.CcpaApplies == true {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	} else if res.CcpaApplies == false {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get ccpa-service.sp-prod.net failed with code: %d", resp3.StatusCode)}
}
