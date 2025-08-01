package upstreams

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/UnlockTests/uts"
	bgptools "github.com/oneclickvirt/backtrace/bgptools"
	backtrace "github.com/oneclickvirt/backtrace/bk"
)

type IpInfo struct {
	Ip      string `json:"ip"`
	City    string `json:"city,omitempty"`
	Region  string `json:"region,omitempty"`
	Country string `json:"country,omitempty"`
	Org     string `json:"org,omitempty"`
}

func fetchIP(ctx context.Context, url string, parse func([]byte) (string, error), ch chan<- string) {
	client := req.C().SetTimeout(3 * time.Second)
	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil || !resp.IsSuccessState() {
		return
	}
	ip, err := parse(resp.Bytes())
	if err == nil && ip != "" && strings.Contains(ip, ".") {
		ch <- ip
	}
}

func UpstreamsCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ipChan := make(chan string, 3)
	go fetchIP(ctx, "https://ipinfo.io", func(b []byte) (string, error) {
		var data IpInfo
		err := json.Unmarshal(b, &data)
		return data.Ip, err
	}, ipChan)
	go fetchIP(ctx, "https://api.ip.sb/ip", func(b []byte) (string, error) {
		return strings.TrimSpace(string(b)), nil
	}, ipChan)
	go fetchIP(ctx, "http://ip-api.com/json/?fields=query", func(b []byte) (string, error) {
		var data struct {
			Query string `json:"query"`
		}
		err := json.Unmarshal(b, &data)
		return data.Query, err
	}, ipChan)
	var ip string
	select {
	case ip = <-ipChan:
	case <-ctx.Done():
	}
	if ip != "" {
		if result, err := bgptools.GetPoPInfo(ip); err == nil {
			fmt.Print(result.Result)
		}
	}
	backtrace.BackTrace(uts.IPV6)
}
