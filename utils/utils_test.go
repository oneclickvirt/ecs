package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	runewidth "github.com/mattn/go-runewidth"
)

func TestCaptureOutputReservesLeadingCell(t *testing.T) {
	output := CaptureOutput(func() {
		fmt.Print("header\nline\n existing\n\n")
	})
	want := " header\n line\n existing\n\n"
	if output != want {
		t.Fatalf("CaptureOutput() = %q, want %q", output, want)
	}
}

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

// TestPrintCenteredTitle_Width verifies that PrintCenteredTitle produces lines
// whose visual display width equals the requested width for both ASCII-only and
// CJK titles (CJK characters each occupy 2 terminal columns).
func TestPrintCenteredTitle_Width(t *testing.T) {
	const width = 80
	titles := []string{
		"VPS融合怪测试",                // mixed CJK + ASCII
		"VPS Fusion Monster Test", // ASCII only
		"流媒体解锁检测",                 // CJK only
		"CPU Performance Test",    // ASCII only
	}

	for _, title := range titles {
		t.Run(title, func(t *testing.T) {
			// Redirect stdout to capture PrintCenteredTitle output.
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("os.Pipe: %v", err)
			}
			origStdout := os.Stdout
			os.Stdout = w

			PrintCenteredTitle(title, width)

			w.Close()
			os.Stdout = origStdout

			var buf strings.Builder
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("reading captured output: %v", err)
			}
			r.Close()

			line := strings.TrimRight(buf.String(), "\n")
			got := runewidth.StringWidth(line)
			if got != width {
				t.Errorf("title %q: visual width = %d, want %d (line: %q)", title, got, width, line)
			}
		})
	}
}
