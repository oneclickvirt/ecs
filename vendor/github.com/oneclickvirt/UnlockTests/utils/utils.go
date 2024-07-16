package utils

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/UnlockTests/model"
	. "github.com/oneclickvirt/defaultset"
)

var ClientProxy = http.ProxyFromEnvironment
var AutoTransport = &http.Transport{
	Proxy:       ClientProxy,
	DialContext: (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
}
var AutoHttpClient = &http.Client{
	Timeout:   30 * time.Second,
	Transport: AutoTransport,
}
var Dialer = &net.Dialer{}
var Ipv4Transport = &http.Transport{
	Proxy: ClientProxy,
	DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 强制使用IPv4
		return Dialer.DialContext(ctx, "tcp4", addr)
	},
	// ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   30 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}
var Ipv4HttpClient = &http.Client{
	Timeout:   30 * time.Second,
	Transport: Ipv4Transport,
}
var Ipv6Transport = &http.Transport{
	Proxy: ClientProxy,
	DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 强制使用IPv4
		return Dialer.DialContext(ctx, "tcp6", addr)
	},
	// ForceAttemptHTTP2:     true,
	MaxIdleConns:           100,
	IdleConnTimeout:        90 * time.Second,
	TLSHandshakeTimeout:    30 * time.Second,
	ExpectContinueTimeout:  1 * time.Second,
	MaxResponseHeaderBytes: 262144,
}
var Ipv6HttpClient = &http.Client{
	Timeout:   30 * time.Second,
	Transport: Ipv6Transport,
}

// ParseInterface 解析网卡IP地址
func ParseInterface(ifaceName, ipAddr, netType string) (*http.Client, error) {
	var localIP net.IP
	if ifaceName != "" {
		// 获取指定网卡的 IP 地址
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			return nil, err
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if (netType == "tcp4" && ipNet.IP.To4() != nil) || (netType == "tcp6" && ipNet.IP.To4() == nil) {
					localIP = ipNet.IP
					break
				}
			}
		}
	} else if ipAddr != "" {
		localIP = net.ParseIP(ipAddr)
		if (netType == "tcp4" && localIP.To4() == nil) || (netType == "tcp6" && localIP.To4() != nil) {
			return nil, fmt.Errorf("IP address does not match the specified netType")
		}
	}
	var dialContext func(ctx context.Context, network, addr string) (net.Conn, error)
	if localIP != nil {
		dialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{
				Timeout:   12 * time.Second,
				KeepAlive: 12 * time.Second,
				LocalAddr: &net.TCPAddr{
					IP: localIP,
				},
			}).DialContext(ctx, netType, addr)
		}
	} else {
		dialContext = func(ctx context.Context, network string, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, netType, addr)
		}
	}
	c := &http.Client{
		Timeout: 12 * time.Second,
		Transport: &http.Transport{
			DialContext: dialContext,
		}}
	return c, nil
}

// Req
// 为 req 设置请求
func Req(c *http.Client) *req.Client {
	client := req.C()
	client.ImpersonateChrome()
	client.Transport.DialContext = c.Transport.(*http.Transport).DialContext
	client.SetProxy(c.Transport.(*http.Transport).Proxy)
	client.R().
		SetRetryCount(2).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetRetryFixedInterval(2 * time.Second)
	client.SetTimeout(10 * time.Second)
	return client
}

// ReqDefault
// 为 req 设置请求
func ReqDefault(c *http.Client) *req.Client {
	client := req.C()
	if client.Headers == nil {
		client.Headers = make(http.Header)
	}
	client.Transport.DialContext = c.Transport.(*http.Transport).DialContext
	client.SetProxy(c.Transport.(*http.Transport).Proxy)
	client.R().
		SetRetryCount(2).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetRetryFixedInterval(2 * time.Second)
	client.SetTimeout(10 * time.Second)
	return client
}

// SetReqHeaders
func SetReqHeaders(client *req.Client, headers map[string]string) *req.Client {
	for key, value := range headers {
		client.Headers.Set(key, value)
	}
	return client
}

