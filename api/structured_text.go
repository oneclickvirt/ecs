package api

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

const structuredLabelWidth = 14

type structuredTextRenderer struct {
	builder strings.Builder
	width   int
	zh      bool
	tcpSort string
}

func renderStructuredRunText(config *Config, dataFiles []DataFileVersion, components []ComponentReport, tcp []TCPReport) string {
	renderer := newStructuredTextRenderer(config)
	renderer.header(config)
	renderer.dataFiles(dataFiles)
	beforeSpeed, speed := partitionSpeedComponents(components)
	renderer.components(beforeSpeed)
	renderer.tcp(tcp)
	renderer.components(speed)
	return renderer.builder.String()
}

func partitionSpeedComponents(components []ComponentReport) ([]ComponentReport, []ComponentReport) {
	beforeSpeed := make([]ComponentReport, 0, len(components))
	speed := make([]ComponentReport, 0, 1)
	for _, component := range components {
		if component.Name == "speed.registry" {
			speed = append(speed, component)
		} else {
			beforeSpeed = append(beforeSpeed, component)
		}
	}
	return beforeSpeed, speed
}

func (renderer *structuredTextRenderer) components(components []ComponentReport) {
	var cpuBenchmark, cpuBurn *ComponentReport
	for index := range components {
		switch components[index].Name {
		case "cputest":
			cpuBenchmark = &components[index]
		case "cputest.burn":
			cpuBurn = &components[index]
		}
	}
	for index := range components {
		component := components[index]
		switch component.Name {
		case "cputest":
			renderer.cpuComponents(cpuBenchmark, cpuBurn)
		case "cputest.burn":
			if cpuBenchmark == nil {
				renderer.cpuComponents(nil, cpuBurn)
			}
		default:
			renderer.component(component)
		}
	}
}

func (renderer *structuredTextRenderer) cpuComponents(benchmark, burn *ComponentReport) {
	renderer.section(renderer.pick("CPU性能测试", "CPU Benchmark"))
	if benchmark != nil {
		renderer.componentState(*benchmark)
		if len(benchmark.Payload) > 0 {
			renderer.cpuPayload(benchmark.Payload, renderer.pick("性能测试", "Benchmark"))
		}
	}
	if burn != nil && burn.Status != ReportStatusSkipped {
		renderer.componentState(*burn)
		if len(burn.Payload) > 0 {
			renderer.cpuPayload(burn.Payload, renderer.pick("压力测试", "Pressure Test"))
		}
	}
}

func appendStructuredTimeText(output string, config *Config, started, finished time.Time) string {
	renderer := newStructuredTextRenderer(config)
	renderer.builder.WriteString(output)
	if output != "" && !strings.HasSuffix(output, "\n") {
		renderer.builder.WriteByte('\n')
	}
	renderer.section("")
	duration := finished.Sub(started)
	if renderer.zh {
		renderer.row("花费", fmt.Sprintf("%d 分 %d 秒", int(duration.Minutes()), int(duration.Seconds())%60))
		renderer.row("时间", finished.Format("Mon Jan 2 15:04:05 MST 2006"))
	} else {
		renderer.row("Cost Time", fmt.Sprintf("%d min %d sec", int(duration.Minutes()), int(duration.Seconds())%60))
		renderer.row("Current Time", finished.Format("Mon Jan 2 15:04:05 MST 2006"))
	}
	renderer.section("")
	return renderer.builder.String()
}

func newStructuredTextRenderer(config *Config) *structuredTextRenderer {
	width := 80
	zh := true
	if config != nil {
		if config.Width > 0 {
			width = config.Width
		}
		zh = config.Language != "en"
	}
	if width < 48 {
		width = 48
	}
	tcpSort := "name"
	if config != nil && config.TCPSortOrder == "latency" {
		tcpSort = "latency"
	}
	return &structuredTextRenderer{width: width, zh: zh, tcpSort: tcpSort}
}

func (renderer *structuredTextRenderer) header(config *Config) {
	version := ""
	if config != nil {
		version = config.EcsVersion
	}
	if renderer.zh {
		renderer.section("VPS融合怪测试")
		renderer.row("版本", version)
		renderer.builder.WriteString(" 测评频道: https://t.me/+UHVoo2U4VyA5NTQ1\n")
		renderer.builder.WriteString(" Go项目地址：https://github.com/oneclickvirt/ecs\n")
		renderer.builder.WriteString(" Shell项目地址：https://github.com/spiritLHLS/ecs\n")
		return
	}
	renderer.section("VPS Fusion Monster Test")
	renderer.row("Version", version)
	renderer.builder.WriteString(" Review Channel: https://t.me/+UHVoo2U4VyA5NTQ1\n")
	renderer.builder.WriteString(" Go Project: https://github.com/oneclickvirt/ecs\n")
	renderer.builder.WriteString(" Shell Project: https://github.com/spiritLHLS/ecs\n")
}

func (renderer *structuredTextRenderer) section(title string) {
	titleWidth := runewidth.StringWidth(title)
	padding := renderer.width - titleWidth
	if padding < 0 {
		padding = 0
	}
	left := padding / 2
	renderer.builder.WriteString(strings.Repeat("-", left))
	renderer.builder.WriteString(truncateDisplay(title, renderer.width))
	renderer.builder.WriteString(strings.Repeat("-", padding-left))
	renderer.builder.WriteByte('\n')
}

func (renderer *structuredTextRenderer) row(label, value string) {
	renderer.rowWithLabelWidth(label, value, structuredLabelWidth)
}

func (renderer *structuredTextRenderer) basicRow(label, value string) {
	if value = compactText(value); value == "" || value == "-" {
		return
	}
	renderer.rowWithLabelWidth(label, value, 20)
}

func (renderer *structuredTextRenderer) rowWithLabelWidth(label, value string, labelWidth int) {
	value = compactText(value)
	if value == "" {
		value = "-"
	}
	prefix := " " + padDisplay(label, labelWidth) + " : "
	continuation := strings.Repeat(" ", runewidth.StringWidth(prefix))
	available := renderer.width - runewidth.StringWidth(prefix)
	if available < 12 {
		available = 12
	}
	lines := wrapDisplay(value, available)
	for index, line := range lines {
		if index == 0 {
			renderer.builder.WriteString(prefix)
		} else {
			renderer.builder.WriteString(continuation)
		}
		renderer.builder.WriteString(line)
		renderer.builder.WriteByte('\n')
	}
}

