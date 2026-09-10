//go:build !ecs_public

package tests

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/oneclickvirt/privatespeedtest/pst"
	"github.com/oneclickvirt/speedtest/model"
	"github.com/oneclickvirt/speedtest/sp"
)

func ShowHead(language string) {
	ShowHeadTo(os.Stdout, language)
}

// ShowHeadTo writes the table header to an isolated writer. This is used by
// concurrent workflow chapters so no process-wide stdout capture is needed.
func ShowHeadTo(writer io.Writer, language string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(writerOrDiscard(writer), "[WARN] speedtest header unavailable")
		}
	}()
	sp.ShowHeadTo(writerOrDiscard(writer), language)
}

func NearbySP() {
	NearbySPWithNetwork("")
}

// NearbySPWithNetwork keeps a nearby Ookla measurement on the requested
// address family. An empty network retains the historical automatic behavior.
func NearbySPWithNetwork(network string) {
	NearbySPWithNetworkContextTo(context.Background(), os.Stdout, network)
}

// NearbySPWithNetworkTo runs the nearby measurement and sends all rendered
// rows to writer.
func NearbySPWithNetworkTo(writer io.Writer, network string) {
	NearbySPWithNetworkContextTo(context.Background(), writer, network)
}

func NearbySPWithNetworkContextTo(ctx context.Context, writer io.Writer, network string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(writerOrDiscard(writer), "[WARN] nearby speedtest unavailable")
		}
	}()
	network = normalizeSpeedNetwork(network)
	if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
		renderNearbySpeedtest(ctx, writer, network, sp.NearbySpeedTestWithNetworkContextTo)
	} else {
		renderNearbySpeedtest(ctx, writer, network, sp.OfficialNearbySpeedTestWithNetworkContextTo)
	}
}

func writerOrDiscard(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}

// formatString 格式化字符串到指定宽度
func formatString(s string, width int) string {
	return fmt.Sprintf("%-*s", width, s)
}

// printTableRow 打印表格行
func printTableRow(result pst.SpeedTestResult) {
	printTableRowTo(os.Stdout, result)
}

func printTableRowTo(writer io.Writer, result pst.SpeedTestResult) {
	writer = writerOrDiscard(writer)
	location := result.City
	if result.CarrierType != "" {
		carrier := result.CarrierType
		switch carrier {
		case "Telecom":
			carrier = "电信"
		case "Unicom":
			carrier = "联通"
		case "Mobile":
			carrier = "移动"
		case "Other":
			carrier = "其他"
		}
		location = fmt.Sprintf("%s%s", carrier, result.City)
	}
	if len(location) > 15 {
		location = location[:15]
	}
	upload := "N/A"
	if result.UploadMbps > 0 {
		upload = fmt.Sprintf("%.2f Mbps", result.UploadMbps)
	}
	download := "N/A"
	if result.DownloadMbps > 0 {
		download = fmt.Sprintf("%.2f Mbps", result.DownloadMbps)
	}
	latency := fmt.Sprintf("%.2f ms", result.PingLatency.Seconds()*1000)
	packetLoss := "N/A"
	fmt.Fprint(writer, formatString(location, 15))
	fmt.Fprint(writer, formatString(upload, 16))
	fmt.Fprint(writer, formatString(download, 16))
	fmt.Fprint(writer, formatString(latency, 16))
	fmt.Fprint(writer, formatString(packetLoss, 16))
	fmt.Fprintln(writer)
}

type PrivateSpeedPreloads struct {
	network    pst.Network
	preloads   map[string]*pst.ServerPreload
	candidates map[string][]pst.ServerConfig
	preloadErr map[string]error
	done       chan struct{}
}

// GlobalSpeedPreload caches the legacy international candidate phase. It is
// intentionally separate from PrivateSpeedPreloads so English mode never
// loads or exposes the managed private registry.
type GlobalSpeedPreload struct {
	preload *sp.CustomSpeedTestPreload
}

