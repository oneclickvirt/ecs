package model

import "time"

const PingTestVersion = "v0.0.4"

var EnableLoger = false

var (
	NetCMCC = "https://raw.githubusercontent.com/spiritLHLS/speedtest.net-CN-ID/main/CN_Mobile.csv"
	NetCT   = "https://raw.githubusercontent.com/spiritLHLS/speedtest.net-CN-ID/main/CN_Telecom.csv"
	NetCU   = "https://raw.githubusercontent.com/spiritLHLS/speedtest.net-CN-ID/main/CN_Unicom.csv"
	CnCMCC  = "https://raw.githubusercontent.com/spiritLHLS/speedtest.cn-CN-ID/main/mobile.csv"
	CnCT    = "https://raw.githubusercontent.com/spiritLHLS/speedtest.cn-CN-ID/main/telecom.csv"
	CnCU    = "https://raw.githubusercontent.com/spiritLHLS/speedtest.cn-CN-ID/main/unicom.csv"
	CdnList = []string{
		"http://cdn0.spiritlhl.top/",
		"http://cdn1.spiritlhl.top/",
		"http://cdn1.spiritlhl.net/",
		"http://cdn3.spiritlhl.net/",
		"http://cdn2.spiritlhl.net/",
	}
)

type Server struct {
	Name string
	IP   string
	Port string
	Avg  time.Duration
}
