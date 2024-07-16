package es

// HBOGO
// api-discovery.hbo.eu 的 host 已经为空了
// func HBOGO(request *gorequest.SuperAgent) model.Result {
// 	name := "HBO Spain"
// 	url := "https://api-discovery.hbo.eu/v1/discover/hbo?language=null&product=hboe"
// 	request = request.Set("User-Agent", model.UA_Browser).Set("X-Client-Name", "web")
// 	resp, body, errs := request.Get(url).Retry(2, 5).End()
// 	if len(errs) > 0 {
// 		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: errs[0]}
// 	}
// 	defer resp.Body.Close()
// 	var hboRes struct {
// 		Country   string `json:"country"`
// 		Territory string `json:"territory"`
// 	}
// 	fmt.Println(body)
// 	if err := json.Unmarshal([]byte(body), &hboRes); err != nil {
// 		return model.Result{Name: name, Status: model.StatusErr, Err: err}
// 	}
// 	if hboRes.Territory == "" {
// 		// 解析不到为空则识别为不解锁
// 		return model.Result{Name: name, Status: model.StatusNo}
// 	}
// 	return model.Result{Name: name, Status: model.StatusYes, Region: strings.ToLower(hboRes.Country)}
// }