// PostJson 向指定的 URL 发送 JSON 格式的 POST 请求，并返回响应、响应体和错误信息
// url: 目标 URL
// payload: 要发送的 JSON 格式的请求体
// headers: 可选的 HTTP 头信息
func PostJson(c *http.Client, url string, payload string, headers map[string]string) (*req.Response, string, error) {
	if model.EnableLoger {
		InitLogger()
		defer Logger.Sync()
	}
	// 构建 POST 请求，设置请求类型为 JSON 并添加请求体
	request := ReqDefault(c)
	// 添加可选的 HTTP 头信息
	if headers != nil {
		request = SetReqHeaders(request, headers)
	}
	resp, err := request.R().SetBodyJsonString(payload).Post(url)
	if err != nil {
		if model.EnableLoger {
			Logger.Info("PostJson failed: " + err.Error())
		}
		return resp, "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		if model.EnableLoger {
			Logger.Info("read resp.Body failed: " + err.Error())
		}
		return resp, "", err
	}
	body := string(b)
	return resp, body, err
}

// GetRegion
// 判断地址是否在允许的地区范围内
func GetRegion(loc string, locationList []string) bool {
	for _, s := range locationList {
		if loc == s {
			return true
		}
	}
	return false
}

// ReParse
// 根据正则表达式提取内容
func ReParse(responseBody, rex string) string {
	re := regexp.MustCompile(rex)
	match := re.FindStringSubmatch(responseBody)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

// CheckDNSIP 检测IP地址是否同子网/在内网
func CheckDNSIP(ipStr string, referenceIP string) int {
	// 解析输入的IP地址字符串
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 1 // 如果IP地址无效，返回1
	}
	if ip.To4() != nil {
		// 处理IPv4地址
		privateIPv4Ranges := []string{
			"10.0.0.0/8",
			"172.16.0.0/12",
			"169.254.0.0/16",
			"192.168.0.0/16",
		}
		// 检查IP是否在私有IPv4地址范围内
		for _, cidr := range privateIPv4Ranges {
			_, ipNet, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			if ipNet.Contains(ip) {
				return 0 // 如果IP在私有地址范围内，返回0
			}
		}
		// 检查IP是否与参考IP在同一子网内
		refIP := net.ParseIP(referenceIP)
		if refIP != nil && ip.Mask(net.CIDRMask(24, 32)).Equal(refIP.Mask(net.CIDRMask(24, 32))) {
			return 0 // 如果在同一子网内，返回0
		}
	} else {
		// 处理IPv6地址
		// 检查IP是否在特殊IPv6地址范围内
		if strings.HasPrefix(ipStr, "fe8") || strings.HasPrefix(ipStr, "FE8") ||
			strings.HasPrefix(ipStr, "fc") || strings.HasPrefix(ipStr, "FC") ||
			strings.HasPrefix(ipStr, "fd") || strings.HasPrefix(ipStr, "FD") ||
			strings.HasPrefix(ipStr, "ff") || strings.HasPrefix(ipStr, "FF") {
			return 0 // 如果IP在特殊IPv6地址范围内，返回0
		}
	}
	return 1 // 如果IP不符合上述条件，返回1
}

// lookupHostWithTimeout 检测网址的IP地址
func lookupHostWithTimeout(hostname string, timeout time.Duration) ([]string, error) {
	// 创建带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 使用默认解析器查找主机地址
	return net.DefaultResolver.LookupHost(ctx, hostname)
}

