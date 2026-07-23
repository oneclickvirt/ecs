package runner

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/oneclickvirt/cputest/cpu"
	"github.com/oneclickvirt/ecs/internal/params"
	pingmodel "github.com/oneclickvirt/pingtest/model"
	"github.com/oneclickvirt/pingtest/pt"
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

func TestBufferedDiskSectionEnglishNeverEmitsAnEmptyChapter(t *testing.T) {
	previous := runLegacyDisk
	defer func() { runLegacyDisk = previous }()
	runLegacyDisk = func(language, method, path string, multi, auto bool) (string, string) {
		if language != "en" || method != "fio" || path != "/fixture" || multi || !auto {
			t.Fatalf("unexpected disk arguments: language=%q method=%q path=%q multi=%t auto=%t", language, method, path, multi, auto)
		}
		return "fio", ""
	}

	text := bufferedDiskSection(context.Background(), &params.Config{
		Language: "en", Width: 64, DiskTestStatus: true, DiskTestMethod: "fio",
		DiskTestPath: "/fixture", AutoChangeDiskMethod: true,
	})
	if !strings.Contains(text, "Disk-Test--fio-Method") || !strings.Contains(text, " Disk test unavailable\n") {
		t.Fatalf("empty English disk result produced an empty chapter: %q", text)
	}
}

func TestBufferedDiskSectionReportsEachUnavailableMethod(t *testing.T) {
	previous := runLegacyDisk
	defer func() { runLegacyDisk = previous }()
	runLegacyDisk = func(_ string, method, _ string, _, _ bool) (string, string) {
		return method, ""
	}

	text := bufferedDiskSection(context.Background(), &params.Config{
		Language: "en", Width: 64, DiskTestStatus: true, AutoChangeDiskMethod: false,
	})
	if strings.Count(text, "Disk test unavailable") != 2 ||
		!strings.Contains(text, "Disk-Test--dd-Method") || !strings.Contains(text, "Disk-Test--fio-Method") {
		t.Fatalf("dual disk method output lost an unavailable result: %q", text)
	}
}

func TestRunDiskTestUsesTheSameEnglishEmptyResultContract(t *testing.T) {
	previous := runLegacyDisk
	defer func() { runLegacyDisk = previous }()
	runLegacyDisk = func(_ string, method, _ string, _, _ bool) (string, string) {
		return method, ""
	}
	cfg := &params.Config{
		Language: "en", Width: 64, DiskTestStatus: true, DiskTestMethod: "fio", AutoChangeDiskMethod: true,
	}
	var outputMutex sync.Mutex
	text := RunDiskTest(context.Background(), cfg, "", "", &outputMutex)
	if !strings.Contains(text, "Disk-Test--fio-Method") || !strings.Contains(text, " Disk test unavailable\n") {
		t.Fatalf("RunDiskTest diverged from buffered disk output: %q", text)
	}
}

