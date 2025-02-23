package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/UnlockTests/uts"
	"github.com/oneclickvirt/basics/ipv6"
	"github.com/oneclickvirt/basics/system"
	. "github.com/oneclickvirt/defaultset"
	"github.com/oneclickvirt/security/network"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

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
		fmt.Println("测评频道: https://t.me/vps_reviews\n" +
			"Go项目地址：https://github.com/oneclickvirt/ecs\n" +
			"Shell项目地址：https://github.com/spiritLHLS/ecs")
	} else {
		PrintCenteredTitle("VPS Fusion Monster Test", width)
		fmt.Printf("Version: %s\n", ecsVersion)
		fmt.Println("Review Channel: https://t.me/vps_reviews\n" +
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

// BasicsAndSecurityCheck 执行安全检查
func BasicsAndSecurityCheck(language, nt3CheckType string, securtyCheckStatus bool) (string, string, string) {
	var wgt sync.WaitGroup
	var ipInfo, securityInfo, systemInfo string
	var err error
	wgt.Add(1)
	go func() {
		defer wgt.Done()
		ipInfo, securityInfo, err = network.NetworkCheck("both", securtyCheckStatus, language)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()
	wgt.Add(1)
	go func() {
		defer wgt.Done()
		systemInfo = system.CheckSystemInfo(language)
	}()
	wgt.Wait()
	ipv6Info, errv6 := ipv6.GetIPv6Mask(language)
	basicInfo := systemInfo + ipInfo
	if errv6 == nil && ipv6Info != "" {
		basicInfo += ipv6Info
		basicInfo += "\n"
	}
	if strings.Contains(ipInfo, "IPV4") && strings.Contains(ipInfo, "IPV6") {
		uts.IPV4 = true
		uts.IPV6 = true
		if nt3CheckType == "" {
			nt3CheckType = "ipv4"
		}
	} else if strings.Contains(ipInfo, "IPV4") {
		uts.IPV4 = true
		uts.IPV6 = false
		if nt3CheckType == "" {
			nt3CheckType = "ipv4"
		}
	} else if strings.Contains(ipInfo, "IPV6") {
		uts.IPV6 = true
		uts.IPV4 = false
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
	return basicInfo, securityInfo, nt3CheckType
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
