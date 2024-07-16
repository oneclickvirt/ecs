package security

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// Abuseipdb 获取 abuseipdb.com 的信息
func Abuseipdb(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityInfo := &model.SecurityInfo{}
	securityScore := &model.SecurityScore{
		AbuseScore: new(int),
	}
	url := fmt.Sprintf("https://api.abuseipdb.com/api/v2/check?ipAddress=%s", ip)
	additionalHeaders := []string{
		"key: e88362808d1219e27a786a465a1f57ec3417b0bdeab46ad670432b7ce1a7fdec0d67b05c3463dd3c",
		"key: a240c11ca3d2f3d58486fa86f1744a143448d3a6fcb2fc1f8880bafd58c3567a0adddcfd7a722364",
	}
	// 生成随机索引
	randomIndex := rand.Intn(len(additionalHeaders))
	// 获取随机元素
	additionalHeader := additionalHeaders[randomIndex]
	data, err := utils.FetchJsonFromURL(url, "tcp4", true, additionalHeader)
	if err != nil {
		return nil, nil, fmt.Errorf("Error fetching Abuseipdb info: %v", err)
	}
	if dataMap, ok := data["data"].(map[string]interface{}); ok {
		if usageType, ok := dataMap["usageType"].(string); ok {
			securityInfo.UsageType = strings.ReplaceAll(usageType, " ", "")
		}
		if abuseConfidenceScore, ok := dataMap["abuseConfidenceScore"].(float64); ok {
			*securityScore.AbuseScore = int(abuseConfidenceScore)
		}
		if isTor, ok := dataMap["isTor"].(bool); ok {
			securityInfo.IsTor = utils.BoolToString(isTor)
		}
	}
	securityScore.Tag = "3"
	securityInfo.Tag = "3"
	return securityScore, securityInfo, nil
}