func StartGlobalSpeedPreload(ctx context.Context, network string) *GlobalSpeedPreload {
	return &GlobalSpeedPreload{preload: sp.StartCustomSpeedTestPreload(ctx, model.NetGlobal, "id", normalizeSpeedNetwork(network))}
}

func (p *GlobalSpeedPreload) Wait(ctx context.Context) error {
	if p == nil || p.preload == nil {
		return fmt.Errorf("国际测速候选预加载不可用")
	}
	return p.preload.Wait(ctx)
}

func RunGlobalSpeedTestWithPreloadTo(ctx context.Context, writer io.Writer, num int, language, network string, preload *GlobalSpeedPreload) error {
	if preload == nil || preload.preload == nil {
		return fmt.Errorf("国际测速候选预加载不可用")
	}
	if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
		return renderFilteredSpeedtestWithError(writer, func(output io.Writer) error {
			return preload.preload.RunCustomSpeedTestContextTo(ctx, output, num, language)
		})
	}
	return renderFilteredSpeedtestWithError(writer, func(output io.Writer) error {
		return preload.preload.RunOfficialCustomSpeedTestContextTo(ctx, output, num, language)
	})
}

const privateSpeedPreloadDeadline = 20 * time.Second

func privateSpeedNetwork(network string) pst.Network {
	switch normalizeSpeedNetwork(network) {
	case "tcp4":
		return pst.NetworkIPv4
	case "tcp6":
		return pst.NetworkIPv6
	default:
		return pst.NetworkAuto
	}
}

func privateSpeedCarrier(operator string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(operator)) {
	case "cmcc":
		return "Mobile", true
	case "cu":
		return "Unicom", true
	case "ct":
		return "Telecom", true
	case "other":
		return "Other", true
	default:
		return "", false
	}
}

// StartPrivateSpeedPreloads starts silent candidate probes for every requested
// carrier concurrently. Call WaitAll before the first throughput measurement
// so all probe traffic has ended before bandwidth sampling begins.
func StartPrivateSpeedPreloads(ctx context.Context, operators []string, network string) *PrivateSpeedPreloads {
	if ctx == nil {
		ctx = context.Background()
	}
	preloads := &PrivateSpeedPreloads{
		network:    privateSpeedNetwork(network),
		preloads:   make(map[string]*pst.ServerPreload, len(operators)),
		candidates: make(map[string][]pst.ServerConfig, len(operators)),
		preloadErr: make(map[string]error, len(operators)),
		done:       make(chan struct{}),
	}
	// Candidate selection is an optimization, not a prerequisite for the
	// overall benchmark. Give it its own bounded lifetime so a large or
	// unreachable registry can never consume the caller's complete test
	// deadline. The parent context still cancels it immediately on shutdown.
	preloadCtx, cancel := context.WithTimeout(ctx, privateSpeedPreloadDeadline)
	go func() {
		defer cancel()
		preloads.load(preloadCtx, operators)
	}()
	return preloads
}

