package datasync

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	// schemaVersion identifies snapshots owned by the goecs repository. The
	// legacy ecs-data/v1 value remains accepted by the runtime loader so older
	// CDN manifests can be consumed during a rolling update.
	schemaVersion       = "goecs-data/v1"
	legacySchemaVersion = "ecs-data/v1"
	syncVersion         = "v0.0.2"
)

var (
	client = &http.Client{Timeout: 20 * time.Second}
)

type sourceSpec struct {
	name           string
	file           string
	url            string
	minimum        int
	transform      func([]byte) (any, int, error)
	health         func(context.Context, any) (any, error)
	validateOutput func([]byte) error
}

type stagedFile struct {
	data   []byte
	count  int
	source string
}

type fileManifest struct {
	SHA256 string `json:"sha256"`
	Count  int    `json:"count"`
	Source string `json:"source"`
}

type manifest struct {
	Schema      string                  `json:"schema"`
	GeneratedAt time.Time               `json:"generated_at"`
	Files       map[string]fileManifest `json:"files"`
}

type tcpTarget struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Category string `json:"category"`
}

type dnsblZone struct {
	Zone string `json:"zone"`
	IPv4 bool   `json:"ipv4"`
	IPv6 bool   `json:"ipv6"`
}

type asnName struct {
	ASN  uint32 `json:"asn"`
	Name string `json:"name"`
}

