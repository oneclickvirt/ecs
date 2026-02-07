package params

import (
	"flag"
	"fmt"
)

// Config holds all configuration parameters
type Config struct {
	EcsVersion           string
	MenuMode             bool
	OnlyChinaTest        bool
	Input                string
	Choice               string
	ShowVersion          bool
	EnableLogger         bool
	Language             string
	CpuTestMethod        string
	CpuTestThreadMode    string
	MemoryTestMethod     string
	DiskTestMethod       string
	DiskTestPath         string
	DiskMultiCheck       bool
	Nt3CheckType         string
	Nt3Location          string
	SpNum                int
	Width                int
	BasicStatus          bool
	CpuTestStatus        bool
	MemoryTestStatus     bool
	DiskTestStatus       bool
	UtTestStatus         bool
	SecurityTestStatus   bool
	EmailTestStatus      bool
	BacktraceStatus      bool
	Nt3Status            bool
	SpeedTestStatus      bool
	PingTestStatus       bool
	TgdcTestStatus       bool
	WebTestStatus        bool
	AutoChangeDiskMethod bool
	FilePath             string
	EnableUpload         bool
	OnlyIpInfoCheck      bool
	Help                 bool
	Finish               bool
	UserSetFlags         map[string]bool
	GoecsFlag            *flag.FlagSet
}

// NewConfig creates a new Config with default values
func NewConfig(version string) *Config {
	return &Config{
		EcsVersion:           version,
		MenuMode:             true,
		OnlyChinaTest:        false,
		Input:                "",
		Choice:               "",
		ShowVersion:          false,
		EnableLogger:         false,
		Language:             "zh",
		CpuTestMethod:        "sysbench",
		CpuTestThreadMode:    "multi",
		MemoryTestMethod:     "stream",
		DiskTestMethod:       "fio",
		DiskTestPath:         "",
		DiskMultiCheck:      false,
		Nt3CheckType:         "ipv4",
		SpNum:                2,
		Width:                82,
		BasicStatus:          true,
		CpuTestStatus:        true,
		MemoryTestStatus:     true,
		DiskTestStatus:       true,
		UtTestStatus:         true,
		SecurityTestStatus:   false,
		EmailTestStatus:      true,
		BacktraceStatus:      true,
		Nt3Status:            true,
		SpeedTestStatus:      true,
		PingTestStatus:       false,
		TgdcTestStatus:       false,
		WebTestStatus:        false,
		AutoChangeDiskMethod: true,
		FilePath:             "goecs.txt",
		EnableUpload:         true,
		OnlyIpInfoCheck:      false,
		Help:                 false,
		Finish:               false,
		UserSetFlags:         make(map[string]bool),
		GoecsFlag:            flag.NewFlagSet("goecs", flag.ContinueOnError),
	}
}

