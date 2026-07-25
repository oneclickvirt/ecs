package tests

import (
	"strings"
	"testing"
)

func TestDiskUnavailableMessageUsesRequestedLanguage(t *testing.T) {
	if got := diskUnavailableMessage("EN"); got != "Disk benchmark returned no usable data.\n" {
		t.Fatalf("English disk fallback = %q", got)
	}
	if got := diskUnavailableMessage("zh"); got != "磁盘测试未返回可用的性能数据。\n" {
		t.Fatalf("Chinese disk fallback = %q", got)
	}
}

func TestDiskTestWithMethodsFallsBackAcrossLanguages(t *testing.T) {
	for _, language := range []string{"zh", "en"} {
		for _, selected := range []string{"fio", "dd"} {
			t.Run(language+"/"+selected, func(t *testing.T) {
				fioCalls, ddCalls := 0, 0
				fio := func(string, bool, string) string {
					fioCalls++
					if selected == "fio" {
						return ""
					}
					return "/ 4k 10.0 MB/s(1) 11.0 MB/s(2)\n"
				}
				dd := func(string, bool, string) string {
					ddCalls++
					if selected == "dd" {
						return "Test Path Block Direct Write(IOPS) Direct Read(IOPS)\n"
					}
					return "/ 4k 12.0 MB/s(1) 13.0 MB/s(2)\n"
				}
				method, output := diskTestWithMethods(language, selected, "/fixture", false, true, fio, dd)
				wantMethod := "dd"
				if selected == "dd" {
					wantMethod = "fio"
				}
				if method != wantMethod || !legacyDiskResultUsable(output) || fioCalls != 1 || ddCalls != 1 {
					t.Fatalf("fallback failed: method=%q output=%q fio=%d dd=%d", method, output, fioCalls, ddCalls)
				}
			})
		}
	}
}

func TestDiskTestWithMethodsRetainsFailureWhenBothMethodsFail(t *testing.T) {
	method, output := diskTestWithMethods("en", "dd", "/fixture", false, true,
		func(string, bool, string) string { return "" },
		func(string, bool, string) string { return "/ 4k Write failed Read failed\n" },
	)
	if method != "dd" || !strings.Contains(output, "Write failed") || !strings.Contains(output, "DD and FIO returned no usable benchmark data.") {
		t.Fatalf("both-failed result = method %q output %q", method, output)
	}
}

func TestLegacyDiskResultUsableRejectsHeadersAndErrors(t *testing.T) {
	for _, output := range []string{"", "Test Path Block Read(IOPS) Write(IOPS)\n", "Unable to create test path\n"} {
		if legacyDiskResultUsable(output) {
			t.Fatalf("failure-only output considered usable: %q", output)
		}
	}
	if !legacyDiskResultUsable("/ 4k 1.2 GiB/s(100) Read failed\n") {
		t.Fatal("partial measured result was rejected")
	}
}