type transferTarget struct {
	ID       string `json:"id"`
	Host     string `json:"host"`
	PortFrom int    `json:"port_from"`
	PortTo   int    `json:"port_to"`
	Provider string `json:"provider"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Status   string `json:"status"`
}

type providerName struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Groups       []string `json:"groups"`
	SupportsIPv6 bool     `json:"supports_ipv6"`
}

type provinceRoute struct {
	Code     string                  `json:"code"`
	Name     string                  `json:"name"`
	Province int                     `json:"province"`
	Short    string                  `json:"short"`
	Targets  []provinceCarrierTarget `json:"targets"`
}

type provinceCarrierTarget struct {
	Carrier string `json:"carrier"`
	IPv4    string `json:"ipv4"`
	IPv6    string `json:"ipv6"`
}

// Version returns the data synchronizer version used by the standalone
// command. Keeping it in the package avoids duplicating release metadata.
func Version() string { return syncVersion }

// Sync fetches, validates and atomically updates the goecs data snapshot.
// Callers should provide a context with an appropriate overall deadline.
func Sync(ctx context.Context, outputDir string) (bool, error) {
	return synchronize(ctx, outputDir, defaultSourceSpecs())
}

// DefaultOutputDir is the checked-in snapshot location used by the command
// and the scheduled workflow.
const DefaultOutputDir = "internal/data/snapshot"

func defaultSourceSpecs() []sourceSpec {
	return []sourceSpec{
		{name: "tcpbench", file: "tcp-targets.json", url: "https://raw.githubusercontent.com/se-tang/TCPbench/main/backend/scripts/run.sh", minimum: 50, transform: parseTCPTargets, validateOutput: validateTCPTargetSchema},
		{name: "provinces", file: "province-routes.json", url: "https://raw.githubusercontent.com/xykt/NetQuality/main/ref/province.json", minimum: 31, transform: parseProvinceRoutes, validateOutput: validateProvinceRouteSchema},
		{name: "speedtest", file: "speedtest-servers.json", url: "https://raw.githubusercontent.com/xykt/NetQuality/main/ref/speedtest_cn.json", minimum: 10, transform: parseSpeedtestServers, health: probeSpeedtestServers, validateOutput: validateSpeedtestServerSchema},
		{name: "transfer", file: "openspeedtest-servers.json", url: "https://raw.githubusercontent.com/xykt/NetQuality/main/ref/iperf.json", minimum: 5, transform: parseTransferTargets, health: probeTransferTargets, validateOutput: validateTransferTargetSchema},
		{name: "dnsbl", file: "dnsbl-zones.json", url: "https://raw.githubusercontent.com/xykt/IPQuality/main/ref/dnsbl.list", minimum: 100, transform: parseDNSBL, health: probeDNSBLZones, validateOutput: validateDNSBLZoneSchema},
		{name: "asn", file: "bgp-asn-map.json", url: "https://raw.githubusercontent.com/xykt/NetQuality/main/ref/AS_Mapping.txt", minimum: 50, transform: parseASNMap, validateOutput: validateASNMapSchema},
		{name: "media", file: "media-providers.json", url: "https://raw.githubusercontent.com/HsukqiLee/MediaUnlockTest/main/pkg/providers/lists.go", minimum: 100, transform: parseProviderNames, validateOutput: validateMediaProviderSchema},
	}
}

func synchronize(ctx context.Context, outputDir string, specs []sourceSpec) (bool, error) {
	if len(specs) == 0 {
		return false, errors.New("no data sources configured")
	}
	staged := make(map[string]stagedFile, len(specs))
	changedAt := time.Now().UTC()
	for _, spec := range specs {
		if spec.name == "" || spec.url == "" || spec.transform == nil {
			return false, errors.New("data source is missing name, URL, or transform")
		}
		if spec.file == "" || filepath.Base(spec.file) != spec.file || filepath.Ext(spec.file) != ".json" || spec.file == "manifest.json" {
			return false, fmt.Errorf("source %s has invalid output file %q", spec.name, spec.file)
		}
		if _, exists := staged[spec.file]; exists {
			return false, fmt.Errorf("duplicate output file %q", spec.file)
		}
		file, err := stageRemoteSource(ctx, spec)
		if err != nil {
			current, currentErr := loadCurrentSource(outputDir, spec)
			if currentErr != nil {
				return false, fmt.Errorf("update %s failed: %v; current snapshot unavailable: %w", spec.name, err, currentErr)
			}
			file = current
		}
		staged[spec.file] = file
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return false, fmt.Errorf("create output directory: %w", err)
	}
	if err := validateCountDrops(outputDir, staged, 0.35); err != nil {
		return false, fmt.Errorf("quantity guard: %w", err)
	}
	if err := validateAvailabilityDrops(outputDir, staged, 0.35); err != nil {
		return false, fmt.Errorf("availability guard: %w", err)
	}
	if semanticDataEqual(outputDir, staged) {
		return false, nil
	}
	m := manifest{Schema: schemaVersion, GeneratedAt: changedAt, Files: make(map[string]fileManifest, len(staged))}
	for name, file := range staged {
		hash := sha256.Sum256(file.data)
		m.Files[name] = fileManifest{SHA256: hex.EncodeToString(hash[:]), Count: file.count, Source: file.source}
	}
	manifestData, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return false, fmt.Errorf("encode manifest: %w", err)
	}
	manifestData = append(manifestData, '\n')
	if err := validateManifest(manifestData, staged); err != nil {
		return false, fmt.Errorf("validate manifest: %w", err)
	}

	if err := commitSnapshot(outputDir, staged, manifestData, nil); err != nil {
		return false, err
	}
	return true, nil
}

func stageRemoteSource(ctx context.Context, spec sourceSpec) (stagedFile, error) {
	raw, err := fetch(ctx, spec.url)
	if err != nil {
		return stagedFile{}, fmt.Errorf("fetch: %w", err)
	}
	value, count, err := spec.transform(raw)
	if err != nil {
		return stagedFile{}, fmt.Errorf("parse: %w", err)
	}
	if count < spec.minimum {
		return stagedFile{}, fmt.Errorf("got %d records, require at least %d", count, spec.minimum)
	}
	if spec.health != nil {
		value, err = spec.health(ctx, value)
		if err != nil {
			return stagedFile{}, fmt.Errorf("health check: %w", err)
		}
		count, err = countJSONArrayValue(value)
		if err != nil {
			return stagedFile{}, fmt.Errorf("count health-checked records: %w", err)
		}
		if count < spec.minimum {
			return stagedFile{}, fmt.Errorf("health check retained %d records, require at least %d", count, spec.minimum)
		}
	}
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return stagedFile{}, fmt.Errorf("encode: %w", err)
	}
	encoded = append(encoded, '\n')
	if err := validateStagedSource(spec, encoded, count); err != nil {
		return stagedFile{}, err
	}
	return stagedFile{data: encoded, count: count, source: spec.url}, nil
}

func loadCurrentSource(outputDir string, spec sourceSpec) (stagedFile, error) {
	manifestData, err := os.ReadFile(filepath.Join(outputDir, "manifest.json"))
	if err != nil {
		return stagedFile{}, err
	}
	var current manifest
	if err := json.Unmarshal(manifestData, &current); err != nil {
		return stagedFile{}, err
	}
	if !supportedCurrentSchema(current.Schema) || current.Files == nil {
		return stagedFile{}, errors.New("current manifest schema is invalid")
	}
	meta, ok := current.Files[spec.file]
	if !ok || meta.Count < spec.minimum || strings.TrimSpace(meta.Source) == "" {
		return stagedFile{}, errors.New("current manifest entry is invalid")
	}
	data, err := os.ReadFile(filepath.Join(outputDir, spec.file))
	if err != nil {
		return stagedFile{}, err
	}
	hash := sha256.Sum256(data)
	if !strings.EqualFold(meta.SHA256, hex.EncodeToString(hash[:])) {
		return stagedFile{}, errors.New("current snapshot SHA-256 mismatch")
	}
	if err := validateStagedSource(spec, data, meta.Count); err != nil {
		return stagedFile{}, fmt.Errorf("current snapshot validation: %w", err)
	}
	return stagedFile{data: data, count: meta.Count, source: meta.Source}, nil
}

func validateStagedSource(spec sourceSpec, data []byte, count int) error {
	if err := validateJSONArray(data, count); err != nil {
		return fmt.Errorf("validate encoded records: %w", err)
	}
	if spec.validateOutput != nil {
		if err := spec.validateOutput(data); err != nil {
			return fmt.Errorf("validate output schema: %w", err)
		}
	}
	return nil
}

func countJSONArrayValue(value any) (int, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return 0, err
	}
	var records []json.RawMessage
	if err := json.Unmarshal(encoded, &records); err != nil {
		return 0, err
	}
	return len(records), nil
}

func validateJSONArray(data []byte, expected int) error {
	if expected < 0 {
		return errors.New("negative record count")
	}
	var records []json.RawMessage
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("decode JSON array: %w", err)
	}
	if len(records) != expected {
		return fmt.Errorf("record count mismatch: got %d want %d", len(records), expected)
	}
	for index, record := range records {
		if len(bytes.TrimSpace(record)) == 0 || bytes.Equal(bytes.TrimSpace(record), []byte("null")) {
			return fmt.Errorf("record %d is empty or null", index)
		}
	}
	return nil
}

func validateManifest(data []byte, staged map[string]stagedFile) error {
	var candidate manifest
	if err := json.Unmarshal(data, &candidate); err != nil {
		return err
	}
	if candidate.Schema != schemaVersion || candidate.GeneratedAt.IsZero() || len(candidate.Files) != len(staged) {
		return errors.New("manifest schema, timestamp, or file set is invalid")
	}
	for name, file := range staged {
		meta, ok := candidate.Files[name]
		if !ok || meta.Count != file.count || meta.Source != file.source {
			return fmt.Errorf("manifest metadata mismatch for %s", name)
		}
		hash := sha256.Sum256(file.data)
		if !strings.EqualFold(meta.SHA256, hex.EncodeToString(hash[:])) {
			return fmt.Errorf("manifest SHA-256 mismatch for %s", name)
		}
	}
	return nil
}

func semanticDataEqual(dir string, staged map[string]stagedFile) bool {
	manifestData, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return false
	}
	var current manifest
	if json.Unmarshal(manifestData, &current) != nil || current.Schema != schemaVersion || len(current.Files) != len(staged) {
		return false
	}
	for name, file := range staged {
		meta, ok := current.Files[name]
		if !ok || meta.Count != file.count || meta.Source != file.source {
			return false
		}
		hash := sha256.Sum256(file.data)
		if !strings.EqualFold(meta.SHA256, hex.EncodeToString(hash[:])) {
			return false
		}
		currentData, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || !bytes.Equal(currentData, file.data) {
			return false
		}
	}
	return true
}

func validateCountDrops(dir string, staged map[string]stagedFile, maximumDrop float64) error {
	if maximumDrop < 0 || maximumDrop >= 1 {
		return errors.New("maximum drop must be between zero and one")
	}
	manifestData, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var current manifest
	if err := json.Unmarshal(manifestData, &current); err != nil {
		return fmt.Errorf("decode current manifest: %w", err)
	}
	if !supportedCurrentSchema(current.Schema) || current.Files == nil {
		return errors.New("current manifest has an invalid schema or file set")
	}
	for name, file := range staged {
		previous, ok := current.Files[name]
		if !ok || previous.Count <= 0 || file.count >= previous.Count {
			continue
		}
		minimumAllowed := int(math.Ceil(float64(previous.Count) * (1 - maximumDrop)))
		if file.count < minimumAllowed {
			return fmt.Errorf("%s dropped from %d to %d records", name, previous.Count, file.count)
		}
	}
	return nil
}

func supportedCurrentSchema(schema string) bool {
	return schema == schemaVersion || schema == legacySchemaVersion
}

// validateAvailabilityDrops protects a known-good registry when a health
// probe suddenly loses most of its reachable nodes. Absolute minimum counts
// are not sufficient because a large registry could otherwise collapse to a
// handful of endpoints and still pass.
func validateAvailabilityDrops(dir string, staged map[string]stagedFile, maximumDrop float64) error {
	if maximumDrop < 0 || maximumDrop >= 1 {
		return errors.New("maximum availability drop must be between zero and one")
	}
	for _, name := range []string{"speedtest-servers.json", "openspeedtest-servers.json"} {
		currentData, err := os.ReadFile(filepath.Join(dir, name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		previousAvailable, previousHasStatus, err := countAvailableRecords(currentData)
		if err != nil {
			return fmt.Errorf("decode current %s: %w", name, err)
		}
		if !previousHasStatus || previousAvailable == 0 {
			continue
		}
		candidate, ok := staged[name]
		if !ok {
			return fmt.Errorf("staged snapshot is missing %s", name)
		}
		currentAvailable, currentHasStatus, err := countAvailableRecords(candidate.data)
		if err != nil {
			return fmt.Errorf("decode staged %s: %w", name, err)
		}
		if !currentHasStatus {
			return fmt.Errorf("staged %s has no availability status", name)
		}
		minimumAllowed := int(math.Ceil(float64(previousAvailable) * (1 - maximumDrop)))
		if currentAvailable < minimumAllowed {
			return fmt.Errorf("%s available nodes dropped from %d to %d", name, previousAvailable, currentAvailable)
		}
	}
	return nil
}

func countAvailableRecords(data []byte) (int, bool, error) {
	var records []struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(data, &records); err != nil {
		return 0, false, err
	}
	hasStatus := false
	available := 0
	for _, record := range records {
		if strings.TrimSpace(record.Status) != "" {
			hasStatus = true
		}
		if strings.EqualFold(strings.TrimSpace(record.Status), "available") {
			available++
		}
	}
	return available, hasStatus, nil
}

func fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "oneclickvirt-goecs-data-sync/1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, errors.New("empty response")
	}
	return data, nil
}

func passJSONArray(data []byte) (any, int, error) {
	var value []map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, 0, err
	}
	return value, len(value), nil
}

func parseProvinceRoutes(data []byte) (any, int, error) {
	var input []struct {
		Code     string `json:"code"`
		Name     string `json:"name"`
		Province int    `json:"province"`
		Short    string `json:"short"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, 0, err
	}
	result := make([]provinceRoute, 0, len(input))
	seenCodes := make(map[string]struct{}, len(input))
	seenProvinces := make(map[int]struct{}, len(input))
	for _, item := range input {
		code := strings.ToLower(strings.TrimSpace(item.Code))
		name := strings.TrimSpace(item.Name)
		if item.Province <= 0 || item.Province >= 70 || code == "" || name == "" {
			continue
		}
		if _, exists := seenCodes[code]; exists {
			continue
		}
		if _, exists := seenProvinces[item.Province]; exists {
			continue
		}
		seenCodes[code] = struct{}{}
		seenProvinces[item.Province] = struct{}{}
		targets := make([]provinceCarrierTarget, 0, 3)
		for _, carrier := range []string{"ct", "cu", "cm"} {
			targets = append(targets, provinceCarrierTarget{
				Carrier: carrier,
				IPv4:    fmt.Sprintf("%s-%s-v4.ip.zstaticcdn.com", code, carrier),
				IPv6:    fmt.Sprintf("%s-%s-v6.ip.zstaticcdn.com", code, carrier),
			})
		}
		result = append(result, provinceRoute{Code: strings.ToUpper(code), Name: name, Province: item.Province, Short: strings.TrimSpace(item.Short), Targets: targets})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Province < result[j].Province })
	return result, len(result), nil
}