func (p *PrivateSpeedPreloads) load(ctx context.Context, operators []string) {
	defer close(p.done)
	serverList, err := privateSpeedServerListWithNetwork(p.network)
	if err != nil {
		for _, operator := range operators {
			operator = strings.ToLower(strings.TrimSpace(operator))
			if operator != "" {
				p.preloadErr[operator] = fmt.Errorf("加载自定义服务器列表失败")
			}
		}
		return
	}

	seen := make(map[string]struct{}, len(operators))
	unique := make([]string, 0, len(operators))
	for _, operator := range operators {
		operator = strings.ToLower(strings.TrimSpace(operator))
		if operator == "" {
			continue
		}
		if _, exists := seen[operator]; exists {
			continue
		}
		seen[operator] = struct{}{}
		unique = append(unique, operator)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, operator := range unique {
		operator := operator
		wg.Add(1)
		go func() {
			defer wg.Done()
			carrier, ok := privateSpeedCarrier(operator)
			if !ok {
				mu.Lock()
				p.preloadErr[operator] = fmt.Errorf("不支持的运营商类型: %s", operator)
				mu.Unlock()
				return
			}
			servers := pst.FilterServersByISP(serverList.Servers, carrier)
			if len(servers) == 0 {
				mu.Lock()
				p.preloadErr[operator] = fmt.Errorf("没有可用的%s测速服务器", carrier)
				mu.Unlock()
				return
			}
			mu.Lock()
			p.candidates[operator] = append([]pst.ServerConfig(nil), servers...)
			mu.Unlock()
			preload, preloadErr := pst.PreloadBestServersWithNetwork(ctx, servers, len(servers), 5*time.Second, false, true, p.network)
			mu.Lock()
			defer mu.Unlock()
			if preloadErr != nil {
				p.preloadErr[operator] = preloadErr
				return
			}
			p.preloads[operator] = preload
		}()
	}
	wg.Wait()
	// ServerPreload launches its probe batch asynchronously. Join each batch
	// here so p.done means all candidate traffic has stopped; this is what keeps
	// the first throughput sample isolated from preload traffic.
	for operator, preload := range p.preloads {
		if _, preloadErr := preload.Wait(ctx); preloadErr != nil {
			mu.Lock()
			if p.preloadErr[operator] == nil {
				p.preloadErr[operator] = preloadErr
			}
			mu.Unlock()
		}
	}
}

// WaitAll joins the preload batch before a throughput test begins. It is safe
// to call more than once and leaves individual carrier errors for Wait.
func (p *PrivateSpeedPreloads) WaitAll(ctx context.Context) error {
	if p == nil || p.done == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-p.done:
	case <-ctx.Done():
		return ctx.Err()
	}
	// PreloadBestServersWithNetwork itself starts a background probe. All of
	// those probes have been launched by load above; joining them here keeps
	// the launch concurrent while guaranteeing no probe traffic remains when
	// the first download/upload sample begins.
	for _, preload := range p.preloads {
		if _, err := preload.Wait(ctx); err != nil && ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return nil
}

// Wait returns preloaded candidates for one carrier. It intentionally blocks
// before the caller starts any download or upload measurement.
func (p *PrivateSpeedPreloads) Wait(ctx context.Context, operator string) ([]pst.ServerWithLatencyInfo, error) {
	if p == nil {
		return nil, fmt.Errorf("测速候选预加载不可用")
	}
	if err := p.WaitAll(ctx); err != nil {
		return nil, err
	}
	operator = strings.ToLower(strings.TrimSpace(operator))
	if err := p.preloadErr[operator]; err != nil {
		return nil, err
	}
	preload := p.preloads[operator]
	if preload == nil {
		return nil, fmt.Errorf("测速候选预加载不可用")
	}
	ranked, err := preload.Wait(ctx)
	if err != nil {
		return nil, err
	}
	// The preload deliberately probes only a bounded front-loaded sample. Keep
	// that ranking, then append every other syntactically eligible carrier node
	// so the real transfer remains the final availability decision.
	return mergePrivateSpeedCandidates(ranked, p.candidates[operator], p.network), nil
}

func mergePrivateSpeedCandidates(ranked []pst.ServerWithLatencyInfo, candidates []pst.ServerConfig, network pst.Network) []pst.ServerWithLatencyInfo {
	merged := append([]pst.ServerWithLatencyInfo(nil), ranked...)
	seen := make(map[string]struct{}, len(ranked))
	for _, candidate := range ranked {
		seen[privateSpeedCandidateKey(candidate)] = struct{}{}
	}
	for _, server := range candidates {
		if !privateSpeedServerSupportsNetwork(server, network) {
			continue
		}
		candidate := pst.ServerWithLatencyInfo{Server: server, Availability: pst.ServerCandidate}
		key := privateSpeedCandidateKey(candidate)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, candidate)
	}
	return merged
}

