package utils

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/UnlockTests/executor"
	bnetwork "github.com/oneclickvirt/basics/network"
	"github.com/oneclickvirt/basics/system"
	butils "github.com/oneclickvirt/basics/utils"
	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/security/network"
)

// 获取本程序本日及总执行的统计信息
type StatsResponse struct {
	Counter   string `json:"counter"`
	Action    string `json:"action"`
	Total     int    `json:"total"`
	Daily     int    `json:"daily"`
	Date      string `json:"date"`
	Timestamp string `json:"timestamp"`
}

// 获取最新的Github的仓库中的版本
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// PrintCenteredTitle 根据指定的宽度打印居中标题
func PrintCenteredTitle(title string, width int) {
	// 计算字符串的字符数
	titleLength := utf8.RuneCountInString(title)
	totalPadding := width - titleLength
	padding := totalPadding / 2
	paddingStr := strings.Repeat("-", padding)
	fmt.Println(paddingStr + title + paddingStr + strings.Repeat("-", totalPadding%2))
}

// PrintHead 根据语言打印头部信息
func PrintHead(language string, width int, ecsVersion string) {
	if language == "zh" {
		PrintCenteredTitle("VPS融合怪测试", width)
		fmt.Printf("版本：%s\n", ecsVersion)
		fmt.Println("测评频道: https://t.me/+UHVoo2U4VyA5NTQ1\n" +
			"Go项目地址：https://github.com/oneclickvirt/ecs\n" +
			"Shell项目地址：https://github.com/spiritLHLS/ecs")
	} else {
		PrintCenteredTitle("VPS Fusion Monster Test", width)
		fmt.Printf("Version: %s\n", ecsVersion)
		fmt.Println("Review Channel: https://t.me/+UHVoo2U4VyA5NTQ1\n" +
			"Go Project: https://github.com/oneclickvirt/ecs\n" +
			"Shell Project: https://github.com/spiritLHLS/ecs")
	}
}

func CheckChina(enableLogger bool) bool {
	if enableLogger {
		InitLogger()
		defer Logger.Sync()
	}
	var selectChina bool
	client := req.C()
	client.SetTimeout(6 * time.Second)
	client.R().
		SetRetryCount(2).
		SetRetryBackoffInterval(1*time.Second, 3*time.Second).
		SetRetryFixedInterval(2 * time.Second)
	ipapiURL := "https://ipapi.co/json"
	ipapiResp, err := client.R().Get(ipapiURL)
	if err != nil {
		if enableLogger {
			Logger.Info("无法获取IP信息:" + err.Error())
		}
		return false
	}
	defer ipapiResp.Body.Close()
	ipapiBody, err := ipapiResp.ToString()
	if err != nil {
		if enableLogger {
			Logger.Info("无法读取IP信息响应:" + err.Error())
		}
		return false
	}
	isInChina := strings.Contains(ipapiBody, "China")
	if isInChina {
		fmt.Println("根据 ipapi.co 提供的信息，当前IP可能在中国")
		var input string
		fmt.Print("是否选用中国专项测试(无流媒体测试，有三网Ping值测试)? ([y]/n) ")
		fmt.Scanln(&input)
		switch strings.ToLower(input) {
		case "yes", "y":
			fmt.Println("使用中国专项测试")
			selectChina = true
		case "no", "n":
			fmt.Println("不使用中国专项测试")
		default:
			fmt.Println("使用中国专项测试")
			selectChina = true
		}
	}
	return selectChina
}

// OnlyBasicsIpInfo 仅检查和输出IP信息
func OnlyBasicsIpInfo(language string) (string, string, string) {
	ipv4, ipv6, ipInfo, _, err := bnetwork.NetworkCheck("both", false, language)
	if err != nil {
		return "", "", ""
	}
	basicInfo := ipInfo
	if strings.Contains(ipInfo, "IPV4") && strings.Contains(ipInfo, "IPV6") && ipv4 != "" && ipv6 != "" {
		executor.IPV4 = true
		executor.IPV6 = true
	} else if strings.Contains(ipInfo, "IPV4") && ipv4 != "" {
		executor.IPV4 = true
		executor.IPV6 = false
	} else if strings.Contains(ipInfo, "IPV6") && ipv6 != "" {
		executor.IPV6 = true
		executor.IPV4 = false
	}
	basicInfo = strings.ReplaceAll(basicInfo, "\n\n", "\n")
	return ipv4, ipv6, basicInfo
}