func (renderer *structuredTextRenderer) table(headers []string, rows [][]string, widths []int) {
	if len(headers) == 0 || len(headers) != len(widths) {
		return
	}
	gapWidth := (len(widths) - 1) * 2
	total := gapWidth + 1
	for _, width := range widths {
		total += width
	}
	if total > renderer.width {
		overflow := total - renderer.width
		for overflow > 0 {
			changed := false
			for index := range widths {
				minimum := 8
				if index == 0 {
					minimum = 12
				}
				if widths[index] > minimum && overflow > 0 {
					widths[index]--
					overflow--
					changed = true
				}
			}
			if !changed {
				break
			}
		}
	}
	renderer.tableLine(headers, widths)
	for _, row := range rows {
		if len(row) == len(widths) {
			renderer.tableLine(row, widths)
		}
	}
}

func (renderer *structuredTextRenderer) tableLine(values []string, widths []int) {
	renderer.builder.WriteByte(' ')
	for index, value := range values {
		if index > 0 {
			renderer.builder.WriteString("  ")
		}
		value = truncateDisplay(compactText(value), widths[index])
		renderer.builder.WriteString(padDisplay(value, widths[index]))
	}
	renderer.builder.WriteByte('\n')
}

func (renderer *structuredTextRenderer) dataFiles(files []DataFileVersion) {
	if len(files) == 0 {
		return
	}
	title := renderer.pick("数据源状态", "Data Sources")
	renderer.section(title)
	rows := make([][]string, 0, len(files))
	for _, file := range files {
		source := file.Source
		if file.Fallback != "" {
			if strings.EqualFold(file.Fallback, file.Source) {
				source += renderer.pick(" (回退)", " (fallback)")
			} else {
				source += " -> " + file.Fallback
			}
		}
		updated := "-"
		if !file.GeneratedAt.IsZero() {
			updated = file.GeneratedAt.Local().Format("01-02 15:04")
		}
		rows = append(rows, []string{strings.TrimSuffix(file.File, ".json"), source, strconv.Itoa(file.Count), updated, renderer.status(file.Status)})
	}
	renderer.table(
		[]string{renderer.pick("数据", "Data"), renderer.pick("来源", "Source"), renderer.pick("数量", "Count"), renderer.pick("更新时间", "Updated"), renderer.pick("状态", "Status")},
		rows, []int{20, 13, 7, 12, 11},
	)
}

func (renderer *structuredTextRenderer) component(component ComponentReport) {
	title := renderer.componentTitle(component.Name)
	if title == "" {
		return
	}
	renderer.section(title)
	renderer.componentState(component)
	if len(component.Payload) == 0 {
		return
	}
	switch component.Name {
	case "basics":
		renderer.basicPayload(component.Payload)
	case "memorytest":
		renderer.memoryPayload(component.Payload)
	case "disktest":
		renderer.diskPayload(component.Payload)
	case "disktest.deep_multi":
		renderer.deepDiskPayload(component.Payload)
	case "unlocktests.media":
		renderer.mediaPayload(component.Payload)
	case "security.evidence":
		renderer.securityPayload(component.Payload)
	case "backtrace.ip_bgp":
		renderer.backtracePayload(component.Payload)
	case "portchecker.email":
		renderer.mailPayload(component.Payload)
	case "nt3.province_latency":
		renderer.latencyPayload(component.Payload)
	case "ping.icmp", "ping.telegram":
		renderer.pingPayload(component.Payload)
	case "ping.web_tcp":
		renderer.tcpPayload(component.Payload)
	case "nt3.province_routes":
		renderer.routePayload(component.Payload)
	case "speed.registry":
		renderer.speedPayload(component.Payload)
	case "gostun.nat":
		renderer.natPayload(component.Payload)
	case "basics.smart_selftest", "basics.gpu_compute":
		renderer.deepToolPayload(component.Payload)
	default:
		renderer.genericPayload(component.Payload)
	}
}

func (renderer *structuredTextRenderer) componentState(component ComponentReport) {
	showState := component.Status != ReportStatusOK && component.Status != ReportStatusSkipped &&
		(component.Status != ReportStatusPartial || len(component.Payload) == 0)
	if showState {
		renderer.row(renderer.pick("状态", "Status"), renderer.status(component.Status))
	}
	if component.Reason != "" && showState {
		renderer.row(renderer.pick("说明", "Reason"), component.Reason)
	}
}

