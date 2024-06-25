package network

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/baseinfo"
	"github.com/oneclickvirt/basics/network/utils"
	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/security/network/printhead"
	. "github.com/oneclickvirt/security/network/security"
)

// sortAndTranslateText 对原始文本进行排序和翻译
func sortAndTranslateText(orginList []string, language string, fields []string) string {
	var result string
	for _, key := range fields {
		var displayKey string
		if language == "zh" {
			displayKey = model.TranslationMap[key]
			if displayKey == "" {
				displayKey = key
			}
		} else {
			displayKey = key
		}
		for _, line := range orginList {
			if strings.Contains(line, key) {
				if displayKey == key {
					result = result + line + "\n"
				} else {
					result = result + strings.ReplaceAll(line, key, displayKey) + "\n"
				}
				break
			}
		}
	}
	return result
}

// fetchAndLogInfo 子函数执行和日志记录
func fetchAndLogInfo(wg *sync.WaitGroup, ip string, scorePtr **model.SecurityScore, infoPtr **model.SecurityInfo, fetchFunc func(string) (*model.SecurityScore, *model.SecurityInfo, error)) {
	defer wg.Done()
	var err error
	if scorePtr != nil && *scorePtr != nil && infoPtr != nil && *infoPtr != nil {
		*scorePtr, *infoPtr, err = fetchFunc(ip)
	} else if scorePtr == nil && infoPtr != nil && *infoPtr != nil {
		_, *infoPtr, err = fetchFunc(ip)
	} else if scorePtr != nil && *scorePtr != nil && infoPtr == nil {
		*scorePtr, _, err = fetchFunc(ip)
	}
	if err != nil {
		if model.EnableLoger {
			Logger.Info(fmt.Sprintf("%s: %s", runtime.FuncForPC(reflect.ValueOf(fetchFunc).Pointer()).Name(), err.Error()))
		}
	}
}

