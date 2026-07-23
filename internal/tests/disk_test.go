package tests

import "testing"

func TestDiskUnavailableMessageUsesRequestedLanguage(t *testing.T) {
	if got := diskUnavailableMessage("EN"); got != "Disk test unavailable\n" {
		t.Fatalf("English disk fallback = %q", got)
	}
	if got := diskUnavailableMessage("zh"); got != "硬盘测试不可用\n" {
		t.Fatalf("Chinese disk fallback = %q", got)
	}
}