func (renderer *structuredTextRenderer) basicPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	cpu := objectValue(root, "cpu")
	memory := objectValue(root, "memory")
	cgroup := objectValue(root, "cgroup")
	virtualization := objectValue(root, "virtualization")
	network := objectValue(root, "network")
	firmware := objectValue(root, "firmware")
	renderer.basicRow("CPU", stringValue(cpu, "model"))
	if value := floatValue(cpu, "frequency_mhz"); value > 0 {
		renderer.basicRow(renderer.pick("CPU频率", "CPU Frequency"), fmt.Sprintf("%.0f MHz", value))
	}
	if value := intValue(cpu, "physical_cores"); value > 0 {
		renderer.basicRow(renderer.pick("CPU物理核心", "CPU Physical Cores"), strconv.Itoa(value))
	}
	if value := intValue(cpu, "logical_cpus"); value > 0 {
		renderer.basicRow(renderer.pick("CPU逻辑线程", "CPU Logical Threads"), strconv.Itoa(value))
	}
	if value := int64Value(memory, "total_bytes"); value > 0 {
		renderer.basicRow(renderer.pick("内存总量", "Memory Total"), formatBytes(value))
	}
	if value := int64Value(memory, "available_bytes"); value > 0 {
		renderer.basicRow(renderer.pick("可用内存", "Memory Available"), formatBytes(value))
	}
	renderer.basicRow(renderer.pick("虚拟化类型", "Virtualization Type"), stringValue(virtualization, "type"))
	renderer.basicRow(renderer.pick("容器运行时", "Container Runtime"), stringValue(virtualization, "container_runtime"))

	// Keep the four TCP tuning rows adjacent, matching the standalone basics
	// component. Each row carries one setting so scanning does not require
	// unpacking a compound value.
	renderer.basicRow(renderer.pick("TCP加速方式", "TCP Acceleration"), stringValue(network, "congestion_control"))
	renderer.basicRow(renderer.pick("TCP队列规则", "TCP Queue Discipline"), stringValue(network, "default_qdisc"))
	renderer.basicRow(renderer.pick("TCP接收缓冲", "TCP Receive Buffer"), tuningTupleValues(int64ArrayValue(network, "tcp_rmem")))
	renderer.basicRow(renderer.pick("TCP发送缓冲", "TCP Send Buffer"), tuningTupleValues(int64ArrayValue(network, "tcp_wmem")))

	renderer.basicRow(renderer.pick("Cgroup版本", "Cgroup Version"), stringValue(cgroup, "version"))
	if value := floatValue(cgroup, "cpu_quota_cores"); value > 0 {
		renderer.basicRow(renderer.pick("Cgroup CPU配额", "Cgroup CPU Quota"), fmt.Sprintf("%.2f cores", value))
	}
	renderer.basicRow(renderer.pick("Cgroup CPU集合", "Cgroup CPU Set"), stringValue(cgroup, "cpuset"))
	if value := int64Value(cgroup, "memory_current_bytes"); value > 0 {
		renderer.basicRow(renderer.pick("Cgroup内存使用", "Cgroup Memory Usage"), formatBytes(value))
	}
	if value := int64Value(cgroup, "memory_limit_bytes"); value > 0 {
		renderer.basicRow(renderer.pick("Cgroup内存上限", "Cgroup Memory Limit"), formatBytes(value))
	}
	if value := int64Value(cgroup, "memory_high_bytes"); value > 0 {
		renderer.basicRow(renderer.pick("Cgroup内存高水位", "Cgroup Memory High"), formatBytes(value))
	}
	if value := int64Value(cgroup, "memory_swap_limit_bytes"); value > 0 {
		renderer.basicRow(renderer.pick("Cgroup交换上限", "Cgroup Swap Limit"), formatBytes(value))
	}
	if value := int64Value(cgroup, "pids_limit"); value > 0 {
		renderer.basicRow(renderer.pick("Cgroup进程上限", "Cgroup PID Limit"), strconv.FormatInt(value, 10))
	}

	for _, field := range []struct{ zh, en, key string }{
		{"主板厂商", "Board Vendor", "board_vendor"}, {"主板型号", "Board Name", "board_name"},
		{"主板版本", "Board Version", "board_version"}, {"BIOS厂商", "BIOS Vendor", "bios_vendor"},
		{"BIOS版本", "BIOS Version", "bios_version"}, {"BIOS日期", "BIOS Date", "bios_date"},
	} {
		renderer.basicRow(renderer.pick(field.zh, field.en), stringValue(firmware, field.key))
	}

	pci := objectValue(root, "pci")
	gpus := arrayValue(root, "gpus")
	pciDevices := arrayValue(pci, "devices")
	renderer.basicRow(renderer.pick("PCI设备数量", "PCI Device Count"), strconv.Itoa(len(pciDevices)))
	renderer.basicRow(renderer.pick("PCI驱动", "PCI Drivers"), structuredDriverList(pciDevices))
	renderer.basicRow(renderer.pick("GPU设备数量", "GPU Device Count"), strconv.Itoa(len(gpus)))
	renderer.basicRow(renderer.pick("GPU驱动", "GPU Drivers"), structuredDriverList(gpus))

	topology := objectValue(root, "memory_topology")
	renderer.basicRow(renderer.pick("NUMA节点数量", "NUMA Node Count"), strconv.Itoa(len(arrayValue(topology, "nodes"))))
	// DIMM capacity intentionally is not summed here: the memory row already
	// reports total memory, and virtualized DMI tables commonly repeat it.
	renderer.basicRow(renderer.pick("DIMM数量", "DIMM Count"), strconv.Itoa(len(arrayValue(topology, "dimms"))))
	if value := int64Value(topology, "hugepages_total"); value > 0 {
		renderer.basicRow(renderer.pick("HugePages总数", "HugePages Total"), strconv.FormatInt(value, 10))
	}
	if value := int64Value(topology, "hugepages_free"); value > 0 {
		renderer.basicRow(renderer.pick("HugePages空闲", "HugePages Free"), strconv.FormatInt(value, 10))
	}
	if value := int64Value(topology, "hugepage_bytes"); value > 0 {
		renderer.basicRow(renderer.pick("HugePage大小", "HugePage Size"), formatBytes(value))
	}
	disks := arrayValue(root, "disks")
	for index, raw := range disks {
		disk, _ := raw.(map[string]any)
		health := objectValue(disk, "health")
		temperature := objectValue(disk, "temperature")
		zhPrefix := fmt.Sprintf("物理盘 %d", index+1)
		enPrefix := fmt.Sprintf("Disk %d", index+1)
		renderer.basicRow(renderer.pick(zhPrefix+" 型号", enPrefix+" Model"), joinNonEmpty(stringValue(disk, "vendor"), stringValue(disk, "model")))
		if value := int64Value(disk, "size_bytes"); value > 0 {
			renderer.basicRow(renderer.pick(zhPrefix+" 容量", enPrefix+" Size"), formatBytes(value))
		}
		renderer.basicRow(renderer.pick(zhPrefix+" 协议", enPrefix+" Protocol"), stringValue(health, "protocol"))
		renderer.basicRow(renderer.pick(zhPrefix+" 健康", enPrefix+" Health"), fallback(stringValue(health, "status"), stringValue(health, "availability")))
		if value := floatValue(temperature, "celsius"); value != 0 {
			renderer.basicRow(renderer.pick(zhPrefix+" 温度", enPrefix+" Temperature"), formatTemperature(value))
		}
	}
	raid := objectValue(root, "raid")
	arrays := arrayValue(raid, "arrays")
	controllers := arrayValue(raid, "controllers")
	renderer.basicRow(renderer.pick("RAID阵列数量", "RAID Arrays"), strconv.Itoa(len(arrays)))
	renderer.basicRow(renderer.pick("RAID级别", "RAID Levels"), structuredUniqueValues(arrays, "level"))
	degraded := 0
	for _, raw := range arrays {
		array, _ := raw.(map[string]any)
		if boolValue(array, "degraded") {
			degraded++
		}
	}
	if degraded > 0 {
		renderer.basicRow(renderer.pick("RAID降级阵列", "RAID Degraded Arrays"), strconv.Itoa(degraded))
	}
	renderer.basicRow(renderer.pick("RAID控制器数量", "RAID Controllers"), strconv.Itoa(len(controllers)))
	renderer.basicRow(renderer.pick("RAID驱动", "RAID Drivers"), structuredDriverList(controllers))
}