func parseSpeedtestServers(data []byte) (any, int, error) {
	var input []map[string]any
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, 0, err
	}
	result := make([]map[string]any, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, server := range input {
		id := strings.TrimSpace(fmt.Sprint(server["id"]))
		host := strings.TrimSpace(fmt.Sprint(server["host"]))
		url := strings.TrimSpace(fmt.Sprint(server["url"]))
		if id == "" || id == "<nil>" || host == "" || host == "<nil>" || url == "" || url == "<nil>" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, server)
	}
	sort.Slice(result, func(i, j int) bool { return fmt.Sprint(result[i]["id"]) < fmt.Sprint(result[j]["id"]) })
	return result, len(result), nil
}

func probeSpeedtestServers(ctx context.Context, value any) (any, error) {
	servers, ok := value.([]map[string]any)
	if !ok {
		return nil, errors.New("speedtest health check received unexpected type")
	}
	addresses := make([]string, len(servers))
	for index, server := range servers {
		addresses[index] = strings.TrimSpace(fmt.Sprint(server["host"]))
	}
	available := probeTCPAddresses(ctx, addresses, 2*time.Second, 8)
	reachable := 0
	for index := range servers {
		status := "unavailable"
		if available[index] {
			status = "available"
			reachable++
		}
		servers[index]["status"] = status
	}
	if reachable < 2 {
		return nil, fmt.Errorf("only %d speedtest servers passed TCP health checks", reachable)
	}
	return servers, nil
}

