package params

import "testing"

func TestAIOnlyUnlockRegionIsValid(t *testing.T) {
	config := NewConfig("test")
	config.ParseFlags([]string{"-utregion", "21"})
	if config.UnlockTestRegion != "21" {
		t.Fatalf("AI-only region was normalized to %q", config.UnlockTestRegion)
	}
}

func TestUnlockNetworkFlagsAreExplicitConfig(t *testing.T) {
	config := NewConfig("test")
	config.ParseFlags([]string{
		"-ut-interface", "eth0", "-ut-dns", "1.1.1.1:53",
		"-ut-http-proxy", "http://127.0.0.1:8080", "-ut-concurrency", "250",
	})
	if config.UnlockTestInterface != "eth0" || config.UnlockTestDNSServers != "1.1.1.1:53" || config.UnlockTestHTTPProxy != "http://127.0.0.1:8080" {
		t.Fatalf("unlock network flags were not retained: %+v", config)
	}
	if config.UnlockTestConcurrency != 100 {
		t.Fatalf("unlock concurrency = %d, want clamp 100", config.UnlockTestConcurrency)
	}
}
