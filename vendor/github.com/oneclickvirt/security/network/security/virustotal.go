package security

import (
	"fmt"
	"math/rand"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
	. "github.com/oneclickvirt/defaultset"
)

// Virustotal 查询 virustotal.com 的信息
func Virustotal(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	securityScore := &model.SecurityScore{
		Reputation:             new(int),
		CommunityVoteHarmless:  new(int),
		CommunityVoteMalicious: new(int),
		HarmlessnessRecords:    new(int),
		MaliciousRecords:       new(int),
		SuspiciousRecords:      new(int),
		NoRecords:              new(int),
	}
	var (
		data             map[string]interface{}
		err              error
		additionalHeader string
		ok               bool
	)
	url := fmt.Sprintf("https://www.virustotal.com/api/v3/ip_addresses/%s", ip)
	additionalHeaders := []string{
		"x-apikey:b14b9651db5022fb8a73c1cfa29041e25786ef78c83c44fb0bfd4a0993688477",
		"x-apikey:9929218dcd124c19bcee49ecd6d7555213de0e8f27d407cc3e85c92c3fc2508e",
		"x-apikey:9655ff318973231a02e65749561436bbc5817f38ee3789ac363e8fa51a41a49d",
		"x-apikey:17f1f9ca3b96391150c71e1ba68adc1e3c0c3bb233011a55e543bffb1211a5ba",
	}
	// 尝试每个密钥
	for len(additionalHeaders) > 0 {
		// 生成随机索引
		randomIndex := rand.Intn(len(additionalHeaders))
		// 获取随机元素
		additionalHeader = additionalHeaders[randomIndex]
		data, err = utils.FetchJsonFromURL(url, "tcp4", true, additionalHeader)
		if err == nil {
			_, ok = data["error"].(map[string]interface{})
			if ok {
				// 如果请求失败，从密钥列表中删除该密钥
				additionalHeaders = append(additionalHeaders[:randomIndex], additionalHeaders[randomIndex+1:]...)
				continue
			} else {
				// 如果请求成功，才不再遍历
				break
			}
		} else {
			// 如果请求失败，从密钥列表中删除该密钥
			additionalHeaders = append(additionalHeaders[:randomIndex], additionalHeaders[randomIndex+1:]...)
		}
	}
	if err != nil && !ok {
		return nil, nil, fmt.Errorf("all keys failed")
	}
	if dataMap, ok := data["data"].(map[string]interface{}); ok {
		if attributesMap, ok := dataMap["attributes"].(map[string]interface{}); ok {
			if reputation, ok := attributesMap["reputation"].(float64); ok {
				*securityScore.Reputation = int(reputation)
			} else {
				if model.EnableLoger {
					Logger.Info(fmt.Sprint("Reputation is not of type float64"))
				}
			}
			if lastAnalysisStats, ok := attributesMap["last_analysis_stats"].(map[string]interface{}); ok {
				if malicious, ok := lastAnalysisStats["malicious"].(float64); ok {
					*securityScore.MaliciousRecords = int(malicious)
				} else {
					if model.EnableLoger {
						Logger.Info(fmt.Sprint("Malicious samples not found"))
					}
				}
				if suspicious, ok := lastAnalysisStats["suspicious"].(float64); ok {
					*securityScore.SuspiciousRecords = int(suspicious)
				} else {
					if model.EnableLoger {
						Logger.Info(fmt.Sprint("Suspicious samples not found"))
					}
				}
				if undetected, ok := lastAnalysisStats["undetected"].(float64); ok {
					*securityScore.NoRecords = int(undetected)
				} else {
					if model.EnableLoger {
						Logger.Info(fmt.Sprint("Undetected samples not found"))
					}
				}
				if harmless, ok := lastAnalysisStats["harmless"].(float64); ok {
					*securityScore.HarmlessnessRecords = int(harmless)
				} else {
					if model.EnableLoger {
						Logger.Info(fmt.Sprint("Harmless samples not found"))
					}
				}
			} else {
				if model.EnableLoger {
					Logger.Info(fmt.Sprint("last_analysis_stats is not of type ap[string]interface{}"))
				}
			}
			if totalVotes, ok := attributesMap["total_votes"].(map[string]interface{}); ok {
				if harmless, ok := totalVotes["harmless"].(float64); ok {
					*securityScore.CommunityVoteHarmless = int(harmless)
				} else {
					if model.EnableLoger {
						Logger.Info(fmt.Sprint("Harmless samples not found"))
					}
				}
				if malicious, ok := totalVotes["malicious"].(float64); ok {
					*securityScore.CommunityVoteMalicious = int(malicious)
				} else {
					if model.EnableLoger {
						Logger.Info(fmt.Sprint("Malicious samples not found"))
					}
				}
			} else {
				if model.EnableLoger {
					Logger.Info(fmt.Sprint("total_votes is not of type ap[string]interface{}"))
				}
			}
		} else {
			if model.EnableLoger {
				Logger.Info(fmt.Sprint("Attributes is not of type map[string]interface{}"))
			}
		}
	} else {
		if model.EnableLoger {
			Logger.Info(fmt.Sprint("Data is not of type map[string]interface{}"))
		}
	}
	securityScore.Tag = "2"
	return securityScore, nil, nil
}
