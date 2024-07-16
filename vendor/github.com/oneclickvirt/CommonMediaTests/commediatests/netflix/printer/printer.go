package printer

import (
	"fmt"

	"github.com/oneclickvirt/CommonMediaTests/commediatests/netflix/verify"
	. "github.com/oneclickvirt/defaultset"
)

func Print(fr verify.FinalResult, language string) (string, error) {
	var result string
	result = printResult("4", fr.Res[1], language)
	result += printResult("6", fr.Res[2], language)
	return result, nil
}

func printResult(ipVersion string, vResponse verify.VerifyResponse, language string) string {
	var result string
	result += Green(fmt.Sprintf("[IPV%s]", ipVersion)) + "\n"
	if language == "zh" {
		switch code := vResponse.StatusCode; {
		case code < -1:
			result += Yellow("您的网络可能没有正常配置IPv"+ipVersion+"，或者没有IPv"+ipVersion+"网络接入") + "\n"
		case code == -1:
			result += Yellow("Netflix在您的出口IP所在的国家不提供服务") + "\n"
		case code == 0:
			result += Blue("Netflix在您的出口IP所在的国家提供服务，但是您的IP疑似代理，无法正常使用服务") + "\n"
			result += Green("NF所识别的IP地域信息："+vResponse.CountryName) + "\n"
		case code == 1:
			result += Blue("您的出口IP可以使用Netflix，但仅可看Netflix自制剧") + "\n"
			result += Green("NF所识别的IP地域信息：") + vResponse.CountryName + "\n"
		case code == 2:
			result += Blue("您的出口IP完整解锁Netflix，支持非自制剧的观看") + "\n"
			result += Green("NF所识别的IP地域信息："+vResponse.CountryName) + "\n"
		case code == 3:
			result += Yellow("您的出口IP无法观看此电影") + "\n"
		case code == 4:
			result += Green("您的出口IP可以观看此电影") + "\n"
			result += Green("NF所识别的IP地域信息：") + vResponse.CountryName + "\n"
		default:
			result += Yellow("解锁检测失败，请稍后重试") + "\n"
		}
	} else if language == "en" {
		switch code := vResponse.StatusCode; {
		case code < -1:
			result += Yellow("Your network may not be properly configured for IPv"+ipVersion+", or there is no IPv"+ipVersion+" network access") + "\n"
		case code == -1:
			result += Yellow("Netflix does not provide service in the country where your exit IP is located") + "\n"
		case code == 0:
			result += Blue("Netflix provides service in the country where your exit IP is located, but your IP seems to be a proxy and cannot be used to access the service normally") + "\n"
			result += Green("Netflix identified IP region: "+vResponse.CountryName) + "\n"
		case code == 1:
			result += Blue("Your exit IP can access Netflix, but you can only watch Netflix originals") + "\n"
			result += Green("Netflix identified IP region: "+vResponse.CountryName) + "\n"
		case code == 2:
			result += Blue("Your exit IP fully unlocks Netflix, supporting the viewing of non-original content") + "\n"
			result += Green("Netflix identified IP region: "+vResponse.CountryName) + "\n"
		case code == 3:
			result += Yellow("Your exit IP cannot watch this movie") + "\n"
		case code == 4:
			result += Green("Your exit IP can watch this movie") + "\n"
			result += Green("Netflix identified IP region: "+vResponse.CountryName) + "\n"
		default:
			result += Yellow("Unlock detection failed, please try again later") + "\n"
		}
	}
	return result
}
