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
	cpuStatsMu        sync.Mutex
	cachedCPUStats    *cpuStatsPayload
	cpuStatsExpireAt  time.Time
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

// GenerateSummary creates a concise post-test summary from final output.
func GenerateSummary(config *params.Config, finalOutput string) string {
	lang := config.Language
	scopes := testedScopes(config)
	bandwidthVals := parseFloatsByRegex(finalOutput, mbpsRe)
	latencyVals := parseFloatsByRegex(finalOutput, msRe)
	cpuLines := summarizeCPUWithRanking(finalOutput, lang)

	if lang == "zh" {
		lines := []string{
			"测试结果总结:",
			fmt.Sprintf("- 本次覆盖: %s", scopesText(scopes, lang)),
		}
		for _, line := range cpuLines {
			lines = append(lines, "- "+line)
		}
		if config.SpeedTestStatus {
			lines = append(lines, "- "+summarizeBandwidth(bandwidthVals, lang))
			lines = append(lines, "- 参考 README_NEW_USER: 一般境外机器带宽 100Mbps 起步，是否够用应以业务下载/传输需求为准。")
		}
		if config.PingTestStatus || config.TgdcTestStatus || config.WebTestStatus || config.BacktraceStatus || config.Nt3Status {
			lines = append(lines, "- "+summarizeLatency(latencyVals, lang))
			lines = append(lines, "- 参考 README_NEW_USER: 延迟 >= 9999ms 可视为目标不可用。")
		}
		lines = append(lines, "- 建议: 结合业务场景(高并发计算/存储/跨境网络)重点参考对应分项。")
		return strings.Join(lines, "\n")
	}

	lines := []string{
		"Test Summary:",
		fmt.Sprintf("- Scope covered: %s", scopesText(scopes, lang)),
	}
	for _, line := range cpuLines {
		lines = append(lines, "- "+line)
	}
	if config.SpeedTestStatus {
		lines = append(lines, "- "+summarizeBandwidth(bandwidthVals, lang))
		lines = append(lines, "- README_NEW_USER note: offshore servers commonly start around 100Mbps; evaluate against your actual workload needs.")
	}
	if config.PingTestStatus || config.TgdcTestStatus || config.WebTestStatus || config.BacktraceStatus || config.Nt3Status {
		lines = append(lines, "- "+summarizeLatency(latencyVals, lang))
		lines = append(lines, "- README_NEW_USER note: latency >= 9999ms should be treated as unavailable target.")
	}
	lines = append(lines, "- Suggestion: prioritize the metrics that match your workload (compute, storage, or cross-region networking).")
	return strings.Join(lines, "\n")
}
