package th

import (
	"github.com/oneclickvirt/UnlockTests/asia"
	"github.com/oneclickvirt/UnlockTests/model"
	"net/http"
)

// BilibiliTH
// 检测泰国B站是否可用
func BilibiliTH(c *http.Client) model.Result {
	return asia.Bilibili(c, "BilibiliTH", "https://api.bilibili.tv/intl/gateway/web/playurl?s_locale=en_US&platform=web&ep_id=10077726")
}