// BasicsAndSecurityCheck 执行安全检查
func BasicsAndSecurityCheck(language, nt3CheckType string, securityCheckStatus bool) (string, string, string, string, string) {
	var wgt sync.WaitGroup
	var ipv4, ipv6, ipInfo, securityInfo, systemInfo string
	wgt.Add(1)
	go func() {
		defer wgt.Done()
		ipv4, ipv6, ipInfo, securityInfo, _ = network.NetworkCheck("both", securityCheckStatus, language)
		// if err != nil {
		// 	fmt.Println(err.Error())
		// }
	}()
	wgt.Add(1)
	go func() {
		defer wgt.Done()
		systemInfo = system.CheckSystemInfo(language)
	}()
	wgt.Wait()
	basicInfo := systemInfo + ipInfo
	if strings.Contains(ipInfo, "IPV4") && strings.Contains(ipInfo, "IPV6") && ipv4 != "" && ipv6 != "" {
		executor.IPV4 = true
		executor.IPV6 = true
		if nt3CheckType == "" {
			nt3CheckType = "ipv4"
		}
	} else if strings.Contains(ipInfo, "IPV4") && ipv4 != "" {
		executor.IPV4 = true
		executor.IPV6 = false
		if nt3CheckType == "" {
			nt3CheckType = "ipv4"
		}
	} else if strings.Contains(ipInfo, "IPV6") && ipv6 != "" {
		executor.IPV6 = true
		executor.IPV4 = false
		if nt3CheckType == "" {
			nt3CheckType = "ipv6"
		}
	}
	if nt3CheckType == "ipv4" && !strings.Contains(ipInfo, "IPV4") && strings.Contains(ipInfo, "IPV6") {
		nt3CheckType = "ipv6"
	} else if nt3CheckType == "ipv6" && !strings.Contains(ipInfo, "IPV6") && strings.Contains(ipInfo, "IPV4") {
		nt3CheckType = "ipv4"
	}
	basicInfo = strings.ReplaceAll(basicInfo, "\n\n", "\n")
	return ipv4, ipv6, basicInfo, securityInfo, nt3CheckType
}

// CaptureOutput 捕获函数输出和错误输出，实时输出，并返回字符串
func CaptureOutput(f func()) string {
	// 保存旧的 stdout 和 stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	// 创建管道
	stdoutPipeR, stdoutPipeW, err := os.Pipe()
	if err != nil {
		return "Error creating stdout pipe"
	}
	stderrPipeR, stderrPipeW, err := os.Pipe()
	if err != nil {
		stdoutPipeW.Close()
		stdoutPipeR.Close()
		return "Error creating stderr pipe"
	}
	// 替换标准输出和标准错误输出为管道写入端
	os.Stdout = stdoutPipeW
	os.Stderr = stderrPipeW
	// 恢复标准输出和标准错误输出
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		stdoutPipeW.Close()
		stderrPipeW.Close()
		stdoutPipeR.Close()
		stderrPipeR.Close()
	}()
	// 缓冲区
	var stdoutBuf, stderrBuf bytes.Buffer
	// 并发读取 stdout 和 stderr
	done := make(chan struct{})
	go func() {
		multiWriter := io.MultiWriter(&stdoutBuf, oldStdout)
		io.Copy(multiWriter, stdoutPipeR)
		done <- struct{}{}
	}()
	go func() {
		multiWriter := io.MultiWriter(&stderrBuf, oldStderr)
		io.Copy(multiWriter, stderrPipeR)
		done <- struct{}{}
	}()
	// 执行函数
	f()
	// 关闭管道写入端，让管道读取端可以读取所有数据
	stdoutPipeW.Close()
	stderrPipeW.Close()
	// 等待两个 goroutine 完成
	<-done
	<-done
	// 返回捕获的输出字符串
	// stderrBuf.String()
	return stdoutBuf.String()
}

// PrintAndCapture 捕获函数输出的同时打印内容
func PrintAndCapture(f func(), tempOutput, output string) string {
	tempOutput = CaptureOutput(f)
	output += tempOutput
	return output
}

