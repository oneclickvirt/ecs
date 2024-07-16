package utils

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// performRequest 发送HTTP请求
func performRequest(requrl, ip, method string, body io.Reader, headers map[string]string) (*http.Response, error) {
	urlValue, err := url.Parse(requrl)
	if err != nil {
		return nil, err
	}
	host := urlValue.Host
	if ip == "" {
		ip = host
	}
	newrequrl := strings.Replace(requrl, host, ip, 1)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{ServerName: host},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		Timeout:       5 * time.Second,
	}
	// 根据传入的 method 创建相应的请求
	req, err := http.NewRequest(method, newrequrl, body)
	if err != nil {
		return nil, fmt.Errorf("%s %s err: %v", method, newrequrl, err.Error())
	}
	req.Host = host
	req.Header.Set("USER-AGENT", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36")
	// 设置额外的请求头
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// RequestDisneyPlusIP 请求Disney+的IP
func RequestDisneyPlusIP(requrl, ip, method string) string {
	urlValue, err := url.Parse(requrl)
	if err != nil {
		return "400"
	}
	headers := map[string]string{
		"Host":          urlValue.Host,
		"authorization": "Bearer ZGlzbmV5JmJyb3dzZXImMS4wLjA.Cu56AgSfBTDag5NiRA81oLHkDZfu5L3CKadnefEAY84",
		"content-type":  "application/x-www-form-urlencoded",
	}
	data := url.Values{"grant_type": {"refresh_token"},
		"refresh_token": {"eyJ6aXAiOiJERUYiLCJraWQiOiJLcTYtNW1Ia3BxOXdzLUtsSUUyaGJHYkRIZFduRjU3UjZHY1h6" +
			"aFlvZi04IiwiY3R5IjoiSldUIiwiZW5jIjoiQzIwUCIsImFsZyI6ImRpciJ9..OdwL8TEIFZouLDJe.wLz6zEC3PlPAGxx4" +
			"X4qyP837lUbFrI_DQGnrJDMtEaQd5gsjHwaYshscoDXCYjMioU8JvsH_HKZga3fzSDEoWuMA5lgv4dyJpoB4Cqi91JjPSkqs" +
			"RHKZ1I-nRoTmnSkcW3RHE-0coAqDWgK7IZ5cPiHQ-9KVRqqZkmTbEHynBdgH2y-FJP8zK0-dAynzR2krlUahhcykp7J7VqhZj" +
			"_l5HVZZkPylZ6eKoK4J8fQvuGJoqMaRZTzrIH4Yk9J3GMbKnYqEG3SKRp5qAuWTtqLDOoGN0wWsUE5VRuCZxRKpxayJWABq2u" +
			"4ABkAtIqUx8CPx77ZXxZVlcjRN1Xa8F2-e2mTxZq_1FgzmWECFg6onkDj_TpfBdeFoxDzhnRNceoQ-iyyNf3sgxJ_nz_bwztV" +
			"Zf0Vt3OR8yBnXfbkuEY7GQ4pvCuy-peW0mwJJCd2eJ9ADwDEGmoY4F47W-8rxdBhgna-0hu0FuLxt9MlmH_tGCmM_T-61xsxym" +
			"LO9tlkwBnxNw4u6T9X2hcvC7-4uzr5cJiaJ3sGPMNo_ixTrP8SG9zCIse-X6_Lq0v3Uo-QOKhcD4N3gIfwZFYEvf-HVGWzFpU6" +
			"83q9CJfTTEXhsufj1URhSis7GdAa3nLZVt7CScsMPcYrMI317PmU-Brdvl_Ic4QeHTeF8-57kzD3mm5mrlQ7kQIXQzzQPqHYt70" +
			"MzxL_scfT90cpYaSOBQnB1l--226h7X51XxSbrOcO-25zS7OSyedya8eMG6zAmgkk1zvZUzdCHZyzYD8-t0KYcfA5AwiLIFHxgq" +
			"L4ni9fVy-SpYTKRwCmkp_pZOPaFwJh8zkhw8QaSLHq7ubko7H1kjJZxzsG1l4Bla1QRlj_-FVoY8GZ6okFk3Ts6A2qOK6v8UT7s" +
			"L_w2zaHDQH1q2o05vsLwqIOxg3Xyey0tahzPbl-In_i1JGGvqGXOiPcKL5uOcTOo1luk32AbCS9i5mkopTS401YYYMH-Sx_krW_" +
			"VJd2czpFefc0dlagtzBytqlcyscscFwq6IE6VHwG2Ij-WfO44G5hGDJFkZMZLeDUnTIyNrLe9hcfJp73koOSFnURsFWFjM2lgUI" +
			"ayiREAl02oh2alUyqnG09gdXufT_2W0DjA4i7qYuv6ol5NIVc389dF3x4a_7dPBvsMU3ppA1rlV04FlK6_fRv-Dk_jclXRZiQ5u" +
			"l2ZO2CQ96LmrzmkdeNxFxcwaNXCJGBiRWXfMunoddIRg_LrVGuqWRgxj4DEnngZ2-qI_dliGiYraIehsHvtWeXIUWNF_FQSnQgZ" +
			"Lg4WPekcluCecE4Iv7Sz36k9GUDqqs8hRWddirhufYem6RC84PyNqafCnwczrx5pOacmVzDl9Oi8OIhdDasdJa7gvsDoFzf6bv5" +
			"st7EvbORkgPs6MK46mDMlwkL7TqjrJnSJzozCX4zLbYeyiWK6EXCehOpImMN262KLYQxnf5ugvk11gIA4NXpTbzyo4hp2LS7u8U" +
			"Ms5_w3t02vizxSQGokp-3qkEWmViy3pup1IXMPrcpS6KWHX0AYi1oRDZB5B8vM04pRHwYjsgMp2L-w4PMaDC4QDRU81IdvQ5VRk" +
			"yLT7CL5hDlq5smXw_7wSFTWxs9vc5PmnrykSAkwFPocORC2j4T96uiu3z4gNoBu_dwKNcPi-dV7myC4iRRTmpm0V5A9IW510RGT" +
			"yso_b-1hUeGvToYl9VwNgN7Impt3PjEQO2HXMU3p96tdulDEA_8bbyPdEGfxxVK3k2n_dxj_GzPKA8V4ESoNMRrV1vCuxPnrzfA" +
			"OhqmNOEewTHqlxENSsZFGvfzVj1KemR7zLky14JMVslILnvxl6vuX7SbfIQ5JDktq9qKtTKo1mFrA-mBS3n00FacjPi364nnugi" +
			"WQN7EwhNdEDH_KtWXGZVh-u2NM5cdoS1kAsOKSLxFTnTDG738LhoB3i_ZOjHFASKiZcsX6yD5csIP21jG5nFF9Qw2qsnqmxRuDL" +
			"ilIoGczEMt2Pfo180CG8Dyr7XtOYNeVU7__h9zBm9CvaAHDoQQhU4KlXM4LsljFeajw5f2wn08OmsdfkSYYl45O718QgzR_RRqw" +
			"DpQH2pyKDJZ9yZt5OCyxcbnCgepjUyp6S-Pigfw73ASoCknhLLheb2mqkWIC-s3NmClpMoK-IyE57AiHHCatZfPGPnNofVioN5S" +
			"bVR08mV7pdyQEhQGxGFM_LTAFFpwC48gOFTq-FWdV58muDULTqO3ImbGG6X3vV-PVbher1oJx0CFnelGGIx9lwM-yHbpVZGq9IX" +
			"nKqoblCHiwuaJgbCKBnTjia2gYPNlN0Ql1ia3vQc7bybDVHyLePAVbOk10MdwHprwMGE__wsXqagElQCGJpU3ytPDktncRPCSQB" +
			"Q3mw94CCIOQYEyhnA1Vik127AznwbR10Xm59diGBtix0Ao-VIrjKzQNw2hXqC_H-IgY46OT5ZndZ02SAe6AVyipq6kTui_ZyuQh" +
			"y-zAOiat4t6qh-LyL1xImBuOZ7e79737LYiLHEIgHOIQ68DKcSmsIuA.gwrRhM5AiYUQ6iAbRZhxlw"},
		"subject_token_type": {"urn:bamtech:params:oauth:token-type:device"}}
	resp, err := performRequest(requrl, ip, "POST", strings.NewReader(data.Encode()), headers)
	if err != nil || resp == nil {
		return "400"
	}
	defer resp.Body.Close()
	switch method {
	case "auth":
		s, _ := io.ReadAll(resp.Body)
		response := string(s)
		if strings.Contains(response, "unauthorized") {
			return "unauthorized"
		}
		return "ok"
	case "query":
		Header := resp.Header
		if cookie := Header.Get("Set-Cookie"); cookie != "" {
			StartIndex := strings.Index(cookie, "y=")
			EndIndex := strings.Index(cookie, ";")
			return cookie[StartIndex+2 : EndIndex]
		} else if location := Header.Get("Location"); location == "https://disneyplus.disney.co.jp/" {
			return "JP"
		} else if location == "https://preview.disneyplus.com/unavailable/" {
			return "Unavailable"
		}
	}
	return "-1"
}

// RequestYoutubeIP 请求Youtube的IP
func RequestYoutubeIP(requrl, ip string) (int, string, string, string) {
	urlValue, err := url.Parse(requrl)
	if err != nil {
		return 400, "", "", ""
	}
	host := urlValue.Host
	if ip == "" {
		ip = host
	}
	resp, err := performRequest(requrl, ip, "GET", nil, nil)
	if err != nil || resp == nil {
		return 400, "", "", ""
	}
	defer resp.Body.Close()
	s, _ := io.ReadAll(resp.Body)
	response := string(s)
	EndLocation := strings.Index(response, "\n")
	response = response[1:EndLocation]
	EndLocation = strings.Index(response, "=> ")
	response = response[EndLocation+3:]
	EndLocation = strings.Index(response, " ")
	response = response[:EndLocation]
	EndLocation = strings.Index(response, "-")

	if EndLocation == -1 {
		method := "Youtube Video Server"
		EndLocation = strings.Index(response, ".")
		airCode := response[EndLocation+1:]
		return 200, method, "", airCode
	}
	method := "Google Global CacheCDN (ISP Cooperation)"
	isp := response[:EndLocation]
	airCode := response[EndLocation+1:]
	return 200, method, isp, airCode
}

// RequestYoutubeIPRegion 请求Youtube的IP区域
func RequestYoutubeIPRegion(requrl, ip string) string {
	urlValue, err := url.Parse(requrl)
	if err != nil {
		return "error"
	}
	headers := map[string]string{"Host": urlValue.Host}
	resp, err := performRequest(requrl, ip, "GET", nil, headers)
	if err != nil || resp == nil {
		return "error"
	}
	defer resp.Body.Close()
	s, _ := io.ReadAll(resp.Body)
	response := string(s)
	EndLocation := strings.Index(response, "\"countryCode\"")
	if EndLocation != -1 {
		return response[EndLocation+15 : EndLocation+17]
	}
	return "null"
}

// RequestNetflixIP 请求Netflix的IP
func RequestNetflixIP(requrl, ip string) (string, error) {
	if ip == "" {
		return "", fmt.Errorf("IP is empty")
	}
	urlValue, err := url.Parse(requrl)
	if err != nil {
		return "", fmt.Errorf("URL parse error")
	}
	headers := map[string]string{"Host": urlValue.Host}
	resp, err := performRequest(requrl, ip, "GET", nil, headers)
	if err != nil || resp == nil {
		return "", err
	}
	defer resp.Body.Close()
	Header := resp.Header
	if Header.Get("X-Robots-Tag") == "index" {
		return "us", nil
	}
	location := Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("Banned")
	}
	return strings.Split(location, "/")[3], nil
}
