package network1

import (
	"fmt"
	"testing"
)

func TestIpv4SecurityCheck(t *testing.T) {
	// 单项测试
	//result1, _ := Ipv4SecurityCheck("8.8.8.8", nil, "zh")
	//fmt.Println(result1)
	//result2, _ := Ipv6SecurityCheck("2001:4860:4860::8844", nil, "zh")
	//fmt.Println(result2)

	// 全项测试
	ipInfo, securityInfo, _ := NetworkCheck("both", true, "zh")
	fmt.Println("--------------------------------------------------")
	fmt.Printf(ipInfo)
	fmt.Println("--------------------------------------------------")
	fmt.Printf(securityInfo)
	fmt.Println("--------------------------------------------------")
}