// UploadText 上传文本内容到指定URL
func UploadText(absPath string) (string, string, error) {
	primaryURL := "http://hpaste.spiritlhl.net/api/UL/upload"
	backupURL := "https://paste.spiritlhl.net/api/UL/upload"
	token := network.SecurityUploadToken
	client := req.C().SetTimeout(6 * time.Second)
	client.R().
		SetRetryCount(2).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetRetryFixedInterval(2 * time.Second)
	// 打开文件
	file, err := os.Open(absPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	// 获取文件信息并检查大小
	fileInfo, err := file.Stat()
	if err != nil {
		return "", "", fmt.Errorf("failed to get file info: %w", err)
	}
	if fileInfo.Size() > 25*1024 { // 25KB
		return "", "", fmt.Errorf("file size exceeds 25KB limit")
	}
	// 上传逻辑
	upload := func(url string) (string, string, error) {
		file, err := os.Open(absPath)
		if err != nil {
			return "", "", fmt.Errorf("failed to re-open file for %s: %w", url, err)
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return "", "", fmt.Errorf("failed to read file content for %s: %w", url, err)
		}
		resp, err := client.R().
			SetHeader("Authorization", token).
			SetFileBytes("file", filepath.Base(absPath), content).
			Post(url)
		if err != nil {
			return "", "", fmt.Errorf("failed to make request to %s: %w", url, err)
		}
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 && resp.String() != "" {
			fileID := strings.TrimSpace(resp.String())
			if strings.Contains(fileID, "show") {
				fileID = fileID[strings.LastIndex(fileID, "/")+1:]
			}
			httpURL := fmt.Sprintf("http://hpaste.spiritlhl.net/#/show/%s", fileID)
			httpsURL := fmt.Sprintf("https://paste.spiritlhl.net/#/show/%s", fileID)
			return httpURL, httpsURL, nil
		}
		return "", "", fmt.Errorf("upload failed for %s with status code: %d", url, resp.StatusCode)
	}
	// 尝试上传到主URL
	httpURL, httpsURL, err := upload(primaryURL)
	if err == nil {
		return httpURL, httpsURL, nil
	}
	// 尝试上传到备份URL
	httpURL, httpsURL, err = upload(backupURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to both primary and backup URLs: %w", err)
	}
	return httpURL, httpsURL, nil
}

// ProcessAndUpload 创建结果文件并上传文件
func ProcessAndUpload(output string, filePath string, enableUplaod bool) (string, string) {
	// 使用 defer 来处理 panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("处理上传时发生错误: %v\n", r)
		}
	}()
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，删除文件
		err = os.Remove(filePath)
		if err != nil {
			fmt.Println("无法删除文件:", err)
			return "", ""
		}
	}
	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("无法创建文件:", err)
		return "", ""
	}
	defer file.Close()
	// 匹配 ANSI 转义序列
	ansiRegex := regexp.MustCompile("\x1B\\[[0-9;]+[a-zA-Z]")
	// 移除 ANSI 转义序列
	cleanedOutput := ansiRegex.ReplaceAllString(output, "")
	// 使用 bufio.Writer 提高写入效率
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(cleanedOutput)
	if err != nil {
		fmt.Println("无法写入文件:", err)
		return "", ""
	}
	// 确保写入缓冲区的数据都刷新到文件中
	err = writer.Flush()
	if err != nil {
		fmt.Println("无法刷新文件缓冲:", err)
		return "", ""
	}
	fmt.Printf("测试结果已写入 %s\n", filePath)
	if enableUplaod {
		// 获取文件的绝对路径
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			fmt.Println("无法获取文件绝对路径:", err)
			return "", ""
		}
		// 上传文件并生成短链接
		http_url, https_url, err := UploadText(absPath)
		if err != nil {
			fmt.Println("上传失败，无法生成链接")
			fmt.Println(err.Error())
			return "", ""
		}
		return http_url, https_url
	}
	return "", ""
}

var StackType string

type NetCheckResult struct {
	HasIPv4   bool
	HasIPv6   bool
	Connected bool
	StackType string // "IPv4", "IPv6", "DualStack", "None"
}

func makeResolver(proto, dnsAddr string) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, proto, dnsAddr)
		},
	}
}