func probeTransferTargets(ctx context.Context, value any) (any, error) {
	targets, ok := value.([]transferTarget)
	if !ok {
		return nil, errors.New("transfer health check received unexpected type")
	}
	addresses := make([]string, len(targets))
	for index, target := range targets {
		addresses[index] = net.JoinHostPort(target.Host, strconv.Itoa(target.PortFrom))
	}
	available := probeTCPAddresses(ctx, addresses, 2*time.Second, 8)
	reachable := 0
	for index := range targets {
		targets[index].Status = "unavailable"
		if available[index] {
			targets[index].Status = "available"
			reachable++
		}
	}
	if reachable < 2 {
		return nil, fmt.Errorf("only %d transfer targets passed TCP health checks", reachable)
	}
	return targets, nil
}

func probeTCPAddresses(ctx context.Context, addresses []string, timeout time.Duration, concurrency int) []bool {
	result := make([]bool, len(addresses))
	if concurrency <= 0 || len(addresses) == 0 {
		return result
	}
	jobs := make(chan int)
	workers := min(concurrency, len(addresses))
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			dialer := net.Dialer{Timeout: timeout}
			for index := range jobs {
				probeCtx, cancel := context.WithTimeout(ctx, timeout)
				connection, err := dialer.DialContext(probeCtx, "tcp", addresses[index])
				cancel()
				if err == nil {
					result[index] = true
					_ = connection.Close()
				}
			}
		}()
	}
	for index := range addresses {
		select {
		case jobs <- index:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return result
		}
	}
	close(jobs)
	wg.Wait()
	return result
}

