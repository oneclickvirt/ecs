package fr

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
)

// Salto
// geo.salto.fr 双栈 get 请求 有问题
// tls验证失败，识别失效，未知原因
func Salto(c *http.Client) model.Result {
	name := "Salto"
	url := "https://www.salto.fr/"
	client := utils.Req(c)
	resp, err := client.R().Get(url)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err == nil {
		fmt.Println(string(b), resp.StatusCode)
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get geo.salto.fr failed with code: %d", resp.StatusCode)}
}
