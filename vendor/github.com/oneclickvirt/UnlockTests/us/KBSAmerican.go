package us

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// KBSAmerican
// vod.kbs.co.kr 仅 ipv4 且 get 请求
func KBSAmerican(c *http.Client) model.Result {
	name := "KBS American"
	hostname := "kbs.co.kr"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://vod.kbs.co.kr/index.html?source=episode&sname=vod&stype=vod&program_code=T2022-0690&program_id=PS-2022164275-01-000&broadcast_complete_yn=N&local_station_code=00&section_code=03"
	client := utils.Req(c)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	body := string(b)
	//fmt.Println(body)
	tempList := strings.Split(body, "\n")
	for _, line := range tempList {
		if strings.Contains(line, "ipck") && strings.Contains(line, "Domestic") {
			tpList := strings.Split(line, "Domestic")
			if strings.Contains(strings.Split(tpList[1], "\"")[1], "false") {
				return model.Result{Name: name, Status: model.StatusNo}
			} else if strings.Contains(strings.Split(tpList[1], "\"")[1], "true") {
				result1, result2, result3 := utils.CheckDNS(hostname)
				unlockType := utils.GetUnlockType(result1, result2, result3)
				return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
			}
		}
	}
	if strings.Contains(body, "해당 영상은 저작권 등의 문제로") && strings.Contains(body, "서비스가 제공되지 않습니다") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get vod.kbs.co.kr failed with code: %d", resp.StatusCode)}
}
