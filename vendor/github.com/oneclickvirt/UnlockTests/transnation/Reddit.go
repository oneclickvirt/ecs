package transnation

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// Reddit
// www.reddit.com 仅 ipv4 且 get 请求
func Reddit(c *http.Client) model.Result {
	name := "Reddit"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://www.reddit.com/"
	client := utils.Req(c)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	body := string(b)
	if resp.StatusCode == 200 || resp.StatusCode == 302 {
		return model.Result{Name: name, Status: model.StatusYes}
	}
	// fmt.Println(body)
	// tempList := strings.Split(body, "\n")
	// for _, l := range tempList {
	// 	if strings.Contains(l, "blocked") {
	// 		fmt.Println(l)
	// 	}
	// }
	if resp.StatusCode == 403 && strings.Contains(body, "been blocked") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.reddit.com failed with code: %d", resp.StatusCode)}
}
