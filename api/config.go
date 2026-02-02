package api

import (
	"github.com/oneclickvirt/ecs/internal/params"
)

// Config 配置接口，导出用于外部调用
type Config = params.Config

// NewConfig 创建默认配置
// version: 版本号字符串
func NewConfig(version string) *Config {
	return params.NewConfig(version)
}

// NewDefaultConfig 创建默认配置（使用默认版本号）
func NewDefaultConfig() *Config {
	return params.NewConfig("v0.1.114")
}

// ConfigOption 配置选项函数类型
type ConfigOption func(*Config)

// WithLanguage 设置语言
func WithLanguage(lang string) ConfigOption {
	return func(c *Config) {
		c.Language = lang
	}
}

// WithCpuTestMethod 设置CPU测试方法
// method: "sysbench" 或 "geekbench"
func WithCpuTestMethod(method string) ConfigOption {
	return func(c *Config) {
		c.CpuTestMethod = method
	}
}

// WithCpuTestThreadMode 设置CPU测试线程模式
// mode: "single" 或 "multi"
func WithCpuTestThreadMode(mode string) ConfigOption {
	return func(c *Config) {
		c.CpuTestThreadMode = mode
	}
}

// WithMemoryTestMethod 设置内存测试方法
// method: "stream", "sysbench", "dd"
func WithMemoryTestMethod(method string) ConfigOption {
	return func(c *Config) {
		c.MemoryTestMethod = method
	}
}

// WithDiskTestMethod 设置硬盘测试方法
// method: "fio" 或 "dd"
func WithDiskTestMethod(method string) ConfigOption {
	return func(c *Config) {
		c.DiskTestMethod = method
	}
}

// WithDiskTestPath 设置硬盘测试路径
func WithDiskTestPath(path string) ConfigOption {
	return func(c *Config) {
		c.DiskTestPath = path
	}
}

// WithDiskMultiCheck 设置是否进行硬盘多路径检测
func WithDiskMultiCheck(enable bool) ConfigOption {
	return func(c *Config) {
		c.DiskMultiCheck = enable
	}
}

// WithSpeedTestNum 设置测速节点数量
func WithSpeedTestNum(num int) ConfigOption {
	return func(c *Config) {
		c.SpNum = num
	}
}

// WithWidth 设置输出宽度
func WithWidth(width int) ConfigOption {
	return func(c *Config) {
		c.Width = width
	}
}

// WithFilePath 设置输出文件路径
func WithFilePath(path string) ConfigOption {
	return func(c *Config) {
		c.FilePath = path
	}
}

// WithEnableUpload 设置是否启用上传
func WithEnableUpload(enable bool) ConfigOption {
	return func(c *Config) {
		c.EnableUpload = enable
	}
}

// WithEnableLogger 设置是否启用日志
func WithEnableLogger(enable bool) ConfigOption {
	return func(c *Config) {
		c.EnableLogger = enable
	}
}

// WithBasicTest 设置是否执行基础信息测试
func WithBasicTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.BasicStatus = enable
	}
}

// WithCpuTest 设置是否执行CPU测试
func WithCpuTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.CpuTestStatus = enable
	}
}

// WithMemoryTest 设置是否执行内存测试
func WithMemoryTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.MemoryTestStatus = enable
	}
}

// WithDiskTest 设置是否执行硬盘测试
func WithDiskTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.DiskTestStatus = enable
	}
}

// WithUnlockTest 设置是否执行流媒体解锁测试
func WithUnlockTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.UtTestStatus = enable
	}
}

// WithSecurityTest 设置是否执行IP质量测试
func WithSecurityTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.SecurityTestStatus = enable
	}
}

// WithEmailTest 设置是否执行邮件端口测试
func WithEmailTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.EmailTestStatus = enable
	}
}

// WithBacktraceTest 设置是否执行回程路由测试
func WithBacktraceTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.BacktraceStatus = enable
	}
}

// WithNt3Test 设置是否执行三网路由测试
func WithNt3Test(enable bool) ConfigOption {
	return func(c *Config) {
		c.Nt3Status = enable
	}
}

// WithSpeedTest 设置是否执行测速测试
func WithSpeedTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.SpeedTestStatus = enable
	}
}

// WithPingTest 设置是否执行PING测试
func WithPingTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.PingTestStatus = enable
	}
}

// WithTgdcTest 设置是否执行Telegram DC测试
func WithTgdcTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.TgdcTestStatus = enable
	}
}

// WithWebTest 设置是否执行网站测试
func WithWebTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.WebTestStatus = enable
	}
}

// WithNt3CheckType 设置三网路由检测类型
// checkType: "ipv4", "ipv6" 或 "auto"
func WithNt3CheckType(checkType string) ConfigOption {
	return func(c *Config) {
		c.Nt3CheckType = checkType
	}
}

// WithNt3Location 设置三网路由检测位置
func WithNt3Location(location string) ConfigOption {
	return func(c *Config) {
		c.Nt3Location = location
	}
}

// WithAutoChangeDiskMethod 设置是否自动切换硬盘测试方法
func WithAutoChangeDiskMethod(enable bool) ConfigOption {
	return func(c *Config) {
		c.AutoChangeDiskMethod = enable
	}
}

// WithOnlyChinaTest 设置是否只进行国内测试
func WithOnlyChinaTest(enable bool) ConfigOption {
	return func(c *Config) {
		c.OnlyChinaTest = enable
	}
}

// WithMenuMode 设置是否启用菜单模式
func WithMenuMode(enable bool) ConfigOption {
	return func(c *Config) {
		c.MenuMode = enable
	}
}

// WithOnlyIpInfoCheck 设置是否只进行IP信息检测
func WithOnlyIpInfoCheck(enable bool) ConfigOption {
	return func(c *Config) {
		c.OnlyIpInfoCheck = enable
	}
}

// WithChoice 设置菜单选择
func WithChoice(choice string) ConfigOption {
	return func(c *Config) {
		c.Choice = choice
	}
}

// ApplyOptions 应用配置选项
func ApplyOptions(config *Config, options ...ConfigOption) *Config {
	for _, opt := range options {
		opt(config)
	}
	return config
}
