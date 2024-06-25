package ntrace

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/nxtrace/NTrace-core/fast_trace"
	"github.com/nxtrace/NTrace-core/ipgeo"
	"github.com/nxtrace/NTrace-core/printer"
	"github.com/nxtrace/NTrace-core/trace"
	"github.com/nxtrace/NTrace-core/tracelog"
	"github.com/nxtrace/NTrace-core/util"
	"github.com/nxtrace/NTrace-core/wshandle"
	"log"
	"os"
	"os/signal"
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

var oe = false

func tracert(f fastTrace.FastTracer, location string, ispCollection fastTrace.ISPCollection) {
	fmt.Fprintf(color.Output, "%s\n", color.New(color.FgYellow, color.Bold).Sprintf("『%s %s 』", location, ispCollection.ISPName))
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

	if oe {
		fp, err := os.OpenFile("/tmp/trace.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		if err != nil {
			return
		}
		defer func(fp *os.File) {
			err := fp.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(fp)

		log.SetOutput(fp)
		log.SetFlags(0)
		log.Printf("『%s %s 』\n", location, ispCollection.ISPName)
		log.Printf("traceroute to %s, %d hops max, %d byte packets\n", ispCollection.IP, f.ParamsFastTrace.MaxHops, f.ParamsFastTrace.PktSize)
		conf.RealtimePrinter = tracelog.RealtimePrinter
	} else {
		conf.RealtimePrinter = printer.RealtimePrinter
	}

	_, err = trace.Traceroute(f.TracerouteMethod, conf)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()
}

func tracert_v6(f fastTrace.FastTracer, location string, ispCollection fastTrace.ISPCollection) {
	fmt.Printf("%s『%s %s 』%s\n", printer.YELLOW_PREFIX, location, ispCollection.ISPName, printer.RESET_PREFIX)
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

	if oe {
		fp, err := os.OpenFile("/tmp/trace.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		if err != nil {
			return
		}
		defer func(fp *os.File) {
			err := fp.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(fp)
		log.SetOutput(fp)
		log.SetFlags(0)
		log.Printf("『%s %s 』\n", location, ispCollection.ISPName)
		log.Printf("traceroute to %s, %d hops max, %d byte packets\n", ispCollection.IPv6, f.ParamsFastTrace.MaxHops, f.ParamsFastTrace.PktSize)
		conf.RealtimePrinter = tracelog.RealtimePrinter
	} else {
		conf.RealtimePrinter = printer.RealtimePrinter
	}

	_, err = trace.Traceroute(f.TracerouteMethod, conf)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
}

func traceroute() {
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
	fmt.Println("TCP v4")
	ft.TracerouteMethod = trace.TCPTrace
	tracert(ft, fastTrace.TestIPsCollection.Beijing.Location, fastTrace.TestIPsCollection.Beijing.EDU)
	//fmt.Println("TCP v6")
	//tracert_v6(ft, fastTrace.TestIPsCollection.Beijing.Location, fastTrace.TestIPsCollection.Beijing.EDU)
	fmt.Println("ICMP v4")
	ft.TracerouteMethod = trace.ICMPTrace
	tracert(ft, fastTrace.TestIPsCollection.Beijing.Location, fastTrace.TestIPsCollection.Beijing.EDU)
	//fmt.Println("ICMP v6")
	//tracert_v6(ft, fastTrace.TestIPsCollection.Beijing.Location, fastTrace.TestIPsCollection.Beijing.EDU)
}
