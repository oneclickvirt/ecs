package de

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// SkyDe
// edge.api.brightcove.com 仅 ipv4 且 get 请求
func SkyDe(c *http.Client) model.Result {
	name := "SKY DE"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://edge.api.brightcove.com/playback/v1/accounts/1050888051001/videos/6247131490001"
	client := utils.Req(c)
	headers := map[string]string{
		"User-Agent": model.UA_Browser,
		"sec-ch-ua":  model.UA_SecCHUA,
		"origin":     "https://www.sky.de",
		"accept":     "application/json;pk=BCpkADawqM0OXCLe4eIkpyuir8Ssf3kIQAM62a1KMa4-1_vTOWQIxoHHD4-oL-dPmlp-rLoS-WIAcaAMKuZVMR57QY4uLAmP4Ov3V416hHbqr0GNNtzVXamJ6d4-rA3Xi98W-8wtypdEyjGEZNepUCt3D7UdMthbsG-Ean3V4cafT4nZX03st5HlyK1chp51SfA-vKcAOhHZ4_Oa9TTN61tEH6YqML9PWGyKrbuN5myICcGsFzP3R2aOF8c5rPCHT2ZAiG7MoavHx8WMjhfB0QdBr2fphX24CSpUKlcjEnQJnBiA1AdLg9iyReWrAdQylX4Eyhw5OwKiCGJznfgY6BDtbUmeq1I9r9RfmhP5bfxVGjILSEFZgXbMqGOvYdrdare0aW2fTCxeHdHt0vyKOWTC6CS1lrGJF2sFPKn1T1csjVR8s4MODqCBY1PTbHY4A9aZ-2MDJUVJDkOK52hGej6aXE5b9N9_xOT2B9wbXL1B1ZB4JLjeAdBuVtaUOJ44N0aCd8Ns0o02E1APxucQqrjnEociLFNB0Bobe1nkGt3PS74IQcs-eBvWYSpolldMH6TKLu8JqgdnM4WIp3FZtTWJRADgAmvF9tVDUG9pcJoRx_CZ4im-rn-AzN3FeOQrM4rTlU3Q8YhSmyEIoxYYqsFDwbFlhsAcvqQkgaElYtuciCL5i3U8N4W9rIhPhQJzsPafmLdWxBP_FXicyek25GHFdQzCiT8nf1o860Jv2cHQ4xUNcnP-9blIkLy9JmuB2RgUXOHzWsrLGGW6hq9wLUtqwEoxcEAAcNJgmoC0k8HE-Ga-NHXng6EFWnqiOg_mZ_MDd7gmHrrKLkQV",
	}
	client = utils.SetReqHeaders(client, headers)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	body := string(b)
	//fmt.Println(body)
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
		return model.Result{Name: name, Status: model.StatusYes, Region: "de"}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get edge.api.brightcove.com with code: %d", resp.StatusCode)}
}
