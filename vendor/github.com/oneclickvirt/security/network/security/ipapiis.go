package security

import (
	"fmt"
	"strings"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// Ipapiis 获取 ipapi.is 的信息
func Ipapiis(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityScore := &model.SecurityScore{}
	securityInfo := &model.SecurityInfo{}
	url := fmt.Sprintf("https://api.ipapi.is/?q=%s", ip)
	data, err := utils.FetchJsonFromURL(url, "tcp4", true, "")
	if err != nil {
		return nil, nil, fmt.Errorf("Error fetching Virustotal info: %v", err)
	}
	if companyMap, ok := data["company"].(map[string]interface{}); ok {
		if companyType, ok := companyMap["type"].(string); ok {
			securityInfo.CompanyType = strings.ReplaceAll(companyType, " ", "")
		}
		if abuseScoreText, ok := companyMap["abuser_score"].(string); ok {
			securityInfo.CompannyAbuseScore = abuseScoreText
		}
	}
	if asnMap, ok := data["asn"].(map[string]interface{}); ok {
		if usageType, ok := asnMap["type"].(string); ok {
			securityInfo.UsageType = strings.ReplaceAll(usageType, " ", "")
		}
		if abuseScoreText, ok := asnMap["abuser_score"].(string); ok {
			securityInfo.ASNAbuseScore = abuseScoreText
		}
	}
	boolFields := []struct {
		key       string
		field     *string
		yesString string
		noString  string
	}{
		{"is_bogon", &securityInfo.IsBogon, "Yes", "No"},
		{"is_mobile", &securityInfo.IsMobile, "Yes", "No"},
		{"is_crawler", &securityInfo.IsCrawler, "Yes", "No"},
		{"is_datacenter", &securityInfo.IsDatacenter, "Yes", "No"},
		{"is_tor", &securityInfo.IsTor, "Yes", "No"},
		{"is_proxy", &securityInfo.IsProxy, "Yes", "No"},
		{"is_vpn", &securityInfo.IsVpn, "Yes", "No"},
		{"is_abuser", &securityInfo.IsAbuser, "Yes", "No"},
	}
	for _, field := range boolFields {
		if value, ok := data[field.key].(bool); ok {
			if value {
				*(field.field) = field.yesString
			} else {
				*(field.field) = field.noString
			}
		}
	}
	securityScore.Tag = "A"
	securityInfo.Tag = "A"
	return securityScore, securityInfo, nil
}
