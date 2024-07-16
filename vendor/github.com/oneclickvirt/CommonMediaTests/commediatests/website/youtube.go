package website

import (
	"fmt"
	"strings"
	"sync"

	"github.com/oneclickvirt/CommonMediaTests/commediatests/utils"
	. "github.com/oneclickvirt/defaultset"
)

func printYoutubeResult(ipVersion, methodV4, ispV4, airCodeV4, RegionCode, language string) string {
	var result string
	result += Green(fmt.Sprintf("[IPV%s]", ipVersion)) + "\n"
	if language == "zh" {
		result += Green("连接方式:") + " " + methodV4 + "\n"
		if ispV4 != "" {
			result += Green("ISP运营商:") + " " + strings.ToUpper(ispV4) + "\n"
		}
		result += Green("视频缓存节点地域:") +
			fmt.Sprintf(" %s(%s)", utils.FindAirCode(airCodeV4), strings.ToUpper(airCodeV4)) + "\n"
		if RegionCode != "" && RegionCode != "null" {
			result += Green("Youtube识别地域:") +
				fmt.Sprintf(" %s(%s)", utils.CountryCodeToCountryName(strings.ToLower(RegionCode)), RegionCode) + "\n"
		}
	} else if language == "en" {
		result += Green("Connection Method:") + " " + methodV4 + "\n"
		if ispV4 != "" {
			result += Green("ISP Provider:") + " " + strings.ToUpper(ispV4) + "\n"
		}
		result += Green("Video Cache Node Region:") +
			fmt.Sprintf(" %s(%s)", utils.FindAirCode(airCodeV4), strings.ToUpper(airCodeV4)) + "\n"
		if RegionCode != "" && RegionCode != "null" {
			result += Green("Youtube Recognized Region:") +
				fmt.Sprintf(" %s(%s)", utils.CountryCodeToCountryName(strings.ToLower(RegionCode)), RegionCode) + "\n"
		}
	}
	return result
}

func YoutubeCheck(language string) (string, error) {
	var (
		ipv4, ipv6, result, methodV4, ispV4, airCodeV4, methodV6, ispV6, airCodeV6, RegionCodeV4, RegionCodeV6 string
		responseCodeV4, responseCodeV6                                                                         int
		err, err4, err6                                                                                        error
		wg                                                                                                     sync.WaitGroup
	)
	dns := "redirector.googlevideo.com"
	ipv4, ipv6, err = lookupIP(dns)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	wg.Add(4)
	go func() {
		defer wg.Done()
		responseCodeV4, methodV4, ispV4, airCodeV4 = utils.RequestYoutubeIP("https://redirector.googlevideo.com/report_mapping", ipv4)
	}()
	go func() {
		defer wg.Done()
		responseCodeV6, methodV6, ispV6, airCodeV6 = utils.RequestYoutubeIP("https://redirector.googlevideo.com/report_mapping", ipv6)
	}()
	go func() {
		defer wg.Done()
		RegionCodeV4, err4 = YoutubeRegionCheck("ipv4")
	}()
	go func() {
		defer wg.Done()
		RegionCodeV6, err6 = YoutubeRegionCheck("ipv6")
	}()
	wg.Wait()
	if responseCodeV4 == 200 {
		if err4 != nil {
			// TODO
		}
		result += printYoutubeResult("4", methodV4, ispV4, airCodeV4, RegionCodeV4, language)
	} else {
		result += Green("[IPV4]") + "\n"
		if language == "zh" {
			result += Yellow("Youtube在您的出口IP所在的国家不提供服务") + "\n"
		} else if language == "en" {
			result += Yellow("Youtube is not available in the country of your export IP") + "\n"
		}
	}
	if responseCodeV6 == 200 {
		if err6 == nil {
			// TODO
		}
		result += printYoutubeResult("6", methodV6, ispV6, airCodeV6, RegionCodeV6, language)
	} else {
		result += Green("[IPV6]") + "\n"
		if language == "zh" {
			result += Yellow("Youtube在您的出口IP所在的国家不提供服务") + "\n"
		} else if language == "en" {
			result += Yellow("Youtube is not available in the country of your export IP") + "\n"
		}
	}
	return result, nil
}
