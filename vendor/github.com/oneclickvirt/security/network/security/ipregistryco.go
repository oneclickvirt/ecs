package security

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// IpregistryCo 获取 ipregistry.co 的信息
func IpregistryCo(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityInfo := &model.SecurityInfo{}
	data, err := fetchDataFromIpregistry(ip)
	if err != nil {
		return nil, nil, fmt.Errorf("Error fetching IpregistryCo info: %v", err)
	}
	parseCompanyInfo(data, securityInfo)
	parseConnectionInfo(data, securityInfo)
	parseSecurityInfo(data, securityInfo)
	securityInfo.Tag = "7"
	return nil, securityInfo, nil
}

// fetchDataFromIpregistry 从 ipregistry.co 获取数据
func fetchDataFromIpregistry(ip string) (map[string]interface{}, error) {
	var (
		data map[string]interface{}
		err  error
	)
	keys := []string{
		"sb69ksjcajfs4c",
		"ing7l12cxp6jaahw",
		"r208izz0q0icseks",
		"szh9vdbsf64ez2bk",
		"vum97powo0pxshko",
		"m7irmmf8ey12rx7o",
		"nd2chql8jm9f7gxa",
		"9mbbr52gsds5xtyb",
		"0xjh6xmh6j0jwsy6",
	}
	// 打乱顺序
	rand.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})
	for _, key := range keys {
		url := fmt.Sprintf("https://api.ipregistry.co/%s?key=%s", ip, key)
		data, err = utils.FetchJsonFromURL(url, "tcp4", true, "")
		if err == nil {
			if codeString, ok := data["code"].(string); ok {
				if strings.Contains(codeString, "FORBIDDEN") {
					continue
				}
			} else {
				return data, nil
			}
		}
	}
	return nil, fmt.Errorf("Error fetching data from Ipregistry")
}

// parseCompanyInfo 解析公司信息
func parseCompanyInfo(data map[string]interface{}, securityInfo *model.SecurityInfo) {
	if company, ok := data["company"].(map[string]interface{}); ok {
		if companyType, ok := company["type"].(string); ok {
			securityInfo.CompanyType = companyType
		}
	}
}

// parseConnectionInfo 解析连接信息
func parseConnectionInfo(data map[string]interface{}, securityInfo *model.SecurityInfo) {
	if connection, ok := data["connection"].(map[string]interface{}); ok {
		if connectionType, ok := connection["type"].(string); ok {
			securityInfo.UsageType = connectionType
		}
	}
}

// parseSecurityInfo 解析安全信息
func parseSecurityInfo(data map[string]interface{}, securityInfo *model.SecurityInfo) {
	if securityData, ok := data["security"].(map[string]interface{}); ok {
		boolFields := []struct {
			key       string
			field     *string
			yesString string
			noString  string
		}{
			{"is_abuser", &securityInfo.IsAbuser, "Yes", "No"},
			{"is_attacker", &securityInfo.IsAttacker, "Yes", "No"},
			{"is_bogon", &securityInfo.IsBogon, "Yes", "No"},
			{"is_cloud_provider", &securityInfo.IsCloudProvider, "Yes", "No"},
			{"is_proxy", &securityInfo.IsProxy, "Yes", "No"},
			{"is_relay", &securityInfo.IsRelay, "Yes", "No"},
			{"is_tor", &securityInfo.IsTor, "Yes", "No"},
			{"is_tor_exit", &securityInfo.IsTorExit, "Yes", "No"},
			{"is_vpn", &securityInfo.IsVpn, "Yes", "No"},
			{"is_anonymous", &securityInfo.IsAnonymous, "Yes", "No"},
			{"is_threat", &securityInfo.IsThreat, "Yes", "No"},
		}
		for _, field := range boolFields {
			if value, ok := securityData[field.key].(bool); ok {
				if value {
					*(field.field) = field.yesString
				} else {
					*(field.field) = field.noString
				}
			}
		}
	}
}
