package ui

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// TestUI 测试界面结构体
type TestUI struct {
	App    fyne.App
	Window fyne.Window

	// 测试选项复选框 - 完整支持所有测试项
	BasicCheck     *widget.Check
	CpuCheck       *widget.Check
	MemoryCheck    *widget.Check
	DiskCheck      *widget.Check
	CommCheck      *widget.Check // 御三家流媒体
	UnlockCheck    *widget.Check // 跨国流媒体解锁
	SecurityCheck  *widget.Check // IP质量检测
	EmailCheck     *widget.Check // 邮件端口检测
	BacktraceCheck *widget.Check // 上游及回程线路
	Nt3Check       *widget.Check // 三网回程路由
	SpeedCheck     *widget.Check // 网络测速
	PingCheck      *widget.Check // 三网PING值
	LogCheck       *widget.Check // 启用日志记录

	// 预设模式选择
	PresetSelect *widget.Select

	// 配置选项
	LanguageSelect     *widget.Select
	CpuMethodSelect    *widget.Select
	MemoryMethodSelect *widget.Select
	DiskMethodSelect   *widget.Select
	DiskPathEntry      *widget.Entry
	ThreadModeSelect   *widget.Select
	Nt3LocationSelect  *widget.Select
	Nt3TypeSelect      *widget.Select
	DiskMultiCheck     *widget.Check
	SpNumEntry         *widget.Entry

	// 控制按钮
	StartButton *widget.Button
	StopButton  *widget.Button

	// 结果显示 - 使用终端输出组件
	Terminal    *TerminalOutput
	ProgressBar *widget.ProgressBar
	StatusLabel *widget.Label

	// 日志相关
	LogViewer *widget.Entry      // 日志查看器
	LogTab    *fyne.Container    // 日志标签页内容
	MainTabs  *container.AppTabs // 主标签页容器

	// 运行状态
	IsRunning bool
	CancelCtx context.Context
	CancelFn  context.CancelFunc
	Mu        sync.Mutex
}
