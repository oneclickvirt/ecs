package uk

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// DiscoveryPlus
// disco-api.discoveryplus.co.uk 仅 ipv4 且 get 请求
func DiscoveryPlus(c *http.Client) model.Result {
	name := "Discovery+ UK"
	hostname := "discoveryplus.co.uk"
	if c == nil {
		return model.Result{Name: name}
	}
	url1 := "https://disco-api.discoveryplus.co.uk/token?realm=questuk&deviceId=61ee588b07c4df08c02861ecc1366a592c4ad02d08e8228ecfee67501d98bf47&shortlived=true"
	client1 := utils.Req(c)
	resp1, err1 := client1.R().Get(url1)
	if err1 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err1}
	}
	defer resp1.Body.Close()
	b1, err1 := io.ReadAll(resp1.Body)
	if err1 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	//body1 := string(b1)
	var res struct {
		Data struct {
			Attributes struct {
				Token string `json:"token"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b1, &res); err != nil {
		return model.Result{Name: name, Status: model.StatusNo, Err: err}
	}
	cookies := "_gcl_au=1.1.858579665.1632206782; _rdt_uuid=1632206782474.6a9ad4f2-8ef7-4a49-9d60-e071bce45e88; " +
		"_scid=d154b864-8b7e-4f46-90e0-8b56cff67d05; " +
		"_pin_unauth=dWlkPU1qWTRNR1ZoTlRBdE1tSXdNaTAwTW1Nd0xUbGxORFV0WWpZMU0yVXdPV1l6WldFeQ; " +
		"_sctr=1|1632153600000; aam_fw=aam%3D9354365%3Baam%3D9040990; " +
		fmt.Sprintf("aam_uuid=24382050115125439381416006538140778858; st=%s; ", res.Data.Attributes.Token) +
		"gi_ls=0; _uetvid=a25161a01aa711ec92d47775379d5e4d; " +
		"AMCV_BC501253513148ED0A490D45%40AdobeOrg=-1124106680%7CMCIDTS%7C18894%7CMCMID%7C24223296309793" +
		"747161435877577673078228%7CMCAAMLH-1633011393%7C9%7CMCAAMB-1633011393%7CRKhpRz8krg2tLO6pguXWp5o" +
		"lkAcUniQYPHaMWWgdJ3xzPWQmdj0y%7CMCOPTOUT-1632413793s%7CNONE%7CvVersion%7C5.2.0; " +
		"ass=19ef15da-95d6-4b1d-8fa2-e9e099c9cc38.1632408400.1632406594"
	headers := map[string]string{
		"Cookie": cookies,
	}
	url2 := "https://disco-api.discoveryplus.co.uk/users/me"
	client2 := utils.Req(c)
	client2 = utils.SetReqHeaders(client2, headers)
	resp2, err2 := client2.R().Get(url2)
	if err2 != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err2}
	}
	defer resp2.Body.Close()
	b2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	var res2 struct {
		Data struct {
			Attributes struct {
				CurrentLocationTerritory string `json:"currentLocationTerritory"`
			} `json:"attributes"`
		} `json:"data"`
	}
	//fmt.Println(string(b2))
	if err = json.Unmarshal(b2, &res2); err != nil {
		return model.Result{Name: name, Status: model.StatusUnexpected, Err: err}
	}
	if res2.Data.Attributes.CurrentLocationTerritory != "" {
		loc := strings.ToLower(res2.Data.Attributes.CurrentLocationTerritory)
		exit := utils.GetRegion(loc, model.DiscoveryPlusSupportCountry)
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		if exit {
			if loc == "gb" {
				return model.Result{Name: name, Status: model.StatusYes, Region: loc, Info: "origin", UnlockType: unlockType}
			} else {
				return model.Result{Name: name, Status: model.StatusYes, Region: loc, Info: "global", UnlockType: unlockType}
			}
		}
		return model.Result{Name: name, Status: model.StatusNo, Region: loc}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get disco-api.discoveryplus.co.uk failed with code: %d", resp2.StatusCode)}
}
