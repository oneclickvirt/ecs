package ntrace

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/nxtrace/NTrace-core/fast_trace"
	"github.com/nxtrace/NTrace-core/ipgeo"
	"github.com/nxtrace/NTrace-core/printer"
	"github.com/nxtrace/NTrace-core/trace"
	"github.com/nxtrace/NTrace-core/util"
	"github.com/nxtrace/NTrace-core/wshandle"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
)

type FastTracer struct {
	TracerouteMethod trace.Method
	ParamsFastTrace  ParamsFastTrace
}

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

type IpListElement struct {
	Ip       string
	Desc     string
	Version4 bool // true for IPv4, false for IPv6
}

func realtimePrinter(res *trace.Result, ttl int) {
	fmt.Printf("%s  ", color.New(color.FgHiYellow, color.Bold).Sprintf("%-2d", ttl+1))
	var latestIP string
	tmpMap := make(map[string][]string)
	for i, v := range res.Hops[ttl] {
		if v.Address == nil && latestIP != "" {
			tmpMap[latestIP] = append(tmpMap[latestIP], fmt.Sprintf("%s ms", "*"))
			continue
		} else if v.Address == nil {
			continue
		}

		if _, exist := tmpMap[v.Address.String()]; !exist {
			tmpMap[v.Address.String()] = append(tmpMap[v.Address.String()], strconv.Itoa(i))
			if latestIP == "" {
				for j := 0; j < i; j++ {
					tmpMap[v.Address.String()] = append(tmpMap[v.Address.String()], fmt.Sprintf("%s ms", "*"))
				}
			}
			latestIP = v.Address.String()
		}

		tmpMap[v.Address.String()] = append(tmpMap[v.Address.String()], fmt.Sprintf("%.2f ms", v.RTT.Seconds()*1000))
	}

	if latestIP == "" {
		fmt.Fprintf(color.Output, "%s\n",
			color.New(color.FgWhite, color.Bold).Sprintf("1*"),
		)
		return
	}

	for ip, v := range tmpMap {
		i, _ := strconv.Atoi(v[0])
		rtt := v[1]

		// 打印RTT
		fmt.Fprintf(color.Output, "%s ", color.New(color.FgHiCyan, color.Bold).Sprintf("%s", rtt))

		// 打印AS号
		if res.Hops[ttl][i].Geo.Asnumber != "" {
			fmt.Fprintf(color.Output, "%s ", color.New(color.FgHiYellow, color.Bold).Sprintf("AS%s", res.Hops[ttl][i].Geo.Asnumber))
		} else {
			fmt.Fprintf(color.Output, "%s ", color.New(color.FgWhite, color.Bold).Sprintf("2*"))
		}

		// 打印地理信息
		if net.ParseIP(ip).To4() != nil {
			fmt.Fprintf(color.Output, "%s, %s, %s, %s",
				color.New(color.FgWhite, color.Bold).Sprintf("%s", res.Hops[ttl][i].Geo.Country),
				color.New(color.FgWhite, color.Bold).Sprintf("%s", res.Hops[ttl][i].Geo.Prov),
				color.New(color.FgWhite, color.Bold).Sprintf("%s", res.Hops[ttl][i].Geo.City),
				color.New(color.FgWhite, color.Bold).Sprintf("%s", res.Hops[ttl][i].Geo.Owner))
		} else {
			fmt.Fprintf(color.Output, "%s, %s, %s, %s",
				color.New(color.FgWhite, color.Bold).Sprintf("%s", res.Hops[ttl][i].Geo.Country),
				color.New(color.FgWhite, color.Bold).Sprintf("%s", res.Hops[ttl][i].Geo.Prov),
				color.New(color.FgWhite, color.Bold).Sprintf("%s", res.Hops[ttl][i].Geo.City),
				color.New(color.FgWhite, color.Bold).Sprintf("%s", res.Hops[ttl][i].Geo.Owner))
		}
		fmt.Println()
	}
}

