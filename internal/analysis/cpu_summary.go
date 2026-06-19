package analysis

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

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