// ParseFlags parses command line flags
func (c *Config) ParseFlags(args []string) {
	c.GoecsFlag.BoolVar(&c.Help, "h", false, "Show help information")
	c.GoecsFlag.BoolVar(&c.Help, "help", false, "Show help information")
	c.GoecsFlag.BoolVar(&c.ShowVersion, "v", false, "Display version information")
	c.GoecsFlag.BoolVar(&c.ShowVersion, "version", false, "Display version information")
	c.GoecsFlag.BoolVar(&c.MenuMode, "menu", true, "Enable/Disable menu mode, disable example: -menu=false")
	c.GoecsFlag.StringVar(&c.Language, "lang", "zh", "Set language (supported: en, zh)")
	c.GoecsFlag.StringVar(&c.Language, "l", "zh", "Set language (supported: en, zh)")
	c.GoecsFlag.BoolVar(&c.BasicStatus, "basic", true, "Enable/Disable basic test")
	c.GoecsFlag.BoolVar(&c.CpuTestStatus, "cpu", true, "Enable/Disable CPU test")
	c.GoecsFlag.BoolVar(&c.MemoryTestStatus, "memory", true, "Enable/Disable memory test")
	c.GoecsFlag.BoolVar(&c.DiskTestStatus, "disk", true, "Enable/Disable disk test")
	c.GoecsFlag.BoolVar(&c.UtTestStatus, "ut", true, "Enable/Disable unlock media test")
	c.GoecsFlag.BoolVar(&c.SecurityTestStatus, "security", false, "Enable/Disable security test")
	c.GoecsFlag.BoolVar(&c.EmailTestStatus, "email", true, "Enable/Disable email port test")
	c.GoecsFlag.BoolVar(&c.BacktraceStatus, "backtrace", true, "Enable/Disable backtrace test (in 'en' language or on windows it always false)")
	c.GoecsFlag.BoolVar(&c.Nt3Status, "nt3", true, "Enable/Disable NT3 test (in 'en' language or on windows it always false)")
	c.GoecsFlag.BoolVar(&c.SpeedTestStatus, "speed", true, "Enable/Disable speed test")
	c.GoecsFlag.BoolVar(&c.PingTestStatus, "ping", false, "Enable/Disable ping test")
	c.GoecsFlag.BoolVar(&c.TgdcTestStatus, "tgdc", false, "Enable/Disable Telegram DC test")
	c.GoecsFlag.BoolVar(&c.WebTestStatus, "web", false, "Enable/Disable popular websites test")
	c.GoecsFlag.StringVar(&c.CpuTestMethod, "cpum", "sysbench", "Set CPU test method (supported: sysbench, geekbench, winsat)")
	c.GoecsFlag.StringVar(&c.CpuTestMethod, "cpu-method", "sysbench", "Set CPU test method (supported: sysbench, geekbench, winsat)")
	c.GoecsFlag.StringVar(&c.CpuTestThreadMode, "cput", "multi", "Set CPU test thread mode (supported: single, multi)")
	c.GoecsFlag.StringVar(&c.CpuTestThreadMode, "cpu-thread", "multi", "Set CPU test thread mode (supported: single, multi)")
	c.GoecsFlag.StringVar(&c.MemoryTestMethod, "memorym", "stream", "Set memory test method (supported: stream, sysbench, dd, winsat, auto)")
	c.GoecsFlag.StringVar(&c.MemoryTestMethod, "memory-method", "stream", "Set memory test method (supported: stream, sysbench, dd, winsat, auto)")
	c.GoecsFlag.StringVar(&c.DiskTestMethod, "diskm", "fio", "Set disk test method (supported: fio, dd, winsat)")
	c.GoecsFlag.StringVar(&c.DiskTestMethod, "disk-method", "fio", "Set disk test method (supported: fio, dd, winsat)")
	c.GoecsFlag.StringVar(&c.DiskTestPath, "diskp", "", "Set disk test path, e.g., -diskp /root")
	c.GoecsFlag.BoolVar(&c.DiskMultiCheck, "diskmc", false, "Enable/Disable multiple disk checks, e.g., -diskmc=false")
	c.GoecsFlag.StringVar(&c.Nt3Location, "nt3loc", "GZ", "Specify NT3 test location (supported: GZ, SH, BJ, CD, ALL for Guangzhou, Shanghai, Beijing, Chengdu and all)")
	c.GoecsFlag.StringVar(&c.Nt3Location, "nt3-location", "GZ", "Specify NT3 test location (supported: GZ, SH, BJ, CD, ALL for Guangzhou, Shanghai, Beijing, Chengdu and all)")
	c.GoecsFlag.StringVar(&c.Nt3CheckType, "nt3t", "ipv4", "Set NT3 test type (supported: both, ipv4, ipv6)")
	c.GoecsFlag.StringVar(&c.Nt3CheckType, "nt3-type", "ipv4", "Set NT3 test type (supported: both, ipv4, ipv6)")
	c.GoecsFlag.IntVar(&c.SpNum, "spnum", 2, "Set the number of servers per operator for speed test")
	c.GoecsFlag.BoolVar(&c.EnableLogger, "log", false, "Enable/Disable logging in the current path")
	c.GoecsFlag.BoolVar(&c.EnableUpload, "upload", true, "Enable/Disable upload the result")
	c.GoecsFlag.Parse(args)

	c.GoecsFlag.Visit(func(f *flag.Flag) {
		c.UserSetFlags[f.Name] = true
	})
}

// HandleHelpAndVersion handles help and version flags
func (c *Config) HandleHelpAndVersion(programName string) bool {
	if c.Help {
		fmt.Printf("Usage: %s [options]\n", programName)
		c.GoecsFlag.PrintDefaults()
		return true
	}
	if c.ShowVersion {
		fmt.Println(c.EcsVersion)
		return true
	}
	return false
}