func tracert(f fastTrace.FastTracer, ispCollection fastTrace.ISPCollection) {
	fmt.Fprintf(color.Output, "%s\n", color.New(color.FgYellow, color.Bold).Sprintf("『%s』", ispCollection.ISPName))
	fmt.Printf("traceroute to %s, %d hops max, %d byte packets\n", ispCollection.IP, f.ParamsFastTrace.MaxHops, f.ParamsFastTrace.PktSize)
	ip, err := util.DomainLookUp(ispCollection.IP, "4", "", true)
	if err != nil {
		log.Fatal(err)
	}
	var conf = trace.Config{
		BeginHop:         f.ParamsFastTrace.BeginHop,
		DestIP:           ip,
		DestPort:         80,
		MaxHops:          f.ParamsFastTrace.MaxHops,
		NumMeasurements:  3,
		ParallelRequests: 18,
		RDns:             f.ParamsFastTrace.RDns,
		AlwaysWaitRDNS:   f.ParamsFastTrace.AlwaysWaitRDNS,
		PacketInterval:   100,
		TTLInterval:      500,
		IPGeoSource:      ipgeo.GetSource("LeoMoeAPI"),
		Timeout:          f.ParamsFastTrace.Timeout,
		SrcAddr:          f.ParamsFastTrace.SrcAddr,
		PktSize:          f.ParamsFastTrace.PktSize,
		Lang:             f.ParamsFastTrace.Lang,
		DontFragment:     f.ParamsFastTrace.DontFragment,
	}
	conf.RealtimePrinter = realtimePrinter
	_, err = trace.Traceroute(f.TracerouteMethod, conf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()
}

func tracert_v6(f fastTrace.FastTracer, ispCollection fastTrace.ISPCollection) {
	fmt.Printf("%s『%s』%s\n", printer.YELLOW_PREFIX, ispCollection.ISPName, printer.RESET_PREFIX)
	fmt.Printf("traceroute to %s, %d hops max, %d byte packets\n", ispCollection.IPv6, f.ParamsFastTrace.MaxHops, f.ParamsFastTrace.PktSize)
	ip, err := util.DomainLookUp(ispCollection.IPv6, "6", "", true)
	if err != nil {
		log.Fatal(err)
	}
	var conf = trace.Config{
		BeginHop:         f.ParamsFastTrace.BeginHop,
		DestIP:           ip,
		DestPort:         80,
		MaxHops:          f.ParamsFastTrace.MaxHops,
		NumMeasurements:  3,
		ParallelRequests: 18,
		RDns:             f.ParamsFastTrace.RDns,
		AlwaysWaitRDNS:   f.ParamsFastTrace.AlwaysWaitRDNS,
		PacketInterval:   100,
		TTLInterval:      500,
		IPGeoSource:      ipgeo.GetSource("LeoMoeAPI"),
		Timeout:          f.ParamsFastTrace.Timeout,
		SrcAddr:          f.ParamsFastTrace.SrcAddr,
		PktSize:          f.ParamsFastTrace.PktSize,
		Lang:             f.ParamsFastTrace.Lang,
		DontFragment:     f.ParamsFastTrace.DontFragment,
	}
	conf.RealtimePrinter = printer.RealtimePrinter
	_, err = trace.Traceroute(f.TracerouteMethod, conf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()
}

func TraceRoute() {
	pFastTrace := fastTrace.ParamsFastTrace{
		SrcDev:         "",
		SrcAddr:        "",
		BeginHop:       1,
		MaxHops:        30,
		RDns:           false,
		AlwaysWaitRDNS: false,
		Lang:           "",
		PktSize:        52,
	}
	ft := fastTrace.FastTracer{ParamsFastTrace: pFastTrace}
	// 建立 WebSocket 连接
	w := wshandle.New()
	w.Interrupt = make(chan os.Signal, 1)
	signal.Notify(w.Interrupt, os.Interrupt)
	defer func() {
		w.Conn.Close()
	}()
	fmt.Println("ICMP v4")
	ft.TracerouteMethod = trace.ICMPTrace
	DX := fastTrace.ISPCollection{
		ISPName: "广州电信",
		IP:      "58.60.188.222",
		IPv6:    "240e:97c:2f:3000::44",
	}
	tracert(ft, DX)
	//fmt.Println("ICMP v6")
	//tracert_v6(ft, DX)
}
