package runner

import (
	"bytes"
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

func TestSpeedCaptureWriterStreamsAndBuffers(t *testing.T) {
	var report, terminal bytes.Buffer
	writer := speedCaptureWriterTo(&report, true, &terminal)
	if _, err := writer.Write([]byte("speed row\n")); err != nil {
		t.Fatal(err)
	}
	if report.String() != "speed row\n" {
		t.Fatalf("report buffer = %q", report.String())
	}
	if terminal.String() != report.String() {
		t.Fatalf("terminal output = %q, report = %q", terminal.String(), report.String())
	}
}

func TestBoundedSpeedTestContextRespectsEarlierParentDeadline(t *testing.T) {
	parentDeadline := time.Now().Add(2 * time.Second)
	parent, parentCancel := context.WithDeadline(context.Background(), parentDeadline)
	defer parentCancel()

	child, childCancel := boundedSpeedTestContext(parent)
	defer childCancel()
	childDeadline, ok := child.Deadline()
	if !ok {
		t.Fatal("bounded speed context has no deadline")
	}
	if childDeadline.After(parentDeadline) {
		t.Fatalf("speed deadline %s exceeded parent deadline %s", childDeadline, parentDeadline)
	}
}

func TestFairSpeedGroupContextReservesTimeForRemainingGroups(t *testing.T) {
	parent, parentCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer parentCancel()
	child, childCancel := fairSpeedGroupContext(parent, 3)
	defer childCancel()
	childDeadline, ok := child.Deadline()
	if !ok {
		t.Fatal("fair speed group context has no deadline")
	}
	remaining := time.Until(childDeadline)
	if remaining < 9*time.Second || remaining > 11*time.Second {
		t.Fatalf("first of three groups received %s, want about one third of remaining time", remaining)
	}
}

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
				Choice:             "7",
				SecurityTestStatus: true,
				BasicStatus:        false,
				OnlyIpInfoCheck:    false,
			},
			want: true,
		},
		{
			name: "ip quality choice without ip-only mode",
			cfg: &params.Config{
				Choice:             "10",
				SecurityTestStatus: true,
				BasicStatus:        false,
				OnlyIpInfoCheck:    false,
			},
			want: true,
		},
		{
			name: "network only with ip-only mode should suppress duplicate",
			cfg: &params.Config{
				Choice:             "7",
				SecurityTestStatus: true,
				BasicStatus:        false,
				OnlyIpInfoCheck:    true,
			},
			want: false,
		},
		{
			name: "basic enabled should not print brief lines",
			cfg: &params.Config{
				Choice:             "7",
				SecurityTestStatus: true,
				BasicStatus:        true,
				OnlyIpInfoCheck:    false,
			},
			want: false,
		},
		{
			name: "security disabled should not print brief lines",
			cfg: &params.Config{
				Choice:             "7",
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

func TestSpeedNetworkForStackPinsDualStackToIPv4(t *testing.T) {
	tests := map[string]string{
		"DualStack": "tcp4",
		"IPv4":      "tcp4",
		"IPv6":      "tcp6",
		"None":      "",
	}
	for stack, want := range tests {
		if got := speedNetworkForStack(stack); got != want {
			t.Fatalf("speedNetworkForStack(%q) = %q, want %q", stack, got, want)
		}
	}
}

func TestChinesePresetSpeedProfilesKeepCompleteAndNearbyScopes(t *testing.T) {
	for _, choice := range []string{"1", "2"} {
		if !usesChineseFullSpeedProfile(&params.Config{Choice: choice}) {
			t.Fatalf("complete preset choice %q was not recognized", choice)
		}
		if usesChineseNearbyCarrierSpeedProfile(&params.Config{Choice: choice}) {
			t.Fatalf("complete preset choice %q was classified as nearby-only", choice)
		}
	}
	for _, choice := range []string{"3", "4", "5", "6", "7"} {
		if !usesChineseNearbyCarrierSpeedProfile(&params.Config{Choice: choice}) {
			t.Fatalf("preset choice %q should use the nearby plus three-carrier profile", choice)
		}
		if usesChineseFullSpeedProfile(&params.Config{Choice: choice}) {
			t.Fatalf("nearby preset choice %q was classified as complete", choice)
		}
	}
	if !usesChineseFullSpeedProfile(&params.Config{Choice: "", MenuMode: false}) {
		t.Fatal("non-menu empty choice should retain the historical complete profile")
	}
	for _, choice := range []string{"", "custom", "manual", "8", "9", "10", "11"} {
		cfg := &params.Config{Choice: choice, MenuMode: true}
		if usesChinesePresetSpeedProfile(cfg) {
			t.Fatalf("custom/menu choice %q should retain its custom speed profile", choice)
		}
	}
}

func TestGlobalSpeedCandidatePreloadMatchesRenderedProfiles(t *testing.T) {
	tests := []struct {
		name string
		cfg  *params.Config
		want bool
	}{
		{name: "nil", cfg: nil, want: false},
		{name: "Chinese complete", cfg: &params.Config{Language: "zh", Choice: "1"}, want: true},
		{name: "Chinese concurrent complete", cfg: &params.Config{Language: "zh", Choice: "2"}, want: true},
		{name: "Chinese compact", cfg: &params.Config{Language: "zh", Choice: "3"}, want: false},
		{name: "Chinese custom", cfg: &params.Config{Language: "zh", Choice: "custom", MenuMode: true}, want: false},
		{name: "Chinese non-menu legacy", cfg: &params.Config{Language: "zh", MenuMode: false}, want: true},
		{name: "English compact", cfg: &params.Config{Language: "en", Choice: "3"}, want: true},
		{name: "unsupported language", cfg: &params.Config{Language: "ja", Choice: "1"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldPreloadGlobalSpeedCandidates(tt.cfg); got != tt.want {
				t.Fatalf("shouldPreloadGlobalSpeedCandidates(%+v) = %t, want %t", tt.cfg, got, tt.want)
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
	if !strings.Contains(text, "Disk-Test--fio-Method") || !strings.Contains(text, " Disk benchmark returned no usable data.\n") {
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
	if strings.Count(text, "Disk benchmark returned no usable data.") != 2 ||
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
	if !strings.Contains(text, "Disk-Test--fio-Method") || !strings.Contains(text, " Disk benchmark returned no usable data.\n") {
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

func TestLegacyWorkflowStartsCandidatePreloadAfterHardware(t *testing.T) {
	hardwareStarted := make(chan struct{}, 1)
	hardwareRelease := make(chan struct{})
	preloadStarted := make(chan struct{}, 1)

	plan := legacyWorkflowPlan{
		basics: func(context.Context) {},
		hardware: []bufferedTask{{name: "cpu", run: func(context.Context) string {
			hardwareStarted <- struct{}{}
			<-hardwareRelease
			return "CPU\n"
		}}},
		preload: func(context.Context) { preloadStarted <- struct{}{} },
	}
	done := make(chan struct{})
	go func() {
		runLegacyWorkflowPlan(context.Background(), plan)
		close(done)
	}()

	select {
	case <-hardwareStarted:
	case <-time.After(time.Second):
		t.Fatal("hardware stage did not start")
	}
	select {
	case <-preloadStarted:
		t.Fatal("candidate preload started while a hardware benchmark was active")
	default:
	}
	close(hardwareRelease)
	select {
	case <-preloadStarted:
	case <-time.After(time.Second):
		t.Fatal("candidate preload did not start after hardware completed")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("workflow did not complete")
	}
}

func TestFullConcurrentLegacyWorkflowStartsBasicAndEveryOtherStageTogether(t *testing.T) {
	started := make(chan string, 4)
	release := map[string]chan struct{}{
		"basic": make(chan struct{}),
		"cpu":   make(chan struct{}),
		"ping":  make(chan struct{}),
		"speed": make(chan struct{}),
	}
	block := func(name, value string) bufferedTask {
		return bufferedTask{name: name, run: func(context.Context) string {
			started <- name
			<-release[name]
			return value
		}}
	}
	var (
		output   strings.Builder
		outputMu sync.Mutex
	)
	identityReady := make(chan struct{}, 1)
	plan := legacyWorkflowPlan{
		fullConcurrent:   true,
		concurrentBasics: ptrBufferedTask(block("basic", "BASIC\n")),
		hardware:         []bufferedTask{block("cpu", "CPU\n")},
		independent:      []bufferedTask{block("ping", "PING\n")},
		concurrentSpeed:  ptrBufferedTask(block("speed", "SPEED\n")),
		identityReady: func(context.Context) {
			identityReady <- struct{}{}
		},
		emit: func(value string) {
			outputMu.Lock()
			output.WriteString(value)
			outputMu.Unlock()
		},
	}
	done := make(chan struct{})
	go func() {
		runLegacyWorkflowPlan(context.Background(), plan)
		close(done)
	}()

	seen := make(map[string]bool, len(release))
	for range release {
		select {
		case name := <-started:
			seen[name] = true
		case <-time.After(time.Second):
			t.Fatalf("full concurrent workflow did not launch every stage before basic completed: %v", seen)
		}
	}
	for name := range release {
		if !seen[name] {
			t.Fatalf("full concurrent workflow did not start %q: %v", name, seen)
		}
	}
	for _, name := range []string{"basic", "cpu", "ping", "speed"} {
		close(release[name])
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("full concurrent workflow did not finish")
	}
	select {
	case <-identityReady:
	default:
		t.Fatal("identity readiness was not signaled after the concurrent workflow")
	}
	outputMu.Lock()
	got := output.String()
	outputMu.Unlock()
	if want := "BASIC\nCPU\nPING\nSPEED\n"; got != want {
		t.Fatalf("full concurrent output order = %q, want %q", got, want)
	}
}

func TestIdentityReadinessBroadcastsToEveryDependentTask(t *testing.T) {
	ready := make(chan struct{})
	ctx := WithIdentityReady(context.Background(), ready)
	const waiters = 4
	finished := make(chan struct{}, waiters)
	for range waiters {
		go func() {
			if !waitIdentityReady(ctx) {
				t.Error("identity wait unexpectedly canceled")
				return
			}
			finished <- struct{}{}
		}()
	}
	select {
	case <-finished:
		t.Fatal("dependent task passed identity barrier before publication")
	case <-time.After(25 * time.Millisecond):
	}
	signalIdentityReady(ctx)
	// A duplicate publication must be harmless and all waiters must observe the
	// same completion edge rather than competing for a single channel token.
	signalIdentityReady(ctx)
	for range waiters {
		select {
		case <-finished:
		case <-time.After(time.Second):
			t.Fatal("dependent task did not observe identity publication")
		}
	}
}

func ptrBufferedTask(value bufferedTask) *bufferedTask { return &value }

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
