package asia

import (
	"github.com/oneclickvirt/UnlockTests/model"
	"net/http"
)

// BilibiliSEA
// 检测东南亚B站是否可用
func BilibiliSEA(c *http.Client) model.Result {
	return Bilibili(c, "BilibiliSEA", "https://api.bilibili.tv/intl/gateway/web/playurl?s_locale=en_US&platform=web&ep_id=347666")
}