func privateSpeedServerSupportsNetwork(server pst.ServerConfig, network pst.Network) bool {
	host := strings.Trim(strings.TrimSpace(server.Host), "[]")
	ip := net.ParseIP(host)
	if ip == nil {
		return true
	}
	switch network {
	case pst.NetworkIPv4:
		return ip.To4() != nil
	case pst.NetworkIPv6:
		return ip.To4() == nil && ip.To16() != nil
	default:
		return true
	}
}

func privateSpeedTest(num int, operator string) (testedCount int, err error) {
	return privateSpeedTestWithNetworkTo(context.Background(), num, operator, "", nil, os.Stdout)
}

func privateSpeedTestWithNetwork(ctx context.Context, num int, operator, network string, preloads *PrivateSpeedPreloads) (testedCount int, err error) {
	return privateSpeedTestWithNetworkTo(ctx, num, operator, network, preloads, os.Stdout)
}

func privateSpeedTestWithNetworkTo(ctx context.Context, num int, operator, network string, preloads *PrivateSpeedPreloads, writer io.Writer) (testedCount int, err error) {
	writer = writerOrDiscard(writer)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(writer, "[WARN] speedtest registry unavailable")
			testedCount = 0
			err = fmt.Errorf("私有测速运行失败")
		}
	}()
	if ctx == nil {
		ctx = context.Background()
	}
	serversPerISP := num
	if serversPerISP <= 0 || serversPerISP > 5 {
		serversPerISP = 2
	}
	operator = strings.ToLower(strings.TrimSpace(operator))
	carrierType, ok := privateSpeedCarrier(operator)
	if !ok {
		return 0, fmt.Errorf("不支持的运营商类型: %s", operator)
	}
	var candidateServers []pst.ServerWithLatencyInfo
	if preloads != nil {
		candidateServers, err = preloads.Wait(ctx, operator)
	}
	if preloads == nil || err != nil || len(candidateServers) == 0 {
		// Preloading is only a sorting optimization. If its short deadline or
		// reachability checks fail, rebuild the ordered candidate list here and
		// let actual download/upload attempts make the final decision.
		candidateServers, err = findPrivateSpeedCandidates(ctx, carrierType, network)
	}
	if err != nil {
		return 0, fmt.Errorf("分组查找失败")
	}
	bestServers := selectPrivateSpeedCandidates(candidateServers, serversPerISP)
	if len(bestServers) == 0 {
		return 0, fmt.Errorf("去重后没有可用的服务器")
	}
	for i, serverInfo := range bestServers {
		if testedCount >= serversPerISP {
			break
		}
		result := pst.RunSpeedTestContextWithNetwork(
			ctx,
			serverInfo.Server,
			false,          // 不禁用下载测试
			false,          // 不禁用上传测试
			6,              // 并发线程数
			12*time.Second, // 超时时间
			&serverInfo,
			false, // 不显示进度条
			privateSpeedNetwork(network),
		)
		if result.UploadMbps > 0 || result.DownloadMbps > 0 {
			printTableRowTo(writer, result)
			testedCount++
		}
		if testedCount < serversPerISP && i < len(bestServers)-1 {
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return testedCount, ctx.Err()
			}
		}
	}
	// 返回实际成功输出的节点数量
	return testedCount, nil
}

func findPrivateSpeedCandidates(ctx context.Context, carrierType, network string) ([]pst.ServerWithLatencyInfo, error) {
	serverList, err := privateSpeedServerListWithNetwork(privateSpeedNetwork(network))
	if err != nil {
		return nil, fmt.Errorf("加载自定义服务器列表失败")
	}
	filteredServers := pst.FilterServersByISP(serverList.Servers, carrierType)
	return pst.FindBestServersContextWithNetwork(
		ctx,
		filteredServers,
		len(filteredServers),
		5*time.Second,
		false,
		true,
		privateSpeedNetwork(network),
	)
}

