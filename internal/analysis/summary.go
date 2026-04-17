package analysis

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/oneclickvirt/ecs/internal/params"
)

var (
	mbpsRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*mbps`)
	msRe   = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*ms`)

	cpuModelZhRe = regexp.MustCompile(`(?im)^\s*CPU\s*型号\s*[:：]\s*(.+?)\s*$`)
	cpuModelEnRe = regexp.MustCompile(`(?im)^\s*CPU\s*Model\s*[:：]\s*(.+?)\s*$`)

	threadScoreEnRe = regexp.MustCompile(`(?im)^\s*(\d+)\s*Thread\(s\)\s*Test\s*:\s*([0-9][0-9,]*(?:\.[0-9]+)?)\s*$`)
	threadScoreZhRe = regexp.MustCompile(`(?im)^\s*(\d+)\s*线程测试\((?:单核|多核)\)得分\s*[:：]\s*([0-9][0-9,]*(?:\.[0-9]+)?)\s*$`)
	gbSingleRe      = regexp.MustCompile(`(?im)^\s*Single-Core\s*Score\s*[:：]\s*([0-9][0-9,]*(?:\.[0-9]+)?)\s*$`)
	gbMultiRe       = regexp.MustCompile(`(?im)^\s*Multi-Core\s*Score\s*[:：]\s*([0-9][0-9,]*(?:\.[0-9]+)?)\s*$`)

	alphaNumRe = regexp.MustCompile(`[a-z0-9]+`)

	// New patterns for condensed summary
	ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	streamFuncRe = regexp.MustCompile(`(?im)^\s*(?:Copy|Scale|Add|Triad)\s*:\s*(\d+(?:\.\d+)?)`)
	anyGbpsRe    = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*GB/s`)
	anyMbpsRe    = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*MB/s`)
)

const (
	cpuStatsPrimaryURL    = "https://raw.githubusercontent.com/oneclickvirt/ecs/ranks/cpu_statistics.json"
	cpuStatsFallbackURL   = "https://github.com/oneclickvirt/ecs/raw/refs/heads/ranks/cpu_statistics.json"
	cpuCDNProbeTestURL    = "https://raw.githubusercontent.com/spiritLHLS/ecs/main/back/test"
	cpuStatsCacheTTL      = 30 * time.Minute
	cpuStatsFailCacheTTL  = 5 * time.Minute
	cpuStatsRequestTimout = 6 * time.Second
)

var cpuStatsCDNList = []string{
	"https://cdn.spiritlhl.net/",
	"http://cdn3.spiritlhl.net/",
	"http://cdn1.spiritlhl.net/",
	"http://cdn2.spiritlhl.net/",
}

type cpuStatsEntry struct {
	CPUPrefix     string  `json:"cpu_prefix"`
	CPUModel      string  `json:"cpu_model"`
	SampleCount   int     `json:"sample_count"`
	MaxSingle     float64 `json:"max_single_score"`
	MaxMulti      float64 `json:"max_multi_score"`
	AvgSingle     float64 `json:"avg_single_score"`
	AvgMulti      float64 `json:"avg_multi_score"`
	Rank          int     `json:"rank"`
	TypicalCores  int     `json:"typical_cores"`
	TypicalThread int     `json:"typical_threads"`
}

type cpuStatsPayload struct {
	CPUStatistics []cpuStatsEntry `json:"cpu_statistics"`
}

var (
	cpuStatsMu       sync.Mutex
	cachedCPUStats   *cpuStatsPayload
	cpuStatsExpireAt time.Time
)

func parseFloatsByRegex(content string, re *regexp.Regexp) []float64 {
	matches := re.FindAllStringSubmatch(content, -1)
	vals := make([]float64, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		v, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			continue
		}
		vals = append(vals, v)
	}
	return vals
}

