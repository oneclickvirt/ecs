package model

import (
	"time"

	fastTrace "github.com/nxtrace/NTrace-core/fast_trace"
)

type ParamsFastTrace struct {
	SrcDev         string
	SrcAddr        string
	BeginHop       int
	MaxHops        int
	RDns           bool
	AlwaysWaitRDNS bool
	Lang           string
	PktSize        int
	Timeout        time.Duration
	File           string
	DontFragment   bool
}

var NextTraceVersion = "v0.0.3"

var EnableLoger = false

var (
	GuangZhouCT = fastTrace.ISPCollection{
		ISPName: "广州电信",
		IP:      "58.60.188.222",
		IPv6:    "240e:97c:2f:3000::44",
	}
	GuangZhouCU = fastTrace.ISPCollection{
		ISPName: "广州联通",
		IP:      "210.21.196.6",
		IPv6:    "2408:8756:f50:1001::c",
	}
	GuangZhouCMCC = fastTrace.ISPCollection{
		ISPName: "广州移动",
		IP:      "120.196.165.24",
		IPv6:    "2409:8c54:871:1001::12",
	}
	ShangHaiCT = fastTrace.ISPCollection{
		ISPName: "上海电信",
		IP:      "202.96.209.133",
		IPv6:    "240e:e1:aa00:4000::24",
	}
	ShangHaiCU = fastTrace.ISPCollection{
		ISPName: "上海联通",
		IP:      "210.22.97.1",
		IPv6:    "2408:80f1:21:5003::a",
	}
	ShangHaiCMCC = fastTrace.ISPCollection{
		ISPName: "上海移动",
		IP:      "211.136.112.200",
		IPv6:    "2409:8c1e:75b0:3003::26",
	}
	BeiJingCT = fastTrace.ISPCollection{
		ISPName: "北京电信",
		IP:      "219.141.140.10",
		IPv6:    "2400:89c0:1053:3::69",
	}
	BeiJingCU = fastTrace.ISPCollection{
		ISPName: "北京联通",
		IP:      "202.106.195.68",
		IPv6:    "2400:89c0:1013:3::54",
	}
	BeiJingCMCC = fastTrace.ISPCollection{
		ISPName: "北京移动",
		IP:      "221.179.155.161",
		IPv6:    "2409:8c00:8421:1303::55",
	}
	ChengDuCT = fastTrace.ISPCollection{
		ISPName: "成都电信",
		IP:      "61.139.2.69",
		IPv6:    "",
	}
	ChengDuCU = fastTrace.ISPCollection{
		ISPName: "成都联通",
		IP:      "119.6.6.6",
		IPv6:    "",
	}
	ChengDuCMCC = fastTrace.ISPCollection{
		ISPName: "成都移动",
		IP:      "211.137.96.205",
		IPv6:    "",
	}
)
