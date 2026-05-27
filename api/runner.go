package api

import (
	"context"
	"sync"
	"time"

	"github.com/oneclickvirt/ecs/internal/runner"
	"github.com/oneclickvirt/ecs/utils"
)

// RunResult 运行结果
type RunResult struct {
	Output    string        // 完整输出
	Duration  time.Duration // 运行时长
	StartTime time.Time     // 开始时间
	EndTime   time.Time     // 结束时间
}

func applyLanguageAndUploadRules(preCheck utils.NetCheckResult, config *Config) {
	if config.Language == "en" {
		config.BacktraceStatus = false
		config.Nt3Status = false
	}
	if !config.EnableUpload {
		config.SecurityTestStatus = false
	}
	if !preCheck.Connected {
		config.EnableUpload = false
	}
}

// RunAllTests 执行所有测试（高级接口）
// preCheck: 网络检查结果
// config: 配置对象
// 返回: 运行结果
func RunAllTests(preCheck utils.NetCheckResult, config *Config) *RunResult {
	var (
		wg1, wg2, wg3                                         sync.WaitGroup
		basicInfo, securityInfo, emailInfo, mediaInfo, ptInfo string
		output, tempOutput                                    string
		outputMutex                                           sync.Mutex
		infoMutex                                             sync.Mutex
	)

	startTime := time.Now()
	applyLanguageAndUploadRules(preCheck, config)

	switch config.Language {
	case "zh":
		runner.RunChineseTests(context.Background(), preCheck, config, &wg1, &wg2, &wg3,
			&basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo,
			&output, tempOutput, startTime, &outputMutex, &infoMutex)
	case "en":
		runner.RunEnglishTests(context.Background(), preCheck, config, &wg1, &wg2, &wg3,
			&basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo,
			&output, tempOutput, startTime, &outputMutex, &infoMutex)
	default:
		runner.RunChineseTests(context.Background(), preCheck, config, &wg1, &wg2, &wg3,
			&basicInfo, &securityInfo, &emailInfo, &mediaInfo, &ptInfo,
			&output, tempOutput, startTime, &outputMutex, &infoMutex)
	}
	if config.AnalyzeResult {
		output = runner.AppendAnalysisSummary(config, output, tempOutput, &outputMutex)
	}

	endTime := time.Now()

	return &RunResult{
		Output:    output,
		Duration:  endTime.Sub(startTime),
		StartTime: startTime,
		EndTime:   endTime,
	}
}

// RunBasicTests 运行基础信息测试
func RunBasicTests(preCheck utils.NetCheckResult, config *Config) string {
	var (
		basicInfo, securityInfo string
		output, tempOutput      string
		outputMutex             sync.Mutex
	)
	return runner.RunBasicTests(context.Background(), preCheck, config, &basicInfo, &securityInfo, output, tempOutput, &outputMutex)
}

// RunCPUTest 运行CPU测试
func RunCPUTest(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunCPUTest(context.Background(), config, output, tempOutput, &outputMutex)
}

// RunMemoryTest 运行内存测试
func RunMemoryTest(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunMemoryTest(context.Background(), config, output, tempOutput, &outputMutex)
}

// RunDiskTest 运行硬盘测试
func RunDiskTest(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunDiskTest(context.Background(), config, output, tempOutput, &outputMutex)
}

// RunIpInfoCheck 执行IP信息检测
func RunIpInfoCheck(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunIpInfoCheck(context.Background(), config, output, tempOutput, &outputMutex)
}

// RunStreamingTests 运行流媒体测试
func RunStreamingTests(config *Config, mediaInfo string) string {
	var (
		wg1                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return runner.RunStreamingTests(context.Background(), config, &wg1, &mediaInfo, output, tempOutput, &outputMutex, &infoMutex)
}

// RunSecurityTests 运行安全测试
func RunSecurityTests(config *Config, securityInfo string) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunSecurityTests(context.Background(), config, securityInfo, output, tempOutput, &outputMutex)
}

// RunEmailTests 运行邮件端口测试
func RunEmailTests(config *Config, emailInfo string) string {
	var (
		wg2                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return runner.RunEmailTests(context.Background(), config, &wg2, &emailInfo, output, tempOutput, &outputMutex, &infoMutex)
}

// RunNetworkTests 运行网络测试（中文模式）
func RunNetworkTests(config *Config, ptInfo string) string {
	var (
		wg3                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
		infoMutex          sync.Mutex
	)
	return runner.RunNetworkTests(context.Background(), config, &wg3, &ptInfo, output, tempOutput, &outputMutex, &infoMutex)
}

// RunSpeedTests 运行测速测试（中文模式）
func RunSpeedTests(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunSpeedTests(context.Background(), config, output, tempOutput, &outputMutex)
}

// RunEnglishNetworkTests 运行网络测试（英文模式）
func RunEnglishNetworkTests(config *Config, ptInfo string) string {
	var (
		wg3                sync.WaitGroup
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunEnglishNetworkTests(context.Background(), config, &wg3, &ptInfo, output, tempOutput, &outputMutex)
}

// RunEnglishSpeedTests 运行测速测试（英文模式）
func RunEnglishSpeedTests(config *Config) string {
	var (
		output, tempOutput string
		outputMutex        sync.Mutex
	)
	return runner.RunEnglishSpeedTests(context.Background(), config, output, tempOutput, &outputMutex)
}

// AppendTimeInfo 添加时间信息
func AppendTimeInfo(config *Config, output string, startTime time.Time) string {
	var (
		tempOutput  string
		outputMutex sync.Mutex
	)
	return runner.AppendTimeInfo(config, output, tempOutput, startTime, &outputMutex)
}

// HandleUploadResults 处理上传结果
func HandleUploadResults(config *Config, output string) {
	runner.HandleUploadResults(config, output)
}
