package eu

// HBOGO
// api.ugw.hbogo.eu 已经 host 为空了 查询不到内容
//func HBOGO(request *gorequest.SuperAgent) model.Result {
//	name := "HBO GO Europe"
//	url := "https://api.ugw.hbogo.eu/v3.0/GeoCheck/json/HUN"
//	request = request.Set("User-Agent", model.UA_Browser)
//	resp, body, errs := request.Get(url).Retry(2, 5).End()
//	if len(errs) > 0 {
//		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: errs[0]}
//	}
//	defer resp.Body.Close()
//	var hboRes struct {
//		Country   string `json:"country"`
//		Territory string `json:"territory"`
//	}
//	fmt.Println(body)
//	if err := json.Unmarshal([]byte(body), &hboRes); err != nil {
//		return model.Result{Name: name, Status: model.StatusErr, Err: err}
//	}
//	if hboRes.Territory == "" {
//		// 解析不到为空则识别为不解锁
//		return model.Result{Name: name, Status: model.StatusNo}
//	}
//	return model.Result{Name: name, Status: model.StatusYes, Region: strings.ToLower(hboRes.Country)}
//}
