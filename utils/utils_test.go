package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestCheckPublicAccess(t *testing.T) {
	timeout := 3 * time.Second
	result := CheckPublicAccess(timeout)
	if result.Connected {
		fmt.Printf("✅ 本机有公网连接，类型: %s\n", result.StackType)
	} else {
		fmt.Println("❌ 本机未检测到公网连接")
	}
}
