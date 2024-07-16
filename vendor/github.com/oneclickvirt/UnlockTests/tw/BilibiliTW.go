package tw

import (
	"github.com/oneclickvirt/UnlockTests/asia"
	"github.com/oneclickvirt/UnlockTests/model"
	"net/http"
)

func BilibiliTW(c *http.Client) model.Result {
	return asia.Bilibili(c, "Bilibili Taiwan Only", "https://api.bilibili.com/pgc/player/web/playurl?avid=50762638&cid=100279344&qn=0&type=&otype=json&ep_id=268176&fourk=1&fnver=0&fnval=16&module=bangumi")
}
