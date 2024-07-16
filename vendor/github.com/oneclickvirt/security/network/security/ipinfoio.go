package security

import (
	"fmt"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// IPInfoIo 从 ipinfo.io 获取信息
func IPInfoIo(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityInfo := &model.SecurityInfo{}
	url := fmt.Sprintf("https://ipinfo.io/widget/demo/%s", ip)
	data, err := utils.FetchJsonFromURL(url, "tcp4", true, "Referer:https://ipinfo.io/")
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching info: %v", err)
	}
	if dataMap, ok := data["data"].(map[string]interface{}); ok {
		if asnMap, ok := dataMap["asn"].(map[string]interface{}); ok {
			if usageType, ok := asnMap["type"].(string); ok {
				securityInfo.UsageType = usageType
			}
		}
		if companyMap, ok := dataMap["company"].(map[string]interface{}); ok {
			if companyType, ok := companyMap["type"].(string); ok {
				securityInfo.CompanyType = companyType
			}
		}
		if privacyMap, ok := dataMap["privacy"].(map[string]interface{}); ok {
			boolFields := []struct {
				key       string
				field     *string
				yesString string
				noString  string
			}{
				{"vpn", &securityInfo.IsVpn, "Yes", "No"},
				{"tor", &securityInfo.IsTor, "Yes", "No"},
				{"proxy", &securityInfo.IsProxy, "Yes", "No"},
				{"relay", &securityInfo.IsRelay, "Yes", "No"},
				{"hosting", &securityInfo.IsDatacenter, "Yes", "No"},
			}
			for _, field := range boolFields {
				if value, ok := privacyMap[field.key].(bool); ok {
					if value {
						*(field.field) = field.yesString
					} else {
						*(field.field) = field.noString
					}
				}
			}
		}
	}
	securityInfo.Tag = "0"
	return nil, securityInfo, nil
}
