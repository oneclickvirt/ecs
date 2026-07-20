package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mattn/go-runewidth"
)

func TestRenderStructuredRunTextKeepsCompactProjectStyle(t *testing.T) {
	config := NewConfig("v-test")
	config.Language = "zh"
	config.Width = 80
	components := []ComponentReport{
		componentFixture(t, "cputest", ReportStatusOK, `{
			"schema_version":"goecs.cpu/v1","status":"ok","requested_threads":8,
			"effective_threads":4,"events_per_second":1234.5,"duration_ms":5000,
			"temperature":{"available":true,"baseline_c":42.0,"max_c":58.5,"delta_c":16.5}
		}`),
		componentFixture(t, "memorytest", ReportStatusOK, `{
			"schema_version":"goecs.memory/v1","status":"ok","working_set_bytes":33554432,
			"sequential_read_mbps":12345.6,"sequential_write_mbps":9876.5,
			"copy_mbps":8765.4,"random_latency_ns":81.2
		}`),
		componentFixture(t, "speed.registry", ReportStatusPartial, `{
			"schema_version":"goecs.speed/v1","benchmarks":[
				{"name":"Local Test Node","status":"available","latency_ms":5.2,"download_mbps":900.1,"upload_mbps":700.2}
			]
		}`),
	}
	reports := []TCPReport{{
		Target: TCPTarget{Name: "Example TCP"}, Attempts: 3, Successful: 2,
		MeanMS: 12.3, P95MS: 18.5, LossPercent: 33.3, Errors: map[string]int{"timeout": 1},
	}}
	text := renderStructuredRunText(config, []DataFileVersion{{
		File: "tcp-targets.json", Source: "embedded", Fallback: "embedded", Count: 64,
		Status: ReportStatusOK, GeneratedAt: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
	}}, components, reports)
	for _, want := range []string{"VPS融合怪测试", "数据源状态", "CPU性能测试", "内存性能测试", "就近节点测速", "TCP连接质量", "Example TCP"} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered text missing %q:\n%s", want, text)
		}
	}
	for _, forbidden := range []string{"schema_version", "events_per_second", "\"benchmarks\"", "{", "}"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("rendered text exposed machine field %q:\n%s", forbidden, text)
		}
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "https://") {
			continue
		}
		if width := runewidth.StringWidth(line); width > config.Width {
			t.Fatalf("line width %d exceeds %d: %q", width, config.Width, line)
		}
	}
}

func TestRenderStructuredRunTextUsesEnglishTitles(t *testing.T) {
	config := NewConfig("v-test")
	config.Language = "en"
	config.Width = 72
	text := renderStructuredRunText(config, nil, []ComponentReport{
		componentFixture(t, "gostun.nat", ReportStatusOK, `{
			"schema_version":"goecs.stun/v1","ip_version":"ipv4","status":"available",
			"successful":1,"failed":0,"mapping_consistency":"available",
			"port_preservation_consistency":"available","hairpin_consistency":"unsupported",
			"results":[{"nat_type":"Full Cone","mapping_behavior":"endpoint-independent","filtering_behavior":"address-dependent"}]
		}`),
	}, nil)
	for _, want := range []string{"VPS Fusion Monster Test", "NAT Type Check", "Mapping Consistency", "Port Preservation", "Hairpin"} {
		if !strings.Contains(text, want) {
			t.Fatalf("English text missing %q:\n%s", want, text)
		}
	}
}