func (renderer *structuredTextRenderer) cpuPayload(payload json.RawMessage, label string) {
	root := payloadObject(payload)
	effective, requested := intValue(root, "effective_threads"), intValue(root, "requested_threads")
	threads := strconv.Itoa(effective)
	if requested > 0 && requested != effective {
		threads = fmt.Sprintf("%d/%d", effective, requested)
	}
	value := fmt.Sprintf("%s / %s %s / %.2f events/s / %d events",
		formatMilliseconds(int64Value(root, "duration_ms")), threads, renderer.pick("线程", "threads"),
		floatValue(root, "events_per_second"), int64Value(root, "events"))
	temperature := objectValue(root, "temperature")
	if boolValue(temperature, "available") {
		value += fmt.Sprintf(" / %.1f->%.1f C (+%.1f C)", floatValue(temperature, "baseline_c"), floatValue(temperature, "max_c"), floatValue(temperature, "delta_c"))
	}
	renderer.row(label, value)
}

func (renderer *structuredTextRenderer) memoryPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	rows := [][]string{
		{renderer.pick("顺序读取", "Sequential Read"), formatRate(floatValue(root, "sequential_read_mbps"))},
		{renderer.pick("顺序写入", "Sequential Write"), formatRate(floatValue(root, "sequential_write_mbps"))},
		{renderer.pick("内存复制", "Memory Copy"), formatRate(floatValue(root, "copy_mbps"))},
		{renderer.pick("随机延迟", "Random Latency"), fmt.Sprintf("%.2f ns", floatValue(root, "random_latency_ns"))},
	}
	renderer.table([]string{renderer.pick("项目", "Metric"), renderer.pick("结果", "Result")}, rows, []int{28, 28})
	renderer.row(renderer.pick("工作集", "Working Set"), formatBytes(int64Value(root, "working_set_bytes")))
}

func (renderer *structuredTextRenderer) diskPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	renderer.diskMetrics(arrayValue(root, "metrics"))
}

func (renderer *structuredTextRenderer) deepDiskPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	paths := arrayValue(root, "paths")
	for index, raw := range paths {
		path, _ := raw.(map[string]any)
		renderer.row(renderer.pick("测试目录", "Test Path"), fmt.Sprintf("#%d %s", index+1, renderer.status(ReportStatus(stringValue(path, "status")))))
		renderer.diskMetrics(arrayValue(path, "metrics"))
	}
}

func (renderer *structuredTextRenderer) diskMetrics(metrics []any) {
	rows := make([][]string, 0, len(metrics))
	for _, raw := range metrics {
		metric, _ := raw.(map[string]any)
		rows = append(rows, []string{
			stringValue(metric, "scenario_id"), fmt.Sprintf("%.0f", floatValue(metric, "iops")),
			formatBytesPerSecond(int64Value(metric, "bandwidth_bytes_per_second")), formatNanoseconds(int64Value(metric, "latency_p50_ns")),
			formatNanoseconds(int64Value(metric, "latency_p95_ns")), formatNanoseconds(int64Value(metric, "latency_p99_ns")),
		})
	}
	renderer.table([]string{renderer.pick("项目", "Scenario"), "IOPS", renderer.pick("带宽", "Bandwidth"), "P50", "P95", "P99"}, rows, []int{20, 10, 14, 10, 10, 10})
}

func (renderer *structuredTextRenderer) mediaPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	results := arrayValue(root, "results")
	rows := make([][]string, 0, len(results))
	for _, raw := range results {
		result, _ := raw.(map[string]any)
		detail := joinNonEmpty(stringValue(result, "region"), stringValue(result, "info"), stringValue(result, "unlock_type"))
		if detail == "" {
			detail = stringValue(result, "error")
		}
		rows = append(rows, []string{stringValue(result, "name"), stringValue(result, "ip_version"), localizedValue(stringValue(result, "status"), renderer.zh), detail})
	}
	renderer.table([]string{renderer.pick("平台", "Platform"), renderer.pick("协议", "IP"), renderer.pick("状态", "Status"), renderer.pick("地区/说明", "Region / Detail")}, rows, []int{24, 8, 14, 28})
}

func (renderer *structuredTextRenderer) securityPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	for _, rawAddress := range arrayValue(root, "addresses") {
		address, _ := rawAddress.(map[string]any)
		label := strings.ToUpper(stringValue(address, "ip_type"))
		if ip := stringValue(address, "ip"); ip != "" {
			label += " " + ip
		}
		renderer.row(renderer.pick("检测地址", "Address"), label)
		providers := arrayValue(address, "providers")
		rows := make([][]string, 0, len(providers))
		for _, rawProvider := range providers {
			provider, _ := rawProvider.(map[string]any)
			note := summarizeSecurityEvidence(provider)
			if note == "" {
				note = stringValue(provider, "error")
			}
			if missing := stringArrayValue(provider, "missing_fields"); len(missing) > 0 {
				note = renderer.pick("缺少: ", "missing: ") + strings.Join(missing, ",")
			}
			rows = append(rows, []string{stringValue(provider, "source"), localizedValue(stringValue(provider, "status"), renderer.zh), note})
		}
		renderer.table([]string{renderer.pick("来源", "Provider"), renderer.pick("状态", "Status"), renderer.pick("说明", "Detail")}, rows, []int{24, 18, 34})
		dnsbl := objectValue(address, "dnsbl")
		if len(dnsbl) > 0 {
			counts := objectValue(dnsbl, "counts")
			renderer.row("DNSBL", formatCounts(counts, renderer.zh))
		}
	}
}

