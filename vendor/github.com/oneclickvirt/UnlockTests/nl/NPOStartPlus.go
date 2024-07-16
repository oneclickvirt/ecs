package nl

import (
	"encoding/json"
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"io"
	"net/http"
	"strings"
)

// NPOStartPlus
// www.npo.nl 双栈 且 get 请求
func NPOStartPlus(c *http.Client) model.Result {
	name := "NPO Start Plus"
	hostname := "npo.nl"
	if c == nil {
		return model.Result{Name: name}
	}
	tokenURL := "https://www.npo.nl/start/api/domain/player-token?productId=LI_NL1_4188102"
	streamURL := "https://prod.npoplayer.nl/stream-link"
	referrerURL := "https://npo.nl/start/live?channel=NPO1"
	client := utils.Req(c)
	resp, err := client.R().Get(tokenURL)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	// body := string(b)
	// fmt.Println(body)
	var res1 struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(b, &res1); err != nil {
		return model.Result{Name: name, Status: model.StatusErr, Err: err}
	}
	if res1.Token != "" {
		headers := map[string]string{
			"Origin":        "https://npo.nl",
			"Referer":       "https://npo.nl/",
			"Content-Type":  "application/json",
			"Authorization": res1.Token,
		}
		client1 := utils.Req(c)
		client1 = utils.SetReqHeaders(client1, headers)
		resp2, err2 := client1.R().
			SetBodyString(`{"profileName":"dash","drmType":"playready","referrerUrl":"` + referrerURL + `"}`).
			Post(streamURL)
		if err2 != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
		}
		defer resp2.Body.Close()
		b, err = io.ReadAll(resp2.Body)
		if err != nil {
			return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
		}
		body := string(b)
		// fmt.Println(body)
		// fmt.Println(resp2.StatusCode)
		// {"status":451,"body":"Dit programma mag niet bekeken worden vanaf jouw locatie."}
		if resp2.StatusCode == 451 || strings.Contains(body, "Dit programma mag niet bekeken worden vanaf jouw locatie.") {
			return model.Result{Name: name, Status: model.StatusNo}
		} else if resp2.StatusCode == 200 {
			result1, result2, result3 := utils.CheckDNS(hostname)
			unlockType := utils.GetUnlockType(result1, result2, result3)
			return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
		} else {
			return model.Result{Name: name, Status: model.StatusNo}
		}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected, Err: fmt.Errorf("Token get null")}
}
