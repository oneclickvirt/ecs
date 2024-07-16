package security

import (
	"fmt"
	"strings"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// IpdataCo 获取 ipdata.co 的信息
func IpdataCo(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityScore := &model.SecurityScore{
		VpnScore:    new(int),
		ProxyScore:  new(int),
		ThreatScore: new(int),
		TrustScore:  new(int),
	}
	securityInfo := &model.SecurityInfo{}
	// 定义要使用的所有密钥
	keys := []string{
		// 优先使用首页script的key
		"eca677b284b3bac29eb72f5e496aa9047f26543605efe99ff2ce35c9",
		// 自定义密钥，请求失败时备用
		"47c090ef820c47af56b382bb08ba863dbd84a0b10b80acd0dd8deb48",
		"c6d4d04d5f11f2cd0839ee03c47c58621d74e361c945b5c1b4f668f3",
	}
	var (
		data, asnMap, threatMap, scores map[string]interface{}
		err                             error
		ok, value                       bool
		usageType                       string
	)
	// 尝试使用所有密钥直到请求成功或没有备用密钥可用
	for _, key := range keys {
		url := fmt.Sprintf("https://api.ipdata.co/%s?api-key=%s", ip, key)
		data, err = utils.FetchJsonFromURL(url, "tcp4", true, "Referer:https://ipdata.co/")
		if err == nil {
			if msg, ok := data["message"]; ok {
				if msg == "IP or domain not in whitelist." {
					continue
				}
			}
			// 如果请求成功，且回传结果不含非白名单的文本
			break
		}
	}
	if err != nil {
		// 所有备用密钥均未成功，返回错误
		return nil, nil, fmt.Errorf("error fetching IpdataCo info: %v", err)
	}
	if asnMap, ok = data["asn"].(map[string]interface{}); ok {
		if usageType, ok = asnMap["type"].(string); ok {
			securityInfo.UsageType = strings.ReplaceAll(usageType, " ", "")
		}
	}
	if threatMap, ok = data["threat"].(map[string]interface{}); ok {
		boolFields := []struct {
			key       string
			field     *string
			yesString string
			noString  string
		}{
			{"is_tor", &securityInfo.IsTor, "Yes", "No"},
			{"is_icloud_relay", &securityInfo.IsRelay, "Yes", "No"},
			{"is_proxy", &securityInfo.IsProxy, "Yes", "No"},
			{"is_datacenter", &securityInfo.IsDatacenter, "Yes", "No"},
			{"is_anonymous", &securityInfo.IsAnonymous, "Yes", "No"},
			{"is_known_abuser", &securityInfo.IsAbuser, "Yes", "No"},
			{"is_known_attacker", &securityInfo.IsAttacker, "Yes", "No"},
			{"is_threat", &securityInfo.IsThreat, "Yes", "No"},
			{"is_bogon", &securityInfo.IsBogon, "Yes", "No"},
		}
		for _, field := range boolFields {
			if value, ok = threatMap[field.key].(bool); ok {
				if value {
					*(field.field) = field.yesString
				} else {
					*(field.field) = field.noString
				}
			}
		}
		if scores, ok = threatMap["scores"].(map[string]interface{}); ok {
			if vpnScore, ok := scores["vpn_score"].(float64); ok {
				*securityScore.VpnScore = int(vpnScore)
			}
			if proxyScore, ok := scores["proxy_score"].(float64); ok {
				*securityScore.ProxyScore = int(proxyScore)
			}
			if threatScore, ok := scores["threat_score"].(float64); ok {
				*securityScore.ThreatScore = int(threatScore)
			}
			if trustScore, ok := scores["trust_score"].(float64); ok {
				*securityScore.TrustScore = int(trustScore)
			}
		}
	}
	securityScore.Tag = "8"
	securityInfo.Tag = "8"
	return securityScore, securityInfo, nil
}