func TestLegacyWorkflowBarriersOrderedDrainAndSpeedIsolation(t *testing.T) {
	hardwareStarted := make(chan string, 3)
	hardwareRelease := map[string]chan struct{}{
		"cpu": make(chan struct{}), "memory": make(chan struct{}), "disk": make(chan struct{}),
	}
	independentStarted := make(chan string, 2)
	independentRelease := map[string]chan struct{}{
		"ping": make(chan struct{}), "tcp": make(chan struct{}),
	}
	emitted := make(chan string, 8)
	speedStarted := make(chan struct{}, 1)
	var (
		mutex      sync.Mutex
		basicsDone bool
		output     strings.Builder
	)
	hardwareTask := func(name, value string) bufferedTask {
		return bufferedTask{name: name, run: func(context.Context) string {
			mutex.Lock()
			if !basicsDone {
				t.Errorf("%s started before basics completed", name)
			}
			mutex.Unlock()
			hardwareStarted <- name
			<-hardwareRelease[name]
			return value
		}}
	}
	independentTask := func(name, value string) bufferedTask {
		return bufferedTask{name: name, run: func(context.Context) string {
			independentStarted <- name
			<-independentRelease[name]
			return value
		}}
	}
	plan := legacyWorkflowPlan{
		basics: func(context.Context) {
			mutex.Lock()
			basicsDone = true
			mutex.Unlock()
		},
		hardware: []bufferedTask{
			hardwareTask("cpu", "CPU\n"), hardwareTask("memory", "MEM\n"), hardwareTask("disk", "DISK\n"),
		},
		independent: []bufferedTask{
			independentTask("ping", "PING-BEGIN PING-END\n"),
			independentTask("tcp", "TCP-BEGIN TCP-END\n"),
		},
		speed: func(context.Context) { speedStarted <- struct{}{} },
		emit: func(value string) {
			mutex.Lock()
			output.WriteString(value)
			mutex.Unlock()
			emitted <- value
		},
	}
	done := make(chan struct{})
	go func() {
		runLegacyWorkflowPlan(context.Background(), plan)
		close(done)
	}()

	for _, name := range []string{"cpu", "memory", "disk"} {
		select {
		case got := <-hardwareStarted:
			if got != name {
				t.Fatalf("hardware started out of order: got %q want %q", got, name)
			}
		case <-time.After(time.Second):
			t.Fatalf("hardware %q did not start", name)
		}
		if name != "disk" {
			select {
			case unexpected := <-hardwareStarted:
				t.Fatalf("hardware %q overlapped %q", unexpected, name)
			default:
			}
		}
		close(hardwareRelease[name])
		select {
		case <-emitted:
		case <-time.After(time.Second):
			t.Fatalf("hardware %q was not emitted after completion", name)
		}
	}

	seen := make(map[string]bool)
	for range 2 {
		select {
		case name := <-independentStarted:
			seen[name] = true
		case <-time.After(time.Second):
			t.Fatal("ping/TCP did not start concurrently after hardware")
		}
	}
	if !seen["ping"] || !seen["tcp"] {
		t.Fatalf("unexpected independent starts: %v", seen)
	}
	close(independentRelease["tcp"])
	select {
	case value := <-emitted:
		t.Fatalf("later TCP chapter overtook ping: %q", value)
	case <-time.After(20 * time.Millisecond):
	}
	select {
	case <-speedStarted:
		t.Fatal("speed started before all non-speed tasks were consumed")
	default:
	}
	close(independentRelease["ping"])
	for _, want := range []string{"PING-BEGIN PING-END\n", "TCP-BEGIN TCP-END\n"} {
		select {
		case got := <-emitted:
			if got != want {
				t.Fatalf("ordered drain emitted %q, want %q", got, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("ordered chapter %q was not emitted immediately", want)
		}
	}
	select {
	case <-speedStarted:
	case <-time.After(time.Second):
		t.Fatal("speed did not start after the non-speed barrier")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("legacy workflow did not complete")
	}
	mutex.Lock()
	gotOutput := output.String()
	mutex.Unlock()
	if want := "CPU\nMEM\nDISK\nPING-BEGIN PING-END\nTCP-BEGIN TCP-END\n"; gotOutput != want {
		t.Fatalf("buffered output interleaved: got %q want %q", gotOutput, want)
	}
}

func TestOrderedBufferedTaskPanicCannotBlockFollowingChapter(t *testing.T) {
	emitted := make(chan string, 1)
	done := make(chan struct{})
	go func() {
		runOrderedBufferedTasks(context.Background(), []bufferedTask{
			{name: "panic", run: func(context.Context) string { panic("fixture") }},
			{name: "next", run: func(context.Context) string { return "next\n" }},
		}, func(value string) { emitted <- value })
		close(done)
	}()
	select {
	case got := <-emitted:
		if got != "next\n" {
			t.Fatalf("unexpected post-panic output %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("provider panic blocked the ordered drain")
	}
	<-done
}

func TestTCPRegistryIsSelectedBeforeSingleProbeRun(t *testing.T) {
	previousLoaded, previousBuiltin := runLoadedTCPRegistry, runBuiltinTCPRegistry
	defer func() {
		runLoadedTCPRegistry, runBuiltinTCPRegistry = previousLoaded, previousBuiltin
	}()
	loadedCalls, builtinCalls := 0, 0
	fixture := []pt.TCPResult{{
		Target:   pingmodel.TCPTarget{Name: "Fixture", Host: "fixture.test", Port: 443},
		Attempts: 1, Successful: 1, SuccessRatePercent: 100,
	}}
	runLoadedTCPRegistry = func(context.Context, pt.TCPProbeConfig) ([]pt.TCPResult, pingmodel.TCPTargetRegistryLoadResult, error) {
		loadedCalls++
		return fixture, pingmodel.TCPTargetRegistryLoadResult{}, nil
	}
	runBuiltinTCPRegistry = func(context.Context, pt.TCPProbeConfig) []pt.TCPResult {
		builtinCalls++
		return fixture
	}
	cfg := &params.Config{Language: "en", Width: 80, TCPProbeStatus: true, TCPSortOrder: "name", TCPTextFormat: "compact"}
	text := bufferedTCPSection(context.Background(), cfg)
	if loadedCalls != 1 || builtinCalls != 0 {
		t.Fatalf("successful loaded registry calls: loaded=%d builtin=%d", loadedCalls, builtinCalls)
	}
	if !strings.Contains(text, "Summary") || !strings.Contains(text, "Platform") {
		t.Fatalf("legacy TCP language was not forwarded: %q", text)
	}

	runLoadedTCPRegistry = func(context.Context, pt.TCPProbeConfig) ([]pt.TCPResult, pingmodel.TCPTargetRegistryLoadResult, error) {
		loadedCalls++
		return nil, pingmodel.TCPTargetRegistryLoadResult{}, errors.New("fixture load failure")
	}
	_ = bufferedTCPSection(context.Background(), cfg)
	if loadedCalls != 2 || builtinCalls != 1 {
		t.Fatalf("fallback registry calls: loaded=%d builtin=%d", loadedCalls, builtinCalls)
	}
}
