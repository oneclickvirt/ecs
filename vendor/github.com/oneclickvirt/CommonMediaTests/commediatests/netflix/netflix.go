package netflix

import (
	"github.com/oneclickvirt/CommonMediaTests/commediatests/netflix/printer"
	"github.com/oneclickvirt/CommonMediaTests/commediatests/netflix/verify"
)

// var cmtnFlag = flag.NewFlagSet("cmtn", flag.ContinueOnError)
// var custom = cmtnFlag.String("custom", "", "自定义测试NF影片ID\n绝命毒师的ID是70143836")
// var address = cmtnFlag.String("address", "", "本机网卡的IP")
// var proxy = cmtnFlag.String("proxy", "", "代理地址")

func Netflix(language string) (string, error) {
	// cmtnFlag.Parse(os.Args[1:])
	// r := verify.NewVerify(verify.Config{
	// 	LocalAddr: *address,
	// 	Custom:    *custom,
	// 	Proxy:     *proxy,
	// })
	r := verify.NewVerify(verify.Config{
		LocalAddr: "",
		Custom:    "",
		Proxy:     "",
	})
	res, _ := printer.Print(*r, language)
	return res, nil
}
