package api

const (
	// Version API版本号
	Version = "v1.0.0"
	
	// DefaultVersion 默认的ECS版本号
	DefaultVersion = "v0.1.114"
)

// 测试方法常量
const (
	// CPU测试方法
	CpuMethodSysbench  = "sysbench"
	CpuMethodGeekbench = "geekbench"
	CpuMethodWinsat    = "winsat"
	
	// 内存测试方法
	MemoryMethodStream   = "stream"
	MemoryMethodSysbench = "sysbench"
	MemoryMethodDD       = "dd"
	MemoryMethodWinsat   = "winsat"
	
	// 硬盘测试方法
	DiskMethodFio    = "fio"
	DiskMethodDD     = "dd"
	DiskMethodWinsat = "winsat"
	
	// 线程模式
	ThreadModeSingle = "single"
	ThreadModeMulti  = "multi"
	
	// 语言选项
	LanguageZH = "zh"
	LanguageEN = "en"
	
	// IP检测类型
	CheckTypeIPv4 = "ipv4"
	CheckTypeIPv6 = "ipv6"
	CheckTypeAuto = "auto"
	
	// 测速平台
	PlatformCN  = "cn"
	PlatformNet = "net"
	
	// 运营商类型
	OperatorCMCC   = "cmcc"   // 中国移动
	OperatorCU     = "cu"     // 中国联通
	OperatorCT     = "ct"     // 中国电信
	OperatorGlobal = "global" // 全球节点
	OperatorOther  = "other"  // 其他
	OperatorHK     = "hk"     // 香港
	OperatorTW     = "tw"     // 台湾
	OperatorJP     = "jp"     // 日本
	OperatorSG     = "sg"     // 新加坡
)
