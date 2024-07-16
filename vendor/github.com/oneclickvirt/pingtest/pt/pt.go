package pt

import (
	"fmt"
	"net"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/pingtest/model"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	ICMPProtocolICMP = 1
	pingCount        = 3
	timeout          = 3 * time.Second
)

func pingServerByGolang(server *model.Server, wg *sync.WaitGroup) {

	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	defer wg.Done()
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		logError("cannot listen for ICMP packets: " + err.Error())
		return
	}
	defer conn.Close()

	target := server.IP
	dst, err := net.ResolveIPAddr("ip4", target)
	if err != nil {
		logError("cannot resolve IP address: " + err.Error())
		return
	}

	var totalRtt time.Duration

	for i := 0; i < pingCount; i++ {
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   i,
				Seq:  i,
				Data: []byte("ping"),
			},
		}

		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			logError("cannot marshal ICMP message: " + err.Error())
			return
		}

		start := time.Now()
		_, err = conn.WriteTo(msgBytes, dst)
		if err != nil {
			logError("cannot send ICMP message: " + err.Error())
			return
		}

		reply := make([]byte, 1500)
		conn.SetReadDeadline(time.Now().Add(timeout))
		n, _, err := conn.ReadFrom(reply)
		if err != nil {
			logError("cannot receive ICMP reply: " + err.Error())
			return
		}

		duration := time.Since(start)
		rm, err := icmp.ParseMessage(ICMPProtocolICMP, reply[:n])
		if err != nil {
			logError("cannot parse ICMP reply: " + err.Error())
			return
		}

		if rm.Type == ipv4.ICMPTypeEchoReply {
			totalRtt += duration
		}
	}

	server.Avg = totalRtt / time.Duration(pingCount)
}

func pingServer(server *model.Server, wg *sync.WaitGroup) {
	// cmd := exec.Command("sudo", "ping", "-h")
	// output, err := cmd.CombinedOutput()
	// if err != nil || (!strings.Contains(string(output), "Usage") && strings.Contains(string(output), "err")) {
	// 	pingServerByGolang(server, wg)
	// } else {
	// 	pingServerByCMD(server, wg)
	// }
	pingServerByCMD(server, wg)
}

func pingServerByCMD(server *model.Server, wg *sync.WaitGroup) {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	defer wg.Done()
	// 执行 ping 命令
	cmd := exec.Command("sudo", "ping", "-c1", "-W3", server.IP)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if model.EnableLoger {
			Logger.Info("cannot ping: " + err.Error())
		}
		return
	}
	if model.EnableLoger {
		Logger.Info(string(output))
	}
	// 解析输出结果
	if !strings.Contains(string(output), "time=") {
		if model.EnableLoger {
			Logger.Info("ping failed without time=")
		}
		return
	}
	var avgTime float64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "time=") {
			matches := strings.Split(line, "time=")
			if len(matches) >= 2 {
				avgTime, err = strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(matches[1], "ms", "")), 64)
				if err != nil {
					if model.EnableLoger {
						Logger.Info("cannot parse avgTime: " + err.Error())
					}
					return
				}
				break
			}
		}
	}
	server.Avg = time.Duration(avgTime * float64(time.Millisecond))
}

func PingTest() string {
	var result string
	servers1 := getServers("cu")
	servers2 := getServers("ct")
	servers3 := getServers("cmcc")

	process := func(servers []*model.Server) []*model.Server {
		var wg sync.WaitGroup
		for i := range servers {
			wg.Add(1)
			go pingServer(servers[i], &wg)
		}
		wg.Wait()
		sort.Slice(servers, func(i, j int) bool {
			return servers[i].Avg < servers[j].Avg
		})
		return servers
	}

	var allServers []*model.Server
	var wga sync.WaitGroup
	wga.Add(3)
	go func() {
		defer wga.Done()
		servers1 = process(servers1)
	}()
	go func() {
		defer wga.Done()
		servers2 = process(servers2)
	}()
	go func() {
		defer wga.Done()
		servers3 = process(servers3)
	}()
	wga.Wait()

	allServers = append(allServers, servers1...)
	allServers = append(allServers, servers2...)
	allServers = append(allServers, servers3...)

	var count int
	for _, server := range allServers {
		if server.Avg.Milliseconds() == 0 {
			continue
		}
		if count > 0 && count%3 == 0 {
			result += "\n"
		}
		count++
		avgStr := fmt.Sprintf("%4d", server.Avg.Milliseconds())
		name := server.Name
		padding := 16 - runewidth.StringWidth(name)
		if padding < 0 {
			padding = 0
		}
		result += fmt.Sprintf("%s%s%4s | ", name, strings.Repeat(" ", padding), avgStr)
	}
	return result
}

func logError(msg string) {
	if model.EnableLoger {
		Logger.Info(msg)
	}
}
