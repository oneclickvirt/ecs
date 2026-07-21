package menu

import (
	"strconv"

	textinput "github.com/charmbracelet/bubbles/textinput"
	"github.com/oneclickvirt/ecs/internal/params"
	"github.com/oneclickvirt/ecs/utils"
)

func defaultMainItems() []mainMenuItem {
	return []mainMenuItem{
		{id: "1", zh: "融合怪完全体(能测全测)", en: "Full Test (All Available Tests)", descZh: "系统信息、CPU、内存、磁盘、解锁、IP质量、邮件端口、回程、NT3、测速、TGDC、网站延迟。", descEn: "Runs all available modules: system, compute, memory, disk, unlock, security, routing and speed.", needNet: false},
		{id: "2", zh: "极简版", en: "Minimal Suite", descZh: "系统信息+CPU+内存+磁盘+测速节点×5，不含解锁/网络/路由测试。", descEn: "System info + CPU + memory + disk + 5 speed nodes. No unlock/network/routing tests.", needNet: false},
		{id: "3", zh: "精简版", en: "Standard Suite", descZh: "系统信息+CPU+内存+磁盘+跨国平台解锁+三网回程路由+测速节点×5。", descEn: "System info + CPU + memory + disk + streaming unlock + 3-network routing + 5 speed nodes.", needNet: false},
		{id: "4", zh: "精简网络版", en: "Network Suite", descZh: "系统信息+CPU+内存+磁盘+上游及三网回程路由+测速节点×5。", descEn: "System info + CPU + memory + disk + upstream/3-network backtrace routing + 5 speed nodes.", needNet: false},
		{id: "5", zh: "精简解锁版", en: "Unlock Suite", descZh: "系统信息+CPU+内存+磁盘IO+跨国平台解锁+测速节点×5。", descEn: "System info + CPU + memory + disk IO + streaming unlock + 5 speed nodes.", needNet: false},
		{id: "6", zh: "网络单项", en: "Network Only", descZh: "仅网络维度：IP质量、回程、NT3、延迟、TGDC、网站和测速。", descEn: "Network-only profile: IP quality, route, latency, TGDC, websites, speed.", needNet: true},
		{id: "7", zh: "解锁单项", en: "Unlock Only", descZh: "仅进行跨国平台解锁与流媒体可用性检测。", descEn: "Unlock-only profile for cross-border media/service availability.", needNet: true},
		{id: "8", zh: "硬件单项", en: "Hardware Only", descZh: "系统信息、CPU、内存、dd/fio 磁盘测试。", descEn: "Hardware-only profile with system, CPU, memory and disk tests.", needNet: false},
		{id: "9", zh: "IP质量检测", en: "IP Quality", descZh: "15个数据库IP质量检测+邮件端口连通性检测。", descEn: "IP quality check across 15 databases + email port connectivity test.", needNet: true},
		{id: "10", zh: "三网回程线路", en: "3-Network Route", descZh: "三网回程、NT3路由、延迟、TGDC、网站延迟专项。", descEn: "3-network backtrace + NT3 route + latency/TGDC/website checks.", needNet: true},
		{id: "custom", zh: ">>> 高级自定义(全参数模式)", en: ">>> Advanced Custom (Full Parameters)", descZh: "按参数逐项配置，支持测试项、方法、路径、上传和结果分析。", descEn: "Configure per-parameter with test toggles, methods, paths, upload and analysis.", needNet: false},
		{id: "0", zh: "退出程序", en: "Exit Program", descZh: "退出当前程序。", descEn: "Exit program.", needNet: false},
	}
}