// Ipv4SecurityCheck 检测 ipv4 安全信息和安全得分
func Ipv4SecurityCheck(ipInfo *model.IpInfo, cheervisionInfo *model.SecurityInfo, language string) (string, error) {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	if ipInfo == nil {
		if model.EnableLoger {
			Logger.Info("ipv4 is not available")
		}
		return "", fmt.Errorf("ipv4 is not available")
	}
	if cheervisionInfo == nil {
		if model.EnableLoger {
			Logger.Info("ipv4 cheervisionInfo nil")
		}
	}
	var (
		ip, temp, orgin, result string
		wg                      sync.WaitGroup
		iPInfoIoInfo            = &model.SecurityInfo{}
		scamalyticsInfo         = &model.SecurityInfo{}
		abuseipdbInfo           = &model.SecurityInfo{}
		ip2locationIoInfo       = &model.SecurityInfo{}
		ipapicomInfo            = &model.SecurityInfo{}
		ipwhoisioInfo           = &model.SecurityInfo{}
		ipregistryCoInfo        = &model.SecurityInfo{}
		ipdataCoInfo            = &model.SecurityInfo{}
		dbIpComInfo             = &model.SecurityInfo{}
		ipapiisInfo             = &model.SecurityInfo{}
		ipapiComInfo            = &model.SecurityInfo{}
		abstractapiInfo         = &model.SecurityInfo{}
		ipqualityscoreComInfo   = &model.SecurityInfo{}

		scamalyticsScore       = &model.SecurityScore{}
		virustotalScore        = &model.SecurityScore{}
		abuseipdbScore         = &model.SecurityScore{}
		dbIpComScore           = &model.SecurityScore{}
		ipapiisScore           = &model.SecurityScore{}
		ipdataCoScore          = &model.SecurityScore{}
		ipapiComScore          = &model.SecurityScore{}
		ipqualityscoreComScore = &model.SecurityScore{}
	)
	ip = ipInfo.Ip
	wg.Add(14)
	go fetchAndLogInfo(&wg, ip, nil, &iPInfoIoInfo, IPInfoIo)
	go fetchAndLogInfo(&wg, ip, &scamalyticsScore, &scamalyticsInfo, Scamalytics)
	go fetchAndLogInfo(&wg, ip, &virustotalScore, nil, Virustotal)
	go fetchAndLogInfo(&wg, ip, &abuseipdbScore, &abuseipdbInfo, Abuseipdb)
	go fetchAndLogInfo(&wg, ip, nil, &ip2locationIoInfo, Ip2locationIo)
	go fetchAndLogInfo(&wg, ip, nil, &ipapicomInfo, IpApiCom)
	go fetchAndLogInfo(&wg, ip, nil, &ipwhoisioInfo, IpwhoisIo)
	go fetchAndLogInfo(&wg, ip, nil, &ipregistryCoInfo, IpregistryCo)
	go fetchAndLogInfo(&wg, ip, &ipdataCoScore, &ipdataCoInfo, IpdataCo)
	go fetchAndLogInfo(&wg, ip, &dbIpComScore, &dbIpComInfo, DbIpCom)
	go fetchAndLogInfo(&wg, ip, &ipapiisScore, &ipapiisInfo, Ipapiis)
	go fetchAndLogInfo(&wg, ip, &ipapiComScore, &ipapiComInfo, IpapiCom)
	go fetchAndLogInfo(&wg, ip, nil, &abstractapiInfo, Abstractapi)
	go fetchAndLogInfo(&wg, ip, &ipqualityscoreComScore, &ipqualityscoreComInfo, IpqualityscoreCom)
	wg.Wait()
	// 构建非空信息
	var allScoreList []model.SecurityScore
	scorePointers := []*model.SecurityScore{virustotalScore, scamalyticsScore, abuseipdbScore, dbIpComScore,
		ipapiisScore, ipapiComScore, ipdataCoScore, ipqualityscoreComScore}
	for _, score := range scorePointers {
		if score != nil {
			allScoreList = append(allScoreList, *score)
		}
	}
	var allInfoList []model.SecurityInfo
	infoPointers := []*model.SecurityInfo{iPInfoIoInfo, scamalyticsInfo, abuseipdbInfo, ip2locationIoInfo, ipapicomInfo,
		ipwhoisioInfo, ipregistryCoInfo, ipdataCoInfo, dbIpComInfo, ipapiisInfo, ipapiComInfo, abstractapiInfo,
		cheervisionInfo, ipqualityscoreComInfo}
	for _, info := range infoPointers {
		if info != nil {
			allInfoList = append(allInfoList, *info)
		}
	}
	// 构建回传的文本内容
	temp += FormatSecurityScore(allScoreList)
	temp += "\n"
	temp += FormatSecurityInfo(allInfoList)
	// 分割输入为行
	lines := strings.Split(temp, "\n")
	// 初始化一个映射用于存储冒号之前的内容及其对应的行数
	contentMap := make(map[string][]int)
	// 遍历每一行，提取冒号之前的内容及其行数
	for i, line := range lines {
		// 如果行为空则跳过
		if line == "" {
			continue
		}
		// 切割行，以冒号为分隔符
		parts := strings.Split(line, ":")
		// 获取冒号之前的内容
		content := parts[0]
		// 将当前行的行号添加到映射中
		contentMap[content] = append(contentMap[content], i)
	}
	// 遍历映射，拼接相同内容的行
	for _, lineNumbers := range contentMap {
		if len(lineNumbers) > 1 { // 只对有多个行的内容进行拼接
			// 初始化一个字符串切片，用于存储拼接后的行
			var mergedLines []string
			// 遍历相同内容的行 添加当前行到拼接后的行中
			for _, lineNumber := range lineNumbers {
				if lineNumber == lineNumbers[0] {
					mergedLines = append(mergedLines, strings.TrimSpace(lines[lineNumber]))
				} else {
					mergedLines = append(mergedLines, strings.TrimSpace(strings.Split(lines[lineNumber], ":")[1]))
				}
			}
			// 将拼接后的行以空格连接起来
			mergedLine := strings.Join(mergedLines, " ")
			// 替换原始行中相同内容的行为拼接后的行，仅替换一次，其他行标注要删除
			isMerged := false
			for _, lineNumber := range lineNumbers {
				if !isMerged {
					lines[lineNumber] = mergedLine
					isMerged = true
				} else {
					lines[lineNumber] += "delete"
				}
			}
		}
	}
	// 删除对应的行，构建原始文本
	for _, line := range lines {
		if !strings.Contains(line, "delete") {
			orgin = orgin + line + "\n"
		}
	}
	orginList := strings.Split(orgin, "\n")
	// 将原始文本按要求进行排序和翻译
	var score model.SecurityScore
	scoreFields := utils.ExtractFieldNames(&score)
	var info model.SecurityInfo
	infoFields := utils.ExtractFieldNames(&info)
	// 拼接安全得分
	if language == "zh" {
		result += Blue("安全得分:") + "\n"
	} else if language == "en" {
		result = Blue("Security Score:") + "\n"
	}
	if len(scoreFields) > 4 {
		result += sortAndTranslateText(orginList, language, scoreFields[:len(scoreFields)-4])
	} else {
		result += sortAndTranslateText(orginList, language, scoreFields)
	}
	// 安全信息中前三个是字符串类型的得分
	result += sortAndTranslateText(orginList, language, infoFields[:3])
	// 需要确保后4个属性都为对应属性时才进行说明的拼接
	if len(scoreFields) > 4 {
		t := ""
		foundKeys := 0
		for _, key := range scoreFields[len(scoreFields)-4:] {
			var displayKey string
			if language == "zh" {
				displayKey = model.TranslationMap[key]
				if displayKey == "" {
					displayKey = key
				}
			} else {
				displayKey = key
			}
			found := false
			for _, line := range orginList {
				if strings.Contains(line, key) {
					key = strings.ReplaceAll(key, ": ", "")
					if displayKey == key {
						t = t + line + " "
					} else {
						t = t + strings.ReplaceAll(line, key, displayKey) + " "
					}
					found = true
					break
				}
			}
			if found {
				foundKeys++
			}
		}
		if foundKeys == 4 {
			if language == "zh" {
				result = result + "黑名单记录统计:(有多少黑名单网站有记录):\n" + t + "\n"
			} else if language == "en" {
				result = result + "Blacklist_Records_Statistics(how many blacklisted websites have records):\n" + t + "\n"
			}
		}
	}
	// 拼接安全信息
	if language == "zh" {
		result += Blue("安全信息:") + "\n"
	} else if language == "en" {
		result += Blue("Security Info:") + "\n"
	}
	result += sortAndTranslateText(orginList, language, infoFields[3:])
	return result, nil
}

