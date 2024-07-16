package sp

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/speedtest/model"
	"github.com/showwin/speedtest-go/speedtest"
	"github.com/showwin/speedtest-go/speedtest/transport"
)

func OfficialAvailableTest() error {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	spvCheck := exec.Command("speedtest", "--version")
	output, err := spvCheck.CombinedOutput()
	if err != nil {
		return err
	} else {
		version := strings.Split(string(output), "\n")[0]
		if strings.Contains(version, "Speedtest by Ookla") && !strings.Contains(version, "err") {
			// 此时确认可使用speedtest命令进行测速
			return nil
		}
	}
	return fmt.Errorf("No match speedtest command")
}

func OfficialNearbySpeedTest() {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	var UPStr, DLStr, Latency, PacketLoss string // serverID,
	// speedtest --progress=no --accept-license --accept-gdpr
	sptCheck := exec.Command("speedtest", "--progress=no", "--accept-license", "--accept-gdpr")
	temp, err := sptCheck.CombinedOutput()
	if err == nil {
		tempList := strings.Split(string(temp), "\n")
		for _, line := range tempList {
			if strings.Contains(line, "Idle Latency") {
				Latency = strings.TrimSpace(strings.Split(strings.Split(line, ":")[1], "(")[0])
			} else if strings.Contains(line, "Download") {
				DLStr = strings.TrimSpace(strings.Split(strings.Split(line, ":")[1], "(")[0])
			} else if strings.Contains(line, "Upload") {
				UPStr = strings.TrimSpace(strings.Split(strings.Split(line, ":")[1], "(")[0])
			} else if strings.Contains(line, "Packet Loss") {
				PacketLoss = strings.TrimSpace(strings.Split(line, ":")[1])
			}
		}
		if Latency != "" && DLStr != "" && UPStr != "" && PacketLoss != "" {
			fmt.Print(formatString("Speedtest.net", 16))
			fmt.Print(formatString(UPStr, 16))
			fmt.Print(formatString(DLStr, 16))
			fmt.Print(formatString(Latency, 16))
			fmt.Print(formatString(PacketLoss, 16))
			fmt.Println()
		}
	}
}

func OfficialCustomSpeedTest(url, byWhat string, num int, language string) {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	if !strings.Contains(url, ".net") {
		fmt.Println("Official speedtest only use .net platform, can not use other platforms.")
		return
	}
	data := getData(url)
	var targets speedtest.Servers
	if byWhat == "id" {
		targets = parseDataFromID(data, url)
	} else if byWhat == "url" {
		targets = parseDataFromURL(data, url)
	}
	var pingList []time.Duration
	var err error
	serverMap := make(map[time.Duration]*speedtest.Server)
	for _, server := range targets {
		err = server.PingTest(nil)
		if err != nil {
			server.Latency = 1000 * time.Millisecond
			if model.EnableLoger {
				Logger.Info(err.Error())
			}
		}
		pingList = append(pingList, server.Latency)
		serverMap[server.Latency] = server
	}
	sort.Slice(pingList, func(i, j int) bool {
		return pingList[i] < pingList[j]
	})
	if num == -1 || num >= len(pingList) {
		num = len(pingList)
	} else if len(pingList) == 0 {
		fmt.Println("No match servers")
		if model.EnableLoger {
			Logger.Info("No match servers")
		}
		return
	}
	var serverName, UPStr, DLStr, Latency, PacketLoss string
	for i := 0; i < len(pingList); i++ {
		server := serverMap[pingList[i]]
		if i < num {
			// speedtest --progress=no --accept-license --accept-gdpr
			sptCheck := exec.Command("speedtest", "--progress=no", "--server-id="+server.ID, "--accept-license", "--accept-gdpr")
			temp, err := sptCheck.CombinedOutput()
			if err == nil {
				serverName = server.Name
				tempList := strings.Split(string(temp), "\n")
				for _, line := range tempList {
					if strings.Contains(line, "Idle Latency") {
						Latency = strings.TrimSpace(strings.Split(strings.Split(line, ":")[1], "(")[0])
					} else if strings.Contains(line, "Download") {
						DLStr = strings.TrimSpace(strings.Split(strings.Split(line, ":")[1], "(")[0])
					} else if strings.Contains(line, "Upload") {
						UPStr = strings.TrimSpace(strings.Split(strings.Split(line, ":")[1], "(")[0])
					} else if strings.Contains(line, "Packet Loss") {
						PacketLoss = strings.TrimSpace(strings.Split(line, ":")[1])
					}
				}
				if Latency != "" && DLStr != "" && UPStr != "" && PacketLoss != "" {
					if language == "zh" {
						fmt.Print(formatString(serverName, 16))
					} else if language == "en" {
						name := serverName
						name = strings.ReplaceAll(name, "中国香港", "HongKong")
						name = strings.ReplaceAll(name, "洛杉矶", "LosAngeles")
						name = strings.ReplaceAll(name, "日本东京", "Tokyo,Japan")
						name = strings.ReplaceAll(name, "新加坡", "Singapore")
						name = strings.ReplaceAll(name, "法兰克福", "Frankfurt")
						fmt.Print(formatString(name, 16))
					}
					fmt.Print(formatString(UPStr, 16))
					fmt.Print(formatString(DLStr, 16))
					fmt.Print(formatString(Latency, 16))
					fmt.Print(formatString(PacketLoss, 16))
					fmt.Println()
				}
			}
		}
	}
}