func defaultTestToggles() []testToggle {
	return []testToggle{
		{key: "basic", nameZh: "基础系统信息", nameEn: "Basic System Info", descZh: "操作系统、CPU型号、内核、虚拟化等基础信息。", descEn: "OS, CPU model, kernel, virtualization and base environment info.", enabled: true, needNet: false},
		{key: "cpu", nameZh: "CPU测试", nameEn: "CPU Test", descZh: "按所选方法执行 CPU 计算性能测试。", descEn: "Run CPU compute benchmarks using selected method.", enabled: true, needNet: false},
		{key: "memory", nameZh: "内存测试", nameEn: "Memory Test", descZh: "按所选方法测试内存吞吐和访问性能。", descEn: "Run memory throughput and access benchmarks by selected method.", enabled: true, needNet: false},
		{key: "disk", nameZh: "磁盘测试", nameEn: "Disk Test", descZh: "按所选方法执行磁盘读写性能测试。", descEn: "Run disk read/write benchmark using selected method/path.", enabled: true, needNet: false},
		{key: "ut", nameZh: "跨国平台解锁", nameEn: "Streaming Unlock", descZh: "检测多类海外流媒体与服务可用性。", descEn: "Check availability of cross-border streaming/services.", enabled: true, needNet: true},
		{key: "security", nameZh: "IP质量检测", nameEn: "IP Quality Check", descZh: "多库 IP 信誉、风险和质量信息检测。", descEn: "IP reputation/risk/quality checks across multiple datasets.", enabled: true, needNet: true},
		{key: "email", nameZh: "邮件端口检测", nameEn: "Email Port Check", descZh: "检查常见邮件相关端口连通能力。", descEn: "Check common mail-related port connectivity.", enabled: true, needNet: true},
		{key: "backtrace", nameZh: "回程路由", nameEn: "Backtrace Route", descZh: "检测上游及三网回程路径。", descEn: "Inspect upstream and 3-network return routes.", enabled: true, needNet: true},
		{key: "nt3", nameZh: "NT3路由", nameEn: "NT3 Route", descZh: "按指定地区与协议执行详细路由追踪。", descEn: "Run detailed route trace by selected location/protocol.", enabled: true, needNet: true},
		{key: "speed", nameZh: "测速", nameEn: "Speed Test", descZh: "测试下载/上传带宽与延迟。", descEn: "Measure download/upload bandwidth and latency.", enabled: true, needNet: true},
		{key: "ping", nameZh: "Ping测试", nameEn: "Ping Test", descZh: "全国/多地区延迟质量测试。", descEn: "Latency quality checks across multiple regions.", enabled: false, needNet: true},
		{key: "tgdc", nameZh: "Telegram DC测试", nameEn: "Telegram DC Test", descZh: "检测各 Telegram 数据中心节点延迟。", descEn: "Measure latency to each Telegram data center node.", enabled: true, needNet: true},
		{key: "web", nameZh: "网站延迟", nameEn: "Website Latency", descZh: "检测常见网站访问延迟。", descEn: "Check latency to commonly used websites.", enabled: true, needNet: true},
	}
}

func applyExplicitConfigToToggles(toggles []testToggle, config *params.Config) {
	if config == nil {
		return
	}
	for i := range toggles {
		if !config.UserSetFlags[toggles[i].key] {
			continue
		}
		switch toggles[i].key {
		case "basic":
			toggles[i].enabled = config.BasicStatus
		case "cpu":
			toggles[i].enabled = config.CpuTestStatus
		case "memory":
			toggles[i].enabled = config.MemoryTestStatus
		case "disk":
			toggles[i].enabled = config.DiskTestStatus
		case "ut":
			toggles[i].enabled = config.UtTestStatus
		case "security":
			toggles[i].enabled = config.SecurityTestStatus
		case "email":
			toggles[i].enabled = config.EmailTestStatus
		case "backtrace":
			toggles[i].enabled = config.BacktraceStatus
		case "nt3":
			toggles[i].enabled = config.Nt3Status
		case "speed":
			toggles[i].enabled = config.SpeedTestStatus
		case "ping":
			toggles[i].enabled = config.PingTestStatus
		case "tgdc":
			toggles[i].enabled = config.TgdcTestStatus
		case "web":
			toggles[i].enabled = config.WebTestStatus
		}
	}
}

func option(value, zh, en, descZh, descEn string) advOption {
	return advOption{value: value, labelZh: zh, labelEn: en, descZh: descZh, descEn: descEn}
}

