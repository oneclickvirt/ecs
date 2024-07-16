package security

import (
	"fmt"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// IpapiCom 获取 ipapi.com 的信息
func IpapiCom(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityScore := &model.SecurityScore{}
	securityInfo := &model.SecurityInfo{}
	url := fmt.Sprintf("https://ipapi.com/ip_api.php?ip=%s", ip)
	data, err := utils.FetchJsonFromURL(url, "tcp4", true, "")
	if err != nil {
		return nil, nil, fmt.Errorf("Error fetching Virustotal info: %v", err)
	}
	if securitys, ok := data["security"].(map[string]interface{}); ok {
		boolFields := []struct {
			key       string
			field     *string
			yesString string
			noString  string
		}{
			{"is_proxy", &securityInfo.IsProxy, "Yes", "No"},
			{"is_crawler", &securityInfo.IsCrawler, "Yes", "No"},
			{"is_tor", &securityInfo.IsTor, "Yes", "No"},
		}
		for _, field := range boolFields {
			if value, ok := securitys[field.key].(bool); ok {
				if value {
					*(field.field) = field.yesString
				} else {
					*(field.field) = field.noString
				}
			}
		}
		if threatLevel, ok := securitys["threat_level"].(string); ok {
			securityInfo.ThreatLevel = threatLevel
		}
	}
	securityScore.Tag = "B"
	securityInfo.Tag = "B"
	return securityScore, securityInfo, nil
}

// IpapiComIpv6 获取 ipapi.com 的IPV6信息
func IpapiComIpv6(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityScore := &model.SecurityScore{}
	securityInfo := &model.SecurityInfo{}
	url := fmt.Sprintf("https://ipapi.com/ip_api.php?ip=%s", ip)
	data, err := utils.FetchJsonFromURL(url, "tcp4", true, "")
	if err != nil {
		return nil, nil, fmt.Errorf("Error fetching Virustotal info: %v", err)
	}
	if securitys, ok := data["security"].(map[string]interface{}); ok {
		boolFields := []struct {
			key       string
			field     *string
			yesString string
			noString  string
		}{
			{"is_proxy", &securityInfo.IsProxy, "Yes", "No"},
			{"is_crawler", &securityInfo.IsCrawler, "Yes", "No"},
			{"is_tor", &securityInfo.IsTor, "Yes", "No"},
		}
		for _, field := range boolFields {
			if value, ok := securitys[field.key].(bool); ok {
				if value {
					*(field.field) = field.yesString
				} else {
					*(field.field) = field.noString
				}
			}
		}
		if threatLevel, ok := securitys["threat_level"].(string); ok {
			securityInfo.ThreatLevel = threatLevel
		}
	}
	securityScore.Tag = "B"
	securityInfo.Tag = "B"
	return securityScore, securityInfo, nil
}
