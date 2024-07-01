package utils

import (
	"bytes"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/UnlockTests/uts"
	"github.com/oneclickvirt/basics/system"
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
			"Go Project URL: https://github.com/oneclickvirt/ecs\n" +
			"Shell Project URL: https://github.com/spiritLHLS/ecs")
	}
}

// SecurityCheck 执行安全检查
func SecurityCheck(language, nt3CheckType string, securtyCheckStatus bool) (string, string, string) {
	var wgt sync.WaitGroup
	var ipInfo, securityInfo, systemInfo string
	var err error
	wgt.Add(2)
	go func() {
		defer wgt.Done()
		ipInfo, securityInfo, err = network.NetworkCheck("both", securtyCheckStatus, language)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()
	go func() {
		defer wgt.Done()
		systemInfo = system.CheckSystemInfo(language)
	}()
	wgt.Wait()
	basicInfo := systemInfo + ipInfo
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

// GetCaptureOutput 仅捕获输出不实时输出
func GetCaptureOutput(f func()) string {
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
	}()
	// 缓冲区
	var stdoutBuf, stderrBuf bytes.Buffer
	// 并发读取 stdout 和 stderr
	done := make(chan struct{})
	go func() {
		io.Copy(&stdoutBuf, stdoutPipeR)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(&stderrBuf, stderrPipeR)
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
	return stdoutBuf.String() + stderrBuf.String()
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
func UploadText(absPath string) (string, error) {
	url := "https://paste.spiritlhl.net/api/upload"
	token := network.SecurityUploadToken
	client := req.DefaultClient()
	client.SetTimeout(6 * time.Second)
	client.R().
		SetRetryCount(2).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetRetryFixedInterval(2 * time.Second)
	file, _ := os.Open(absPath)
	resp, err := client.R().
		SetHeader("Authorization", token).
		SetHeader("Format", "RANDOM").
		SetHeader("Max-Views", "0").
		SetHeader("UploadText", "true").
		SetHeader("Content-Type", "multipart/form-data").
		SetHeader("No-JSON", "true").
		SetFileReader("file", "goecs.txt", file).
		Post(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return resp.String(), nil
	} else {
		return "", fmt.Errorf("upload failed with status code: %d", resp.StatusCode)
	}
}

// ProcessAndUpload 创建结果文件并上传文件
func ProcessAndUpload(output string, filePath string, enableUplaod bool) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，删除文件
		err = os.Remove(filePath)
		if err != nil {
			fmt.Println("Cannot delete file:", err)
			return
		}
	}
	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Cannot create file:", err)
		return
	}
	defer file.Close()
	// 匹配 ANSI 转义序列
	ansiRegex := regexp.MustCompile("\x1B\\[[0-9;]+[a-zA-Z]")
	// 移除 ANSI 转义序列
	cleanedOutput := ansiRegex.ReplaceAllString(output, "")
	// 写入文件
	_, err = file.WriteString(cleanedOutput)
	if err != nil {
		fmt.Println("Cannot write to file:", err)
		return
	} else {
		fmt.Println("Write test result in ", filePath)
	}
	if enableUplaod {
		// 获取文件的绝对路径
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			fmt.Println("Failed to get absolute file path:", err)
			return
		}
		// 上传文件并生成短链接
		shorturl, err := UploadText(absPath)
		if err != nil {
			fmt.Println("Upload failed, cannot generate short URL.")
			fmt.Println(err.Error())
			return
		}
		fmt.Println("Upload successful, short URL:", shorturl)
	}
}
