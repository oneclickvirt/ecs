package nt

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	fastTrace "github.com/nxtrace/NTrace-core/fast_trace"
	"github.com/nxtrace/NTrace-core/ipgeo"
	"github.com/nxtrace/NTrace-core/trace"
	"github.com/nxtrace/NTrace-core/util"
	"github.com/nxtrace/NTrace-core/wshandle"
	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/nt3/model"
)

func realtimePrinter(res *trace.Result, ttl int) {
	//fmt.Printf("%s  ", color.New(color.FgHiYellow, color.Bold).Sprintf("%-2d", ttl+1))
	var latestIP string
	tmpMap := make(map[string][]string)
	for i, v := range res.Hops[ttl] {
		if v.Address == nil && latestIP != "" {
			tmpMap[latestIP] = append(tmpMap[latestIP], fmt.Sprintf("%-10s", fmt.Sprintf("%.2f ms", v.RTT.Seconds()*1000)))
			continue
		} else if v.Address == nil {
			continue
		}
		if _, exist := tmpMap[v.Address.String()]; !exist {
			tmpMap[v.Address.String()] = append(tmpMap[v.Address.String()], strconv.Itoa(i))
			if latestIP == "" {
				for j := 0; j < i; j++ {
					tmpMap[v.Address.String()] = append(tmpMap[v.Address.String()], fmt.Sprintf("%-10s", fmt.Sprintf("%.2f ms", v.RTT.Seconds()*1000)))
				}
			}
			latestIP = v.Address.String()
		}
		tmpMap[v.Address.String()] = append(tmpMap[v.Address.String()], fmt.Sprintf("%-10s", fmt.Sprintf("%.2f ms", v.RTT.Seconds()*1000)))
	}
	if latestIP == "" {
		fmt.Printf(White("*") + "\n")
		return
	}
	for ip, v := range tmpMap {
		i, _ := strconv.Atoi(v[0])
		rtt := v[1]
		// 打印RTT
		fmt.Printf(Cyan("%-12s "), rtt)
		// 打印AS号
		if res.Hops[ttl][i].Geo.Asnumber != "" {
			fmt.Printf(Yellow("%-10s "), fmt.Sprintf("AS%s", res.Hops[ttl][i].Geo.Asnumber))
		} else {
			fmt.Printf(White("%-10s "), "*")
		}
		// 打印地理信息
		if net.ParseIP(ip).To4() != nil {
			whoisFormat := strings.Split(res.Hops[ttl][i].Geo.Whois, "-")
			if len(whoisFormat) > 1 {
				whoisFormat[0] = strings.Join(whoisFormat[:2], "-")
			}
			if whoisFormat[0] != "" {
				//如果以RFC或DOD开头那么为空
				if !(strings.HasPrefix(whoisFormat[0], "RFC") ||
					strings.HasPrefix(whoisFormat[0], "DOD")) {
					whoisFormat[0] = "[" + whoisFormat[0] + "]"
				} else {
					whoisFormat[0] = ""
				}
			}
			// CMIN2, CUII, CN2, CUG 改为壕金色高亮
			switch {
			case res.Hops[ttl][i].Geo.Asnumber == "58807":
				fallthrough
			case res.Hops[ttl][i].Geo.Asnumber == "10099":
				fallthrough
			case res.Hops[ttl][i].Geo.Asnumber == "4809":
				fallthrough
			case res.Hops[ttl][i].Geo.Asnumber == "9929":
				fallthrough
			case res.Hops[ttl][i].Geo.Asnumber == "23764":
				fallthrough
			case whoisFormat[0] == "[CTG-CN]":
				fallthrough
			case whoisFormat[0] == "[CNC-BACKBONE]":
				fallthrough
			case whoisFormat[0] == "[CUG-BACKBONE]":
				fallthrough
			case whoisFormat[0] == "[CMIN2-NET]":
				fallthrough
			case strings.HasPrefix(res.Hops[ttl][i].Address.String(), "59.43."):
				fmt.Printf(Yellow("%s "), fmt.Sprintf("%-18s", whoisFormat[0]))
			default:
				fmt.Printf(Green("%s "), fmt.Sprintf("%-18s", whoisFormat[0]))
			}
			var parts []string
			country := res.Hops[ttl][i].Geo.Country
			prov := res.Hops[ttl][i].Geo.Prov
			city := res.Hops[ttl][i].Geo.City
			owner := res.Hops[ttl][i].Geo.Owner
			if country != "" {
				parts = append(parts, White(country))
			}
			if prov != "" {
				parts = append(parts, White(prov))
			}
			if city != "" {
				parts = append(parts, White(city))
			}
			if owner != "" {
				parts = append(parts, White(owner))
			}
			if len(parts) > 0 {
				fmt.Printf(strings.Join(parts, ", "))
			}
		}
		fmt.Println()
	}
}

