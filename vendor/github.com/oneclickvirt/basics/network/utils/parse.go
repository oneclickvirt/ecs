package utils

import (
	"strconv"
	"strings"

	"github.com/oneclickvirt/basics/model"
)

func ParseIpInfo(data map[string]interface{}) *model.IpInfo {
	ipInfo := &model.IpInfo{}
	if ip, ok := data["ip"].(string); ok {
		ipInfo.Ip = ip
	}
	if location, ok := data["location"].(map[string]interface{}); ok {
		if city, ok := location["city"].(string); ok {
			ipInfo.City = city
		}
		if region, ok := location["region"].(map[string]interface{}); ok {
			if name, ok := region["name"].(string); ok {
				ipInfo.Region = name
			}
		}
		if country, ok := location["country"].(map[string]interface{}); ok {
			if name, ok := country["name"].(string); ok {
				ipInfo.Country = name
			}
		}
	}
	if connection, ok := data["connection"].(map[string]interface{}); ok {
		if asn, ok := connection["asn"].(float64); ok {
			ipInfo.ASN = strconv.Itoa(int(asn))
		}
		if org, ok := connection["organization"].(string); ok {
			ipInfo.Org = org
		}
	}
	return ipInfo
}

func ParseSecurityInfo(data map[string]interface{}) *model.SecurityInfo {
	securityInfo := &model.SecurityInfo{}
	if security, ok := data["security"].(map[string]interface{}); ok {
		if isAbuser, ok := security["is_abuser"].(bool); ok {
			securityInfo.IsAbuser = BoolToString(isAbuser)
		}
		if isAttacker, ok := security["is_attacker"].(bool); ok {
			securityInfo.IsAttacker = BoolToString(isAttacker)
		}
		if isBogon, ok := security["is_bogon"].(bool); ok {
			securityInfo.IsBogon = BoolToString(isBogon)
		}
		if isCloudProvider, ok := security["is_cloud_provider"].(bool); ok {
			securityInfo.IsCloudProvider = BoolToString(isCloudProvider)
		}
		if isProxy, ok := security["is_proxy"].(bool); ok {
			securityInfo.IsProxy = BoolToString(isProxy)
		}
		if isRelay, ok := security["is_relay"].(bool); ok {
			securityInfo.IsRelay = BoolToString(isRelay)
		}
		if isTor, ok := security["is_tor"].(bool); ok {
			securityInfo.IsTor = BoolToString(isTor)
		}
		if isTorExit, ok := security["is_tor_exit"].(bool); ok {
			securityInfo.IsTorExit = BoolToString(isTorExit)
		}
		if isVpn, ok := security["is_vpn"].(bool); ok {
			securityInfo.IsVpn = BoolToString(isVpn)
		}
		if isAnonymous, ok := security["is_anonymous"].(bool); ok {
			securityInfo.IsAnonymous = BoolToString(isAnonymous)
		}
		if isThreat, ok := security["is_threat"].(bool); ok {
			securityInfo.IsThreat = BoolToString(isThreat)
		}
	}
	return securityInfo
}

// ParseYesNo 检测文本内容含No则返回No，否则返回Yes
func ParseYesNo(text string) string {
	if strings.Contains(strings.ToLower(text), "no") {
		return "No"
	}
	return "Yes"
}
