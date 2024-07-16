package transnation

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// WikipediaEditable
// zh.wikipedia.org 双栈 get 请求
func WikipediaEditable(c *http.Client) model.Result {
	name := "Wikipedia Editability"
	if c == nil {
		return model.Result{Name: name}
	}
	url := "https://zh.wikipedia.org/w/index.php?title=Wikipedia%3A%E6%B2%99%E7%9B%92&action=edit"
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
	if strings.Contains(body, "Banned") {
		return model.Result{Name: name, Status: model.StatusNo}
	} else if resp.StatusCode == 200 {
		return model.Result{Name: name, Status: model.StatusYes}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get zh.wikipedia.org failed with code: %d", resp.StatusCode)}
}
