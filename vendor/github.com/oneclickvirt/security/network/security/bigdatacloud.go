package security

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// Bigdatacloud 获取 www.bigdatacloud.com 的信息
func Bigdatacloud(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	securityInfo := &model.SecurityInfo{}
	additionalKeys := []string{
		"bdc_59177983ae3a48578324dbac416f6ccc",
		"bdc_a3fd490909df46c3b6e259e86e2425e3",
		"bdc_1d5bbb07eb374d60ac3c0ddb6950972e",
		"bdc_c13e3a1984864b699e461a25f5a138ed",
		"bdc_0737aab69de84723b4ad0805cba82523",
		"bdc_4422bb94409c46e986818d3e9f3b2bc2",
		"bdc_8ae05a7492c64933ab7b03ac107cf100",
	}
	// 生成随机索引
	randomIndex := rand.Intn(len(additionalKeys))
	var resp *req.Response
	var err error
	for retry := 0; retry < len(additionalKeys); retry++ {
		// 获取随机元素
		additionalKey := additionalKeys[(randomIndex+retry)%len(additionalKeys)]
		url := fmt.Sprintf("https://api-bdc.net/data/ip-geolocation-full?ip=%s&localityLanguage=en&key=%s", ip, additionalKey)
		client := req.C()
		client.ImpersonateChrome()
		client.SetTimeout(6 * time.Second)
		client.R().
			SetRetryCount(2).
			SetRetryBackoffInterval(1*time.Second, 5*time.Second).
			SetRetryFixedInterval(2 * time.Second)
		resp, err = client.R().Get(url)
		if err != nil {
			// 如果请求失败，则尝试下一个密钥
			continue
		}
		if resp.StatusCode == 200 {
			// 处理成功的情况
			break
		}
	}
	defer resp.Body.Close()
	if err != nil || resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("error fetching info: all keys failed")
	}
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, nil, fmt.Errorf("Error decoding Bigdatacloud info: %v ", err)
	}
	if usageType, ok := data["securityThreat"].(string); ok {
		securityInfo.UsageType = usageType
	}
	if securitys, ok := data["hazardReport"].(map[string]interface{}); ok {
		if isTor, ok := securitys["isKnownAsTorServer"].(bool); ok {
			securityInfo.IsTor = utils.BoolToString(isTor)
		}
		if isVpn, ok := securitys["isKnownAsVpn"].(bool); ok {
			securityInfo.IsVpn = utils.BoolToString(isVpn)
		}
		if isProxy, ok := securitys["isKnownAsProxy"].(bool); ok {
			securityInfo.IsProxy = utils.BoolToString(isProxy)
		}
		if isSpamhausDrop, ok := securitys["isSpamhausDrop"].(bool); ok {
			securityInfo.IsAbuser = utils.BoolToString(isSpamhausDrop)
		}
		if isSpamhausEdrop, ok := securitys["isSpamhausEdrop"].(bool); ok {
			if securityInfo.IsAbuser == "" || securityInfo.IsAbuser == "No" {
				securityInfo.IsAbuser = utils.BoolToString(isSpamhausEdrop)
			}
		}
		if isSpamhausAsnDrop, ok := securitys["isSpamhausAsnDrop"].(bool); ok {
			if securityInfo.IsAbuser == "" || securityInfo.IsAbuser == "No" {
				securityInfo.IsAbuser = utils.BoolToString(isSpamhausAsnDrop)
			}
		}
		if isBlacklistedUceprotect, ok := securitys["isBlacklistedUceprotect"].(bool); ok {
			if securityInfo.IsThreat == "" || securityInfo.IsThreat == "No" {
				securityInfo.IsThreat = utils.BoolToString(isBlacklistedUceprotect)
			}
		}
		if isBlacklistedBlocklistDe, ok := securitys["isBlacklistedBlocklistDe"].(bool); ok {
			if securityInfo.IsThreat == "" || securityInfo.IsThreat == "No" {
				securityInfo.IsThreat = utils.BoolToString(isBlacklistedBlocklistDe)
			}
		}
		if isBogon, ok := securitys["isBogon"].(bool); ok {
			if securityInfo.IsBogon == "" || securityInfo.IsBogon == "No" {
				securityInfo.IsBogon = utils.BoolToString(isBogon)
			}
		}
		if isUnreachable, ok := securitys["isUnreachable"].(bool); ok {
			if securityInfo.IsBogon == "" || securityInfo.IsBogon == "No" {
				securityInfo.IsBogon = utils.BoolToString(isUnreachable)
			}
		}
		if isHostingAsn, ok := securitys["isHostingAsn"].(bool); ok {
			securityInfo.IsDatacenter = utils.BoolToString(isHostingAsn)
		}
		if isCellular, ok := securitys["isCellular"].(bool); ok {
			securityInfo.IsMobile = utils.BoolToString(isCellular)
		}
		if iCloudPrivateRelay, ok := securitys["iCloudPrivateRelay"].(bool); ok {
			securityInfo.IsRelay = utils.BoolToString(iCloudPrivateRelay)
		}
	}
	securityInfo.Tag = "C"
	return nil, securityInfo, nil
}
