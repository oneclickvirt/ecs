package speedtest

import (
	"fmt"
	"github.com/showwin/speedtest-go/speedtest"
	"github.com/showwin/speedtest-go/speedtest/transport"
	"log"
	"strings"
	"time"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func NearbySpeedTest() {
	fmt.Println("位置\t\t 上传速度\t 下载速度\t 延迟\t  丢包率")
	var speedtestClient = speedtest.New()
	serverList, _ := speedtestClient.FetchServers()
	targets, _ := serverList.FindServer([]int{})
	analyzer := speedtest.NewPacketLossAnalyzer(nil)
	var LowestLatency time.Duration
	var NearbyServer *speedtest.Server
	var PacketLoss string
	for _, server := range targets {
		server.PingTest(nil)
		if LowestLatency == 0 && NearbyServer == nil {
			LowestLatency = server.Latency
			NearbyServer = server
		} else if server.Latency < LowestLatency && NearbyServer != nil {
			NearbyServer = server
		}
		server.Context.Reset()
	}
	if NearbyServer != nil {
		NearbyServer.DownloadTest()
		NearbyServer.UploadTest()
		err := analyzer.Run(NearbyServer.Host, func(packetLoss *transport.PLoss) {
			PacketLoss = strings.ReplaceAll(packetLoss.String(), "Packet Loss: ", "")
		})
		checkError(err)
		fmt.Printf("%s\t\t %s\t %s\t %s\t  %s\n",
			NearbyServer.Name, NearbyServer.ULSpeed, NearbyServer.DLSpeed, NearbyServer.Latency, PacketLoss)
		NearbyServer.Context.Reset()
	}
	// https://github.com/spiritLHLS/speedtest.net-CN-ID
	// https://github.com/spiritLHLS/speedtest.cn-CN-ID
}
