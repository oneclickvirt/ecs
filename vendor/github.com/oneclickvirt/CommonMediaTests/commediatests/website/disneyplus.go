package website

import (
	"fmt"
	"sync"

	. "github.com/oneclickvirt/defaultset"
)

func printDisneyplusResult(ipVersion string, QueryStatus string, VerifyStatus int, language string) string {
	var result string
	if QueryStatus != "" {
		result += Green(fmt.Sprintf("[IPV%s]", ipVersion)) + "\n"
	}
	if language == "zh" {
		switch QueryStatus {
		case "400":
			result += Yellow("DisneyPlus在您的出口IP所在的国家不提供服务") + "\n"
			break
		case "Unavailable":
			result += Yellow("当前出口不在DisneyPlus所支持的地区") + "\n"
			break
		case "-1":
			result += Blue("当前IPv4出口所在地区即将开通DisneyPlus") + "\n"
			break
		default:
			if VerifyStatus == -1 {
				result += Yellow("当前出口所在地区不能解锁DisneyPlus") + "\n"
			} else {
				result += Green("当前出口所在地区解锁DisneyPlus") + "\n"
				result += Green("区域：") + QueryStatus + Green(" 区") + "\n"
			}
		}
	} else if language == "en" {
		switch QueryStatus {
		case "400":
			result += Yellow("DisneyPlus does not provide service in the country where your exit IP is located") + "\n"
			break
		case "Unavailable":
			result += Yellow("The current exit is not supported in the region of DisneyPlus") + "\n"
			break
		case "-1":
			result += Blue("DisneyPlus will soon be available in the current region of the IPv4 exit") + "\n"
			break
		default:
			if VerifyStatus == -1 {
				result += Yellow("The current exit region cannot unlock DisneyPlus") + "\n"
			} else {
				result += Green("The current exit region unlocks DisneyPlus") + "\n"
				result += Green("Region: ") + QueryStatus + Green(" region") + "\n"
			}
		}
	}
	return result
}

func Disneyplus(language string) (string, error) {
	var (
		result, QueryStatusv4, QueryStatusv6 string
		VerifyStatus                         int
		wg                                   sync.WaitGroup
	)
	wg.Add(3)
	go func() {
		defer wg.Done()
		QueryStatusv4 = DisneyplusQueryAreaAvailable("ipv4")
	}()
	go func() {
		defer wg.Done()
		QueryStatusv6 = DisneyplusQueryAreaAvailable("ipv6")
	}()
	go func() {
		defer wg.Done()
		VerifyStatus = DisneyplusVerifyAuthorized()
	}()
	wg.Wait()
	if VerifyStatus == -2 {
		if language == "zh" {
			result += Purple("无法获取DisneyPlus权验接口信息，当前测试可能会不准确") + "\n"
		} else if language == "en" {
			result += Purple("Unable to obtain DisneyPlus authentication interface information, "+
				"current tests may be inaccurate") + "\n"
		}
	}
	result += printDisneyplusResult("4", QueryStatusv4, VerifyStatus, language)
	result += printDisneyplusResult("6", QueryStatusv6, VerifyStatus, language)
	return result, nil
}