func selectPrivateSpeedCandidates(candidates []pst.ServerWithLatencyInfo, serversPerISP int) []pst.ServerWithLatencyInfo {
	if serversPerISP <= 0 {
		return nil
	}
	attemptLimit := serversPerISP * 2
	if attemptLimit > len(candidates) {
		attemptLimit = len(candidates)
	}
	// City diversity remains a ranking preference only. Precheck failures stay
	// eligible inside the bounded 2*limit real-transfer window.
	selected := pst.SelectDistinctCityServers(candidates, attemptLimit)
	if len(selected) >= attemptLimit {
		return selected
	}
	used := make(map[string]struct{}, len(selected))
	for _, candidate := range selected {
		used[privateSpeedCandidateKey(candidate)] = struct{}{}
	}
	// Prefer city diversity, then retain same-city endpoints as real-transfer
	// fallbacks instead of discarding them after a failed reachability check.
	for _, candidate := range candidates {
		key := privateSpeedCandidateKey(candidate)
		if _, exists := used[key]; exists {
			continue
		}
		used[key] = struct{}{}
		selected = append(selected, candidate)
		if len(selected) >= attemptLimit {
			break
		}
	}
	return selected
}

func privateSpeedCandidateKey(candidate pst.ServerWithLatencyInfo) string {
	server := candidate.Server
	return server.ID + "\x00" + server.URL + "\x00" + server.Host + "\x00" + fmt.Sprint(server.Port)
}

func privateSpeedTestWithFallback(num int, operator, language string) {
	privateSpeedTestWithFallbackWithNetwork(num, operator, language, "")
}

func privateSpeedTestWithFallbackWithNetwork(num int, operator, language, network string) {
	privateSpeedTestWithFallbackWithNetworkContextTo(context.Background(), os.Stdout, num, operator, language, network)
}

func privateSpeedTestWithFallbackWithNetworkTo(writer io.Writer, num int, operator, language, network string) {
	privateSpeedTestWithFallbackWithNetworkContextTo(context.Background(), writer, num, operator, language, network)
}

func privateSpeedTestWithFallbackWithNetworkContextTo(ctx context.Context, writer io.Writer, num int, operator, language, network string) {
	writer = writerOrDiscard(writer)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(writer, "[WARN] preferred speedtest unavailable; using fallback")
		}
	}()
	testedCount, err := privateSpeedTestWithNetworkTo(ctx, num, operator, network, nil, writer)
	if err != nil || testedCount == 0 {
		var url, parseType string
		url = model.NetGlobal
		parseType = "id"
		if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
			renderFilteredSpeedtest(writer, func(output io.Writer) {
				sp.CustomSpeedTestWithNetworkContextTo(ctx, output, url, parseType, num, language, normalizeSpeedNetwork(network))
			})
		} else {
			renderFilteredSpeedtest(writer, func(output io.Writer) {
				sp.OfficialCustomSpeedTestWithNetworkContextTo(ctx, output, url, parseType, num, language, normalizeSpeedNetwork(network))
			})
		}
	}
}

func CustomSP(platform, operator string, num int, language string) {
	CustomSPWithNetwork(platform, operator, num, language, "")
}

// CustomSPWithNetwork keeps public and managed carrier tests on one explicit
// address family when requested.
func CustomSPWithNetwork(platform, operator string, num int, language, network string) {
	CustomSPWithNetworkContextTo(context.Background(), os.Stdout, platform, operator, num, language, network)
}

// CustomSPWithNetworkTo is the writer-aware public/custom speed entry point.
func CustomSPWithNetworkTo(writer io.Writer, platform, operator string, num int, language, network string) {
	CustomSPWithNetworkContextTo(context.Background(), writer, platform, operator, num, language, network)
}

func CustomSPWithNetworkContextTo(ctx context.Context, writer io.Writer, platform, operator string, num int, language, network string) {
	customSPWithNetworkTo(ctx, writer, platform, operator, num, language, network, nil)
}

