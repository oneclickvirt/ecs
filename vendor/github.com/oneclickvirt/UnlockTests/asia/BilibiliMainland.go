package asia

import (
	"github.com/oneclickvirt/UnlockTests/model"
	"net/http"
)

// BilibiliMainland
// 检测大陆B站是否可用
func BilibiliMainland(c *http.Client) model.Result {
	return Bilibili(c, "BilibiliMainland", "https://api.bilibili.com/pgc/player/web/playurl?avid=82846771&qn=0&type=&otype=json&ep_id=307247&fourk=1&fnver=0&fnval=16&session=$r_session&module=bangumi")
}
