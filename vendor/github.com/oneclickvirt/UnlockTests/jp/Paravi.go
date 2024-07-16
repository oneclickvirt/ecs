package jp

// Paravi
// api.paravi.jp 仅 ipv4 且 post 请求
//func Paravi(request *gorequest.SuperAgent) model.Result {
//	name := "Paravi"
//if request == nil {
//return model.Result{Name: name}
//}
//	resp, bodyBytes, errs := utils.PostJson(request, "https://api.paravi.jp/api/v1/playback/auth",
//		`{"meta_id":17414,"vuid":"3b64a775a4e38d90cc43ea4c7214702b","device_code":1,"app_id":1}`,
// nil,
//	)
//	if len(errs) > 0 {
//		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: errs[0]}
//	}
//	defer resp.Body.Close()
//	if resp.StatusCode == 403 {
//		return model.Result{Name: name, Status: model.StatusNo}
//	}
//	var res struct {
//		Error struct {
//			Type string `json:"type"`
//		} `json:"error"`
//	}
//	if err := json.Unmarshal(bodyBytes, &res); err != nil {
//		if strings.Contains(string(bodyBytes), "Forbidden") {
//			return model.Result{Name: name, Status: model.StatusNo}
//		}
//		if strings.Contains(string(bodyBytes), "Unauthorized") {
//			return model.Result{Name: name, Status: model.StatusYes}
//		}
//		return model.Result{Name: name, Status: model.StatusErr, Err: err}
//	}
//	if res.Error.Type == "Unauthorized" {
//		return model.Result{Name: name, Status: model.StatusYes}
//	}
//	if res.Error.Type == "Forbidden" {
//		return model.Result{Name: name, Status: model.StatusNo}
//	}
//	return model.Result{Name: name, Status: model.StatusUnexpected,
//		Err: fmt.Errorf("get api.paravi.jp failed with code: %d", resp.StatusCode)}
//}