var shellQuoted = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)

func parseTCPTargets(data []byte) (any, int, error) {
	var names, hosts []string
	for _, line := range strings.Split(string(data), "\n") {
		switch {
		case strings.HasPrefix(line, "NAMES=("):
			names = parseQuotedArray(line)
		case strings.HasPrefix(line, "HOSTS=("):
			hosts = parseQuotedArray(line)
		}
	}
	if len(names) == 0 || len(names) != len(hosts) {
		return nil, 0, fmt.Errorf("name/host length mismatch: %d/%d", len(names), len(hosts))
	}
	result := make([]tcpTarget, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for i := range names {
		host := strings.ToLower(strings.TrimSpace(hosts[i]))
		if host == "" {
			continue
		}
		if _, exists := seen[host]; exists {
			continue
		}
		seen[host] = struct{}{}
		result = append(result, tcpTarget{ID: slug(names[i]), Name: names[i], Host: host, Port: 443, Category: "global"})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, len(result), nil
}

func parseQuotedArray(line string) []string {
	matches := shellQuoted.FindAllStringSubmatch(line, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		value, err := strconv.Unquote(`"` + match[1] + `"`)
		if err == nil {
			result = append(result, value)
		}
	}
	return result
}

func parseDNSBL(data []byte) (any, int, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	seen := make(map[string]struct{})
	for scanner.Scan() {
		zone := strings.ToLower(strings.TrimSpace(scanner.Text()))
		zone = strings.TrimSuffix(zone, ".")
		if zone == "" || strings.HasPrefix(zone, "#") || !strings.Contains(zone, ".") || !validDNSName(zone) {
			continue
		}
		seen[zone] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}
	result := make([]dnsblZone, 0, len(seen))
	for zone := range seen {
		if concatenatedDNSBLZone(zone, seen) {
			continue
		}
		ipv6 := strings.Contains(zone, "v6") || strings.Contains(zone, "ipv6")
		// The upstream list does not publish address-family metadata. Only
		// explicitly IPv6-named zones are marked v6-capable; ordinary zones are
		// conservatively IPv4-only instead of claiming unsupported v6 queries.
		ipv4 := !ipv6
		result = append(result, dnsblZone{Zone: zone, IPv4: ipv4, IPv6: ipv6})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Zone < result[j].Zone })
	return result, len(result), nil
}

func concatenatedDNSBLZone(zone string, zones map[string]struct{}) bool {
	for suffix := range zones {
		if suffix == zone || len(suffix) >= len(zone) || !strings.HasSuffix(zone, suffix) {
			continue
		}
		start := len(zone) - len(suffix)
		if start <= 0 || zone[start-1] == '.' {
			continue
		}
		prefix := strings.TrimSuffix(zone[:start], ".")
		if strings.Contains(prefix, ".") && validDNSName(prefix) {
			return true
		}
	}
	return false
}

var errDNSBLZoneInvalid = errors.New("DNSBL base domain is invalid")

var lookupDNSBLZone = func(ctx context.Context, zone string) error {
	base := dnsblHealthKey(zone)
	var lastErr error
	for _, endpoint := range []string{"https://dns.google/resolve", "https://cloudflare-dns.com/dns-query"} {
		probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err := lookupDNSBLBaseDOH(probeCtx, endpoint, base)
		cancel()
		if err == nil || errors.Is(err, errDNSBLZoneInvalid) {
			return err
		}
		lastErr = err
	}
	return lastErr
}

func lookupDNSBLBaseDOH(ctx context.Context, endpoint, base string) error {
	requestURL := endpoint + "?name=" + url.QueryEscape(base) + "&type=NS"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/dns-json")
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("DNS-over-HTTPS returned HTTP %d", response.StatusCode)
	}
	var payload struct {
		Status int `json:"Status"`
		Answer []struct {
			Type int `json:"type"`
		} `json:"Answer"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(&payload); err != nil {
		return err
	}
	if payload.Status != 0 {
		return fmt.Errorf("%w: DNS-over-HTTPS status %d", errDNSBLZoneInvalid, payload.Status)
	}
	for _, answer := range payload.Answer {
		if answer.Type == 2 {
			return nil
		}
	}
	return fmt.Errorf("%w: DNS-over-HTTPS response has no NS answer", errDNSBLZoneInvalid)
}

func dnsblHealthKey(zone string) string {
	base, err := publicsuffix.EffectiveTLDPlusOne(strings.TrimSuffix(strings.ToLower(strings.TrimSpace(zone)), "."))
	if err == nil && base != "" {
		return base
	}
	return zone
}

func probeDNSBLZones(ctx context.Context, value any) (any, error) {
	zones, ok := value.([]dnsblZone)
	if !ok {
		return nil, errors.New("DNSBL health input has an invalid type")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	groups := make(map[string][]dnsblZone, len(zones))
	for _, zone := range zones {
		key := dnsblHealthKey(zone.Zone)
		groups[key] = append(groups[key], zone)
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	workers := min(16, len(keys))
	if workers == 0 {
		return zones, nil
	}
	type healthResult struct {
		key string
		err error
	}
	jobs := make(chan string)
	results := make(chan healthResult, len(keys))
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for key := range jobs {
				probeCtx, cancel := context.WithTimeout(ctx, 7*time.Second)
				err := lookupDNSBLZone(probeCtx, groups[key][0].Zone)
				cancel()
				results <- healthResult{key: key, err: err}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, key := range keys {
			select {
			case jobs <- key:
			case <-ctx.Done():
				return
			}
		}
	}()
	wg.Wait()
	close(results)
	kept := make([]dnsblZone, 0, len(zones))
	var healthErr error
	for result := range results {
		if result.err == nil {
			kept = append(kept, groups[result.key]...)
		} else if !errors.Is(result.err, errDNSBLZoneInvalid) && healthErr == nil {
			healthErr = fmt.Errorf("probe %s: %w", result.key, result.err)
		}
	}
	if healthErr != nil {
		return nil, healthErr
	}
	sort.Slice(kept, func(i, j int) bool { return kept[i].Zone < kept[j].Zone })
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return kept, nil
}

func validDNSName(name string) bool {
	name = strings.TrimSuffix(strings.TrimSpace(name), ".")
	if name == "" || len(name) > 253 {
		return false
	}
	for _, label := range strings.Split(name, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
				continue
			}
			return false
		}
	}
	return true
}

func parseASNMap(data []byte) (any, int, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	seen := make(map[uint32]string)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || !strings.HasPrefix(strings.ToUpper(fields[0]), "AS") {
			continue
		}
		n, err := strconv.ParseUint(strings.TrimPrefix(strings.ToUpper(fields[0]), "AS"), 10, 32)
		if err != nil {
			continue
		}
		if _, exists := seen[uint32(n)]; !exists {
			seen[uint32(n)] = strings.Join(fields[1:], " ")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}
	result := make([]asnName, 0, len(seen))
	for asn, name := range seen {
		result = append(result, asnName{ASN: asn, Name: name})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ASN < result[j].ASN })
	return result, len(result), nil
}

func parseTransferTargets(data []byte) (any, int, error) {
	var input []struct {
		Code        int    `json:"code"`
		Server      string `json:"server"`
		PortFrom    int    `json:"portl"`
		PortTo      int    `json:"portu"`
		Provider    string `json:"provider"`
		CountryCode string `json:"countrycode"`
		City        string `json:"city"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, 0, err
	}
	result := make([]transferTarget, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, item := range input {
		host := strings.ToLower(strings.TrimSpace(item.Server))
		city := strings.TrimSpace(item.City)
		if item.Code <= 0 || host == "" || city == "" || item.PortFrom <= 0 || item.PortFrom > 65535 || item.PortTo < item.PortFrom || item.PortTo > 65535 {
			continue
		}
		id := fmt.Sprintf("%d-%s", item.Code, slug(city))
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, transferTarget{
			ID: id, Host: host,
			PortFrom: item.PortFrom, PortTo: item.PortTo, Provider: strings.TrimSpace(item.Provider),
			Country: strings.ToUpper(strings.TrimSpace(item.CountryCode)), City: city, Status: "candidate",
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, len(result), nil
}

func parseProviderNames(data []byte) (any, int, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "lists.go", data, 0)
	if err != nil {
		return nil, 0, err
	}
	type providerMetadata struct {
		name         string
		groups       map[string]struct{}
		supportsIPv6 bool
	}
	seen := make(map[string]*providerMetadata)
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, spec := range general.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for index, expression := range values.Values {
				literal, ok := expression.(*ast.CompositeLit)
				if !ok || !isTestItemArray(literal) {
					continue
				}
				group := "other"
				if index < len(values.Names) {
					group = mediaProviderGroup(values.Names[index].Name)
				}
				for _, entry := range literal.Elts {
					item, ok := entry.(*ast.CompositeLit)
					if !ok || len(item.Elts) < 2 || isNilExpression(item.Elts[1]) {
						continue
					}
					nameLiteral, ok := item.Elts[0].(*ast.BasicLit)
					if !ok || nameLiteral.Kind != token.STRING {
						continue
					}
					name, err := strconv.Unquote(nameLiteral.Value)
					if err != nil {
						continue
					}
					name = strings.TrimSpace(name)
					id := slug(name)
					if name == "" || id == "" {
						continue
					}
					metadata := seen[id]
					if metadata == nil {
						metadata = &providerMetadata{name: name, groups: make(map[string]struct{})}
						seen[id] = metadata
					}
					metadata.groups[group] = struct{}{}
					if len(item.Elts) >= 3 {
						if enabled, ok := item.Elts[2].(*ast.Ident); ok && enabled.Name == "true" {
							metadata.supportsIPv6 = true
						}
					}
				}
			}
		}
	}
	if len(seen) == 0 {
		return nil, 0, errors.New("no TestItem provider entries found")
	}
	result := make([]providerName, 0, len(seen))
	for id, metadata := range seen {
		groups := make([]string, 0, len(metadata.groups))
		for group := range metadata.groups {
			groups = append(groups, group)
		}
		sort.Strings(groups)
		result = append(result, providerName{ID: id, Name: metadata.name, Groups: groups, SupportsIPv6: metadata.supportsIPv6})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, len(result), nil
}

func isTestItemArray(literal *ast.CompositeLit) bool {
	array, ok := literal.Type.(*ast.ArrayType)
	if !ok {
		return false
	}
	element, ok := array.Elt.(*ast.Ident)
	return ok && element.Name == "TestItem"
}

func isNilExpression(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == "nil"
}

func mediaProviderGroup(name string) string {
	groups := map[string]string{
		"GlobeTests": "global", "HongKongTests": "hk", "TaiwanTests": "tw",
		"JapanTests": "jp", "KoreaTests": "kr", "NorthAmericaTests": "na",
		"SouthAmericaTests": "sa", "EuropeTests": "eu", "AfricaTests": "africa",
		"SouthEastAsiaTests": "sea", "OceaniaTests": "oceania", "AITests": "ai",
	}
	if group := groups[name]; group != "" {
		return group
	}
	return "other"
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".goecs-data-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

type snapshotBackup struct {
	data   []byte
	exists bool
}

// commitSnapshot writes all payloads, removes obsolete generated JSON files,
// and writes the manifest last. Every touched file participates in rollback.
func commitSnapshot(outputDir string, staged map[string]stagedFile, manifestData []byte, writer func(string, []byte) error) error {
	return commitSnapshotWithRemove(outputDir, staged, manifestData, writer, nil)
}

func commitSnapshotWithRemove(outputDir string, staged map[string]stagedFile, manifestData []byte, writer func(string, []byte) error, remover func(string) error) error {
	if writer == nil {
		writer = atomicWrite
	}
	if remover == nil {
		remover = os.Remove
	}
	payloadNames := make([]string, 0, len(staged))
	for name := range staged {
		payloadNames = append(payloadNames, name)
	}
	sort.Strings(payloadNames)
	staleNames, err := obsoleteSnapshotNames(outputDir, staged)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(payloadNames)+len(staleNames)+1)
	names = append(names, payloadNames...)
	names = append(names, staleNames...)
	names = append(names, "manifest.json")
	backups := make(map[string]snapshotBackup, len(names))
	for _, name := range names {
		path := filepath.Join(outputDir, name)
		data, err := os.ReadFile(path)
		switch {
		case err == nil:
			backups[path] = snapshotBackup{data: data, exists: true}
		case errors.Is(err, os.ErrNotExist):
			backups[path] = snapshotBackup{}
		default:
			return fmt.Errorf("read existing %s: %w", name, err)
		}
	}
	restore := func() error {
		var restoreErr error
		for _, name := range names {
			path := filepath.Join(outputDir, name)
			backup := backups[path]
			if backup.exists {
				if err := atomicWrite(path, backup.data); err != nil && restoreErr == nil {
					restoreErr = err
				}
				continue
			}
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) && restoreErr == nil {
				restoreErr = err
			}
		}
		return restoreErr
	}
	rollback := func(operation, name string, err error) error {
		if restoreErr := restore(); restoreErr != nil {
			return fmt.Errorf("%s %s: %v (rollback failed: %v)", operation, name, err, restoreErr)
		}
		return fmt.Errorf("%s %s: %w", operation, name, err)
	}
	for _, name := range payloadNames {
		if err := writer(filepath.Join(outputDir, name), staged[name].data); err != nil {
			return rollback("write", name, err)
		}
	}
	for _, name := range staleNames {
		if err := remover(filepath.Join(outputDir, name)); err != nil {
			return rollback("remove", name, err)
		}
	}
	if err := writer(filepath.Join(outputDir, "manifest.json"), manifestData); err != nil {
		return rollback("write", "manifest.json", err)
	}
	return nil
}

func obsoleteSnapshotNames(outputDir string, staged map[string]stagedFile) ([]string, error) {
	manifestData, err := os.ReadFile(filepath.Join(outputDir, "manifest.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read current manifest: %w", err)
	}
	var current manifest
	if err := json.Unmarshal(manifestData, &current); err != nil {
		return nil, fmt.Errorf("decode current manifest: %w", err)
	}
	if !supportedCurrentSchema(current.Schema) || current.Files == nil {
		return nil, errors.New("current manifest has an invalid schema or file set")
	}
	staleNames := make([]string, 0)
	for name := range current.Files {
		if _, ok := staged[name]; ok {
			continue
		}
		if name == "manifest.json" || filepath.Base(name) != name || filepath.Ext(name) != ".json" {
			return nil, fmt.Errorf("current manifest has invalid payload name %q", name)
		}
		info, err := os.Lstat(filepath.Join(outputDir, name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspect obsolete %s: %w", name, err)
		}
		if info.Mode().IsRegular() {
			staleNames = append(staleNames, name)
		}
	}
	sort.Strings(staleNames)
	return staleNames, nil
}
