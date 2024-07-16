package pt

import (
	"context"
	"encoding/csv"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/pingtest/model"
)

func getData(endpoint string) string {
	client := req.C()
	client.SetTimeout(6 * time.Second)
	client.R().
		SetRetryCount(2).
		SetRetryBackoffInterval(1*time.Second, 3*time.Second).
		SetRetryFixedInterval(2 * time.Second)
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	for _, baseUrl := range model.CdnList {
		url := baseUrl + endpoint
		resp, err := client.R().Get(url)
		if err == nil {
			defer resp.Body.Close()
			b, err := io.ReadAll(resp.Body)
			if err == nil {
				return string(b)
			}
		}
		if model.EnableLoger {
			Logger.Info(err.Error())
		}
	}
	return ""
}

func resolveIP(name string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r := net.Resolver{}
	ips, err := r.LookupIP(ctx, "ip", name)
	if err != nil {
		return ""
	}
	if len(ips) > 0 {
		return ips[0].String()
	}
	return ""
}

func parseCSVData(data, platform, operator string) []*model.Server {
	var servers []*model.Server
	r := csv.NewReader(strings.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		if model.EnableLoger {
			Logger.Info(err.Error())
		}
		return servers
	}
	if len(records) > 0 && (records[0][6] == "country_code" || records[0][1] == "country_code") {
		records = records[1:]
	}
	var head string
	switch operator {
	case "cmccc":
		head = "移动"
	case "ct":
		head = "电信"
	case "cu":
		head = "联通"
	}
	if platform == "net" {
		for _, record := range records {
			if len(record) >= 8 {
				servers = append(servers, &model.Server{
					Name: head + record[3],
					IP:   record[4],
					Port: record[6],
				})
			}
		}
	} else if platform == "cn" {
		var name, ip string
		for _, record := range records {
			if len(record) >= 8 {
				ip = strings.Split(record[5], ":")[0]
				if net.ParseIP(ip) == nil {
					ip = resolveIP(ip)
					if ip == "" {
						continue
					}
				}
				name = record[10] + record[8]
				if !strings.Contains(name, head) {
					name = head + name
				}
				servers = append(servers, &model.Server{
					Name: name,
					IP:   ip,
					Port: strings.Split(record[5], ":")[1],
				})
			}
		}
	}
	return servers
}

func getServers(operator string) []*model.Server {
	netList := []string{model.NetCMCC, model.NetCT, model.NetCU}
	cnList := []string{model.CnCMCC, model.CnCT, model.CnCU}
	var servers []*model.Server
	var wg sync.WaitGroup
	dataCh := make(chan []*model.Server, 2)
	// 定义一个函数来获取数据并解析
	fetchData := func(data string, dataType, operator string) {
		defer wg.Done()
		if data != "" {
			parsedData := parseCSVData(data, dataType, operator)
			dataCh <- parsedData
		}
	}
	appendData := func(data1, data2, operator string) {
		wg.Add(2)
		go fetchData(data1, "net", operator)
		go fetchData(data2, "cn", operator)
	}
	switch operator {
	case "cmcc":
		appendData(getData(netList[0]), getData(cnList[0]), operator)
	case "ct":
		appendData(getData(netList[1]), getData(cnList[1]), operator)
	case "cu":
		appendData(getData(netList[2]), getData(cnList[2]), operator)
	}
	go func() {
		wg.Wait()
		close(dataCh)
	}()
	for data := range dataCh {
		servers = append(servers, data...)
	}
	// 去重IP
	uniqueServers := make(map[string]*model.Server)
	for _, server := range servers {
		uniqueServers[server.IP] = server
	}
	servers = []*model.Server{}
	for _, server := range uniqueServers {
		servers = append(servers, server)
	}
	// 去重地址
	uniqueServers = make(map[string]*model.Server)
	for _, server := range servers {
		uniqueServers[server.Name] = server
	}
	servers = []*model.Server{}
	for _, server := range uniqueServers {
		servers = append(servers, server)
	}
	return servers
}
