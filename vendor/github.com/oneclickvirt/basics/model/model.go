package model

const BasicsVersion = "v0.0.6"

var EnableLoger bool

var MacOSInfo []string

type IpInfo struct {
	Ip      string
	ASN     string
	Org     string
	Country string
	Region  string
	City    string
}

type SecurityScore struct {
	Tag                    string
	Reputation             *int
	TrustScore             *int
	VpnScore               *int
	ProxyScore             *int
	CommunityVoteHarmless  *int
	CommunityVoteMalicious *int
	CloudFlareRisk         *int // 还没有加入
	ThreatScore            *int
	FraudScore             *int
	AbuseScore             *int
	HarmlessnessRecords    *int
	MaliciousRecords       *int
	SuspiciousRecords      *int
	NoRecords              *int
}

type SecurityInfo struct {
	Tag                string
	ASNAbuseScore      string // 这三个实际是得分类型，但由于是字符串所以还在这解析
	CompannyAbuseScore string
	ThreatLevel        string
	UsageType          string // connection_type、usage_type、asn_type
	CompanyType        string // company type
	IsCloudProvider    string
	IsDatacenter       string // datacenter、server、hosting
	IsMobile           string
	IsProxy            string // Public Proxy、Web Proxy
	IsVpn              string
	IsTor              string
	IsTorExit          string
	IsCrawler          string
	IsAnonymous        string
	IsAttacker         string
	IsAbuser           string
	IsThreat           string
	IsRelay            string // icloud_relay、is_relay
	IsBogon            string
	IsBot              string // Search Engine Robot
}

// TranslationMap 定义英文到中文的映射表
var TranslationMap = map[string]string{
	"Reputation":             "声誉(越高越好)",
	"TrustScore":             "信任得分(越高越好)",
	"VpnScore":               "VPN得分(越低越好)",
	"ProxyScore":             "代理得分(越低越好)",
	"CommunityVoteHarmless":  "社区投票-无害",
	"CommunityVoteMalicious": "社区投票-恶意",
	"CloudFlareRisk":         "CloudFlare风险(越低越好)",
	"ThreatScore":            "威胁得分(越低越好)",
	"FraudScore":             "欺诈得分(越低越好)",
	"AbuseScore":             "滥用得分(越低越好)",
	"HarmlessnessRecords":    "无害记录数",
	"MaliciousRecords":       "恶意记录数",
	"SuspiciousRecords":      "可疑记录数",
	"NoRecords":              "无记录数",
	"ASNAbuseScore":          "ASN滥用得分(越低越好)",
	"CompannyAbuseScore":     "公司滥用得分(越低越好)",
	"ThreatLevel":            "威胁级别",
	"UsageType":              "使用类型",
	"CompanyType":            "公司类型",
	"IsCloudProvider":        "是否云提供商",
	"IsDatacenter":           "是否数据中心",
	"IsMobile":               "是否移动设备",
	"IsProxy":                "是否代理",
	"IsVpn":                  "是否VPN",
	"IsTor":                  "是否Tor",
	"IsTorExit":              "是否Tor出口",
	"IsCrawler":              "是否网络爬虫",
	"IsAnonymous":            "是否匿名",
	"IsAttacker":             "是否攻击者",
	"IsAbuser":               "是否滥用者",
	"IsThreat":               "是否威胁",
	"IsRelay":                "是否中继",
	"IsBogon":                "是否Bogon",
	"IsBot":                  "是否机器人",
}

type CpuInfo struct {
	CpuModel string
	CpuCores string
	CpuCache string
	CpuAesNi string
	CpuVAH   string
}

type MemoryInfo struct {
	MemoryUsage string
	MemoryTotal string
	SwapUsage   string
	SwapTotal   string
}

type DiskInfo struct {
	DiskUsage  string
	DiskTotal  string
	Percentage string
	BootPath   string
}

type SystemInfo struct {
	CpuInfo
	MemoryInfo
	DiskInfo
	Platform              string // 系统名字 Distro1
	PlatformVersion       string // 系统版本 Distro2
	Kernel                string // 系统内核
	Arch                  string //
	Uptime                string // 正常运行时间
	TimeZone              string // 系统时区
	VmType                string // 虚拟化架构
	Load                  string // load1 load2 load3
	NatType               string // stun
	VirtioBalloon         string // 气球驱动
	KSM                   string // 内存合并
	TcpAccelerationMethod string // TCP拥塞控制
}

type Win32_Processor struct {
	L2CacheSize uint32
	L3CacheSize uint32
}

type Win32_ComputerSystem struct {
	SystemType string
}

type Win32_OperatingSystem struct {
	BuildType string
}

type Win32_TimeZone struct {
	Caption string
}