func parseFloatString(s string) (float64, bool) {
	clean := strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	v, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func extractCPUModel(output string) string {
	for _, re := range []*regexp.Regexp{cpuModelZhRe, cpuModelEnRe} {
		m := re.FindStringSubmatch(output)
		if len(m) >= 2 {
			model := strings.TrimSpace(m[1])
			if model != "" {
				return model
			}
		}
	}
	return ""
}

func extractCPUScores(output string) (single float64, singleOK bool, multi float64, multiOK bool) {
	for _, re := range []*regexp.Regexp{threadScoreEnRe, threadScoreZhRe} {
		matches := re.FindAllStringSubmatch(output, -1)
		for _, m := range matches {
			if len(m) < 3 {
				continue
			}
			threads, err := strconv.Atoi(strings.TrimSpace(m[1]))
			if err != nil {
				continue
			}
			score, ok := parseFloatString(m[2])
			if !ok {
				continue
			}
			if threads == 1 {
				single, singleOK = score, true
				continue
			}
			if threads > 1 && (!multiOK || score > multi) {
				multi, multiOK = score, true
			}
		}
	}

	if !singleOK {
		if m := gbSingleRe.FindStringSubmatch(output); len(m) >= 2 {
			if v, ok := parseFloatString(m[1]); ok {
				single, singleOK = v, true
			}
		}
	}
	if !multiOK {
		if m := gbMultiRe.FindStringSubmatch(output); len(m) >= 2 {
			if v, ok := parseFloatString(m[1]); ok {
				multi, multiOK = v, true
			}
		}
	}

	return
}

func normalizeCPUString(s string) string {
	s = strings.ToLower(s)
	b := strings.Builder{}
	b.Grow(len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func cpuTokens(s string) []string {
	lower := strings.ToLower(s)
	raw := alphaNumRe.FindAllString(lower, -1)
	if len(raw) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		if len(t) < 2 {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func fuzzyScoreCPUModel(model string, entry cpuStatsEntry) float64 {
	nm := normalizeCPUString(model)
	ne := normalizeCPUString(entry.CPUModel)
	np := normalizeCPUString(entry.CPUPrefix)
	if nm == "" || (ne == "" && np == "") {
		return 0
	}

	if nm == ne || nm == np {
		return 1
	}

	containsScore := 0.0
	for _, candidate := range []string{ne, np} {
		if candidate == "" {
			continue
		}
		if strings.Contains(candidate, nm) || strings.Contains(nm, candidate) {
			shortLen := len(nm)
			if len(candidate) < shortLen {
				shortLen = len(candidate)
			}
			longLen := len(nm)
			if len(candidate) > longLen {
				longLen = len(candidate)
			}
			if longLen > 0 {
				ratio := float64(shortLen) / float64(longLen)
				if ratio > containsScore {
					containsScore = ratio
				}
			}
		}
	}

	modelTokens := cpuTokens(model)
	if len(modelTokens) == 0 {
		return containsScore
	}

	entryTokenSet := make(map[string]struct{})
	for _, t := range cpuTokens(entry.CPUModel + " " + entry.CPUPrefix) {
		entryTokenSet[t] = struct{}{}
	}
	overlap := 0
	for _, t := range modelTokens {
		if _, ok := entryTokenSet[t]; ok {
			overlap++
		}
	}
	overlapScore := float64(overlap) / float64(len(modelTokens))

	if containsScore > overlapScore {
		return containsScore
	}
	return overlapScore
}

func loadCPUStats() *cpuStatsPayload {
	cpuStatsMu.Lock()
	defer cpuStatsMu.Unlock()

	now := time.Now()
	if now.Before(cpuStatsExpireAt) {
		return cachedCPUStats
	}

	client := req.C()
	client.SetTimeout(cpuStatsRequestTimout)
	endpoints := []string{cpuStatsPrimaryURL, cpuStatsFallbackURL}

	availableCDN := detectAvailableCPUCDN(client)
	for _, endpoint := range endpoints {
		urls := []string{}
		if availableCDN != "" {
			urls = append(urls, availableCDN+endpoint)
		}
		urls = append(urls, endpoint)

		for _, u := range urls {
			payload := tryDecodeCPUStatsFromURL(client, u)
			if payload == nil {
				continue
			}
			cachedCPUStats = payload
			cpuStatsExpireAt = now.Add(cpuStatsCacheTTL)
			return cachedCPUStats
		}
	}

	cachedCPUStats = nil
	cpuStatsExpireAt = now.Add(cpuStatsFailCacheTTL)
	return nil
}

func detectAvailableCPUCDN(client *req.Client) string {
	for _, baseURL := range cpuStatsCDNList {
		if checkCPUCDN(client, baseURL) {
			return baseURL
		}
		time.Sleep(500 * time.Millisecond)
	}
	return ""
}

func checkCPUCDN(client *req.Client, baseURL string) bool {
	resp, err := client.R().SetHeader("User-Agent", "goecs-summary/1.0").Get(baseURL + cpuCDNProbeTestURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	if err != nil {
		return false
	}
	return strings.Contains(string(b), "success")
}

func tryDecodeCPUStatsFromURL(client *req.Client, u string) *cpuStatsPayload {
	resp, err := client.R().SetHeader("User-Agent", "goecs-summary/1.0").Get(u)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}

	var payload cpuStatsPayload
	dec := json.NewDecoder(io.LimitReader(resp.Body, 8<<20))
	if err := dec.Decode(&payload); err != nil {
		return nil
	}
	if len(payload.CPUStatistics) == 0 {
		return nil
	}
	return &payload
}

func matchCPUStatsEntry(model string, payload *cpuStatsPayload) *cpuStatsEntry {
	if payload == nil || model == "" || len(payload.CPUStatistics) == 0 {
		return nil
	}

	trimModel := strings.TrimSpace(model)
	normModel := normalizeCPUString(trimModel)

	for i := range payload.CPUStatistics {
		entry := &payload.CPUStatistics[i]
		if strings.EqualFold(strings.TrimSpace(entry.CPUModel), trimModel) {
			return entry
		}
	}

	for i := range payload.CPUStatistics {
		entry := &payload.CPUStatistics[i]
		if normModel == normalizeCPUString(entry.CPUModel) || normModel == normalizeCPUString(entry.CPUPrefix) {
			return entry
		}
	}

	bestIdx := -1
	bestScore := 0.0
	for i := range payload.CPUStatistics {
		score := fuzzyScoreCPUModel(trimModel, payload.CPUStatistics[i])
		if score > bestScore {
			bestScore = score
			bestIdx = i
			continue
		}
		if score == bestScore && bestIdx >= 0 && payload.CPUStatistics[i].SampleCount > payload.CPUStatistics[bestIdx].SampleCount {
			bestIdx = i
		}
	}

	if bestIdx >= 0 && bestScore >= 0.45 {
		return &payload.CPUStatistics[bestIdx]
	}
	return nil
}

func cpuTierText(score float64, lang string) string {
	if lang == "zh" {
		switch {
		case score >= 5000:
			return "按 README_NEW_USER 的 Sysbench 口径，单核 >5000 可视为高性能第一梯队。"
		case score < 500:
			return "按 README_NEW_USER 的 Sysbench 口径，单核 <500 属于偏弱性能。"
		default:
			return "按 README_NEW_USER 的 Sysbench 口径，可按每约 1000 分视作一个性能档位。"
		}
	}

	switch {
	case score >= 5000:
		return "Per README_NEW_USER Sysbench guidance, single-core > 5000 is considered first-tier high performance."
	case score < 500:
		return "Per README_NEW_USER Sysbench guidance, single-core < 500 is considered weak performance."
	default:
		return "Per README_NEW_USER Sysbench guidance, roughly every 1000 points is about one performance tier."
	}
}

func summarizeCPUWithRanking(finalOutput, lang string) []string {
	model := extractCPUModel(finalOutput)
	single, singleOK, multi, multiOK := extractCPUScores(finalOutput)
	if !singleOK && !multiOK {
		return nil
	}

	stats := loadCPUStats()
	entry := matchCPUStatsEntry(model, stats)

	var score float64
	var avg float64
	var max float64
	kind := "single"

	if singleOK && entry != nil && entry.AvgSingle > 0 && entry.MaxSingle > 0 {
		score, avg, max = single, entry.AvgSingle, entry.MaxSingle
	} else if multiOK && entry != nil && entry.AvgMulti > 0 && entry.MaxMulti > 0 {
		score, avg, max = multi, entry.AvgMulti, entry.MaxMulti
		kind = "multi"
	} else if singleOK {
		score = single
	} else {
		score = multi
		kind = "multi"
	}

	lines := make([]string, 0, 4)
	if lang == "zh" {
		if kind == "single" {
			lines = append(lines, fmt.Sprintf("CPU: 检测到单核得分 %.2f。", score))
		} else {
			lines = append(lines, fmt.Sprintf("CPU: 检测到多核得分 %.2f。", score))
		}
	} else {
		if kind == "single" {
			lines = append(lines, fmt.Sprintf("CPU: detected single-core score %.2f.", score))
		} else {
			lines = append(lines, fmt.Sprintf("CPU: detected multi-core score %.2f.", score))
		}
	}

	if kind == "single" {
		lines = append(lines, cpuTierText(score, lang))
	}

	if entry == nil || avg <= 0 || max <= 0 {
		if lang == "zh" {
			if model != "" {
				lines = append(lines, fmt.Sprintf("CPU 对标: 未在在线榜单中稳定匹配到型号 \"%s\"，已仅给出本机分数解读。", model))
			} else {
				lines = append(lines, "CPU 对标: 未提取到 CPU 型号，已仅给出本机分数解读。")
			}
		} else {
			if model != "" {
				lines = append(lines, fmt.Sprintf("CPU ranking: no reliable online match found for model \"%s\"; local score interpretation only.", model))
			} else {
				lines = append(lines, "CPU ranking: CPU model not found in output; local score interpretation only.")
			}
		}
		return lines
	}

	reachAvg := score >= avg
	gapToMax := max - score
	fullBlood := false
	if max > 0 {
		ratioDiff := (score - max) / max
		if ratioDiff < 0 {
			ratioDiff = -ratioDiff
		}
		fullBlood = ratioDiff <= 0.05
	}
	pctOfAvg := score / avg * 100
	pctOfMax := score / max * 100

	if lang == "zh" {
		lines = append(lines,
			fmt.Sprintf("CPU 对标: 匹配 \"%s\"（样本 %d，排名 #%d）。", entry.CPUModel, entry.SampleCount, entry.Rank),
			fmt.Sprintf("平均分达标: %s（本机 %.2f，均值 %.2f，达成率 %.2f%%）。", map[bool]string{true: "是", false: "否"}[reachAvg], score, avg, pctOfAvg),
			fmt.Sprintf("满血对比: 满血分 %.2f，本机为 %.2f%%，差值 %.2f。", max, pctOfMax, gapToMax),
			fmt.Sprintf("满血判定(±5%%波动): %s。", map[bool]string{true: "是", false: "否"}[fullBlood]),
		)
	} else {
		lines = append(lines,
			fmt.Sprintf("CPU ranking: matched \"%s\" (samples %d, rank #%d).", entry.CPUModel, entry.SampleCount, entry.Rank),
			fmt.Sprintf("Average-level check: %s (local %.2f vs avg %.2f, %.2f%% of avg).", map[bool]string{true: "pass", false: "below avg"}[reachAvg], score, avg, pctOfAvg),
			fmt.Sprintf("Full-blood comparison: max %.2f, local is %.2f%% of max, gap %.2f.", max, pctOfMax, gapToMax),
			fmt.Sprintf("Full-blood status (within ±5%%): %s.", map[bool]string{true: "yes", false: "no"}[fullBlood]),
		)
	}

	return lines
}

func summarizeBandwidth(vals []float64, lang string) string {
	if len(vals) == 0 {
		if lang == "zh" {
			return "测速: 未检测到有效 Mbps 数据。"
		}
		return "Speed: no valid Mbps values found."
	}
	sort.Float64s(vals)
	maxV := vals[len(vals)-1]
	if lang == "zh" {
		switch {
		case maxV >= 2000:
			return fmt.Sprintf("测速: 峰值约 %.2f Mbps，属于高带宽网络。", maxV)
		case maxV >= 800:
			return fmt.Sprintf("测速: 峰值约 %.2f Mbps，带宽表现较好。", maxV)
		case maxV >= 200:
			return fmt.Sprintf("测速: 峰值约 %.2f Mbps，带宽中等可用。", maxV)
		default:
			return fmt.Sprintf("测速: 峰值约 %.2f Mbps，带宽偏低，建议关注线路与机型。", maxV)
		}
	}
	switch {
	case maxV >= 2000:
		return fmt.Sprintf("Speed: peak around %.2f Mbps, high-bandwidth profile.", maxV)
	case maxV >= 800:
		return fmt.Sprintf("Speed: peak around %.2f Mbps, strong bandwidth performance.", maxV)
	case maxV >= 200:
		return fmt.Sprintf("Speed: peak around %.2f Mbps, moderate and usable bandwidth.", maxV)
	default:
		return fmt.Sprintf("Speed: peak around %.2f Mbps, relatively limited bandwidth.", maxV)
	}
}

func summarizeLatency(vals []float64, lang string) string {
	if len(vals) == 0 {
		if lang == "zh" {
			return "延迟: 未检测到有效 ms 数据。"
		}
		return "Latency: no valid ms values found."
	}
	sort.Float64s(vals)
	minV := vals[0]
	if lang == "zh" {
		switch {
		case minV <= 15:
			return fmt.Sprintf("延迟: 最优约 %.2f ms，实时交互体验优秀。", minV)
		case minV <= 45:
			return fmt.Sprintf("延迟: 最优约 %.2f ms，整体交互体验良好。", minV)
		case minV <= 90:
			return fmt.Sprintf("延迟: 最优约 %.2f ms，可用但有一定时延。", minV)
		default:
			return fmt.Sprintf("延迟: 最优约 %.2f ms，时延偏高，建议优化线路。", minV)
		}
	}
	switch {
	case minV <= 15:
		return fmt.Sprintf("Latency: best around %.2f ms, excellent for interactive workloads.", minV)
	case minV <= 45:
		return fmt.Sprintf("Latency: best around %.2f ms, generally responsive.", minV)
	case minV <= 90:
		return fmt.Sprintf("Latency: best around %.2f ms, usable with moderate delay.", minV)
	default:
		return fmt.Sprintf("Latency: best around %.2f ms, relatively high and may impact responsiveness.", minV)
	}
}

func testedScopes(config *params.Config) []string {
	scopes := make([]string, 0, 8)
	if config.BasicStatus {
		scopes = append(scopes, "basic")
	}
	if config.CpuTestStatus {
		scopes = append(scopes, "cpu")
	}
	if config.MemoryTestStatus {
		scopes = append(scopes, "memory")
	}
	if config.DiskTestStatus {
		scopes = append(scopes, "disk")
	}
	if config.UtTestStatus {
		scopes = append(scopes, "unlock")
	}
	if config.SecurityTestStatus {
		scopes = append(scopes, "security")
	}
	if config.Nt3Status || config.BacktraceStatus || config.PingTestStatus || config.TgdcTestStatus || config.WebTestStatus {
		scopes = append(scopes, "network")
	}
	if config.SpeedTestStatus {
		scopes = append(scopes, "speed")
	}
	return scopes
}

func scopesText(scopes []string, lang string) string {
	if len(scopes) == 0 {
		if lang == "zh" {
			return "无"
		}
		return "none"
	}
	labelsZh := map[string]string{
		"basic": "系统基础", "cpu": "CPU", "memory": "内存", "disk": "磁盘", "unlock": "解锁", "security": "IP质量", "network": "网络路由", "speed": "带宽测速",
	}
	labelsEn := map[string]string{
		"basic": "system basics", "cpu": "CPU", "memory": "memory", "disk": "disk", "unlock": "unlock", "security": "IP quality", "network": "network route", "speed": "bandwidth",
	}
	out := make([]string, 0, len(scopes))
	for _, s := range scopes {
		if lang == "zh" {
			out = append(out, labelsZh[s])
		} else {
			out = append(out, labelsEn[s])
		}
	}
	return strings.Join(out, ", ")
}

// stripAnsiCodes removes ANSI escape codes from a string.
func stripAnsiCodes(s string) string {
	return ansiEscapeRe.ReplaceAllString(s, "")
}

// extractAllSectionContent returns the concatenated text of every section whose
// header contains markerZh or markerEn (case-sensitive). Section headers are
// lines that start with three or more dashes (produced by PrintCenteredTitle).
func extractAllSectionContent(output, markerZh, markerEn string) string {
	lines := strings.Split(output, "\n")
	var sb strings.Builder
	inSection := false
	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "---") {
			if strings.Contains(stripped, markerZh) || (markerEn != "" && strings.Contains(stripped, markerEn)) {
				inSection = true
				continue
			}
			// Any other section header ends the current one
			inSection = false
			continue
		}
		if inSection {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// extractMaxMemoryBandwidth returns the highest bandwidth value (in MB/s) found
// inside the memory-test section(s) of the captured output.
func extractMaxMemoryBandwidth(output string) float64 {
	content := extractAllSectionContent(output, "内存测试", "Memory-Test")
	if content == "" {
		return 0
	}
	maxMbps := 0.0
	// STREAM format: "Copy:  12345.6 ..." – the first number is Best Rate MB/s
	for _, m := range streamFuncRe.FindAllStringSubmatch(content, -1) {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil && v > maxMbps {
			maxMbps = v
		}
	}
	// Values reported as GB/s (some dd / mbw outputs)
	for _, m := range anyGbpsRe.FindAllStringSubmatch(content, -1) {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			if mbps := v * 1024; mbps > maxMbps {
				maxMbps = mbps
			}
		}
	}
	// Values reported as MB/s (mbw, sysbench, dd, winsat …)
	for _, m := range anyMbpsRe.FindAllStringSubmatch(content, -1) {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil && v > maxMbps {
			maxMbps = v
		}
	}
	return maxMbps
}

// inferMemoryDDRAndChannels converts a memory bandwidth (MB/s) to a human-readable
// DDR type + channel string using the thresholds from README_NEW_USER.
//
//	DDR3 single: 10240–17408 MB/s
//	DDR4 single: 17408–34816 MB/s
//	DDR4 dual:   34816–51200 MB/s
//	DDR5 single: 51200–77824 MB/s
//	DDR5 dual:   ≥77824 MB/s
func inferMemoryDDRAndChannels(mbps float64, lang string) string {
	type tier struct {
		minMbps         float64
		ddr, chZh, chEn string
	}
	tiers := []tier{
		{77824, "DDR5", "双通道", "Dual-Channel"},
		{51200, "DDR5", "单通道", "Single-Channel"},
		{34816, "DDR4", "双通道", "Dual-Channel"},
		{17408, "DDR4", "单通道", "Single-Channel"},
		{0, "DDR3", "单通道", "Single-Channel"},
	}
	for _, t := range tiers {
		if mbps >= t.minMbps {
			if lang == "zh" {
				return t.ddr + " " + t.chZh
			}
			return t.ddr + " " + t.chEn
		}
	}
	if lang == "zh" {
		return "DDR3 单通道"
	}
	return "DDR3 Single-Channel"
}

// extractDiskTypeAndCount scans all disk-test section(s) for fio 4K rows and
// returns the 4K read speed (MB/s) and the number of unique test paths found.
// Falls back to any MB/s / GB/s value in the disk section when no 4K rows exist.
func extractDiskTypeAndCount(output string) (readMbps float64, pathCount int) {
	content := extractAllSectionContent(output, "硬盘测试", "Disk-Test")
	if content == "" {
		return 0, 0
	}
	pathSet := make(map[string]struct{})
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		// fio output row: <path> <blocksize> <readSpeed> <readUnit(iops)> ...
		if len(fields) < 4 {
			continue
		}
		if !strings.EqualFold(fields[1], "4k") {
			continue
		}
		pathSet[fields[0]] = struct{}{}
		val, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			continue
		}
		unit := strings.ToUpper(fields[3])
		var mbps float64
		if strings.HasPrefix(unit, "GB") {
			mbps = val * 1024
		} else {
			mbps = val
		}
		if mbps > readMbps {
			readMbps = mbps
		}
	}
	pathCount = len(pathSet)
	// Fallback: no 4K rows found – use any speed value present (dd / winsat)
	if readMbps == 0 {
		for _, m := range anyGbpsRe.FindAllStringSubmatch(content, -1) {
			if v, err := strconv.ParseFloat(m[1], 64); err == nil {
				if mbps := v * 1024; mbps > readMbps {
					readMbps = mbps
				}
			}
		}
		for _, m := range anyMbpsRe.FindAllStringSubmatch(content, -1) {
			if v, err := strconv.ParseFloat(m[1], 64); err == nil && v > readMbps {
				readMbps = v
			}
		}
	}
	if pathCount == 0 && readMbps > 0 {
		pathCount = 1
	}
	return readMbps, pathCount
}

// inferDiskType classifies a disk by its 4K (or sequential fallback) read speed.
//
//	NVMe SSD : ≥200 MB/s
//	SATA SSD : 50–200 MB/s
//	HDD      : 10–50 MB/s
func inferDiskType(readMbps float64, lang string) string {
	switch {
	case readMbps >= 200:
		return "NVMe SSD"
	case readMbps >= 50:
		if lang == "zh" {
			return "SATA SSD"
		}
		return "SATA SSD"
	case readMbps >= 10:
		return "HDD"
	default:
		if lang == "zh" {
			return "低性能磁盘"
		}
		return "Low-Perf Disk"
	}
}

// extractISPRanking parses the backtrace section and returns a ranking string
// like "电信 > 联通 > 移动" based on the best route quality detected per ISP.
// Quality tiers: [精品线路]=3, [优质线路]=2, [普通线路]=1.
func extractISPRanking(output, lang string) string {
	content := extractAllSectionContent(output, "上游及回程线路检测", "Upstream")
	if content == "" {
		return ""
	}
	scores := map[string]int{"电信": 0, "联通": 0, "移动": 0}
	for _, raw := range strings.Split(content, "\n") {
		line := stripAnsiCodes(raw)
		var q int
		switch {
		case strings.Contains(line, "[精品线路]"):
			q = 3
		case strings.Contains(line, "[优质线路]"):
			q = 2
		case strings.Contains(line, "[普通线路]"):
			q = 1
		default:
			continue
		}
		for isp := range scores {
			if strings.Contains(line, isp) && q > scores[isp] {
				scores[isp] = q
			}
		}
	}
	if scores["电信"] == 0 && scores["联通"] == 0 && scores["移动"] == 0 {
		return ""
	}
	isps := []string{"电信", "联通", "移动"}
	order := map[string]int{"电信": 0, "联通": 1, "移动": 2}
	sort.Slice(isps, func(i, j int) bool {
		si, sj := scores[isps[i]], scores[isps[j]]
		if si != sj {
			return si > sj
		}
		return order[isps[i]] < order[isps[j]]
	})
	return strings.Join(isps, " > ")
}

// extractCPURankCondensed returns "CPU排名 #N 为满血性能的XX.XX%" (zh) or the
// English equivalent, using the same CPU stats lookup as summarizeCPUWithRanking.
func extractCPURankCondensed(finalOutput, lang string) string {
	model := extractCPUModel(finalOutput)
	single, singleOK, multi, multiOK := extractCPUScores(finalOutput)
	if !singleOK && !multiOK {
		return ""
	}
	stats := loadCPUStats()
	entry := matchCPUStatsEntry(model, stats)
	if entry == nil || entry.Rank <= 0 {
		return ""
	}
	var score, maxScore float64
	if singleOK && entry.MaxSingle > 0 {
		score, maxScore = single, entry.MaxSingle
	} else if multiOK && entry.MaxMulti > 0 {
		score, maxScore = multi, entry.MaxMulti
	} else {
		return ""
	}
	pct := score / maxScore * 100
	if lang == "zh" {
		return fmt.Sprintf("CPU排名 #%d 为满血性能的%.2f%%", entry.Rank, pct)
	}
	return fmt.Sprintf("CPU rank #%d is %.2f%% of full performance", entry.Rank, pct)
}

// extractBandwidthCondensed returns the peak bandwidth as a human-readable
// "网络峰值带宽大于 X.XXGbps" / "> X.XXGbps" string.
func extractBandwidthCondensed(vals []float64, lang string) string {
	if len(vals) == 0 {
		return ""
	}
	sort.Float64s(vals)
	maxV := vals[len(vals)-1]
	if lang == "zh" {
		if maxV >= 1000 {
			return fmt.Sprintf("网络峰值带宽大于 %.2fGbps", maxV/1000)
		}
		return fmt.Sprintf("网络峰值带宽大于 %.2fMbps", maxV)
	}
	if maxV >= 1000 {
		return fmt.Sprintf("Peak bandwidth > %.2fGbps", maxV/1000)
	}
	return fmt.Sprintf("Peak bandwidth > %.2fMbps", maxV)
}

// sectionExists reports whether the output contains a section header matching
// the given Chinese or English marker.
func sectionExists(output, markerZh, markerEn string) bool {
	for _, line := range strings.Split(output, "\n") {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "---") {
			if strings.Contains(stripped, markerZh) || (markerEn != "" && strings.Contains(stripped, markerEn)) {
				return true
			}
		}
	}
	return false
}

// GenerateSummary creates a concise one-line post-test summary from final output.
func GenerateSummary(config *params.Config, finalOutput string) string {
	lang := config.Language
	parts := make([]string, 0, 5)

	// helper: localised N/A string
	na := func() string {
		if lang == "zh" {
			return "无有效数据"
		}
		return "N/A"
	}

	// 1. CPU rank and full-blood percentage
	if config.CpuTestStatus {
		s := extractCPURankCondensed(finalOutput, lang)
		if s != "" {
			parts = append(parts, s)
		} else if sectionExists(finalOutput, "CPU测试", "CPU-Test") {
			// Section ran but no rank could be derived (test failure or no DB match)
			if lang == "zh" {
				parts = append(parts, "CPU排名: "+na())
			} else {
				parts = append(parts, "CPU rank: "+na())
			}
		}
		// If the section doesn't appear at all, the test simply wasn't run – omit silently.
	}

	// 2. Memory DDR type and channels, with average-level check
	if config.MemoryTestStatus {
		bw := extractMaxMemoryBandwidth(finalOutput)
		if bw > 0 {
			mem := inferMemoryDDRAndChannels(bw, lang)
			// README_NEW_USER threshold: < 10240 MB/s (≈10 GB/s) indicates overselling risk
			const memAvgThreshMbps = 10240.0
			if lang == "zh" {
				if bw >= memAvgThreshMbps {
					parts = append(parts, "内存为 "+mem+"(达标)")
				} else {
					parts = append(parts, "内存为 "+mem+"(未达标)")
				}
			} else {
				if bw >= memAvgThreshMbps {
					parts = append(parts, "Memory: "+mem+"(pass)")
				} else {
					parts = append(parts, "Memory: "+mem+"(below avg)")
				}
			}
		} else if sectionExists(finalOutput, "内存测试", "Memory-Test") {
			if lang == "zh" {
				parts = append(parts, "内存: "+na())
			} else {
				parts = append(parts, "Memory: "+na())
			}
		}
	}

	// 3. Disk type and path count, with average-level check
	if config.DiskTestStatus {
		readMbps, pathCount := extractDiskTypeAndCount(finalOutput)
		if readMbps > 0 || pathCount > 0 {
			if pathCount <= 0 {
				pathCount = 1
			}
			dtype := inferDiskType(readMbps, lang)
			// README_NEW_USER: < 10 MB/s = poor performance / severe overselling
			diskOK := readMbps >= 10
			if lang == "zh" {
				var label string
				if diskOK {
					label = "(达标)"
				} else {
					label = "(未达标)"
				}
				parts = append(parts, fmt.Sprintf("硬盘IO为 %s %d路%s", dtype, pathCount, label))
			} else {
				var label string
				if diskOK {
					label = "(pass)"
				} else {
					label = "(below avg)"
				}
				parts = append(parts, fmt.Sprintf("Disk IO: %s %d path(s)%s", dtype, pathCount, label))
			}
		} else if sectionExists(finalOutput, "硬盘测试", "Disk-Test") {
			if lang == "zh" {
				parts = append(parts, "硬盘IO: "+na())
			} else {
				parts = append(parts, "Disk IO: "+na())
			}
		}
	}

	// 4. Network peak bandwidth
	if config.SpeedTestStatus {
		bwVals := parseFloatsByRegex(finalOutput, mbpsRe)
		s := extractBandwidthCondensed(bwVals, lang)
		if s != "" {
			parts = append(parts, s)
		} else if sectionExists(finalOutput, "测速", "Speed-Test") {
			if lang == "zh" {
				parts = append(parts, "网络带宽: "+na())
			} else {
				parts = append(parts, "Bandwidth: "+na())
			}
		}
	}

	// 5. Domestic ISP ranking — only meaningful in Chinese mode (backtrace targets CN ISPs)
	if lang == "zh" && config.BacktraceStatus {
		if ranking := extractISPRanking(finalOutput, lang); ranking != "" {
			parts = append(parts, "国内三大运营商推荐排名为 "+ranking)
		} else if sectionExists(finalOutput, "上游及回程线路检测", "") {
			parts = append(parts, "国内三大运营商推荐排名: "+na())
		}
	}

	if len(parts) == 0 {
		if lang == "zh" {
			return "测试总结: 无足够数据生成摘要。"
		}
		return "Test Summary: insufficient data for summary."
	}

	prefix := "测试总结: "
	if lang != "zh" {
		prefix = "Test Summary: "
	}
	return prefix + strings.Join(parts, " | ")
}