func (renderer *structuredTextRenderer) backtracePayload(payload json.RawMessage) {
	root := payloadObject(payload)
	rows := make([][]string, 0)
	var details []map[string]any
	for _, raw := range arrayValue(root, "reports") {
		report, _ := raw.(map[string]any)
		details = append(details, report)
		rir := objectValue(report, "rir")
		source := stringValue(report, "prefix_source")
		if source == "" {
			switch {
			case len(objectValue(report, "rdap")) > 0:
				source = "RDAP"
			case len(objectValue(report, "whois")) > 0:
				source = "WHOIS"
			}
		}
		registered := stringValue(report, "registration_date")
		if len(registered) > 10 {
			registered = registered[:10]
		}
		rows = append(rows, []string{
			stringValue(report, "ip"), stringValue(report, "asn"), stringValue(rir, "name"), strings.Join(stringArrayValue(report, "prefixes"), ","), source, registered,
		})
	}
	renderer.table([]string{"IP", "ASN", "RIR", renderer.pick("前缀", "Prefix"), renderer.pick("来源", "Source"), renderer.pick("注册日期", "Registered")}, rows, []int{20, 10, 8, 18, 10, 12})
	for _, report := range details {
		relationships := objectValue(report, "relationships")
		asn := fallback(stringValue(report, "asn"), "ASN")
		if len(relationships) > 0 {
			renderer.row(asn+" "+renderer.pick("上游", "Upstream"), summarizeRelationships(arrayValue(relationships, "upstreams")))
			renderer.row(asn+" "+renderer.pick("对等", "Peers"), summarizeRelationships(arrayValue(relationships, "peers")))
			renderer.row(asn+" IXP", summarizeRelationships(arrayValue(relationships, "ixps")))
		}
		if geofeeds := arrayValue(report, "geofeeds"); len(geofeeds) > 0 {
			counts := make(map[string]int)
			for _, raw := range geofeeds {
				geofeed, _ := raw.(map[string]any)
				counts[localizedValue(stringValue(geofeed, "status"), renderer.zh)]++
			}
			renderer.row("Geofeed", formatIntCounts(counts))
		}
	}
}

func (renderer *structuredTextRenderer) mailPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	groups := []struct{ key, zh, en string }{
		{"local", "本地监听", "Local"}, {"outbound_smtp25", "出站25", "Outbound 25"}, {"mx", "MX 25", "MX 25"}, {"fixed", "固定端点", "Fixed"},
	}
	rows := make([][]string, 0, len(groups))
	for _, group := range groups {
		items := arrayValue(root, group.key)
		counts := make(map[string]int)
		for _, raw := range items {
			item, _ := raw.(map[string]any)
			counts[stringValue(item, "status")]++
		}
		rows = append(rows, []string{renderer.pick(group.zh, group.en), fmt.Sprintf("%d/%d", counts["available"], len(items)), formatIntCounts(counts)})
	}
	renderer.table([]string{renderer.pick("类型", "Type"), renderer.pick("可用", "Available"), renderer.pick("状态分布", "Status Counts")}, rows, []int{22, 14, 36})
}

func (renderer *structuredTextRenderer) latencyPayload(payload json.RawMessage) {
	var values []any
	if err := json.Unmarshal(payload, &values); err != nil {
		return
	}
	rows := make([][]string, 0, len(values))
	for _, raw := range values {
		result, _ := raw.(map[string]any)
		target := objectValue(result, "target")
		name := fallback(stringValue(target, "name"), stringValue(target, "province_name"), stringValue(target, "id"))
		carrier := stringValue(target, "carrier")
		if carrier != "" {
			name = joinNonEmpty(name, strings.ToUpper(carrier))
		}
		sent := intValue(result, "sent")
		received := intValue(result, "received")
		if sent == 0 {
			sent, received = intValue(result, "attempts"), intValue(result, "successful")
		}
		status := stringValue(result, "status")
		if status == "" {
			if received > 0 {
				status = "ok"
			} else {
				status = "unavailable"
			}
		}
		rows = append(rows, []string{
			name, localizedValue(status, renderer.zh), fmt.Sprintf("%d/%d", received, sent),
			formatJSONDuration(result["mean"]), formatJSONDuration(result["p95"]), fmt.Sprintf("%.0f%%", floatValue(result, "loss_percent")),
		})
	}
	renderer.table([]string{renderer.pick("目标", "Target"), renderer.pick("状态", "Status"), renderer.pick("成功", "Success"), renderer.pick("平均", "Mean"), "P95", renderer.pick("丢包", "Loss")}, rows, []int{24, 14, 10, 10, 10, 10})
}

func (renderer *structuredTextRenderer) pingPayload(payload json.RawMessage) {
	var values []map[string]any
	if err := json.Unmarshal(payload, &values); err != nil {
		return
	}
	const columns = 3
	cellWidth := (renderer.width - 1 - (columns-1)*2) / columns
	head := renderer.pick("平台 | 平均/P95 | 丢包", "Target | Mean/P95 | Loss")
	headCells := []string{head, head, head}
	renderer.compactColumns(headCells, cellWidth)
	for index := 0; index < len(values); index += columns {
		cells := make([]string, columns)
		for offset := 0; offset < columns && index+offset < len(values); offset++ {
			result := values[index+offset]
			target := objectValue(result, "target")
			name := fallback(stringValue(target, "name"), stringValue(target, "province_name"), stringValue(target, "id"))
			carrier := strings.ToUpper(stringValue(target, "carrier"))
			name = joinNonEmpty(name, carrier)
			cells[offset] = fmt.Sprintf("%s | %s/%s | %.0f%%", name, formatJSONDuration(result["mean"]), formatJSONDuration(result["p95"]), floatValue(result, "loss_percent"))
		}
		renderer.compactColumns(cells, cellWidth)
	}
}

func (renderer *structuredTextRenderer) compactColumns(values []string, width int) {
	renderer.builder.WriteByte(' ')
	for index, value := range values {
		if index > 0 {
			renderer.builder.WriteString("  ")
		}
		renderer.builder.WriteString(padDisplay(truncateDisplay(compactText(value), width), width))
	}
	renderer.builder.WriteByte('\n')
}

func (renderer *structuredTextRenderer) tcpPayload(payload json.RawMessage) {
	var values []map[string]any
	if err := json.Unmarshal(payload, &values); err != nil {
		return
	}
	rows := make([][]string, 0, len(values))
	for _, result := range values {
		target := objectValue(result, "target")
		rows = append(rows, []string{
			fallback(stringValue(target, "name"), stringValue(target, "id")), fmt.Sprintf("%d/%d", intValue(result, "successful"), intValue(result, "attempts")),
			formatJSONDuration(result["mean"]), formatJSONDuration(result["p95"]), fmt.Sprintf("%.0f%%", floatValue(result, "loss_percent")), formatCounts(objectValue(result, "error_counts"), renderer.zh),
		})
	}
	renderer.table([]string{renderer.pick("目标", "Target"), renderer.pick("成功", "Success"), renderer.pick("平均", "Mean"), "P95", renderer.pick("丢包", "Loss"), renderer.pick("错误", "Errors")}, rows, []int{24, 10, 10, 10, 10, 18})
}

