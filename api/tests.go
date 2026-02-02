package api

import (
	"github.com/oneclickvirt/ecs/internal/tests"
)

// TestResult 测试结果结构
type TestResult struct {
	TestMethod string // 实际使用的测试方法
	Output     string // 测试输出结果
	Success    bool   // 是否成功
	Error      error  // 错误信息
}

// CpuTest CPU测试公共接口
// language: 语言 ("zh" 或 "en")
// testMethod: 测试方法 ("sysbench" 或 "geekbench")
// testThread: 线程模式 ("single" 或 "multi")
// 返回: (实际测试方法, 测试结果)
func CpuTest(language, testMethod, testThread string) (string, string) {
	return tests.CpuTest(language, testMethod, testThread)
}

// MemoryTest 内存测试公共接口
// language: 语言 ("zh" 或 "en")
// testMethod: 测试方法 ("stream", "sysbench", "dd")
// 返回: (实际测试方法, 测试结果)
func MemoryTest(language, testMethod string) (string, string) {
	return tests.MemoryTest(language, testMethod)
}

// DiskTest 硬盘测试公共接口
// language: 语言 ("zh" 或 "en")
// testMethod: 测试方法 ("fio" 或 "dd")
// testPath: 测试路径
// isMultiCheck: 是否多路径检测
// autoChange: 是否自动切换方法
// 返回: (实际测试方法, 测试结果)
func DiskTest(language, testMethod, testPath string, isMultiCheck, autoChange bool) (string, string) {
	return tests.DiskTest(language, testMethod, testPath, isMultiCheck, autoChange)
}

// MediaTest 流媒体解锁测试公共接口
// language: 语言 ("zh" 或 "en")
// 返回: 测试结果
func MediaTest(language string) string {
	return tests.MediaTest(language)
}

// SpeedTestShowHead 显示测速表头
// language: 语言 ("zh" 或 "en")
func SpeedTestShowHead(language string) {
	tests.ShowHead(language)
}

// SpeedTestNearby 就近节点测速
func SpeedTestNearby() {
	tests.NearbySP()
}

// SpeedTestCustom 自定义测速
// platform: 平台 ("cn" 或 "net")
// operator: 运营商 ("cmcc", "cu", "ct", "global", "other" 等)
// num: 测试节点数量
// language: 语言 ("zh" 或 "en")
func SpeedTestCustom(platform, operator string, num int, language string) {
	tests.CustomSP(platform, operator, num, language)
}

// NextTrace3Check 三网路由追踪测试
// language: 语言 ("zh" 或 "en")
// location: 位置
// checkType: 检测类型 ("ipv4", "ipv6")
func NextTrace3Check(language, location, checkType string) {
	tests.NextTrace3Check(language, location, checkType)
}

// UpstreamsCheck 上游及回程线路检测
func UpstreamsCheck() {
	tests.UpstreamsCheck()
}

// GetIPv4Address 获取当前IPv4地址
func GetIPv4Address() string {
	return tests.IPV4
}

// GetIPv6Address 获取当前IPv6地址
func GetIPv6Address() string {
	return tests.IPV6
}

// SetIPv4Address 设置IPv4地址（用于测试）
func SetIPv4Address(ipv4 string) {
	tests.IPV4 = ipv4
}

// SetIPv6Address 设置IPv6地址（用于测试）
func SetIPv6Address(ipv6 string) {
	tests.IPV6 = ipv6
}
