package backtrace

import (
	"testing"
)

//func TestGeneratePrefixMap(t *testing.T) {
//	prefix := "223.119.8.0/21"
//	prefixList := GeneratePrefixList(prefix)
//	if prefixList != nil {
//		// 打印生成的IP地址前缀列表
//		for _, ip := range prefixList {
//			fmt.Println(ip)
//		}
//	}
//}

// 本包仅测试，无实际使用
func TestBackTrace(t *testing.T) {
	BackTrace(false)
}
