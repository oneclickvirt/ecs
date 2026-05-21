package utils

import (
	"strings"
	"testing"
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

func TestBasicsAndSecurityCheck_SecurityDisabled(t *testing.T) {
	originalFn := networkCheckFn
	t.Cleanup(func() { networkCheckFn = originalFn })

	var receivedSecurityStatus bool
	networkCheckFn = func(checkType string, securityCheckStatus bool, language string) (string, string, string, string, error) {
		receivedSecurityStatus = securityCheckStatus
		return "1.1.1.1", "", "IPV4: 1.1.1.1\n", "", nil
	}

	_, _, basicInfo, securityInfo, nt3CheckType := BasicsAndSecurityCheck("zh", "ipv4", false)
	if receivedSecurityStatus {
		t.Fatalf("security check should remain disabled")
	}
	if securityInfo != "" {
		t.Fatalf("expected empty security output when disabled, got: %q", securityInfo)
	}
	if !strings.Contains(basicInfo, "IPV4: 1.1.1.1") {
		t.Fatalf("expected basic info to include ipv4 output, got: %q", basicInfo)
	}
	if nt3CheckType != "ipv4" {
		t.Fatalf("expected nt3CheckType to remain ipv4, got: %q", nt3CheckType)
	}
}

func TestBasicsAndSecurityCheck_SecurityEnabled(t *testing.T) {
	originalFn := networkCheckFn
	t.Cleanup(func() { networkCheckFn = originalFn })

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
		t.Fatalf("security check should run when enabled")
	}
	if !strings.Contains(securityInfo, "mock-security") {
		t.Fatalf("expected security output, got: %q", securityInfo)
	}
}
