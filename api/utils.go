package api

import (
	"time"

	"github.com/oneclickvirt/ecs/utils"
)

// NetCheckResult 网络检查结果
type NetCheckResult = utils.NetCheckResult

// StatsResponse 统计信息响应
type StatsResponse = utils.StatsResponse

// GitHubRelease GitHub发布信息
type GitHubRelease = utils.GitHubRelease

// CheckPublicAccess 检查公网访问能力
// timeout: 超时时间
// 返回: 网络检查结果
func CheckPublicAccess(timeout time.Duration) NetCheckResult {
	return utils.CheckPublicAccess(timeout)
}

// GetGoescStats 获取goecs统计信息
// 返回: (统计响应, 错误)
func GetGoescStats() (*StatsResponse, error) {
	return utils.GetGoescStats()
}

// GetLatestEcsRelease 获取最新的ECS版本信息
// 返回: (GitHub发布信息, 错误)
func GetLatestEcsRelease() (*GitHubRelease, error) {
	return utils.GetLatestEcsRelease()
}

// PrintHead 打印程序头部信息
// language: 语言 ("zh" 或 "en")
// width: 显示宽度
// version: 版本号
func PrintHead(language string, width int, version string) {
	utils.PrintHead(language, width, version)
}

// PrintCenteredTitle 打印居中标题
// title: 标题文本
// width: 显示宽度
func PrintCenteredTitle(title string, width int) {
	utils.PrintCenteredTitle(title, width)
}

// ProcessAndUpload 处理并上传结果
// output: 输出内容
// filePath: 文件路径
// enableUpload: 是否启用上传
// 返回: (HTTP URL, HTTPS URL)
func ProcessAndUpload(output, filePath string, enableUpload bool) (string, string) {
	return utils.ProcessAndUpload(output, filePath, enableUpload)
}

// BasicsAndSecurityCheck 基础信息和安全检查
// language: 语言
// checkType: 检查类型
// securityTestStatus: 是否执行安全测试
// 返回: (IPv4地址, IPv6地址, 基础信息, 安全信息, 检查类型)
func BasicsAndSecurityCheck(language, checkType string, securityTestStatus bool) (string, string, string, string, string) {
	return utils.BasicsAndSecurityCheck(language, checkType, securityTestStatus)
}

// OnlyBasicsIpInfo 仅获取基础IP信息
// language: 语言
// 返回: (IPv4地址, IPv6地址, IP信息)
func OnlyBasicsIpInfo(language string) (string, string, string) {
	return utils.OnlyBasicsIpInfo(language)
}

// FormatGoecsNumber 格式化数字显示
// num: 数字
// 返回: 格式化后的字符串
func FormatGoecsNumber(num int) string {
	return utils.FormatGoecsNumber(num)
}

// PrintAndCapture 打印并捕获输出
// fn: 执行的函数
// tempOutput: 临时输出
// existingOutput: 现有输出
// 返回: 捕获的输出
func PrintAndCapture(fn func(), tempOutput, existingOutput string) string {
	return utils.PrintAndCapture(fn, tempOutput, existingOutput)
}