// CustomSPWithNetworkAndPreloads consumes a batch started earlier by
// StartPrivateSpeedPreloads. It waits for the relevant carrier before running
// transfer traffic.
func CustomSPWithNetworkAndPreloads(ctx context.Context, platform, operator string, num int, language, network string, preloads *PrivateSpeedPreloads) {
	CustomSPWithNetworkAndPreloadsTo(os.Stdout, ctx, platform, operator, num, language, network, preloads)
}

// CustomSPWithNetworkAndPreloadsTo consumes an earlier preload batch and
// writes the complete section to writer.
func CustomSPWithNetworkAndPreloadsTo(writer io.Writer, ctx context.Context, platform, operator string, num int, language, network string, preloads *PrivateSpeedPreloads) {
	customSPWithNetworkTo(ctx, writer, platform, operator, num, language, network, preloads)
}

func customSPWithNetwork(ctx context.Context, platform, operator string, num int, language, network string, preloads *PrivateSpeedPreloads) {
	customSPWithNetworkTo(ctx, os.Stdout, platform, operator, num, language, network, preloads)
}

func customSPWithNetworkTo(ctx context.Context, writer io.Writer, platform, operator string, num int, language, network string, preloads *PrivateSpeedPreloads) {
	writer = writerOrDiscard(writer)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(writer, "[WARN] custom speedtest unavailable")
		}
	}()
	network = normalizeSpeedNetwork(network)
	opLower := strings.ToLower(operator)
	if opLower == "cmcc" || opLower == "cu" || opLower == "ct" || opLower == "other" {
		testedCount, err := privateSpeedTestWithNetworkTo(ctx, num, opLower, network, preloads, writer)
		if err != nil {
			fmt.Fprintln(writer, "[WARN] preferred speedtest unavailable; using fallback")
		} else if testedCount >= num {
			return
		} else if testedCount > 0 {
			fmt.Fprintf(writer, "[INFO] 私有节点仅测试了 %d 个，补充 %d 个公共节点\n", testedCount, num-testedCount)
			num = num - testedCount
		} else {
			// testedCount == 0，继续使用公共节点
		}
	}

	var url, parseType string
	if strings.ToLower(platform) == "cn" {
		if strings.ToLower(operator) == "cmcc" {
			url = model.CnCMCC
		} else if strings.ToLower(operator) == "cu" {
			url = model.CnCU
		} else if strings.ToLower(operator) == "ct" {
			url = model.CnCT
		} else if strings.ToLower(operator) == "hk" {
			url = model.CnHK
		} else if strings.ToLower(operator) == "tw" {
			url = model.CnTW
		} else if strings.ToLower(operator) == "jp" {
			url = model.CnJP
		} else if strings.ToLower(operator) == "sg" {
			url = model.CnSG
		}
		parseType = "url"
	} else if strings.ToLower(platform) == "net" {
		if strings.ToLower(operator) == "cmcc" {
			url = model.NetCMCC
		} else if strings.ToLower(operator) == "cu" {
			url = model.NetCU
		} else if strings.ToLower(operator) == "ct" {
			url = model.NetCT
		} else if strings.ToLower(operator) == "hk" {
			url = model.NetHK
		} else if strings.ToLower(operator) == "tw" {
			url = model.NetTW
		} else if strings.ToLower(operator) == "jp" {
			url = model.NetJP
		} else if strings.ToLower(operator) == "sg" {
			url = model.NetSG
		} else if strings.ToLower(operator) == "global" || strings.ToLower(operator) == "other" {
			url = model.NetGlobal
		}
		parseType = "id"
	}
	if runtime.GOOS == "windows" || sp.OfficialAvailableTest() != nil {
		renderFilteredSpeedtest(writer, func(output io.Writer) {
			sp.CustomSpeedTestWithNetworkContextTo(ctx, output, url, parseType, num, language, network)
		})
	} else {
		renderFilteredSpeedtest(writer, func(output io.Writer) {
			sp.OfficialCustomSpeedTestWithNetworkContextTo(ctx, output, url, parseType, num, language, network)
		})
	}
}