func TestRenderStructuredRunTextCoversEveryStructuredSection(t *testing.T) {
	config := NewConfig("v-test")
	config.Width = 100
	fixtures := []ComponentReport{
		componentFixture(t, "basics", ReportStatusOK, `{"cpu":{"model":"Fixture CPU","logical_cpus":4},"memory":{"total_bytes":1073741824},"cgroup":{"version":"v2"},"virtualization":{"type":"kvm"},"network":{},"firmware":{},"pci":{},"memory_topology":{},"raid":{}}`),
		componentFixture(t, "disktest", ReportStatusOK, `{"metrics":[{"scenario_id":"4k-q1-read","iops":1000,"bandwidth_bytes_per_second":1048576,"latency_p50_ns":1000,"latency_p95_ns":2000,"latency_p99_ns":3000}]}`),
		componentFixture(t, "unlocktests.media", ReportStatusPartial, `{"results":[{"name":"Dola AI","status":"Yes","region":"US"},{"name":"X","status":"RateLimited"}]}`),
		componentFixture(t, "security.evidence", ReportStatusPartial, `{"addresses":[{"ip_type":"ipv4","ip":"198.51.100.2","providers":[{"source":"fixture","status":"available","score":{"FraudScore":5}}],"dnsbl":{"counts":{"clean":2,"listed":1}}}]}`),
		componentFixture(t, "backtrace.ip_bgp", ReportStatusOK, `{"reports":[{"ip":"198.51.100.2","asn":"64500","prefixes":["198.51.100.0/24"],"rir":{"name":"TEST"},"relationships":{"upstreams":[],"peers":[],"ixps":[]}}]}`),
		componentFixture(t, "portchecker.email", ReportStatusPartial, `{"local":[{"status":"available"}],"outbound_smtp25":[{"status":"timeout"}],"mx":[],"fixed":[]}`),
		componentFixture(t, "nt3.province_latency", ReportStatusOK, `[{"target":{"province_name":"广东","carrier":"ct"},"attempts":2,"successful":2,"mean":1000000,"p95":2000000}]`),
		componentFixture(t, "nt3.province_routes", ReportStatusOK, `[{"target":{"province_name":"广东","carrier":"ct","ip_version":"ipv4"},"status":"ok","duration":1000000,"hops":[{}]}]`),
		componentFixture(t, "ping.icmp", ReportStatusOK, `[{"target":{"name":"Fixture"},"status":"ok","sent":3,"received":3,"mean":1000000,"p95":2000000}]`),
		componentFixture(t, "ping.telegram", ReportStatusOK, `[{"target":{"name":"DC1"},"status":"ok","sent":3,"received":3,"mean":1000000,"p95":2000000}]`),
		componentFixture(t, "ping.web_tcp", ReportStatusOK, `[{"target":{"name":"Website"},"attempts":3,"successful":3,"mean":1000000,"p95":2000000}]`),
		componentFixture(t, "gostun.nat", ReportStatusOK, `{"successful":1,"failed":0,"mapping_consistency":"available","port_preservation_consistency":"available","hairpin_consistency":"unsupported","results":[]}`),
		componentFixture(t, "disktest.deep_multi", ReportStatusOK, `{"paths":[{"status":"ok","metrics":[]}]}`),
		componentFixture(t, "basics.smart_selftest", ReportStatusSkipped, `{"results":[]}`),
		componentFixture(t, "cputest.burn", ReportStatusOK, `{"effective_threads":2,"events_per_second":200,"duration_ms":1000,"temperature":{}}`),
		componentFixture(t, "basics.gpu_compute", ReportStatusSkipped, `{"status":"skipped","target":"gpu0"}`),
	}
	text := renderStructuredRunText(config, nil, fixtures, nil)
	for _, want := range []string{
		"系统基础信息", "磁盘性能测试", "跨国平台解锁", "IP质量检测", "上游及注册信息", "邮件端口检测",
		"全国三网延迟", "全国三网详细路由", "PING值检测", "Telegram DC延迟", "网站连接延迟", "NAT类型检测",
		"多目录深度磁盘测试", "SMART自检", "CPU压力测试", "GPU计算测试",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("all-section render missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "schema_version") || strings.Contains(text, "\"results\"") || strings.Contains(text, "{") {
		t.Fatalf("all-section render exposed raw JSON:\n%s", text)
	}
}

func TestPrivacyModeCanRenderRedactedStructuredText(t *testing.T) {
	report := &StructuredReport{
		Components: []ComponentReport{componentFixture(t, "backtrace.ip_bgp", ReportStatusOK, `{
			"schema_version":"goecs.backtrace/v1","reports":[{"ip":"203.0.113.9","asn":"64500","prefixes":["203.0.113.0/24"],"rir":{"name":"TEST"}}]
		}`)},
		TCP: []TCPReport{{Target: TCPTarget{Name: "203.0.113.9 service", Host: "203.0.113.9"}, Attempts: 1, Successful: 1}},
	}
	applyStructuredPrivacy(report)
	config := NewConfig("v-test")
	config.PrivacyMode = true
	text := renderStructuredRunText(config, report.DataFiles, report.Components, report.TCP)
	if strings.Contains(text, "203.0.113.9") || strings.Contains(text, "203.0.113.0") {
		t.Fatalf("privacy text leaked an address:\n%s", text)
	}
	if !strings.Contains(text, "[redacted") {
		t.Fatalf("privacy text did not preserve a redacted result:\n%s", text)
	}
}

func componentFixture(t *testing.T, name string, status ReportStatus, payload string) ComponentReport {
	t.Helper()
	if !json.Valid([]byte(payload)) {
		t.Fatalf("invalid fixture payload for %s", name)
	}
	return ComponentReport{Name: name, Status: status, Payload: json.RawMessage(payload)}
}
