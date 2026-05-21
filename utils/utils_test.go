package utils

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// func TestCheckPublicAccess(t *testing.T) {
// 	timeout := 3 * time.Second
// 	result := CheckPublicAccess(timeout)
// 	if result.Connected {
// 		fmt.Print("✅ 本机有公网连接，类型: %s\n", result.StackType)
// 	} else {
// 		fmt.Println("❌ 本机未检测到公网连接")
// 	}
// }

func TestBasicsAndSecurityCheck(t *testing.T) {
	timeout := 3 * time.Second
	result := CheckPublicAccess(timeout)
	if result.Connected {
		fmt.Printf("✅ 本机有公网连接，类型: %s\n", result.StackType)
	} else {
		fmt.Println("❌ 本机未检测到公网连接")
	}
	_, _, basicInfo, securityInfo, nt3CheckType := BasicsAndSecurityCheck("zh", "ipv4", false)
	fmt.Println(basicInfo)
	fmt.Println(securityInfo)
	fmt.Println(nt3CheckType)
}

func TestBasicsAndSecurityCheck_ProtectedModeByDefault(t *testing.T) {
	originalFn := networkCheckFn
	t.Cleanup(func() {
		networkCheckFn = originalFn
		_ = os.Unsetenv(unsafeSecurityCheckEnv)
	})

	_ = os.Unsetenv(unsafeSecurityCheckEnv)
	calledSecurityTrue := false
	networkCheckFn = func(checkType string, securityCheckStatus bool, language string) (string, string, string, string, error) {
		if securityCheckStatus {
			calledSecurityTrue = true
		}
		return "1.1.1.1", "", "IPV4: 1.1.1.1\n", "", nil
	}

	_, _, _, securityInfo, _ := BasicsAndSecurityCheck("zh", "ipv4", true)
	if calledSecurityTrue {
		t.Fatalf("security check should stay disabled in protected mode")
	}
	if !strings.Contains(securityInfo, "保护模式") {
		t.Fatalf("unexpected fallback message: %q", securityInfo)
	}
}

func TestBasicsAndSecurityCheck_UnsafeModeEnabled(t *testing.T) {
	originalFn := networkCheckFn
	t.Cleanup(func() {
		networkCheckFn = originalFn
		_ = os.Unsetenv(unsafeSecurityCheckEnv)
	})

	if err := os.Setenv(unsafeSecurityCheckEnv, "1"); err != nil {
		t.Fatalf("set env failed: %v", err)
	}
	calledSecurityTrue := false
	networkCheckFn = func(checkType string, securityCheckStatus bool, language string) (string, string, string, string, error) {
		if securityCheckStatus {
			calledSecurityTrue = true
			return "1.1.1.1", "", "IPV4: 1.1.1.1\n", "mock-security\n", nil
		}
		return "1.1.1.1", "", "IPV4: 1.1.1.1\n", "", nil
	}

	_, _, _, securityInfo, _ := BasicsAndSecurityCheck("en", "ipv4", true)
	if !calledSecurityTrue {
		t.Fatalf("security check should run when unsafe mode is enabled")
	}
	if !strings.Contains(securityInfo, "mock-security") {
		t.Fatalf("expected unsafe-mode security output, got: %q", securityInfo)
	}
}