func (renderer *structuredTextRenderer) routePayload(payload json.RawMessage) {
	var values []map[string]any
	if err := json.Unmarshal(payload, &values); err != nil {
		return
	}
	rows := make([][]string, 0, len(values))
	for _, result := range values {
		target := objectValue(result, "target")
		name := joinNonEmpty(stringValue(target, "province_name"), strings.ToUpper(stringValue(target, "carrier")), stringValue(target, "ip_version"))
		rows = append(rows, []string{name, localizedValue(stringValue(result, "status"), renderer.zh), strconv.Itoa(len(arrayValue(result, "hops"))), formatJSONDuration(result["duration"]), stringValue(result, "error")})
	}
	renderer.table([]string{renderer.pick("目标", "Target"), renderer.pick("状态", "Status"), renderer.pick("跳数", "Hops"), renderer.pick("耗时", "Duration"), renderer.pick("说明", "Detail")}, rows, []int{28, 14, 8, 12, 20})
}

func (renderer *structuredTextRenderer) speedPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	benchmarks := append(arrayValue(root, "benchmarks"), arrayValue(root, "private_benchmarks")...)
	rows := make([][]string, 0, len(benchmarks))
	for _, raw := range benchmarks {
		benchmark, _ := raw.(map[string]any)
		rows = append(rows, []string{
			fallback(stringValue(benchmark, "name"), stringValue(benchmark, "id")), fallback(stringValue(benchmark, "source"), "speedtest"), localizedValue(stringValue(benchmark, "status"), renderer.zh),
			fmt.Sprintf("%.2f ms", floatValue(benchmark, "latency_ms")), fmt.Sprintf("%.2f Mbps", floatValue(benchmark, "download_mbps")), fmt.Sprintf("%.2f Mbps", floatValue(benchmark, "upload_mbps")),
		})
	}
	renderer.table([]string{renderer.pick("节点", "Server"), renderer.pick("来源", "Source"), renderer.pick("状态", "Status"), renderer.pick("延迟", "Latency"), renderer.pick("下载", "Download"), renderer.pick("上传", "Upload")}, rows, []int{20, 14, 12, 10, 12, 12})
	if len(rows) == 0 {
		available := 0
		for _, raw := range arrayValue(root, "nodes") {
			node, _ := raw.(map[string]any)
			if stringValue(node, "availability") == "available" {
				available++
			}
		}
		renderer.row(renderer.pick("节点探活", "Node Probes"), fmt.Sprintf("%d/%d", available, len(arrayValue(root, "nodes"))))
	}
}

func (renderer *structuredTextRenderer) natPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	renderer.row(renderer.pick("网络类型", "IP Version"), stringValue(root, "ip_version"))
	renderer.row(renderer.pick("探测结果", "Probe Result"), fmt.Sprintf("%d/%d", intValue(root, "successful"), intValue(root, "successful")+intValue(root, "failed")))
	renderer.row(renderer.pick("映射一致性", "Mapping Consistency"), localizedValue(stringValue(root, "mapping_consistency"), renderer.zh))
	renderer.row(renderer.pick("端口保持", "Port Preservation"), localizedValue(stringValue(root, "port_preservation_consistency"), renderer.zh))
	renderer.row("Hairpin", localizedValue(stringValue(root, "hairpin_consistency"), renderer.zh))
	results := arrayValue(root, "results")
	if len(results) > 0 {
		first, _ := results[0].(map[string]any)
		renderer.row(renderer.pick("映射/过滤", "Mapping / Filtering"), joinNonEmpty(stringValue(first, "mapping_behavior"), stringValue(first, "filtering_behavior")))
		renderer.row(renderer.pick("NAT类型", "NAT Type"), stringValue(first, "nat_type"))
	}
}

func (renderer *structuredTextRenderer) deepToolPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	results := arrayValue(root, "results")
	if len(results) == 0 {
		results = []any{root}
	}
	rows := make([][]string, 0, len(results))
	for _, raw := range results {
		result, _ := raw.(map[string]any)
		rows = append(rows, []string{stringValue(result, "target"), localizedValue(stringValue(result, "status"), renderer.zh), formatMilliseconds(int64Value(result, "duration_ms")), fallback(stringValue(result, "error"), stringValue(result, "output"))})
	}
	renderer.table([]string{renderer.pick("目标", "Target"), renderer.pick("状态", "Status"), renderer.pick("耗时", "Duration"), renderer.pick("说明", "Detail")}, rows, []int{22, 14, 12, 28})
}