// 前置联网能力检测
func CheckPublicAccess(timeout time.Duration) NetCheckResult {
	if timeout < 2*time.Second {
		timeout = 2 * time.Second
	}
	var wg sync.WaitGroup
	resultChan := make(chan string, 8)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	checks := []struct {
		Tag  string
		Addr string
		Kind string // udp4, udp6, http4, http6
	}{
		// UDP DNS
		{"IPv4", "223.5.5.5:53", "udp4"},              // 阿里 DNS
		{"IPv4", "8.8.8.8:53", "udp4"},                // Google DNS
		{"IPv6", "[2400:3200::1]:53", "udp6"},         // 阿里 IPv6 DNS
		{"IPv6", "[2001:4860:4860::8888]:53", "udp6"}, // Google IPv6 DNS
		// HTTP HEAD
		{"IPv4", "https://www.baidu.com", "http4"},     // 百度
		{"IPv4", "https://1.1.1.1", "http4"},           // Cloudflare
		{"IPv6", "https://[2400:3200::1]", "http6"},    // 阿里 IPv6
		{"IPv6", "https://[2606:4700::1111]", "http6"}, // Cloudflare IPv6
	}
	for _, check := range checks {
		wg.Add(1)
		go func(tag, addr, kind string) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
				}
			}()
			switch kind {
			case "udp4", "udp6":
				dialer := &net.Dialer{
					Timeout: timeout / 4,
				}
				conn, err := dialer.DialContext(ctx, kind, addr)
				if err == nil && conn != nil {
					conn.Close()
					select {
					case resultChan <- tag:
					case <-ctx.Done():
						return
					}
				}
			case "http4", "http6":
				var resolver *net.Resolver
				if kind == "http4" {
					resolver = makeResolver("udp4", "223.5.5.5:53")
				} else {
					resolver = makeResolver("udp6", "[2400:3200::1]:53")
				}
				dialer := &net.Dialer{
					Timeout:  timeout / 4,
					Resolver: resolver,
				}
				transport := &http.Transport{
					DialContext:           dialer.DialContext,
					MaxIdleConns:          1,
					MaxIdleConnsPerHost:   1,
					IdleConnTimeout:       time.Second,
					TLSHandshakeTimeout:   timeout / 4,
					ResponseHeaderTimeout: timeout / 4,
					DisableKeepAlives:     true,
				}
				client := &http.Client{
					Timeout:   timeout / 4,
					Transport: transport,
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
				}
				req, err := http.NewRequestWithContext(ctx, "HEAD", addr, nil)
				if err != nil {
					return
				}
				resp, err := client.Do(req)
				if err == nil && resp != nil {
					if resp.Body != nil {
						resp.Body.Close()
					}
					if resp.StatusCode < 500 {
						select {
						case resultChan <- tag:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}(check.Tag, check.Addr, check.Kind)
	}
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	hasV4 := false
	hasV6 := false
	for {
		select {
		case res, ok := <-resultChan:
			if !ok {
				goto result
			}
			if res == "IPv4" {
				hasV4 = true
			}
			if res == "IPv6" {
				hasV6 = true
			}
		case <-ctx.Done():
			goto result
		}
	}
result:
	stack := "None"
	if hasV4 && hasV6 {
		stack = "DualStack"
	} else if hasV4 {
		stack = "IPv4"
	} else if hasV6 {
		stack = "IPv6"
	}
	StackType = stack
	butils.CheckPublicAccess(3 * time.Second) // 设置basics检测，避免部分测试未启用
	return NetCheckResult{
		HasIPv4:   hasV4,
		HasIPv6:   hasV6,
		Connected: hasV4 || hasV6,
		StackType: stack,
	}
}

// 获取每日/总的程序执行统计信息
func GetGoescStats() (*StatsResponse, error) {
	client := req.C().SetTimeout(5 * time.Second)
	var stats StatsResponse
	resp, err := client.R().
		SetSuccessResult(&stats).
		Get("https://hits.spiritlhl.net/goecs")
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return &stats, nil
}

// 统计结果单位转换
func FormatGoecsNumber(num int) string {
	if num >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	} else if num >= 1000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	}
	return fmt.Sprintf("%d", num)
}

// 通过Github的API检索仓库最新TAG的版本
func GetLatestEcsRelease() (*GitHubRelease, error) {
	urls := []string{
		"https://api.github.com/repos/oneclickvirt/ecs/releases/latest",
		"https://fd.spiritlhl.top/https://api.github.com/repos/oneclickvirt/ecs/releases/latest",
		"https://githubapi.spiritlhl.top/repos/oneclickvirt/ecs/releases/latest",
		"https://githubapi.spiritlhl.workers.dev/repos/oneclickvirt/ecs/releases/latest",
	}
	client := req.C().SetTimeout(3 * time.Second)
	for _, url := range urls {
		var release GitHubRelease
		resp, err := client.R().
			SetSuccessResult(&release).
			Get(url)
		if err != nil {
			continue
		}
		if resp.IsSuccessState() && release.TagName != "" {
			return &release, nil
		}
	}
	return nil, fmt.Errorf("failed to fetch release from all sources")
}

// 比较程序版本是否需要升级
func CompareVersions(v1, v2 string) int {
	normalize := func(s string) []int {
		s = strings.TrimPrefix(strings.ToLower(s), "v")
		parts := strings.Split(s, ".")
		result := make([]int, 3)
		for i := 0; i < 3 && i < len(parts); i++ {
			n, _ := strconv.Atoi(parts[i])
			result[i] = n
		}
		return result
	}
	a := normalize(v1)
	b := normalize(v2)
	for i := 0; i < 3; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