// CheckDNS 三个检测DNS的逻辑并发检测
func CheckDNS(hostname string) (string, string, string) {
	//if strings.Contains(hostname, "https://") {
	//	hostname = strings.ReplaceAll(hostname, "https://", "")
	//	hostname = strings.Split(hostname, "/")[0]
	//}
	var wg sync.WaitGroup
	var result1, result2, result3 string
	wg.Add(3)
	// 内网/同网IP检测
	go func() {
		defer wg.Done()
		addrs, err := lookupHostWithTimeout(hostname, 5*time.Second)
		if err != nil || len(addrs) == 0 {
			result1 = ""
			return
		}
		result1 = "1"
		for i := 0; i < len(addrs); i++ {
			for j := i + 1; j < len(addrs); j++ {
				//fmt.Printf("Checking %s and %s\n", addrs[i], addrs[j])
				if CheckDNSIP(addrs[i], addrs[j]) == 0 {
					result1 = "0"
				}
			}
		}
	}()
	//主域名DNS解析检测
	go func() {
		defer wg.Done()
		addrs, err := lookupHostWithTimeout(hostname, 5*time.Second)
		if err != nil {
			result2 = ""
			return
		}
		// 判断实际的回答数量
		var result2Value string
		switch len(addrs) {
		case 0, 1, 2:
			result2Value = "0"
		default:
			result2Value = "1"
		}
		result2 = result2Value
	}()
	//随机前缀DNS解析检测
	go func() {
		defer wg.Done()
		testDomain := fmt.Sprintf("test%d.%s", rand.Int(), hostname)
		//fmt.Println(testDomain)
		addrs, err := lookupHostWithTimeout(testDomain, 5*time.Second)
		if err != nil || len(addrs) == 0 {
			result3 = "1"
			return
		}
		result3 = "0"
	}()
	wg.Wait()
	return result1, result2, result3
}

// GetUnlockType 获取解锁的类型
func GetUnlockType(results ...string) string {
	// 检查结果中是否有空值
	for _, result := range results {
		if result == "" {
			return ""
		}
	}
	// 检查结果中是否有"0"
	for _, result := range results {
		if result == "0" {
			return "Via DNS"
		}
	}
	return "Native"
}

// 通过Info标记要被插入的行的下一行包含什么文本内容
func PrintCA(c *http.Client) model.Result {
	return model.Result{Name: "Canada", Status: model.PrintHead, Info: "HotStar"}
}

func PrintGB(c *http.Client) model.Result {
	return model.Result{Name: "England", Status: model.PrintHead, Info: "HotStar"}
}

func PrintFR(c *http.Client) model.Result {
	return model.Result{Name: "France", Status: model.PrintHead, Info: "Canal+"}
}

func PrintDE(c *http.Client) model.Result {
	return model.Result{Name: "Germany", Status: model.PrintHead, Info: "Joyn"}
}

func PrintNL(c *http.Client) model.Result {
	return model.Result{Name: "Netherlands", Status: model.PrintHead, Info: "NLZIET"}
}

func PrintES(c *http.Client) model.Result {
	return model.Result{Name: "Spain", Status: model.PrintHead, Info: "Movistar+"}
}

func PrintIT(c *http.Client) model.Result {
	return model.Result{Name: "Italy", Status: model.PrintHead, Info: "Rai Play"}
}

func PrintCH(c *http.Client) model.Result {
	return model.Result{Name: "Switzerland", Status: model.PrintHead, Info: "SKY CH"}
}

func PrintRU(c *http.Client) model.Result {
	return model.Result{Name: "Russia", Status: model.PrintHead, Info: "Amediateka"}
}

func PrintAU(c *http.Client) model.Result {
	return model.Result{Name: "Australia", Status: model.PrintHead, Info: "Stan"}
}

func PrintNZ(c *http.Client) model.Result {
	return model.Result{Name: "New Zealand", Status: model.PrintHead, Info: "Neon TV"}
}

func PrintSG(c *http.Client) model.Result {
	return model.Result{Name: "Singapore", Status: model.PrintHead, Info: "MeWatch"}
}

func PrintTH(c *http.Client) model.Result {
	return model.Result{Name: "Thailand", Status: model.PrintHead, Info: "AIS Play"}
}

func PrintGame(c *http.Client) model.Result {
	return model.Result{Name: "Game", Status: model.PrintHead, Info: "Kancolle Japan"}
}

func PrintMusic(c *http.Client) model.Result {
	return model.Result{Name: "Music", Status: model.PrintHead, Info: "Mora"}
}

func PrintForum(c *http.Client) model.Result {
	return model.Result{Name: "Forum", Status: model.PrintHead, Info: "EroGameSpace"}
}