func (renderer *structuredTextRenderer) genericPayload(payload json.RawMessage) {
	root := payloadObject(payload)
	keys := make([]string, 0, len(root))
	for key, value := range root {
		if key == "schema_version" || key == "status" || key == "error" {
			continue
		}
		switch value.(type) {
		case string, float64, bool:
			keys = append(keys, key)
		case []any:
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := root[key]
		if array, ok := value.([]any); ok {
			renderer.row(humanizeKey(key), strconv.Itoa(len(array)))
			continue
		}
		renderer.row(humanizeKey(key), fmt.Sprint(value))
	}
}

func (renderer *structuredTextRenderer) tcp(reports []TCPReport) {
	if len(reports) == 0 {
		return
	}
	renderer.section(renderer.pick("TCP握手延迟", "TCP Handshake Latency"))
	ordered := append([]TCPReport(nil), reports...)
	sortTCPReports(ordered, renderer.tcpSort)
	summary := summarizeStructuredTCP(ordered)
	failed := max(0, summary.attempts-summary.successful)
	summaryValue := fmt.Sprintf(renderer.pick("目标:%d  握手:%d/%d  成功率:%.1f%%  失败:%d", "Targets:%d  Handshakes:%d/%d  Success:%.1f%%  Failed:%d"), len(ordered), summary.successful, summary.attempts, summary.successRate, failed)
	if failures := formatStructuredTCPFailures(summary.dns, summary.refused, summary.timeout, summary.other, renderer.zh); failures != "" {
		summaryValue += "  " + failures
	}
	renderer.row(renderer.pick("汇总", "Summary"), summaryValue)
	rows := make([][]string, 0, len(ordered))
	for _, report := range ordered {
		rows = append(rows, []string{
			fallback(report.Target.Name, report.Target.ID),
			fmt.Sprintf("%d/%d", report.Successful, report.Attempts),
			fmt.Sprintf("%.1f%%", structuredTCPLoss(report)),
			fmt.Sprintf("%s; %s", formatTCPMilliseconds(report.MinMS, report.MeanMS, report.P50MS, report.P95MS, report.MaxMS), formatTCPErrorCounts(report.Errors)),
		})
	}
	renderer.table(
		[]string{renderer.pick("平台", "Platform"), renderer.pick("成功/尝试", "Success"), renderer.pick("丢包", "Loss"), "Min/Avg/P50/P95/Max; D/R/T/O"},
		rows, []int{20, 9, 7, 37},
	)
}

func structuredTCPLoss(report TCPReport) float64 {
	if report.LossPercent > 0 || report.Attempts <= 0 || report.Successful >= report.Attempts {
		return report.LossPercent
	}
	return float64(report.Attempts-report.Successful) * 100 / float64(report.Attempts)
}

type structuredTCPSummary struct {
	attempts, successful, dns, refused, timeout, other int
	successRate                                        float64
}

func summarizeStructuredTCP(reports []TCPReport) structuredTCPSummary {
	var summary structuredTCPSummary
	for _, report := range reports {
		summary.attempts += report.Attempts
		summary.successful += report.Successful
		for class, count := range report.Errors {
			switch class {
			case "dns":
				summary.dns += count
			case "refused":
				summary.refused += count
			case "timeout":
				summary.timeout += count
			default:
				summary.other += count
			}
		}
	}
	if summary.attempts > 0 {
		summary.successRate = float64(summary.successful) * 100 / float64(summary.attempts)
	}
	return summary
}

func formatTCPMilliseconds(values ...float64) string {
	maximum := float64(0)
	for _, value := range values {
		if value > maximum {
			maximum = value
		}
	}
	precision := 1
	suffix := " ms"
	if maximum >= 100 {
		precision = 0
		suffix = "ms"
	}
	parts := make([]string, len(values))
	for index, value := range values {
		if value <= 0 {
			parts[index] = "-"
		} else {
			if precision == 0 {
				value = math.Round(value)
			}
			parts[index] = strconv.FormatFloat(value, 'f', precision, 64)
		}
	}
	return strings.Join(parts, "/") + suffix
}

func formatTCPErrorCounts(errors map[string]int) string {
	counts := [4]int{}
	for class, count := range errors {
		switch class {
		case "dns":
			counts[0] += count
		case "refused":
			counts[1] += count
		case "timeout":
			counts[2] += count
		default:
			counts[3] += count
		}
	}
	return fmt.Sprintf("%d/%d/%d/%d", counts[0], counts[1], counts[2], counts[3])
}

func formatStructuredTCPFailures(dns, refused, timeout, other int, zh bool) string {
	labels := [4]string{"DNS", "R", "T", "O"}
	if zh {
		labels = [4]string{"DNS", "拒", "超", "其"}
	}
	counts := [4]int{dns, refused, timeout, other}
	parts := make([]string, 0, len(counts))
	for index, count := range counts {
		if count > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", labels[index], count))
		}
	}
	return strings.Join(parts, " ")
}

func (renderer *structuredTextRenderer) componentTitle(name string) string {
	titles := map[string][2]string{
		"basics": {"系统基础信息", "System Basic Information"}, "cputest": {"CPU性能测试", "CPU Benchmark"},
		"memorytest": {"内存性能测试", "Memory Benchmark"}, "disktest": {"磁盘性能测试", "Disk Benchmark"},
		"unlocktests.media": {"跨国平台解锁", "Cross-Border Platform Unlock"}, "security.evidence": {"IP质量检测", "IP Quality Check"},
		"backtrace.ip_bgp": {"上游及注册信息", "Upstream and Registry"}, "portchecker.email": {"邮件端口检测", "Email Port Check"},
		"nt3.province_latency": {"全国三网延迟", "Province Carrier Latency"}, "nt3.province_routes": {"全国三网详细路由", "Province Carrier Routes"},
		"gostun.nat": {"NAT类型检测", "NAT Type Check"}, "speed.registry": {"就近节点测速", "Speed Test"},
		"ping.icmp": {"PING值检测", "PING Test"}, "ping.telegram": {"Telegram DC延迟", "Telegram DC Latency"},
		"ping.web_tcp": {"网站连接延迟", "Website TCP Latency"}, "disktest.deep_multi": {"多目录深度磁盘测试", "Deep Multi-Path Disk Test"},
		"basics.smart_selftest": {"SMART自检", "SMART Self-Test"},
		"basics.gpu_compute":    {"GPU计算测试", "GPU Compute Test"},
	}
	value, ok := titles[name]
	if !ok {
		return ""
	}
	if renderer.zh {
		return value[0]
	}
	return value[1]
}

func (renderer *structuredTextRenderer) status(status ReportStatus) string {
	return localizedValue(string(status), renderer.zh)
}

func (renderer *structuredTextRenderer) pick(zh, en string) string {
	if renderer.zh {
		return zh
	}
	return en
}

func payloadObject(payload json.RawMessage) map[string]any {
	var result map[string]any
	_ = json.Unmarshal(payload, &result)
	return result
}

func objectValue(value map[string]any, key string) map[string]any {
	result, _ := value[key].(map[string]any)
	return result
}

func arrayValue(value map[string]any, key string) []any {
	result, _ := value[key].([]any)
	return result
}

func stringArrayValue(value map[string]any, key string) []string {
	raw := arrayValue(value, key)
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
			result = append(result, text)
		}
	}
	return result
}

func int64ArrayValue(value map[string]any, key string) []int64 {
	raw := arrayValue(value, key)
	result := make([]int64, 0, len(raw))
	for _, item := range raw {
		if number, ok := item.(float64); ok {
			result = append(result, int64(number))
		}
	}
	return result
}

func stringValue(value map[string]any, key string) string {
	text, _ := value[key].(string)
	return strings.TrimSpace(text)
}

func floatValue(value map[string]any, key string) float64 {
	number, _ := value[key].(float64)
	return number
}

func intValue(value map[string]any, key string) int { return int(floatValue(value, key)) }

func int64Value(value map[string]any, key string) int64 { return int64(floatValue(value, key)) }

func boolValue(value map[string]any, key string) bool {
	result, _ := value[key].(bool)
	return result
}

