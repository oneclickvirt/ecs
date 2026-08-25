package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	runewidth "github.com/mattn/go-runewidth"
	dnsresolver "github.com/oneclickvirt/basics/network/resolver"
	butils "github.com/oneclickvirt/basics/utils"
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

func TestCaptureOutputSanitizesCompleteLinesBeforeDisplayAndRetention(t *testing.T) {
	SetOutputSanitizer(func(value string) string {
		return strings.ReplaceAll(value, "https://private.example/list?token=secret", "[remote-url]")
	})
	t.Cleanup(func() { SetOutputSanitizer(nil) })
	output := CaptureOutput(func() {
		fmt.Print("registry https://private.")
		fmt.Print("example/list?token=secret\n")
	})
	if output != " registry [remote-url]\n" {
		t.Fatalf("CaptureOutput() = %q", output)
	}
}

func TestDNSStatusText(t *testing.T) {
	tests := []struct {
		name     string
		preCheck NetCheckResult
		language string
		want     string
	}{
		{name: "not configured", want: ""},
		{
			name:     "system resolver",
			preCheck: NetCheckResult{DNSActive: "system"},
			language: "en",
			want:     "[DNS] Using the system resolver",
		},
		{
			name:     "automatic fallback",
			preCheck: NetCheckResult{DNSActive: "doh", DNSFallback: true, DNSProvider: "Cloudflare"},
			language: "en",
			want:     "[DNS] System DNS unavailable; built-in DoH fallback active (Cloudflare)",
		},
		{
			name:     "automatic DoT fallback",
			preCheck: NetCheckResult{DNSActive: "dot", DNSFallback: true, DNSProvider: "Google"},
			language: "en",
			want:     "[DNS] System DNS unavailable; built-in DoT fallback active (Google)",
		},
		{
			name:     "inconclusive system resolver",
			preCheck: NetCheckResult{DNSActive: "system", DNSReason: "system DNS probe inconclusive; preserving system resolver"},
			language: "en",
			want:     "[DNS] System resolver probe was inconclusive; preserving system resolver",
		},
		{
			name:     "unavailable",
			preCheck: NetCheckResult{DNSActive: "unavailable", DNSReason: "DoH probe failed"},
			language: "en",
			want:     "[DNS] Resolver unavailable: DoH probe failed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := DNSStatusText(test.preCheck, test.language); got != test.want {
				t.Fatalf("DNSStatusText() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestConfigureDNSUsesNormalPathWithoutBootstrapProbe(t *testing.T) {
	originalConfigure := dnsConfigureFn
	originalShutdown := dnsShutdownFn
	originalBootstrapReachable := dnsBootstrapReachableFn
	t.Cleanup(func() {
		dnsConfigureFn = originalConfigure
		dnsShutdownFn = originalShutdown
		dnsBootstrapReachableFn = originalBootstrapReachable
	})
	configureCalls := 0
	shutdownCalls := 0
	bootstrapCalls := 0
	dnsConfigureFn = func(_ context.Context, config dnsresolver.Config) dnsresolver.Status {
		configureCalls++
		if config.Mode != dnsresolver.ModeAuto {
			t.Fatalf("resolver mode = %q, want auto", config.Mode)
		}
		return dnsresolver.Status{Requested: dnsresolver.ModeAuto, Active: dnsresolver.ModeSystem, SystemAvailable: true}
	}
	dnsShutdownFn = func() { shutdownCalls++ }
	dnsBootstrapReachableFn = func(context.Context, dnsresolver.Config) (string, bool) {
		bootstrapCalls++
		return "", false
	}
	preCheck := &NetCheckResult{Connected: true}

	status := ConfigureDNS(context.Background(), "auto", preCheck)
	if configureCalls != 1 || bootstrapCalls != 0 || shutdownCalls != 0 {
		t.Fatalf("DNS calls = configure:%d bootstrap:%d shutdown:%d, want configure:1 bootstrap:0 shutdown:0", configureCalls, bootstrapCalls, shutdownCalls)
	}
	if status.Active != dnsresolver.ModeSystem || preCheck.DNSActive != string(dnsresolver.ModeSystem) {
		t.Fatalf("normal DNS status = %#v, precheck = %#v", status, preCheck)
	}
}

func TestConfigureDNSSkipsAutoFallbackWhenBootstrapIsUnreachable(t *testing.T) {
	originalConfigure := dnsConfigureFn
	originalShutdown := dnsShutdownFn
	originalBootstrapReachable := dnsBootstrapReachableFn
	t.Cleanup(func() {
		dnsConfigureFn = originalConfigure
		dnsShutdownFn = originalShutdown
		dnsBootstrapReachableFn = originalBootstrapReachable
	})
	configureCalls := 0
	shutdownCalls := 0
	bootstrapCalls := 0
	dnsConfigureFn = func(context.Context, dnsresolver.Config) dnsresolver.Status {
		configureCalls++
		return dnsresolver.Status{}
	}
	dnsShutdownFn = func() { shutdownCalls++ }
	dnsBootstrapReachableFn = func(context.Context, dnsresolver.Config) (string, bool) {
		bootstrapCalls++
		return "", false
	}
	preCheck := &NetCheckResult{Connected: false}

	status := ConfigureDNS(context.Background(), "auto", preCheck)
	if configureCalls != 0 || bootstrapCalls != 1 || shutdownCalls != 1 {
		t.Fatalf("DNS calls = configure:%d bootstrap:%d shutdown:%d, want configure:0 bootstrap:1 shutdown:1", configureCalls, bootstrapCalls, shutdownCalls)
	}
	if status.Active != dnsresolver.ModeUnavailable || preCheck.DNSActive != string(dnsresolver.ModeUnavailable) {
		t.Fatalf("offline DNS status = %#v, precheck = %#v", status, preCheck)
	}
}

func TestConfigureDNSUsesAutoFallbackAfterReachableBootstrap(t *testing.T) {
	originalConfigure := dnsConfigureFn
	originalShutdown := dnsShutdownFn
	originalBootstrapReachable := dnsBootstrapReachableFn
	originalStackType := StackType
	originalBasicsStackType := butils.StackType
	t.Cleanup(func() {
		dnsConfigureFn = originalConfigure
		dnsShutdownFn = originalShutdown
		dnsBootstrapReachableFn = originalBootstrapReachable
		StackType = originalStackType
		butils.StackType = originalBasicsStackType
	})
	configureCalls := 0
	shutdownCalls := 0
	bootstrapCalls := 0
	dnsConfigureFn = func(_ context.Context, config dnsresolver.Config) dnsresolver.Status {
		configureCalls++
		if config.Mode != dnsresolver.ModeAuto {
			t.Fatalf("resolver mode = %q, want auto", config.Mode)
		}
		return dnsresolver.Status{Requested: dnsresolver.ModeAuto, Active: dnsresolver.ModeDoH, DoHAvailable: true, Fallback: true, Stack: "IPv4"}
	}
	dnsShutdownFn = func() { shutdownCalls++ }
	dnsBootstrapReachableFn = func(_ context.Context, config dnsresolver.Config) (string, bool) {
		bootstrapCalls++
		if config.Mode != dnsresolver.ModeAuto {
			t.Fatalf("bootstrap mode = %q, want auto", config.Mode)
		}
		return "IPv4", true
	}
	preCheck := &NetCheckResult{Connected: false, StackType: "None"}

	status := ConfigureDNS(context.Background(), "auto", preCheck)
	if configureCalls != 1 || bootstrapCalls != 1 || shutdownCalls != 0 {
		t.Fatalf("DNS calls = configure:%d bootstrap:%d shutdown:%d, want configure:1 bootstrap:1 shutdown:0", configureCalls, bootstrapCalls, shutdownCalls)
	}
	if status.Active != dnsresolver.ModeDoH || !preCheck.Connected || !preCheck.HasIPv4 || preCheck.HasIPv6 || preCheck.StackType != "IPv4" || butils.StackType != "IPv4" {
		t.Fatalf("successful auto fallback did not promote network state: status=%#v precheck=%#v", status, preCheck)
	}
}

func TestConfigureDNSAttemptsForcedDoHWithoutConnectivity(t *testing.T) {
	originalConfigure := dnsConfigureFn
	originalShutdown := dnsShutdownFn
	originalBootstrapReachable := dnsBootstrapReachableFn
	originalStackType := StackType
	originalBasicsStackType := butils.StackType
	t.Cleanup(func() {
		dnsConfigureFn = originalConfigure
		dnsShutdownFn = originalShutdown
		dnsBootstrapReachableFn = originalBootstrapReachable
		StackType = originalStackType
		butils.StackType = originalBasicsStackType
	})
	configureCalls := 0
	shutdownCalls := 0
	bootstrapCalls := 0
	dnsConfigureFn = func(_ context.Context, config dnsresolver.Config) dnsresolver.Status {
		configureCalls++
		if config.Mode != dnsresolver.ModeDoH {
			t.Fatalf("resolver mode = %q, want doh", config.Mode)
		}
		return dnsresolver.Status{Requested: dnsresolver.ModeDoH, Active: dnsresolver.ModeDoH, DoHAvailable: true, Stack: "IPv6"}
	}
	dnsShutdownFn = func() { shutdownCalls++ }
	dnsBootstrapReachableFn = func(context.Context, dnsresolver.Config) (string, bool) {
		bootstrapCalls++
		return "", false
	}
	preCheck := &NetCheckResult{Connected: false}

	status := ConfigureDNS(context.Background(), "doh", preCheck)
	if configureCalls != 1 || bootstrapCalls != 0 || shutdownCalls != 0 {
		t.Fatalf("DNS calls = configure:%d bootstrap:%d shutdown:%d, want configure:1 bootstrap:0 shutdown:0", configureCalls, bootstrapCalls, shutdownCalls)
	}
	if status.Active != dnsresolver.ModeDoH || preCheck.DNSActive != string(dnsresolver.ModeDoH) {
		t.Fatalf("forced DoH status = %#v, precheck = %#v", status, preCheck)
	}
	if !preCheck.Connected || !preCheck.HasIPv6 || preCheck.HasIPv4 || preCheck.StackType != "IPv6" || butils.StackType != "IPv6" {
		t.Fatalf("successful DoH did not promote network state: %#v", preCheck)
	}
}

func TestConfigureDNSAttemptsForcedDoTWithoutConnectivity(t *testing.T) {
	originalConfigure := dnsConfigureFn
	originalShutdown := dnsShutdownFn
	originalBootstrapReachable := dnsBootstrapReachableFn
	originalStackType := StackType
	originalBasicsStackType := butils.StackType
	t.Cleanup(func() {
		dnsConfigureFn = originalConfigure
		dnsShutdownFn = originalShutdown
		dnsBootstrapReachableFn = originalBootstrapReachable
		StackType = originalStackType
		butils.StackType = originalBasicsStackType
	})
	configureCalls := 0
	dnsConfigureFn = func(_ context.Context, config dnsresolver.Config) dnsresolver.Status {
		configureCalls++
		if config.Mode != dnsresolver.ModeDoT {
			t.Fatalf("resolver mode = %q, want dot", config.Mode)
		}
		return dnsresolver.Status{Requested: dnsresolver.ModeDoT, Active: dnsresolver.ModeDoT, DoTAvailable: true, Stack: "IPv4"}
	}
	dnsShutdownFn = func() { t.Fatal("forced DoT must not shut down a successful resolver") }
	dnsBootstrapReachableFn = func(context.Context, dnsresolver.Config) (string, bool) {
		t.Fatal("forced DoT must bypass the auto bootstrap gate")
		return "", false
	}
	preCheck := &NetCheckResult{Connected: false}
	status := ConfigureDNS(context.Background(), "dot", preCheck)
	if configureCalls != 1 || status.Active != dnsresolver.ModeDoT || preCheck.DNSActive != string(dnsresolver.ModeDoT) {
		t.Fatalf("forced DoT status = %#v, precheck = %#v", status, preCheck)
	}
	if !preCheck.Connected || !preCheck.HasIPv4 || preCheck.HasIPv6 || preCheck.StackType != "IPv4" {
		t.Fatalf("successful DoT did not promote network state: %#v", preCheck)
	}
}

func TestConfigureDNSDoesNotFallbackInSystemMode(t *testing.T) {
	originalConfigure := dnsConfigureFn
	originalShutdown := dnsShutdownFn
	originalBootstrapReachable := dnsBootstrapReachableFn
	t.Cleanup(func() {
		dnsConfigureFn = originalConfigure
		dnsShutdownFn = originalShutdown
		dnsBootstrapReachableFn = originalBootstrapReachable
	})
	configureCalls := 0
	shutdownCalls := 0
	bootstrapCalls := 0
	dnsConfigureFn = func(context.Context, dnsresolver.Config) dnsresolver.Status {
		configureCalls++
		return dnsresolver.Status{}
	}
	dnsShutdownFn = func() { shutdownCalls++ }
	dnsBootstrapReachableFn = func(context.Context, dnsresolver.Config) (string, bool) {
		bootstrapCalls++
		return "IPv4", true
	}

	status := ConfigureDNS(context.Background(), "system", &NetCheckResult{Connected: false})
	if configureCalls != 0 || bootstrapCalls != 0 || shutdownCalls != 1 {
		t.Fatalf("DNS calls = configure:%d bootstrap:%d shutdown:%d, want configure:0 bootstrap:0 shutdown:1", configureCalls, bootstrapCalls, shutdownCalls)
	}
	if status.Requested != dnsresolver.ModeSystem || status.Active != dnsresolver.ModeUnavailable {
		t.Fatalf("system-only status = %#v", status)
	}
}

func TestEnsureDNSReusesMatchingHealthyResolver(t *testing.T) {
	originalConfigure := dnsConfigureFn
	originalCurrent := dnsCurrentStatusFn
	t.Cleanup(func() {
		dnsConfigureFn = originalConfigure
		dnsCurrentStatusFn = originalCurrent
	})
	configureCalls := 0
	dnsConfigureFn = func(context.Context, dnsresolver.Config) dnsresolver.Status {
		configureCalls++
		return dnsresolver.Status{}
	}
	dnsCurrentStatusFn = func() dnsresolver.Status {
		return dnsresolver.Status{Requested: dnsresolver.ModeAuto, Active: dnsresolver.ModeSystem, SystemAvailable: true}
	}
	preCheck := &NetCheckResult{Connected: true}

	status := EnsureDNS(context.Background(), "auto", preCheck)
	if configureCalls != 0 {
		t.Fatalf("matching resolver was configured again %d times", configureCalls)
	}
	if status.Active != dnsresolver.ModeSystem || preCheck.DNSActive != string(dnsresolver.ModeSystem) {
		t.Fatalf("reused DNS status = %#v, precheck = %#v", status, preCheck)
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

func TestSecurityInfoCheckRunsOnlyDeferredSecurityProbe(t *testing.T) {
	originalFn := networkCheckFn
	t.Cleanup(func() { networkCheckFn = originalFn })

	calls := 0
	networkCheckFn = func(checkType string, securityCheckStatus bool, language string) (string, string, string, string, error) {
		calls++
		if checkType != "both" || !securityCheckStatus || language != "en" {
			t.Fatalf("unexpected deferred security request: type=%q enabled=%v language=%q", checkType, securityCheckStatus, language)
		}
		return "1.1.1.1", "", "ignored basic output", "deferred security output", nil
	}

	if got := SecurityInfoCheck("en"); got != "deferred security output" {
		t.Fatalf("SecurityInfoCheck() = %q", got)
	}
	if calls != 1 {
		t.Fatalf("network check calls = %d, want 1", calls)
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
