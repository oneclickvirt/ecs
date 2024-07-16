package security

import (
	"strconv"
	"strings"

	"github.com/oneclickvirt/basics/model"
)

// FormatSecurityScore 格式化给定属性的安全得分
func FormatSecurityScore(attributes []model.SecurityScore) string {
	result := ""
	valueMap := make(map[string][]string) // 用于记录每个属性值出现的标签序号
	for _, attr := range attributes {
		if attributes == nil {
			continue
		}
		// 遍历属性结构体的字段
		fields := []struct {
			name  string
			value *int
		}{
			{"Reputation", attr.Reputation},
			{"AbuseScore", attr.AbuseScore},
			{"FraudScore", attr.FraudScore},
			{"VpnScore", attr.VpnScore},
			{"ProxyScore", attr.ProxyScore},
			{"ThreatScore", attr.ThreatScore},
			{"TrustScore", attr.TrustScore},
			{"CloudFlareRisk", attr.CloudFlareRisk},
			{"CommunityVoteHarmless", attr.CommunityVoteHarmless},
			{"CommunityVoteMalicious", attr.CommunityVoteMalicious},
			{"HarmlessnessRecords", attr.HarmlessnessRecords},
			{"MaliciousRecords", attr.MaliciousRecords},
			{"SuspiciousRecords", attr.SuspiciousRecords},
			{"NoRecords", attr.NoRecords},
		}
		for _, field := range fields {
			if field.value != nil {
				key := field.name + ": " + strconv.Itoa(*field.value)
				if tags, ok := valueMap[key]; ok {
					valueMap[key] = append(tags, attr.Tag)
				} else {
					valueMap[key] = []string{attr.Tag}
				}
			}
		}
	}
	// 构建结果字符串
	for key, tags := range valueMap {
		result += key
		for i, tag := range tags {
			// 追加标签序号
			if tag == tags[0] && len(tags) != 1 {
				result += " [" + tag
			} else if tag == tags[0] && len(tags) == 1 {
				result += " [" + tag + "] "
			} else if tag == tags[len(tags)-1] {
				result += tag + "] "
			} else {
				result += tag
			}
			if i < len(tags)-1 {
				result += " "
			}
		}
		result += "\n"
	}
	return strings.TrimSpace(result)
}

// FormatSecurityInfo 格式化给定属性的安全信息
func FormatSecurityInfo(attributes []model.SecurityInfo) string {
	result := ""
	valueMap := make(map[string][]string) // 用于记录每个属性值出现的标签序号
	for _, attr := range attributes {
		if attributes == nil {
			continue
		}
		// 遍历属性结构体的字段
		fields := []struct {
			name  string
			value string
		}{
			{"ASNAbuseScore", attr.ASNAbuseScore},
			{"CompannyAbuseScore", attr.CompannyAbuseScore},
			{"ThreatLevel", attr.ThreatLevel},
			{"UsageType", attr.UsageType},
			{"CompanyType", attr.CompanyType},
			{"IsCloudProvider", attr.IsCloudProvider},
			{"IsDatacenter", attr.IsDatacenter},
			{"IsMobile", attr.IsMobile},
			{"IsProxy", attr.IsProxy},
			{"IsVpn", attr.IsVpn},
			{"IsTor", attr.IsTor},
			{"IsTorExit", attr.IsTorExit},
			{"IsCrawler", attr.IsCrawler},
			{"IsAnonymous", attr.IsAnonymous},
			{"IsAttacker", attr.IsAttacker},
			{"IsAbuser", attr.IsAbuser},
			{"IsThreat", attr.IsThreat},
			{"IsRelay", attr.IsRelay},
			{"IsBogon", attr.IsBogon},
			{"IsBot", attr.IsBot},
		}
		for _, field := range fields {
			if field.value != "" {
				key := field.name + ": " + field.value
				if tags, ok := valueMap[key]; ok {
					valueMap[key] = append(tags, attr.Tag)
				} else {
					valueMap[key] = []string{attr.Tag}
				}
			}
		}
	}
	// 构建结果字符串
	for key, tags := range valueMap {
		result += key
		for i, tag := range tags {
			// 追加标签序号
			if tag == tags[0] && len(tags) != 1 {
				result += " [" + tag
			} else if tag == tags[0] && len(tags) == 1 {
				result += " [" + tag + "] "
			} else if tag == tags[len(tags)-1] {
				result += tag + "] "
			} else {
				result += tag
			}
			if i < len(tags)-1 {
				result += " "
			}
		}
		result += "\n"
	}
	return strings.TrimSpace(result)
}

// TODO
// ip2location.io 仅使用首页token请求
// https://ping0.cc/ js动态加载