// Ipv6SecurityCheck 检测 ipv4 安全信息和安全得分
func Ipv6SecurityCheck(ipInfo *model.IpInfo, cheervisionInfo *model.SecurityInfo, language string) (string, error) {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	if ipInfo == nil {
		if model.EnableLoger {
			Logger.Info("ipv6 is not available")
		}
		return "", fmt.Errorf("ipv6 is not available")
	}
	if cheervisionInfo == nil {
		if model.EnableLoger {
			Logger.Info("ipv6 cheervisionInfo nil")
		}
	}
	var (
		ip, temp, orgin, result string
		wg                      sync.WaitGroup
		scamalyticsInfo         = &model.SecurityInfo{}
		abuseipdbInfo           = &model.SecurityInfo{}
		ipapiisInfo             = &model.SecurityInfo{}
		ipapiComInfo            = &model.SecurityInfo{}
		scamalyticsScore        = &model.SecurityScore{}
		abuseipdbScore          = &model.SecurityScore{}
		ipapiisScore            = &model.SecurityScore{}
		ipapiComScore           = &model.SecurityScore{}
	)
	ip = ipInfo.Ip
	wg.Add(4)
	go fetchAndLogInfo(&wg, ip, &scamalyticsScore, &scamalyticsInfo, Scamalytics)
	go fetchAndLogInfo(&wg, ip, &abuseipdbScore, &abuseipdbInfo, Abuseipdb)
	go fetchAndLogInfo(&wg, ip, &ipapiisScore, &ipapiisInfo, Ipapiis)
	go fetchAndLogInfo(&wg, ip, &ipapiComScore, &ipapiComInfo, IpapiComIpv6)
	wg.Wait()
	// 构建非空信息
	var allScoreList []model.SecurityScore
	scorePointers := []*model.SecurityScore{scamalyticsScore, abuseipdbScore, ipapiisScore, ipapiComScore}
	for _, score := range scorePointers {
		if score != nil {
			allScoreList = append(allScoreList, *score)
		}
	}
	var allInfoList []model.SecurityInfo
	infoPointers := []*model.SecurityInfo{scamalyticsInfo, abuseipdbInfo, ipapiisInfo, ipapiComInfo, cheervisionInfo}
	for _, info := range infoPointers {
		if info != nil {
			allInfoList = append(allInfoList, *info)
		}
	}
	// 构建回传的文本内容
	temp += FormatSecurityScore(allScoreList)
	temp += "\n"
	temp += FormatSecurityInfo(allInfoList)
	// 分割输入为行
	lines := strings.Split(temp, "\n")
	// 初始化一个映射用于存储冒号之前的内容及其对应的行数
	contentMap := make(map[string][]int)
	// 遍历每一行，提取冒号之前的内容及其行数
	for i, line := range lines {
		// 如果行为空则跳过
		if line == "" {
			continue
		}
		// 切割行，以冒号为分隔符
		parts := strings.Split(line, ":")
		// 获取冒号之前的内容
		content := parts[0]
		// 将当前行的行号添加到映射中
		contentMap[content] = append(contentMap[content], i)
	}
	// 遍历映射，拼接相同内容的行
	for _, lineNumbers := range contentMap {
		if len(lineNumbers) > 1 { // 只对有多个行的内容进行拼接
			// 初始化一个字符串切片，用于存储拼接后的行
			var mergedLines []string
			// 遍历相同内容的行 添加当前行到拼接后的行中
			for _, lineNumber := range lineNumbers {
				if lineNumber == lineNumbers[0] {
					mergedLines = append(mergedLines, strings.TrimSpace(lines[lineNumber]))
				} else {
					mergedLines = append(mergedLines, strings.TrimSpace(strings.Split(lines[lineNumber], ":")[1]))
				}
			}
			// 将拼接后的行以空格连接起来
			mergedLine := strings.Join(mergedLines, " ")
			// 替换原始行中相同内容的行为拼接后的行，仅替换一次，其他行标注要删除
			isMerged := false
			for _, lineNumber := range lineNumbers {
				if !isMerged {
					lines[lineNumber] = mergedLine
					isMerged = true
				} else {
					lines[lineNumber] += "delete"
				}
			}
		}
	}
	// 删除对应的行，构建原始文本
	for _, line := range lines {
		if !strings.Contains(line, "delete") {
			orgin = orgin + line + "\n"
		}
	}
	orginList := strings.Split(orgin, "\n")
	var score model.SecurityScore
	scoreFields := utils.ExtractFieldNames(&score)
	var info model.SecurityInfo
	infoFields := utils.ExtractFieldNames(&info)
	// 拼接安全得分
	if language == "zh" {
		result += Blue("安全得分:") + "\n"
	} else if language == "en" {
		result = Blue("Security Score:") + "\n"
	}
	result += sortAndTranslateText(orginList, language, scoreFields)
	result += sortAndTranslateText(orginList, language, infoFields[:3])
	// 拼接完整安全信息
	if language == "zh" {
		result += Blue("安全信息:") + "\n"
	} else if language == "en" {
		result += Blue("Security Info:") + "\n"
	}
	result += sortAndTranslateText(orginList, language, infoFields[3:])
	return result, nil
}

