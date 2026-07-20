package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/oneclickvirt/ecs/api"
)

func main() {
	cdn := flag.String("data-cdn", "https://cdn.spiritlhl.net/https://raw.githubusercontent.com/oneclickvirt/ecs-data/main/data", "ecs-data CDN base")
	offline := flag.Bool("offline", false, "force embedded snapshot fallback")
	standard := flag.Bool("standard", false, "run the complete standard structured profile")
	tcp := flag.Bool("tcp", false, "run structured TCP probes")
	hardware := flag.Bool("hardware", false, "run the bounded structured hardware stage")
	deepDiskPath := flag.String("deep-disk-path", "", "run the deep fio matrix on one explicit mounted directory")
	deepBurnDuration := flag.Duration("deep-burn-duration", 0, "run an explicit bounded CPU burn")
	maxDuration := flag.Duration("canary-deadline", 30*time.Second, "canary deadline")
	flag.Parse()
	if *offline {
		*cdn = "http://127.0.0.1:1"
	}
	config := canaryConfig(*cdn, *offline, *standard, *tcp, *hardware, *deepDiskPath, *deepBurnDuration, *maxDuration)
	preCheck := api.NetCheckResult{Connected: !*offline, StackType: "IPv4", HasIPv4: true}
	ctx, cancel := context.WithTimeout(context.Background(), *maxDuration)
	defer cancel()
	started := time.Now()
	var report *api.StructuredReport
	if canaryRunsHardware(config) {
		report = api.RunAllTestsContext(ctx, preCheck, config).Report
	} else {
		report = api.CollectStructuredReport(ctx, preCheck, config, "", started, started)
	}
	finished := time.Now()
	report.FinishedAt = finished
	report.DurationMS = finished.Sub(started).Milliseconds()
	encoded, err := json.Marshal(report)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(encoded))
}

func canaryConfig(cdn string, offline, standard, tcp, hardware bool, deepDiskPath string, deepBurnDuration, maxDuration time.Duration) *api.Config {
	config := api.NewConfig("canary")
	config.BasicStatus = hardware || standard
	config.CpuTestStatus = hardware || standard
	config.MemoryTestStatus = hardware || standard
	config.DiskTestStatus = hardware || standard
	config.UtTestStatus = standard
	config.SecurityTestStatus = standard
	config.EmailTestStatus = standard
	config.BacktraceStatus = standard
	config.Nt3Status = standard
	config.PingTestStatus = standard
	config.SpeedTestStatus = standard
	config.TgdcTestStatus = standard
	config.WebTestStatus = standard
	config.DataCDNBase = cdn
	config.DataOffline = offline
	config.TCPProbeStatus = tcp || standard
	config.DeepMode = deepDiskPath != "" || deepBurnDuration > 0
	config.DeepDiskPaths = deepDiskPath
	config.DeepBurnDuration = deepBurnDuration
	// Canary intentionally exposes no SMART or GPU selector.
	config.DeepSMARTDevices = ""
	config.DeepGPUDevice = ""
	config.PrivacyMode = true
	config.EnableUpload = false
	config.MaxDuration = maxDuration
	config.HardwareBudget = min(2*time.Minute, maxDuration)
	config.ValidateParams()
	return config
}

func canaryRunsHardware(config *api.Config) bool {
	return config != nil && (config.BasicStatus || config.CpuTestStatus || config.MemoryTestStatus || config.DiskTestStatus ||
		(config.DeepMode && (config.DeepDiskPaths != "" || config.DeepBurnDuration > 0)))
}
