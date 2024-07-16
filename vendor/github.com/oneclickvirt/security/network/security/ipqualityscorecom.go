package security

import (
	"fmt"
	"math/rand"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// IpqualityscoreCom 获取 ipqualityscore.com 的信息 需要优化
func IpqualityscoreCom(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityScore := &model.SecurityScore{
		FraudScore: new(int),
	}
	securityInfo := &model.SecurityInfo{}
	additionalKeys := []string{
		"O3AT606i4GvBJBC5ePetx3rGXzIMxapj",
		"DJcK0cFyVvqBhR4RYqqsELIxHz0clzja",
		"VloVwex6edRJfsgR41xZO6vmsIrxbqqR",
		"yx3K0H8IoNTqtGO7oGfNr4ntbp53fnkP",
		"U2g1jJPWvfWKpD5jijAI4QuykqBPXhoU",
		"RO2MscN92tfRKSzl81BqNyo4D7zYtiHF",
	}
	var (
		data               map[string]interface{}
		err                error
		additionalKey      string
		success, ok, value bool
	)
	// 尝试每个密钥
	for len(additionalKeys) > 0 {
		// 生成随机索引
		randomIndex := rand.Intn(len(additionalKeys))
		// 获取随机元素
		additionalKey = additionalKeys[randomIndex]
		url := fmt.Sprintf("https://www.ipqualityscore.com/api/json/ip/%s/%s?strictness=0&allow_public_access_points=true&fast=true&lighter_penalties=true&mobile=true", additionalKey, ip)
		data, err = utils.FetchJsonFromURL(url, "tcp4", true, "")
		if err == nil {
			success, ok = data["success"].(bool)
			if ok {
				if !success {
					// 如果请求失败，从密钥列表中删除该密钥
					additionalKeys = append(additionalKeys[:randomIndex], additionalKeys[randomIndex+1:]...)
					continue
				} else {
					// 如果请求成功，才不再遍历
					break
				}
			}
		} else {
			// 如果请求失败，从密钥列表中删除该密钥
			additionalKeys = append(additionalKeys[:randomIndex], additionalKeys[randomIndex+1:]...)
		}
	}
	if err != nil || !success {
		return nil, nil, fmt.Errorf("all keys failed")
	}
	boolFields := []struct {
		key       string
		field     *string
		yesString string
		noString  string
	}{
		{"is_crawler", &securityInfo.IsCrawler, "Yes", "No"},
		{"mobile", &securityInfo.IsMobile, "Yes", "No"},
		{"proxy", &securityInfo.IsProxy, "Yes", "No"},
		{"vpn", &securityInfo.IsVpn, "Yes", "No"},
		{"tor", &securityInfo.IsTor, "Yes", "No"},
		{"recent_abuse", &securityInfo.IsAbuser, "Yes", "No"},
		{"bot_status", &securityInfo.IsBot, "Yes", "No"},
	}
	for _, field := range boolFields {
		if value, ok = data[field.key].(bool); ok {
			if value {
				*(field.field) = field.yesString
			} else {
				*(field.field) = field.noString
			}
		}
	}
	if fraudScore, ok := data["fraud_score"].(float64); ok {
		*securityScore.FraudScore = int(fraudScore)
	}
	securityScore.Tag = "E"
	securityInfo.Tag = "E"
	return securityScore, securityInfo, nil
}