func defaultAdvSettings(config *params.Config) []advSetting {
	adv := []advSetting{
		{
			key: "cpum", nameZh: "CPU测试方法", nameEn: "CPU Method", kind: "option",
			descZh: "选择 CPU 基准测试工具（sysbench/geekbench/winsat）。",
			descEn: "Choose CPU benchmark tool (sysbench/geekbench/winsat).",
			options: []advOption{
				option("sysbench", "Sysbench", "Sysbench", "通用 CPU 基准测试工具。", "General-purpose CPU benchmark tool."),
				option("geekbench", "Geekbench", "Geekbench", "综合场景 CPU 基准测试工具。", "Synthetic benchmark simulating real-world application workloads."),
				option("winsat", "WinSAT", "WinSAT", "Windows 环境下的 CPU 基准测试工具。", "CPU benchmark tool for Windows environments."),
			},
		},
		{
			key: "cput", nameZh: "CPU线程模式", nameEn: "CPU Thread Mode", kind: "option",
			descZh: "单线程: 测试单核最高运算速度; 多线程: 测试全核并发吞吐。",
			descEn: "Single-thread: peak single-core speed; Multi-thread: full-core parallel throughput.",
			options: []advOption{
				option("multi", "多线程", "Multi-thread", "测试所有核心并发运算吞吐。", "Measure parallel compute throughput across all cores."),
				option("single", "单线程", "Single-thread", "测试单核最高运算速度。", "Measure peak single-core compute speed."),
			},
		},
		{
			key: "memorym", nameZh: "内存测试方法", nameEn: "Memory Method", kind: "option",
			descZh: "选择内存基准测试工具。",
			descEn: "Choose memory benchmark tool.",
			options: []advOption{
				option("stream", "STREAM", "STREAM", "专项内存带宽基准测试工具（STREAM）。", "Memory bandwidth benchmark tool (STREAM)."),
				option("sysbench", "Sysbench", "Sysbench", "通用内存基准测试工具。", "General-purpose memory benchmark tool."),
				option("dd", "dd", "dd", "使用 dd 命令测量内存顺序读写。", "Measure memory sequential R/W using dd command."),
				option("winsat", "WinSAT", "WinSAT", "Windows 环境内存基准测试工具。", "Memory benchmark tool for Windows environments."),
				option("auto", "自动", "Auto", "按优先级自动选择可用测试工具。", "Automatically select the preferred available tool."),
			},
		},
		{
			key: "diskm", nameZh: "磁盘测试方法", nameEn: "Disk Method", kind: "option",
			descZh: "选择磁盘基准测试工具。",
			descEn: "Choose disk benchmark tool.",
			options: []advOption{
				option("fio", "FIO", "FIO", "多队列深度顺序/随机 I/O 全面基准测试。", "Comprehensive sequential/random I/O benchmark with multiple queue depths."),
				option("dd", "dd", "dd", "使用 dd 命令进行顺序读写基准测试。", "Sequential read/write benchmark using dd command."),
				option("winsat", "WinSAT", "WinSAT", "Windows 环境磁盘基准测试工具。", "Disk benchmark tool for Windows environments."),
			},
		},
		{
			key: "diskp", nameZh: "磁盘测试路径", nameEn: "Disk Test Path", kind: "text",
			descZh:  "自定义磁盘测试目录。留空表示默认路径。",
			descEn:  "Custom disk test directory. Empty means default path.",
			textVal: config.DiskTestPath,
		},
		{
			key: "diskmc", nameZh: "多磁盘检测", nameEn: "Multi-Disk Check", kind: "bool",
			descZh:  "启用后检测并测试所有已挂载磁盘路径。",
			descEn:  "When enabled, detect and benchmark all mounted disk paths.",
			boolVal: config.DiskMultiCheck,
		},
		{
			key: "autodiskm", nameZh: "磁盘方法失败自动切换", nameEn: "Auto Switch Disk Method", kind: "bool",
			descZh:  "所选磁盘测试方法失败时自动切换为其他可用方法。",
			descEn:  "Automatically try another available disk method if the selected method fails.",
			boolVal: config.AutoChangeDiskMethod,
		},
		{
			key: "deep", nameZh: "深度测试", nameEn: "Deep Mode", kind: "bool",
			descZh:  "仅运行下方明确填写目标的高负载深度项目。",
			descEn:  "Run high-load deep tests only for explicitly configured targets.",
			boolVal: config.DeepMode,
		},
		{key: "deepdiskpaths", nameZh: "深度多盘目录", nameEn: "Deep Disk Paths", kind: "text", descZh: "逗号分隔的已挂载普通目录；留空关闭。", descEn: "Comma-separated mounted directories; empty disables.", textVal: config.DeepDiskPaths},
		{key: "deepsmartdevices", nameZh: "SMART自检设备", nameEn: "SMART Devices", kind: "text", descZh: "逗号分隔的显式设备；留空关闭。", descEn: "Comma-separated explicit devices; empty disables.", textVal: config.DeepSMARTDevices},
		{key: "deepburnduration", nameZh: "CPU烤机时长", nameEn: "CPU Burn Duration", kind: "text", descZh: "例如30s或2m；留空关闭。", descEn: "For example 30s or 2m; empty disables.", textVal: config.DeepBurnDuration.String()},
		{key: "deepgpudevice", nameZh: "GPU设备选择器", nameEn: "GPU Device", kind: "text", descZh: "显式GPU设备；留空关闭。", descEn: "Explicit GPU selector; empty disables.", textVal: config.DeepGPUDevice},
		{key: "timeout", nameZh: "全局截止时间", nameEn: "Global Deadline", kind: "text", descZh: "最长15m，例如10m。", descEn: "Up to 15m, for example 10m.", textVal: config.MaxDuration.String()},
		{key: "hardwarebudget", nameZh: "硬件阶段预算", nameEn: "Hardware Budget", kind: "text", descZh: "标准模式最长2m；深度模式不超过全局截止时间。", descEn: "Up to 2m in standard mode; deep mode is capped by the global deadline.", textVal: config.HardwareBudget.String()},
		{
			key: "nt3loc", nameZh: "NT3测试地区", nameEn: "NT3 Location", kind: "option",
			descZh: "选择路由追踪地区。显示中文全称，内部仍使用标准参数值。",
			descEn: "Choose route trace region. Full names are shown while preserving standard values.",
			options: []advOption{
				option("GZ", "广州", "Guangzhou", "从广州节点进行追踪。", "Trace from Guangzhou node."),
				option("SH", "上海", "Shanghai", "从上海节点进行追踪。", "Trace from Shanghai node."),
				option("BJ", "北京", "Beijing", "从北京节点进行追踪。", "Trace from Beijing node."),
				option("CD", "成都", "Chengdu", "从成都节点进行追踪。", "Trace from Chengdu node."),
				option("ALL", "全部地区", "All Regions", "依次测试全部地区节点。", "Run route traces from all supported regions."),
			},
		},
		{
			key: "nt3t", nameZh: "NT3协议类型", nameEn: "NT3 Protocol", kind: "option",
			descZh: "指定 NT3 路由检测协议栈。",
			descEn: "Select protocol stack used by NT3 route checks.",
			options: []advOption{
				option("ipv4", "仅 IPv4", "IPv4 Only", "仅测试 IPv4 路由路径。", "Test IPv4 routing only."),
				option("ipv6", "仅 IPv6", "IPv6 Only", "仅测试 IPv6 路由路径。", "Test IPv6 routing only."),
				option("both", "IPv4 + IPv6", "IPv4 + IPv6", "同时测试 IPv4 与 IPv6。", "Test both IPv4 and IPv6."),
			},
		},
		{
			key: "spnum", nameZh: "测速节点数/运营商", nameEn: "Speed Nodes per ISP", kind: "option",
			descZh: "每个运营商参与测速的节点数量。",
			descEn: "Number of speed test nodes per ISP.",
			options: []advOption{
				option("1", "1 个", "1 node", "每运营商1节点，耗时最短，覆盖面最小。", "1 node per ISP, shortest runtime, least coverage."),
				option("2", "2 个", "2 nodes", "每运营商2节点（默认值）。", "2 nodes per ISP (default)."),
				option("3", "3 个", "3 nodes", "每运营商3节点，覆盖面扩大，耗时增加。", "3 nodes per ISP, wider coverage, longer runtime."),
				option("4", "4 个", "4 nodes", "每运营商4节点。", "4 nodes per ISP."),
				option("5", "5 个", "5 nodes", "每运营商5节点，覆盖面宽，耗时较高。", "5 nodes per ISP, wide coverage, higher runtime."),
				option("6", "6 个", "6 nodes", "每运营商6节点。", "6 nodes per ISP."),
				option("7", "7 个", "7 nodes", "每运营商7节点。", "7 nodes per ISP."),
				option("8", "8 个", "8 nodes", "每运营商8节点。", "8 nodes per ISP."),
				option("9", "9 个", "9 nodes", "每运营商9节点。", "9 nodes per ISP."),
				option("10", "10 个", "10 nodes", "每运营商10节点，覆盖面最宽，耗时最高。", "10 nodes per ISP, widest coverage, longest runtime."),
			},
		},
		{
			key: "unlockregion", nameZh: "解锁检测地区", nameEn: "Unlock Region", kind: "option",
			descZh: "选择跨国平台解锁检测覆盖的地区组合（仅在启用解锁检测时生效）。",
			descEn: "Select region combination for streaming unlock test (only when unlock test is enabled).",
			options: []advOption{
				option("0", "跨国平台", "Global", "仅检测跨国流媒体平台（默认）。", "Check global/international streaming platforms only (default)."),
				option("1", "跨国 + 台湾", "Global + Taiwan", "跨国平台 + 台湾本地平台。", "Global + Taiwan local platforms."),
				option("2", "跨国 + 香港", "Global + Hong Kong", "跨国平台 + 香港本地平台。", "Global + Hong Kong local platforms."),
				option("3", "跨国 + 日本", "Global + Japan", "跨国平台 + 日本本地平台。", "Global + Japan local platforms."),
				option("4", "跨国 + 韩国", "Global + Korea", "跨国平台 + 韩国本地平台。", "Global + Korea local platforms."),
				option("5", "跨国 + 北美", "Global + North America", "跨国平台 + 北美本地平台。", "Global + North America local platforms."),
				option("6", "跨国 + 南美", "Global + South America", "跨国平台 + 南美本地平台。", "Global + South America local platforms."),
				option("7", "跨国 + 欧洲", "Global + Europe", "跨国平台 + 欧洲本地平台。", "Global + Europe local platforms."),
				option("8", "跨国 + 非洲", "Global + Africa", "跨国平台 + 非洲本地平台。", "Global + Africa local platforms."),
				option("9", "跨国 + 大洋洲", "Global + Oceania", "跨国平台 + 大洋洲本地平台。", "Global + Oceania local platforms."),
				option("10", "仅台湾", "Taiwan Only", "仅检测台湾本地平台。", "Taiwan local platforms only."),
				option("11", "仅香港", "Hong Kong Only", "仅检测香港本地平台。", "Hong Kong local platforms only."),
				option("12", "仅日本", "Japan Only", "仅检测日本本地平台。", "Japan local platforms only."),
				option("13", "仅韩国", "Korea Only", "仅检测韩国本地平台。", "Korea local platforms only."),
				option("14", "仅北美", "North America Only", "仅检测北美本地平台。", "North America local platforms only."),
				option("15", "仅南美", "South America Only", "仅检测南美本地平台。", "South America local platforms only."),
				option("16", "仅欧洲", "Europe Only", "仅检测欧洲本地平台。", "Europe local platforms only."),
				option("17", "仅非洲", "Africa Only", "仅检测非洲本地平台。", "Africa local platforms only."),
				option("18", "仅大洋洲", "Oceania Only", "仅检测大洋洲本地平台。", "Oceania local platforms only."),
				option("19", "仅体育", "Sports Only", "仅检测体育类平台。", "Sports platforms only."),
				option("20", "全部平台", "All Platforms", "检测所有地区全部平台（耗时最长）。", "Check all platforms across all regions (longest runtime)."),
				option("21", "仅 AI 平台", "AI Only", "仅检测 AI 平台。", "Check AI platforms only."),
			},
		},
		{
			key: "unlockshowip", nameZh: "解锁IP标签兼容开关", nameEn: "Legacy IP Label Switch", kind: "bool",
			descZh:  "兼容旧配置；当前解锁输出会自动在小节标题中标注 IPV4/IPV6。",
			descEn:  "Legacy compatibility; unlock output now marks IPV4/IPV6 in section headers automatically.",
			boolVal: config.UnlockTestShowIP,
		},
		{
			key: "unlockipver", nameZh: "解锁测试IP版本", nameEn: "Unlock Test IP Version", kind: "option",
			descZh: "选择解锁测试使用的IP版本，默认自动测试所有可用版本，测不到时静默跳过。",
			descEn: "Select which IP version to use for unlock tests. Default auto-tests all available; silently skips unavailable versions.",
			options: []advOption{
				option("auto", "自动(全部)", "Auto (Both)", "自动测试所有可用IP版本（默认）。", "Test all available IP versions (default)."),
				option("ipv4", "仅 IPv4", "IPv4 Only", "仅使用 IPv4 进行解锁测试。", "Only test using IPv4."),
				option("ipv6", "仅 IPv6", "IPv6 Only", "仅使用 IPv6 进行解锁测试。", "Only test using IPv6."),
			},
		},
		{key: "utinterface", nameZh: "解锁源接口或IP", nameEn: "Unlock Source Interface/IP", kind: "text", descZh: "留空使用默认路由。", descEn: "Empty uses the default route.", textVal: config.UnlockTestInterface},
		{key: "utdns", nameZh: "解锁DNS服务器", nameEn: "Unlock DNS Servers", kind: "text", descZh: "多个DNS用逗号分隔；留空使用系统DNS。", descEn: "Comma-separated DNS servers; empty uses system DNS.", textVal: config.UnlockTestDNSServers},
		{key: "uthttpproxy", nameZh: "解锁HTTP代理", nameEn: "Unlock HTTP Proxy", kind: "text", descZh: "留空关闭。", descEn: "Empty disables.", textVal: config.UnlockTestHTTPProxy},
		{key: "utsocksproxy", nameZh: "解锁SOCKS5代理", nameEn: "Unlock SOCKS5 Proxy", kind: "text", descZh: "留空关闭。", descEn: "Empty disables.", textVal: config.UnlockTestSOCKSProxy},
		{key: "utconcurrency", nameZh: "解锁并发数", nameEn: "Unlock Concurrency", kind: "text", descZh: "范围1到100。", descEn: "Range 1 to 100.", textVal: strconv.Itoa(config.UnlockTestConcurrency)},
		{
			key: "log", nameZh: "调试日志", nameEn: "Debug Logger", kind: "bool",
			descZh:  "启用后输出更多调试日志，便于排障。",
			descEn:  "Enable verbose logs for troubleshooting.",
			boolVal: config.EnableLogger,
		},
		{
			key: "upload", nameZh: "上传并生成分享链接", nameEn: "Upload & Share Link", kind: "bool",
			descZh:  "启用后上传测试结果并生成可分享链接。",
			descEn:  "Upload final result and generate a shareable link.",
			boolVal: config.EnableUpload,
		},
		{
			key: "analysis", nameZh: "测试后结果总结分析", nameEn: "Post-Test Summary Analysis", kind: "bool",
			descZh:  "测试结束后输出简明总结（含CPU排名、带宽和延迟数据）。",
			descEn:  "Output a concise summary after tests (CPU rank, bandwidth, latency scores).",
			boolVal: config.AnalyzeResult,
		},
		{
			key: "filepath", nameZh: "结果文件名", nameEn: "Result File Name", kind: "text",
			descZh:  "上传前本地结果文件名。",
			descEn:  "Local result filename used before upload.",
			textVal: config.FilePath,
		},
		{key: "privacy", nameZh: "隐私模式", nameEn: "Privacy Mode", kind: "bool", descZh: "隐藏敏感硬件标识并禁止上传。", descEn: "Hide sensitive hardware identifiers and disable upload.", boolVal: config.PrivacyMode},
		{key: "tcp", nameZh: "TCP握手探针", nameEn: "TCP Handshake Probe", kind: "bool", descZh: "追加TCP握手延迟与错误分类单项。", descEn: "Append the TCP latency and error classification section.", boolVal: config.TCPProbeStatus},
		{
			key: "tcpformat", nameZh: "TCP输出明细", nameEn: "TCP Output Detail", kind: "option",
			descZh: "紧凑模式显示全局与分类汇总及少量异常/最慢目标；完整模式逐目标显示。",
			descEn: "Compact shows overall/category summaries and selected exceptions; full lists every target.",
			options: []advOption{
				option("compact", "紧凑汇总", "Compact", "默认，保留完整统计且减少屏幕占用。", "Default; complete aggregate coverage with fewer rows."),
				option("full", "完整逐项", "Full", "显示全部目标明细。", "Show every target row."),
			},
		},
		{key: "jsonpath", nameZh: "JSON结果路径", nameEn: "JSON Result Path", kind: "text", descZh: "留空关闭；使用-输出到标准输出。", descEn: "Empty disables; use - for stdout.", textVal: config.JSONPath},
		{key: "dataoffline", nameZh: "仅使用内置数据", nameEn: "Embedded Data Only", kind: "bool", descZh: "禁止远程数据请求并使用内置最新有效快照。", descEn: "Disable remote data requests and use the embedded valid snapshot.", boolVal: config.DataOffline},
		{key: "datacdn", nameZh: "数据CDN地址", nameEn: "Data CDN Base", kind: "text", descZh: "Go ECS内置快照目录的CDN基础地址。", descEn: "CDN base URL for the Go ECS snapshot directory.", textVal: config.DataCDNBase},
	}

	for i := range adv {
		switch adv[i].key {
		case "cpum":
			adv[i].current = optionIndexByValue(adv[i].options, config.CpuTestMethod)
		case "cput":
			adv[i].current = optionIndexByValue(adv[i].options, config.CpuTestThreadMode)
		case "memorym":
			adv[i].current = optionIndexByValue(adv[i].options, config.MemoryTestMethod)
		case "diskm":
			adv[i].current = optionIndexByValue(adv[i].options, config.DiskTestMethod)
		case "nt3loc":
			adv[i].current = optionIndexByValue(adv[i].options, config.Nt3Location)
		case "nt3t":
			adv[i].current = optionIndexByValue(adv[i].options, config.Nt3CheckType)
		case "spnum":
			adv[i].current = optionIndexByValue(adv[i].options, strconv.Itoa(config.SpNum))
		case "unlockregion":
			adv[i].current = optionIndexByValue(adv[i].options, config.UnlockTestRegion)
		case "unlockipver":
			adv[i].current = optionIndexByValue(adv[i].options, config.UnlockTestIPVersion)
		case "tcpformat":
			adv[i].current = optionIndexByValue(adv[i].options, config.TCPTextFormat)
		}
	}

	return adv
}

