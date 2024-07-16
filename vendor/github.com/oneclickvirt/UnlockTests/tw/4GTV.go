package tw

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"net/url"
)

// Tw4gtv
// api2.4gtv.tv 仅 ipv4 且 post 请求
func Tw4gtv(c *http.Client) model.Result {
	name := "4GTV.TV"
	hostname := "4gtv.tv"
	if c == nil {
		return model.Result{Name: name}
	}
	url1 := "https://api2.4gtv.tv//Vod/GetVodUrl3"
	data := url.Values{
		"value": {"D33jXJ0JVFkBqV%2BZSi1mhPltbejAbPYbDnyI9hmfqjKaQwRQdj7ZKZRAdb16%2FRUrE8vGXLFfNKBLKJv%2BfDSiD%2BZJlUa5Msps2P4IWuTrUP1%2BCnS255YfRadf%2BKLUhIPj"},
	}
	client := utils.Req(c)
	resp, err := client.R().SetFormDataFromValues(data).Post(url1)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: fmt.Errorf("can not parse body")}
	}
	//body := string(b)
	//fmt.Println(body)
	var res struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res.Success {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	} else if res.Success == false || resp.StatusCode == 403 || resp.StatusCode == 404 {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get api2.4gtv.tv failed with code: %d", resp.StatusCode)}
}