func NearbySpeedTest() {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
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
		if err == nil {
			fmt.Print(formatString("Speedtest.net", 16))
			fmt.Print(formatString(fmt.Sprintf("%-8s", fmt.Sprintf("%.2f", NearbyServer.ULSpeed.Mbps())+" Mbps"), 16))
			fmt.Print(formatString(fmt.Sprintf("%-8s", fmt.Sprintf("%.2f", NearbyServer.DLSpeed.Mbps())+" Mbps"), 16))
			fmt.Print(formatString(fmt.Sprintf("%s", NearbyServer.Latency), 16))
			fmt.Print(formatString(PacketLoss, 16))
			fmt.Println()
			NearbyServer.Context.Reset()
		} else if model.EnableLoger {
			Logger.Info(err.Error())
		}
	}
}

func CustomSpeedTest(url, byWhat string, num int, language string) {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	data := getData(url)
	var targets speedtest.Servers
	if byWhat == "id" {
		targets = parseDataFromID(data, url)
	} else if byWhat == "url" {
		targets = parseDataFromURL(data, url)
	}
	var pingList []time.Duration
	var err, err1, err2, err3 error
	serverMap := make(map[time.Duration]*speedtest.Server)
	for _, server := range targets {
		err = server.PingTest(nil)
		if err != nil {
			server.Latency = 1000 * time.Millisecond
			if model.EnableLoger {
				Logger.Info(err.Error())
			}
		}
		pingList = append(pingList, server.Latency)
		serverMap[server.Latency] = server
	}
	sort.Slice(pingList, func(i, j int) bool {
		return pingList[i] < pingList[j]
	})
	analyzer := speedtest.NewPacketLossAnalyzer(nil)
	var PacketLoss string
	if num == -1 || num >= len(pingList) {
		num = len(pingList)
	} else if len(pingList) == 0 {
		fmt.Println("No match servers")
		if model.EnableLoger {
			Logger.Info("No match servers")
		}
		return
	}
	for i := 0; i < len(pingList); i++ {
		server := serverMap[pingList[i]]
		if i < num {
			err1 = server.DownloadTest()
			err2 = server.UploadTest()
			err3 = analyzer.Run(server.Host, func(packetLoss *transport.PLoss) {
				PacketLoss = strings.ReplaceAll(packetLoss.String(), "Packet Loss: ", "")
			})
			if err3 != nil {
				if model.EnableLoger {
					Logger.Info(server.ID)
					Logger.Info(err3.Error())
				}
				PacketLoss = "N/A"
			}
			if err1 != nil {
				if model.EnableLoger {
					Logger.Info(server.ID)
					Logger.Info(err1.Error())
				}
				server.Context.Reset()
				continue
			}
			if err2 != nil {
				if model.EnableLoger {
					Logger.Info(server.ID)
					Logger.Info(err2.Error())
				}
				server.Context.Reset()
				continue
			}
			if language == "zh" {
				fmt.Print(formatString(server.Name, 16))
			} else if language == "en" {
				name := server.Name
				name = strings.ReplaceAll(name, "中国香港", "HongKong")
				name = strings.ReplaceAll(name, "洛杉矶", "LosAngeles")
				name = strings.ReplaceAll(name, "日本东京", "Tokyo,Japan")
				name = strings.ReplaceAll(name, "新加坡", "Singapore")
				name = strings.ReplaceAll(name, "法兰克福", "Frankfurt")
				fmt.Print(formatString(name, 16))
			}
			fmt.Print(formatString(fmt.Sprintf("%-8s", fmt.Sprintf("%.2f", server.ULSpeed.Mbps())+" Mbps"), 16))
			fmt.Print(formatString(fmt.Sprintf("%-8s", fmt.Sprintf("%.2f", server.DLSpeed.Mbps())+" Mbps"), 16))
			fmt.Print(formatString(fmt.Sprintf("%s", server.Latency), 16))
			fmt.Print(formatString(PacketLoss, 16))
			fmt.Println()
		}
		server.Context.Reset()
	}
}