func optionIndexByValue(options []advOption, value string) int {
	for i, opt := range options {
		if opt.value == value {
			return i
		}
	}
	return 0
}

func newTuiModel(preCheck utils.NetCheckResult, config *params.Config, langPreset bool, statsTotal, statsDaily int, hasStats bool, cmpVersion int, newVersion string) tuiModel {
	toggles := defaultTestToggles()
	applyExplicitConfigToToggles(toggles, config)
	advanced := defaultAdvSettings(config)
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = ""
	ti.CharLimit = 255
	ti.Width = 45
	m := tuiModel{
		config:         config,
		preCheck:       preCheck,
		langPreset:     langPreset,
		mainItems:      defaultMainItems(),
		mainAnalyze:    config.AnalyzeResult,
		mainUpload:     config.EnableUpload,
		mainExtraTotal: 2,
		toggles:        toggles,
		advanced:       advanced,
		customTotal:    len(toggles) + len(advanced) + 1,
		statsTotal:     statsTotal,
		statsDaily:     statsDaily,
		hasStats:       hasStats,
		cmpVersion:     cmpVersion,
		newVersion:     newVersion,
		width:          config.Width,
		height:         24,
		textInput:      ti,
	}
	if langPreset {
		m.phase = phaseMain
		m.result.language = config.Language
	} else {
		m.phase = phaseLang
	}
	return m
}