// processPrintIPInfo 处理IP信息
func processPrintIPInfo(headASNString string, headLocationString string, ipResult *model.IpInfo) string {
	var info string
	// 处理ASN信息
	if ipResult.ASN != "" || ipResult.Org != "" {
		info += headASNString
		if ipResult.ASN != "" {
			info += "AS" + ipResult.ASN
			if ipResult.Org != "" {
				info += " "
			}
		}
		info += ipResult.Org + "\n"
	}
	// 处理位置信息
	if ipResult.City != "" || ipResult.Region != "" || ipResult.Country != "" {
		info += headLocationString
		if ipResult.City != "" {
			info += ipResult.City + " / "
		}
		if ipResult.Region != "" {
			info += ipResult.Region + " / "
		}
		if ipResult.Country != "" {
			info += ipResult.Country
		}
		info += "\n"
	}
	return info
}

// NetworkCheck 查询网络信息
// checkType 可选 both ipv4 ipv6
// enableSecurityCheck 可选 true false
// language 暂时仅支持 en 或 zh
// 回传 ipInfo securityInfo err
func NetworkCheck(checkType string, enableSecurityCheck bool, language string) (string, string, error) {
	var ipInfo, securityInfo string
	if checkType == "both" {
		ipInfoV4Result, cheervisionInfoV4, ipInfoV6Result, cheervisionInfoV6, _ := baseinfo.RunIpCheck("both")
		if ipInfoV4Result != nil {
			ipInfo += processPrintIPInfo(" IPV4 ASN            : ", " IPV4 Location       : ", ipInfoV4Result)
		}
		if ipInfoV6Result != nil {
			ipInfo += processPrintIPInfo(" IPV6 ASN            : ", " IPV6 Location       : ", ipInfoV6Result)
		}
		// 检测是否需要查询相关安全信息
		if enableSecurityCheck {
			var (
				wg                                       sync.WaitGroup
				ipv4Res, ipv6Res, ipv4DNSRes, ipv6DNSRes string
				err1, err2                               error
			)
			wg.Add(4)
			go func() {
				defer wg.Done()
				ipv4DNSRes = BlackList(ipInfoV4Result, "ipv4", language)
			}()
			go func() {
				defer wg.Done()
				ipv6DNSRes = BlackList(ipInfoV6Result, "ipv6", language)
			}()
			go func() {
				defer wg.Done()
				ipv4Res, err1 = Ipv4SecurityCheck(ipInfoV4Result, cheervisionInfoV4, language)
			}()
			go func() {
				defer wg.Done()
				ipv6Res, err2 = Ipv6SecurityCheck(ipInfoV6Result, cheervisionInfoV6, language)
			}()
			wg.Wait()
			securityHead, err3 := printhead.PrintDatabaseInfo(language)
			if err1 == nil && err2 == nil && err3 == nil {
				securityInfo = securityHead + Green("IPV4:") + "\n" + ipv4Res + ipv4DNSRes + Green("IPV6:") + "\n" + ipv6Res + ipv6DNSRes
				return ipInfo, securityInfo, nil
			} else if err1 == nil && err2 != nil && err3 == nil {
				securityInfo = securityHead + Green("IPV4:") + "\n" + ipv4Res + ipv4DNSRes
				return ipInfo, securityInfo, nil
			} else if err1 != nil && err2 == nil && err3 == nil {
				securityInfo = securityHead + Green("IPV6:") + "\n" + ipv6Res + ipv6DNSRes
				return ipInfo, securityInfo, nil
			} else {
				return ipInfo, "", nil
			}
		} else {
			return ipInfo, "", nil
		}
	} else if checkType == "ipv4" {
		ipInfoV4Result, cheervisionInfoV4, _, _, _ := baseinfo.RunIpCheck("ipv4")
		if ipInfoV4Result != nil {
			ipInfo += processPrintIPInfo(" IPV4 ASN            : ", " IPV4 Location       : ", ipInfoV4Result)
		}
		if enableSecurityCheck {
			var (
				wg                  sync.WaitGroup
				ipv4Res, ipv4DNSRes string
				err1                error
			)
			wg.Add(2)
			go func() {
				defer wg.Done()
				ipv4DNSRes = BlackList(ipInfoV4Result, "ipv4", language)
			}()
			go func() {
				defer wg.Done()
				ipv4Res, err1 = Ipv4SecurityCheck(ipInfoV4Result, cheervisionInfoV4, language)
			}()
			wg.Wait()
			securityHead, err2 := printhead.PrintDatabaseInfo(language)
			if err1 == nil && err2 == nil {
				securityInfo = securityHead + Green("IPV4:") + "\n" + ipv4Res + ipv4DNSRes
				return ipInfo, securityInfo, nil
			} else {
				return ipInfo, "", nil
			}
		} else {
			return ipInfo, "", nil
		}
	} else if checkType == "ipv6" {
		_, _, ipInfoV6Result, cheervisionInfoV6, _ := baseinfo.RunIpCheck("ipv6")
		if ipInfoV6Result != nil {
			ipInfo += processPrintIPInfo(" IPV6 ASN            : ", " IPV6 Location       : ", ipInfoV6Result)
		}
		if enableSecurityCheck {
			var (
				wg                  sync.WaitGroup
				ipv6Res, ipv6DNSRes string
				err1                error
			)
			wg.Add(2)
			go func() {
				defer wg.Done()
				ipv6DNSRes = BlackList(ipInfoV6Result, "ipv6", language)
			}()
			go func() {
				defer wg.Done()
				ipv6Res, err1 = Ipv6SecurityCheck(ipInfoV6Result, cheervisionInfoV6, language)
			}()
			wg.Wait()
			securityHead, err2 := printhead.PrintDatabaseInfo(language)
			if err1 == nil && err2 == nil {
				securityInfo = securityHead + Green("IPV6:") + "\n" + ipv6Res + ipv6DNSRes
				return ipInfo, securityInfo, nil
			} else {
				return ipInfo, "", nil
			}
		} else {
			return ipInfo, "", nil
		}
	}
	return "", "", fmt.Errorf("wrong in NetworkCheck")
}