func localizedValue(value string, zh bool) string {
	value = strings.TrimSpace(value)
	if !zh {
		if value == "" {
			return "-"
		}
		return strings.ReplaceAll(value, "_", " ")
	}
	values := map[string]string{
		"ok": "正常", "available": "可用", "unavailable": "不可用", "partial": "部分可用", "timeout": "超时",
		"canceled": "已取消", "error": "错误", "skipped": "已跳过", "unsupported": "不支持", "rate_limited": "限流",
		"missing_fields": "字段缺失", "permission_denied": "权限不足", "clean": "正常", "listed": "列入名单", "marked": "已标记",
		"Yes": "解锁", "No": "不解锁", "Restricted": "受限", "Banned": "封禁", "CDN Relay": "CDN中转", "RateLimited": "限流",
	}
	if result, ok := values[value]; ok {
		return result
	}
	if value == "" {
		return "-"
	}
	return strings.ReplaceAll(value, "_", " ")
}

func formatCounts(counts map[string]any, zh bool) string {
	converted := make(map[string]int, len(counts))
	for key, value := range counts {
		if number, ok := value.(float64); ok && number > 0 {
			converted[localizedValue(key, zh)] = int(number)
		}
	}
	return formatIntCounts(converted)
}

func formatIntCounts(counts map[string]int) string {
	keys := make([]string, 0, len(counts))
	for key, count := range counts {
		if count > 0 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s:%d", key, counts[key]))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}

func summarizeSecurityEvidence(provider map[string]any) string {
	var parts []string
	appendFields := func(values map[string]any, fields ...string) {
		for _, field := range fields {
			value, ok := values[field]
			if !ok || value == nil || fmt.Sprint(value) == "" {
				continue
			}
			parts = append(parts, humanizeKey(field)+"="+fmt.Sprint(value))
			if len(parts) >= 3 {
				return
			}
		}
	}
	appendFields(objectValue(provider, "score"), "FraudScore", "AbuseScore", "Reputation", "TrustScore", "VpnScore", "ProxyScore", "HumanPercent", "ThreatScore")
	if len(parts) < 3 {
		appendFields(objectValue(provider, "info"), "ThreatLevel", "UsageType", "IsProxy", "IsVpn", "IsTor", "IsDatacenter", "IsCloudProvider")
	}
	return strings.Join(parts, " ")
}

func summarizeRelationships(values []any) string {
	if len(values) == 0 {
		return "-"
	}
	parts := make([]string, 0, min(len(values), 3))
	for _, raw := range values {
		value, _ := raw.(map[string]any)
		parts = append(parts, fallback(stringValue(value, "asn"), stringValue(value, "name"), stringValue(value, "ixp_id")))
		if len(parts) == 3 {
			break
		}
	}
	result := strings.Join(parts, ", ")
	if len(values) > len(parts) {
		result += fmt.Sprintf(" (+%d)", len(values)-len(parts))
	}
	return result
}

func formatJSONDuration(value any) string {
	number, _ := value.(float64)
	if number <= 0 {
		return "-"
	}
	return formatDuration(time.Duration(int64(number)))
}

func formatDuration(value time.Duration) string {
	if value >= time.Second {
		return fmt.Sprintf("%.2f s", value.Seconds())
	}
	if value >= time.Millisecond {
		return fmt.Sprintf("%.2f ms", float64(value)/float64(time.Millisecond))
	}
	return fmt.Sprintf("%.2f us", float64(value)/float64(time.Microsecond))
}

func formatMilliseconds(value int64) string {
	if value <= 0 {
		return "-"
	}
	return formatDuration(time.Duration(value) * time.Millisecond)
}

func formatNanoseconds(value int64) string {
	if value <= 0 {
		return "-"
	}
	return formatDuration(time.Duration(value))
}

func formatBytes(value int64) string {
	if value <= 0 {
		return "-"
	}
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	number := float64(value)
	unit := 0
	for number >= 1024 && unit < len(units)-1 {
		number /= 1024
		unit++
	}
	return fmt.Sprintf("%.2f %s", number, units[unit])
}

func formatBytesPerSecond(value int64) string {
	if value <= 0 {
		return "-"
	}
	return formatBytes(value) + "/s"
}

func formatRate(value float64) string {
	if value <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f MiB/s", value)
}

func formatTemperature(value float64) string {
	if value == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f C", value)
}

func tuningTupleValues(values []int64) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if value > 0 {
			parts = append(parts, formatBytes(value))
		}
	}
	return strings.Join(parts, "/")
}

func structuredDriverList(items []any) string {
	values := make(map[string]struct{})
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		if value := strings.TrimSpace(stringValue(item, "driver")); value != "" {
			values[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	if len(result) > 4 {
		result = append(result[:4], fmt.Sprintf("+%d", len(result)-4))
	}
	return strings.Join(result, ",")
}

func structuredUniqueValues(items []any, key string) string {
	values := make(map[string]struct{})
	for _, raw := range items {
		item, _ := raw.(map[string]any)
		if value := strings.TrimSpace(stringValue(item, key)); value != "" {
			values[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return strings.Join(result, ",")
}

func fallback(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return "-"
}

func joinNonEmpty(values ...string) string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" && value != "-" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return strings.Join(result, " / ")
}

func humanizeKey(value string) string {
	words := strings.Fields(strings.ReplaceAll(value, "_", " "))
	for index := range words {
		words[index] = strings.ToUpper(words[index][:1]) + words[index][1:]
	}
	return strings.Join(words, " ")
}

func compactText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func truncateDisplay(value string, width int) string {
	if width <= 0 || runewidth.StringWidth(value) <= width {
		return value
	}
	if width <= 3 {
		return runewidth.Truncate(value, width, "")
	}
	return runewidth.Truncate(value, width-3, "") + "..."
}

func padDisplay(value string, width int) string {
	current := runewidth.StringWidth(value)
	if current >= width {
		return value
	}
	return value + strings.Repeat(" ", width-current)
}

func wrapDisplay(value string, width int) []string {
	if value == "" {
		return []string{""}
	}
	var result []string
	remaining := value
	for runewidth.StringWidth(remaining) > width {
		line := runewidth.Truncate(remaining, width, "")
		if line == "" {
			break
		}
		result = append(result, line)
		remaining = strings.TrimSpace(strings.TrimPrefix(remaining, line))
	}
	result = append(result, remaining)
	return result
}
