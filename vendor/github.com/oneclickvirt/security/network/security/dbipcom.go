package security

import (
	"fmt"
	"strings"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// DbIpCom 获取 https://db-ip.com/ 的信息
func DbIpCom(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityScore := &model.SecurityScore{}
	securityInfo := &model.SecurityInfo{}
	url := fmt.Sprintf("https://db-ip.com/demo/home.php?s=%s", ip)
	data, err := utils.FetchJsonFromURL(url, "tcp4", true, "")
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching DbIpCom info: %v", err)
	}
	if demoInfo, ok := data["demoInfo"].(map[string]interface{}); ok {
		if e, ok := demoInfo["error"].(string); ok {
			if strings.Contains(e, "over query limit") {
				return nil, nil, fmt.Errorf("over query limit DbIpCom: %v", err)
			}
		}
	}
	if demoInfo, ok := data["demoInfo"].(map[string]interface{}); ok {
		if usageType, ok := demoInfo["usageType"].(string); ok {
			securityInfo.UsageType = usageType
		}
		boolFields := []struct {
			key       string
			field     *string
			yesString string
			noString  string
		}{
			{"isCrawler", &securityInfo.IsCrawler, "Yes", "No"},
			{"isProxy", &securityInfo.IsProxy, "Yes", "No"},
		}
		for _, field := range boolFields {
			if value, ok := demoInfo[field.key].(bool); ok {
				if value {
					*(field.field) = field.yesString
				} else {
					*(field.field) = field.noString
				}
			}
		}
		if threatLevel, ok := demoInfo["threatLevel"].(string); ok {
			securityInfo.ThreatLevel = threatLevel
		}
	}
	securityScore.Tag = "9"
	securityInfo.Tag = "9"
	return securityScore, securityInfo, nil
}
