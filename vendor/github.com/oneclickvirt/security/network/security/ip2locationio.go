package security

import (
	"fmt"
	"math/rand"

	"github.com/oneclickvirt/basics/model"
	"github.com/oneclickvirt/basics/network/utils"
)

// Ip2locationIo 获取 ip2location.io 的信息 需要优化
func Ip2locationIo(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
	if ip == "" {
		return nil, nil, fmt.Errorf("IP地址为空")
	}
	securityInfo := &model.SecurityInfo{}
	additionalKeys := []string{
		"EA80FE926355332BE6006367A092348E",
		"0d4f60641cd9b95ff5ac9b4d866a0655",
		"7C5384E65E3B5B520A588FB8F9281719",
		"4E191A613023EA66D24E35E41C870D3B",
		"3D07E2EAAF55940AF44734C3F2AC7C1A",
		"32D24DBFB5C3BFFDEF5FE9331F93BA5B",
		"28cc35ee8608480fa7087be0e435320c",
	}
	var (
		data          map[string]interface{}
		err           error
		additionalKey string
		ok            bool
	)
	// 尝试每个密钥
	for len(additionalKeys) > 0 {
		// 生成随机索引
		randomIndex := rand.Intn(len(additionalKeys))
		// 获取随机元素
		additionalKey = additionalKeys[randomIndex]
		url := fmt.Sprintf("https://api.ip2location.io/?key=%s&ip=%s", additionalKey, ip)
		data, err = utils.FetchJsonFromURL(url, "tcp4", true, "")
		if err == nil {
			_, ok = data["error"].(map[string]interface{})
			if ok {
				// 如果请求失败，从密钥列表中删除该密钥
				additionalKeys = append(additionalKeys[:randomIndex], additionalKeys[randomIndex+1:]...)
				continue
			} else {
				// 请求成功，直接跳出循环
				break
			}
		} else {
			// 如果请求失败，从密钥列表中删除该密钥
			additionalKeys = append(additionalKeys[:randomIndex], additionalKeys[randomIndex+1:]...)
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("all keys failed")
	}
	if isProxy, ok := data["is_proxy"].(bool); ok {
		securityInfo.IsProxy = utils.BoolToString(isProxy)
	}
	securityInfo.Tag = "4"
	return nil, securityInfo, nil
}

// Ip2locationIo 获取 ip2location.io 的信息 session还是过期，没整明白
// func Ip2locationIo(ip string) (*model.SecurityScore, *model.SecurityInfo, error) {
// 	if ip == "" {
// 		return nil, nil, fmt.Errorf("IP地址为空")
// 	}
// 	const (
// 		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0"
// 		accept    = "application/json, text/javascript, */*; q=0.01"
// 		referer   = "https://ip2location.io"
// 		host      = "www.ip2location.io"
// 	)
// 	request := gorequest.New().Get("https://ip2location.io")
// 	request.Set("User-Agent", userAgent)
// 	request.Set("Accept", accept)
// 	request.Set("Referer", referer)
// 	request.Set("Host", host)
// 	response, body, err := request.End()
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("Can not load response body: %v", err)
// 	}
// 	if response.StatusCode != http.StatusOK {
// 		return nil, nil, fmt.Errorf("HTTP request failed: %d", response.StatusCode)
// 	}
// 	doc, readErr := goquery.NewDocumentFromReader(strings.NewReader(body))
// 	if readErr != nil {
// 		return nil, nil, fmt.Errorf("Can not parse page: %v", readErr.Error())
// 	}
// 	tokenConfig := doc.Find("div.input-group-append").Parent().Next()
// 	if tokenConfig.Is("input") {
// 		token, exit := tokenConfig.Attr("value")
// 		if exit {
// 			a := strings.Split(token, ";")
// 			l, _ := strconv.Atoi(a[1])
// 			t := a[0][10:20] + a[0][len(a[0])-10:] + a[0][0:10] + strings.ReplaceAll(a[0][20:(20+l)], "#", "=")
// 			realToken, decodeErr := base64.StdEncoding.DecodeString(t)
// 			if decodeErr != nil {
// 				return nil, nil, fmt.Errorf("Can not decode token: %v", decodeErr)
// 			}
// 			cookie := strings.ReplaceAll(response.Header.Get("Set-Cookie"),
// 				"path=/; secure; HttpOnly; SameSite=LAX", "")
// 			cookie = "_ga_M904Z63N7V=GS1.1.1714095355.7.1.1714096650.0.0.0; _ga=GA1.1.739884093.1713917871; " +
// 				"__site=c3adbb276a9cdd9e4579ff558949224416bc1514; " +
// 				"CookieConsent={necessary:true,preferences:true,statistics:true,marketing:true}; " + cookie
// 			url := fmt.Sprintf("https://www.ip2location.io/lookup-ip.json?ip=%s&token=%s", ip, string(realToken))
// 			requestToken := gorequest.New().
// 				Get(url).Retry(3, 6*time.Second, http.StatusBadRequest, http.StatusInternalServerError)
// 			requestToken.Set("User-Agent", userAgent)
// 			requestToken.Set("Accept", accept)
// 			requestToken.Set("Referer", referer)
// 			requestToken.Set("Cookie", cookie)
// 			request.Set("Host", host)
// 			responseToken, bodyToken, errToken := requestToken.End()
// 			if errToken != nil {
// 				return nil, nil, fmt.Errorf("Can not load response body: %v", errToken)
// 			}
// 			if responseToken.StatusCode != http.StatusOK {
// 				return nil, nil, fmt.Errorf("HTTP request failed: %d", responseToken.StatusCode)
// 			}
// 			fmt.Println(url)
// 			fmt.Println(cookie)
// 			fmt.Println(string(realToken))
// 			fmt.Println(bodyToken)
// 		}
// 	}
// 	return nil, nil, nil
// }