func tracert(f fastTrace.FastTracer, ispCollection fastTrace.ISPCollection) {
	fmt.Printf("traceroute to %s, %d hops max, %d byte packets\n", ispCollection.IP, f.ParamsFastTrace.MaxHops, f.ParamsFastTrace.PktSize)
	ip, err := util.DomainLookUp(ispCollection.IP, "4", "", true)
	if err != nil {
		log.Fatal(err)
	}
	var conf = trace.Config{
		BeginHop:         1,
		DestIP:           ip,
		DestPort:         80,
		MaxHops:          30,
		NumMeasurements:  3,
		ParallelRequests: 18,
		RDns:             f.ParamsFastTrace.RDns,
		AlwaysWaitRDNS:   f.ParamsFastTrace.AlwaysWaitRDNS,
		PacketInterval:   50,
		TTLInterval:      50,
		IPGeoSource:      ipgeo.GetSource("LeoMoeAPI"),
		Timeout:          time.Duration(1000) * time.Millisecond,
		SrcAddr:          f.ParamsFastTrace.SrcAddr,
		PktSize:          52,
		Lang:             f.ParamsFastTrace.Lang,
		DontFragment:     f.ParamsFastTrace.DontFragment,
	}
	conf.RealtimePrinter = realtimePrinter
	//conf.RealtimePrinter = printer.RealtimePrinter
	//conf.RealtimePrinter = tracelog.RealtimePrinter
	_, err = trace.Traceroute(f.TracerouteMethod, conf)
	if err != nil && model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
		Logger.Info("trace failed: " + err.Error())
	}
}

func tracert_v6(f fastTrace.FastTracer, ispCollection fastTrace.ISPCollection) {
	fmt.Printf("traceroute to %s, %d hops max, %d byte packets\n", ispCollection.IPv6, f.ParamsFastTrace.MaxHops, f.ParamsFastTrace.PktSize)
	ip, err := util.DomainLookUp(ispCollection.IPv6, "6", "", true)
	if err != nil {
		log.Fatal(err)
	}
	var conf = trace.Config{
		BeginHop:         1,
		DestIP:           ip,
		DestPort:         80,
		MaxHops:          30,
		NumMeasurements:  3,
		ParallelRequests: 18,
		RDns:             f.ParamsFastTrace.RDns,
		AlwaysWaitRDNS:   f.ParamsFastTrace.AlwaysWaitRDNS,
		PacketInterval:   50,
		TTLInterval:      50,
		IPGeoSource:      ipgeo.GetSource("LeoMoeAPI"),
		Timeout:          time.Duration(1000) * time.Millisecond,
		SrcAddr:          f.ParamsFastTrace.SrcAddr,
		PktSize:          52,
		Lang:             f.ParamsFastTrace.Lang,
		DontFragment:     f.ParamsFastTrace.DontFragment,
	}
	conf.RealtimePrinter = realtimePrinter
	//conf.RealtimePrinter = printer.RealtimePrinter
	//conf.RealtimePrinter = tracelog.RealtimePrinter
	_, err = trace.Traceroute(f.TracerouteMethod, conf)
	if err != nil && model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
		Logger.Info("trace failed: " + err.Error())
	}
}

func TraceRoute(language, location, testType string) {
	if language == "zh" || language == "" {
		language = "cn"
	} else if language != "en" {
		fmt.Println("Invalid language.")
		return
	}
	var TL []fastTrace.ISPCollection
	if location == "GZ" {
		TL = []fastTrace.ISPCollection{model.GuangZhouCT, model.GuangZhouCU, model.GuangZhouCMCC}
	} else if location == "BJ" {
		TL = []fastTrace.ISPCollection{model.BeiJingCT, model.BeiJingCU, model.BeiJingCMCC}
	} else if location == "SH" {
		TL = []fastTrace.ISPCollection{model.ShangHaiCT, model.ShangHaiCU, model.ShangHaiCMCC}
	} else if location == "CD" {
		TL = []fastTrace.ISPCollection{model.ChengDuCT, model.ChengDuCU, model.ChengDuCMCC}
	} else {
		fmt.Println("Invalid location.")
		return
	}
	pFastTrace := fastTrace.ParamsFastTrace{
		SrcDev:         "",
		SrcAddr:        "",
		BeginHop:       1,
		MaxHops:        30,
		RDns:           false,
		AlwaysWaitRDNS: false,
		Lang:           language,
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
	ft.TracerouteMethod = trace.ICMPTrace
	if TL != nil {
		for _, T := range TL {
			if testType == "both" {
				fmt.Printf(Yellow("%s - "), fmt.Sprintf("%s - ICMP v4", T.ISPName))
				tracert(ft, T)
				fmt.Printf(Yellow("%s - "), fmt.Sprintf("%s - ICMP v6", T.ISPName))
				tracert_v6(ft, T)
			} else if testType == "ipv4" {
				fmt.Printf(Yellow("%s - "), fmt.Sprintf("%s - ICMP v4", T.ISPName))
				tracert(ft, T)
			} else if testType == "ipv6" {
				fmt.Printf(Yellow("%s - "), fmt.Sprintf("%s - ICMP v6", T.ISPName))
				tracert_v6(ft, T)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}