// SaveUserSetParams saves user-set parameters
func (c *Config) SaveUserSetParams() map[string]interface{} {
	saved := make(map[string]interface{})

	if c.UserSetFlags["basic"] {
		saved["basic"] = c.BasicStatus
	}
	if c.UserSetFlags["cpu"] {
		saved["cpu"] = c.CpuTestStatus
	}
	if c.UserSetFlags["memory"] {
		saved["memory"] = c.MemoryTestStatus
	}
	if c.UserSetFlags["disk"] {
		saved["disk"] = c.DiskTestStatus
	}
	if c.UserSetFlags["ut"] {
		saved["ut"] = c.UtTestStatus
	}
	if c.UserSetFlags["security"] {
		saved["security"] = c.SecurityTestStatus
	}
	if c.UserSetFlags["email"] {
		saved["email"] = c.EmailTestStatus
	}
	if c.UserSetFlags["backtrace"] {
		saved["backtrace"] = c.BacktraceStatus
	}
	if c.UserSetFlags["nt3"] {
		saved["nt3"] = c.Nt3Status
	}
	if c.UserSetFlags["speed"] {
		saved["speed"] = c.SpeedTestStatus
	}
	if c.UserSetFlags["ping"] {
		saved["ping"] = c.PingTestStatus
	}
	if c.UserSetFlags["tgdc"] {
		saved["tgdc"] = c.TgdcTestStatus
	}
	if c.UserSetFlags["web"] {
		saved["web"] = c.WebTestStatus
	}
	if c.UserSetFlags["cpum"] || c.UserSetFlags["cpu-method"] {
		saved["cpum"] = c.CpuTestMethod
	}
	if c.UserSetFlags["cput"] || c.UserSetFlags["cpu-thread"] {
		saved["cput"] = c.CpuTestThreadMode
	}
	if c.UserSetFlags["memorym"] || c.UserSetFlags["memory-method"] {
		saved["memorym"] = c.MemoryTestMethod
	}
	if c.UserSetFlags["diskm"] || c.UserSetFlags["disk-method"] {
		saved["diskm"] = c.DiskTestMethod
	}
	if c.UserSetFlags["diskp"] {
		saved["diskp"] = c.DiskTestPath
	}
	if c.UserSetFlags["diskmc"] {
		saved["diskmc"] = c.DiskMultiCheck
	}
	if c.UserSetFlags["nt3loc"] || c.UserSetFlags["nt3-location"] {
		saved["nt3loc"] = c.Nt3Location
	}
	if c.UserSetFlags["nt3t"] || c.UserSetFlags["nt3-type"] {
		saved["nt3t"] = c.Nt3CheckType
	}
	if c.UserSetFlags["spnum"] {
		saved["spnum"] = c.SpNum
	}

	return saved
}

// RestoreUserSetParams restores user-set parameters
func (c *Config) RestoreUserSetParams(saved map[string]interface{}) {
	if val, ok := saved["basic"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.BasicStatus = boolVal
		}
	}
	if val, ok := saved["cpu"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.CpuTestStatus = boolVal
		}
	}
	if val, ok := saved["memory"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.MemoryTestStatus = boolVal
		}
	}
	if val, ok := saved["disk"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.DiskTestStatus = boolVal
		}
	}
	if val, ok := saved["ut"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.UtTestStatus = boolVal
		}
	}
	if val, ok := saved["security"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.SecurityTestStatus = boolVal
		}
	}
	if val, ok := saved["email"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.EmailTestStatus = boolVal
		}
	}
	if val, ok := saved["backtrace"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.BacktraceStatus = boolVal
		}
	}
	if val, ok := saved["nt3"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.Nt3Status = boolVal
		}
	}
	if val, ok := saved["speed"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.SpeedTestStatus = boolVal
		}
	}
	if val, ok := saved["ping"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.PingTestStatus = boolVal
		}
	}
	if val, ok := saved["tgdc"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.TgdcTestStatus = boolVal
		}
	}
	if val, ok := saved["web"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.WebTestStatus = boolVal
		}
	}
	if val, ok := saved["cpum"]; ok {
		if strVal, ok := val.(string); ok {
			c.CpuTestMethod = strVal
		}
	}
	if val, ok := saved["cput"]; ok {
		if strVal, ok := val.(string); ok {
			c.CpuTestThreadMode = strVal
		}
	}
	if val, ok := saved["memorym"]; ok {
		if strVal, ok := val.(string); ok {
			c.MemoryTestMethod = strVal
		}
	}
	if val, ok := saved["diskm"]; ok {
		if strVal, ok := val.(string); ok {
			c.DiskTestMethod = strVal
		}
	}
	if val, ok := saved["diskp"]; ok {
		if strVal, ok := val.(string); ok {
			c.DiskTestPath = strVal
		}
	}
	if val, ok := saved["diskmc"]; ok {
		if boolVal, ok := val.(bool); ok {
			c.DiskMultiCheck = boolVal
		}
	}
	if val, ok := saved["nt3loc"]; ok {
		if c.Choice != "10" {
			if strVal, ok := val.(string); ok {
				c.Nt3Location = strVal
			}
		}
	}
	if val, ok := saved["nt3t"]; ok {
		if strVal, ok := val.(string); ok {
			c.Nt3CheckType = strVal
		}
	}
	if val, ok := saved["spnum"]; ok {
		if intVal, ok := val.(int); ok {
			c.SpNum = intVal
		}
	}

	c.ValidateParams()
}

