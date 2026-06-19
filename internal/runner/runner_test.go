package runner

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

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
