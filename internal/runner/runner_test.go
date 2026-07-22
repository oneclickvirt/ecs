package runner

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/oneclickvirt/cputest/cpu"
	"github.com/oneclickvirt/ecs/internal/params"
)

func TestShouldPrintBriefIPLinesInBasicStage(t *testing.T) {
	tests := []struct {
		name string
		cfg  *params.Config
		want bool
	}{
		{
			name: "nil config",
			cfg:  nil,
			want: false,
		},
		{
			name: "network only choice without ip-only mode",
			cfg: &params.Config{
				Choice:             "6",
				SecurityTestStatus: true,
				BasicStatus:        false,
				OnlyIpInfoCheck:    false,
			},
			want: true,
		},
		{
			name: "ip quality choice without ip-only mode",
			cfg: &params.Config{
				Choice:             "9",
				SecurityTestStatus: true,
				BasicStatus:        false,
				OnlyIpInfoCheck:    false,
			},
			want: true,
		},
		{
			name: "network only with ip-only mode should suppress duplicate",
			cfg: &params.Config{
				Choice:             "6",
				SecurityTestStatus: true,
				BasicStatus:        false,
				OnlyIpInfoCheck:    true,
			},
			want: false,
		},
		{
			name: "basic enabled should not print brief lines",
			cfg: &params.Config{
				Choice:             "6",
				SecurityTestStatus: true,
				BasicStatus:        true,
				OnlyIpInfoCheck:    false,
			},
			want: false,
		},
		{
			name: "security disabled should not print brief lines",
			cfg: &params.Config{
				Choice:             "6",
				SecurityTestStatus: false,
				BasicStatus:        false,
				OnlyIpInfoCheck:    false,
			},
			want: false,
		},
		{
			name: "other choice should not print brief lines",
			cfg: &params.Config{
				Choice:             "3",
				SecurityTestStatus: true,
				BasicStatus:        false,
				OnlyIpInfoCheck:    false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldPrintBriefIPLinesInBasicStage(tt.cfg)
			if got != tt.want {
				t.Fatalf("shouldPrintBriefIPLinesInBasicStage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldPrintPingInfoSection(t *testing.T) {
	tests := []struct {
		name string
		cfg  *params.Config
		info string
		want bool
	}{
		{name: "nil config", cfg: nil, info: "x", want: false},
		{name: "empty info", cfg: &params.Config{OnlyChinaTest: true}, info: "", want: false},
		{name: "china-only prints", cfg: &params.Config{OnlyChinaTest: true, PingTestStatus: false}, info: "ok", want: true},
		{name: "ping-only prints", cfg: &params.Config{OnlyChinaTest: false, PingTestStatus: true}, info: "ok", want: true},
		{name: "both flags print once path", cfg: &params.Config{OnlyChinaTest: true, PingTestStatus: true}, info: "ok", want: true},
		{name: "neither flag", cfg: &params.Config{OnlyChinaTest: false, PingTestStatus: false}, info: "ok", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldPrintPingInfoSection(tt.cfg, tt.info)
			if got != tt.want {
				t.Fatalf("shouldPrintPingInfoSection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldPrintPingExtraSectionWithoutInfo(t *testing.T) {
	tests := []struct {
		name string
		cfg  *params.Config
		want bool
	}{
		{name: "nil config", cfg: nil, want: false},
		{name: "both disabled no extras", cfg: &params.Config{}, want: false},
		{name: "tgdc only prints", cfg: &params.Config{TgdcTestStatus: true}, want: true},
		{name: "web only prints", cfg: &params.Config{WebTestStatus: true}, want: true},
		{name: "ping enabled should not use extra path", cfg: &params.Config{PingTestStatus: true, TgdcTestStatus: true}, want: false},
		{name: "china-only should not use extra path", cfg: &params.Config{OnlyChinaTest: true, WebTestStatus: true}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldPrintPingExtraSectionWithoutInfo(tt.cfg)
			if got != tt.want {
				t.Fatalf("shouldPrintPingExtraSectionWithoutInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunNetworkTestsWaitsBeforeReadingPingInfo(t *testing.T) {
	cfg := &params.Config{
		Language:       "zh",
		Width:          40,
		PingTestStatus: true,
	}
	var (
		wg          sync.WaitGroup
		ptInfo      string
		outputMutex sync.Mutex
		infoMutex   sync.Mutex
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		infoMutex.Lock()
		ptInfo = "late ping result"
		infoMutex.Unlock()
	}()

	output := RunNetworkTests(context.Background(), cfg, &wg, &ptInfo, "", "", &outputMutex, &infoMutex)

	if !strings.Contains(output, "late ping result") {
		t.Fatalf("expected delayed ping info in output, got %q", output)
	}
}

func TestRunEnglishNetworkTestsPrintsPingInfo(t *testing.T) {
	cfg := &params.Config{
		Language:       "en",
		Width:          40,
		PingTestStatus: true,
	}
	var (
		wg          sync.WaitGroup
		ptInfo      string
		outputMutex sync.Mutex
		infoMutex   sync.Mutex
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		infoMutex.Lock()
		ptInfo = "english ping result"
		infoMutex.Unlock()
	}()

	output := RunEnglishNetworkTests(context.Background(), cfg, &wg, &ptInfo, "", "", &outputMutex, &infoMutex)

	if !strings.Contains(output, "english ping result") {
		t.Fatalf("expected English ping info in output, got %q", output)
	}
}

func TestRunCPUTestMergesConfiguredBurnIntoCPUSection(t *testing.T) {
	previous := runLegacyCPUBurn
	defer func() { runLegacyCPUBurn = previous }()
	var captured cpu.BurnConfig
	runLegacyCPUBurn = func(_ context.Context, config cpu.BurnConfig) cpu.BurnResult {
		captured = config
		return cpu.BurnResult{Status: "ok", EffectiveThreads: 2, DurationMS: 20000, Events: 42, EventsPerSecond: 2.1}
	}

	cfg := &params.Config{Language: "zh", Width: 80, DeepMode: true, DeepBurnDuration: 20 * time.Second}
	var outputMutex sync.Mutex
	output := RunCPUTest(context.Background(), cfg, "", "", &outputMutex)
	if captured.Duration != 20*time.Second || captured.MaxPrime != 50000 || captured.Threads <= 0 {
		t.Fatalf("unexpected burn config: %+v", captured)
	}
	if !strings.Contains(output, "CPU测试") || !strings.Contains(output, "压力测试") || !strings.Contains(output, "20s / 2 线程 / 2.10 次/秒 / 42 次") {
		t.Fatalf("burn output was not merged into compact CPU section: %q", output)
	}
	if strings.Contains(output, "CPU压力测试") || strings.Contains(output, "状态") || strings.Contains(output, "ok") {
		t.Fatalf("burn output retained a redundant section or status row: %q", output)
	}
}

func TestRunCPUBurnTestOnlyPrintsFailureReason(t *testing.T) {
	previous := runLegacyCPUBurn
	defer func() { runLegacyCPUBurn = previous }()
	runLegacyCPUBurn = func(context.Context, cpu.BurnConfig) cpu.BurnResult {
		return cpu.BurnResult{Status: "canceled", Error: "deadline exceeded"}
	}
	cfg := &params.Config{Language: "zh", DeepMode: true, DeepBurnDuration: time.Second}
	var outputMutex sync.Mutex
	output := RunCPUBurnTest(context.Background(), cfg, "", "", &outputMutex)
	if output != " 压力测试            : deadline exceeded\n" {
		t.Fatalf("unexpected failed burn output: %q", output)
	}
}

func TestRunCPUBurnTestSkipsOrdinaryDefaults(t *testing.T) {
	called := false
	previous := runLegacyCPUBurn
	defer func() { runLegacyCPUBurn = previous }()
	runLegacyCPUBurn = func(context.Context, cpu.BurnConfig) cpu.BurnResult {
		called = true
		return cpu.BurnResult{}
	}
	cfg := &params.Config{Language: "zh", Width: 40}
	var outputMutex sync.Mutex
	if output := RunCPUBurnTest(context.Background(), cfg, "existing", "", &outputMutex); output != "existing" || called {
		t.Fatalf("ordinary CLI default unexpectedly ran deep burn: output=%q called=%t", output, called)
	}
}