// ValidateParams validates parameter values
func (c *Config) ValidateParams() {
	validCpuMethods := map[string]bool{"sysbench": true, "geekbench": true, "winsat": true}
	if !validCpuMethods[c.CpuTestMethod] {
		if c.Language == "zh" {
			fmt.Printf("警告: CPU测试方法 '%s' 无效，使用默认值 'sysbench'\n", c.CpuTestMethod)
		} else {
			fmt.Printf("Warning: Invalid CPU test method '%s', using default 'sysbench'\n", c.CpuTestMethod)
		}
		c.CpuTestMethod = "sysbench"
	}

	validThreadModes := map[string]bool{"single": true, "multi": true}
	if !validThreadModes[c.CpuTestThreadMode] {
		if c.Language == "zh" {
			fmt.Printf("警告: CPU线程模式 '%s' 无效，使用默认值 'multi'\n", c.CpuTestThreadMode)
		} else {
			fmt.Printf("Warning: Invalid CPU thread mode '%s', using default 'multi'\n", c.CpuTestThreadMode)
		}
		c.CpuTestThreadMode = "multi"
	}

	validMemoryMethods := map[string]bool{"stream": true, "sysbench": true, "dd": true, "winsat": true, "auto": true}
	if !validMemoryMethods[c.MemoryTestMethod] {
		if c.Language == "zh" {
			fmt.Printf("警告: 内存测试方法 '%s' 无效，使用默认值 'stream'\n", c.MemoryTestMethod)
		} else {
			fmt.Printf("Warning: Invalid memory test method '%s', using default 'stream'\n", c.MemoryTestMethod)
		}
		c.MemoryTestMethod = "stream"
	}

	validDiskMethods := map[string]bool{"fio": true, "dd": true, "winsat": true}
	if !validDiskMethods[c.DiskTestMethod] {
		if c.Language == "zh" {
			fmt.Printf("警告: 磁盘测试方法 '%s' 无效，使用默认值 'fio'\n", c.DiskTestMethod)
		} else {
			fmt.Printf("Warning: Invalid disk test method '%s', using default 'fio'\n", c.DiskTestMethod)
		}
		c.DiskTestMethod = "fio"
	}

	validNt3Locations := map[string]bool{"GZ": true, "SH": true, "BJ": true, "CD": true, "ALL": true}
	if !validNt3Locations[c.Nt3Location] {
		if c.Language == "zh" {
			fmt.Printf("警告: NT3测试位置 '%s' 无效，使用默认值 'GZ'\n", c.Nt3Location)
		} else {
			fmt.Printf("Warning: Invalid NT3 location '%s', using default 'GZ'\n", c.Nt3Location)
		}
		c.Nt3Location = "GZ"
	}

	validNt3Types := map[string]bool{"both": true, "ipv4": true, "ipv6": true}
	if !validNt3Types[c.Nt3CheckType] {
		if c.Language == "zh" {
			fmt.Printf("警告: NT3测试类型 '%s' 无效，使用默认值 'ipv4'\n", c.Nt3CheckType)
		} else {
			fmt.Printf("Warning: Invalid NT3 check type '%s', using default 'ipv4'\n", c.Nt3CheckType)
		}
		c.Nt3CheckType = "ipv4"
	}

	if c.SpNum < 0 {
		if c.Language == "zh" {
			fmt.Printf("警告: 测速节点数量 '%d' 无效，使用默认值 2\n", c.SpNum)
		} else {
			fmt.Printf("Warning: Invalid speed test node count '%d', using default 2\n", c.SpNum)
		}
		c.SpNum = 2
	}

	validLanguages := map[string]bool{"zh": true, "en": true}
	if !validLanguages[c.Language] {
		fmt.Printf("Warning: Invalid language '%s', using default 'zh'\n", c.Language)
		c.Language = "zh"
	}
}
