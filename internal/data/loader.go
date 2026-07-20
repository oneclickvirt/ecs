package data

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//go:embed snapshot/manifest.json snapshot/*.json
var embedded embed.FS

const (
	manifestPath   = "snapshot/manifest.json"
	dataBaseURL    = "https://raw.githubusercontent.com/oneclickvirt/ecs-data/main/data/"
	expectedSchema = "ecs-data/v1"
	requestTimeout = 12 * time.Second
	maxPayloadSize = 16 << 20
)

// KnownFiles is the set of data files understood by this build. Keep this
// list explicit so a newly published manifest cannot silently introduce an
// unsupported payload into a report or runtime path.
var knownFiles = []string{
	"bgp-asn-map.json",
	"dnsbl-zones.json",
	"media-providers.json",
	"openspeedtest-servers.json",
	"province-routes.json",
	"speedtest-servers.json",
	"tcp-targets.json",
}

// KnownFiles returns a stable copy of all supported ecs-data payload names.
func KnownFiles() []string {
	return append([]string(nil), knownFiles...)
}

type FileMeta struct {
	SHA256 string `json:"sha256"`
	Count  int    `json:"count"`
	Source string `json:"source"`
}

type Manifest struct {
	Schema      string              `json:"schema"`
	GeneratedAt time.Time           `json:"generated_at"`
	Files       map[string]FileMeta `json:"files"`
}

type Loader struct {
	Client  *http.Client
	CDNBase string
	RawBase string
}

type Result struct {
	Name       string
	Data       []byte
	Manifest   Manifest
	UsedRemote bool
	Fallback   string
	Source     string
}

func NewLoader(client *http.Client, cdnBase string) *Loader {
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	return &Loader{Client: client, CDNBase: strings.TrimRight(cdnBase, "/"), RawBase: strings.TrimRight(dataBaseURL, "/")}
}

func (l *Loader) Load(ctx context.Context, name string) (Result, error) {
	results, err := l.LoadMany(ctx, []string{name})
	if err != nil {
		return Result{}, err
	}
	result, ok := results[name]
	if !ok {
		return Result{}, fmt.Errorf("data file %q was not loaded", name)
	}
	return result, nil
}

// LoadMany loads an internally consistent set of files. All returned results
// are verified against the same manifest; if any requested payload fails, the
// complete candidate is rejected before trying the next manifest source.
func (l *Loader) LoadMany(ctx context.Context, names []string) (map[string]Result, error) {
	names, err := normalizedDataNames(names)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, manifestBase := range l.manifestBases() {
		manifestData, err := l.fetch(ctx, strings.TrimRight(manifestBase, "/")+"/manifest.json")
		if err != nil {
			lastErr = err
			continue
		}
		var m Manifest
		if err := json.Unmarshal(manifestData, &m); err != nil {
			lastErr = fmt.Errorf("decode remote manifest from %s: %w", manifestBase, err)
			continue
		}
		if m.Schema != expectedSchema || m.GeneratedAt.IsZero() || m.Files == nil {
			lastErr = fmt.Errorf("invalid remote manifest from %s", manifestBase)
			continue
		}
		candidate := make(map[string]Result, len(names))
		candidateValid := true
		manifestSource := sourceForBase(l.CDNBase, l.RawBase, manifestBase)
		manifestFallback := manifestSource == "cdn" && strings.TrimRight(l.RawBase, "/") != "" && strings.TrimRight(l.RawBase, "/") != strings.TrimRight(l.CDNBase, "/")
		for _, name := range names {
			meta, ok := m.Files[name]
			if !ok {
				lastErr = fmt.Errorf("data file %q is not in manifest from %s", name, manifestBase)
				candidateValid = false
				break
			}
			var loaded bool
			for _, dataBase := range l.bases() {
				data, dataErr := l.fetch(ctx, strings.TrimRight(dataBase, "/")+"/"+name)
				if dataErr == nil {
					dataErr = verify(data, meta.SHA256)
				}
				if dataErr == nil {
					dataErr = verifyCount(data, meta.Count)
				}
				if dataErr == nil {
					dataErr = verifySchema(name, data)
				}
				if dataErr != nil {
					lastErr = fmt.Errorf("validate %s from %s using manifest %s: %w", name, dataBase, manifestBase, dataErr)
					continue
				}
				dataSource := sourceForBase(l.CDNBase, l.RawBase, dataBase)
				fallback := ""
				if dataSource == "raw" {
					fallback = "raw"
				} else if manifestFallback {
					fallback = "cdn_manifest"
				}
				candidate[name] = Result{Name: name, Data: data, Manifest: m, UsedRemote: true, Source: dataSource, Fallback: fallback}
				loaded = true
				break
			}
			if !loaded {
				candidateValid = false
				break
			}
		}
		if candidateValid && len(candidate) == len(names) {
			return candidate, nil
		}
	}
	embeddedResults, embeddedErr := l.loadEmbeddedMany(names)
	if embeddedErr != nil && lastErr != nil {
		return nil, fmt.Errorf("remote data failed: %v; embedded data failed: %w", lastErr, embeddedErr)
	}
	return embeddedResults, embeddedErr
}

func (l *Loader) loadEmbedded(name string) (Result, error) {
	results, err := l.loadEmbeddedMany([]string{name})
	if err != nil {
		return Result{}, err
	}
	return results[name], nil
}

func (l *Loader) loadEmbeddedMany(names []string) (map[string]Result, error) {
	names, err := normalizedDataNames(names)
	if err != nil {
		return nil, err
	}
	manifestData, err := embedded.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("load embedded manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		return nil, fmt.Errorf("decode embedded manifest: %w", err)
	}
	if m.Schema != expectedSchema || m.GeneratedAt.IsZero() || m.Files == nil {
		return nil, errors.New("embedded manifest is missing schema or files")
	}
	results := make(map[string]Result, len(names))
	for _, name := range names {
		meta, ok := m.Files[name]
		if !ok {
			return nil, fmt.Errorf("embedded data file %q is not in manifest", name)
		}
		data, err := embedded.ReadFile("snapshot/" + name)
		if err != nil {
			return nil, fmt.Errorf("load embedded %s: %w", name, err)
		}
		if err := verify(data, meta.SHA256); err != nil {
			return nil, fmt.Errorf("verify embedded %s: %w", name, err)
		}
		if err := verifyCount(data, meta.Count); err != nil {
			return nil, fmt.Errorf("verify embedded %s: %w", name, err)
		}
		if err := verifySchema(name, data); err != nil {
			return nil, fmt.Errorf("verify embedded %s schema: %w", name, err)
		}
		results[name] = Result{Name: name, Data: data, Manifest: m, Fallback: "embedded", Source: "embedded"}
	}
	return results, nil
}

func normalizedDataNames(names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, errors.New("no data files requested")
	}
	result := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || filepath.Base(name) != name || filepath.Ext(name) != ".json" {
			return nil, fmt.Errorf("invalid data file name %q", name)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("duplicate data file name %q", name)
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result, nil
}

func sourceForBase(cdnBase, rawBase, base string) string {
	base = strings.TrimRight(base, "/")
	if base != "" && base == strings.TrimRight(cdnBase, "/") {
		return "cdn"
	}
	if base != "" && base == strings.TrimRight(rawBase, "/") {
		return "raw"
	}
	return "remote"
}

func (l *Loader) fetch(ctx context.Context, name string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, name, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "oneclickvirt-goecs-data/1")
	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxPayloadSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxPayloadSize {
		return nil, fmt.Errorf("response exceeds %d bytes", maxPayloadSize)
	}
	return data, nil
}

func (l *Loader) bases() []string {
	result := make([]string, 0, 2)
	seen := make(map[string]struct{}, 2)
	for _, base := range []string{l.CDNBase, l.RawBase} {
		base = strings.TrimRight(base, "/")
		if base == "" {
			continue
		}
		if _, ok := seen[base]; ok {
			continue
		}
		seen[base] = struct{}{}
		result = append(result, base)
	}
	return result
}

// manifestBases prefers the GitHub Raw manifest because it is the canonical
// generation marker. The CDN manifest is retained only as a last remote
// resort when Raw is unavailable; payloads are still validated against the
// selected manifest before they can be returned.
func (l *Loader) manifestBases() []string {
	result := make([]string, 0, 2)
	seen := make(map[string]struct{}, 2)
	for _, base := range []string{l.RawBase, l.CDNBase} {
		base = strings.TrimRight(base, "/")
		if base == "" {
			continue
		}
		if _, ok := seen[base]; ok {
			continue
		}
		seen[base] = struct{}{}
		result = append(result, base)
	}
	return result
}

func verify(data []byte, expected string) error {
	if expected == "" {
		return errors.New("missing SHA-256")
	}
	hash := sha256.Sum256(data)
	if !strings.EqualFold(hex.EncodeToString(hash[:]), expected) {
		return errors.New("SHA-256 mismatch")
	}
	return nil
}

func verifyCount(data []byte, expected int) error {
	if expected < 0 {
		return errors.New("invalid negative record count")
	}
	var records []json.RawMessage
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("decode records: %w", err)
	}
	if len(records) != expected {
		return fmt.Errorf("record count mismatch: got %d want %d", len(records), expected)
	}
	return nil
}

func verifySchema(name string, data []byte) error {
	switch name {
	case "tcp-targets.json":
		var records []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Host string `json:"host"`
			Port int    `json:"port"`
		}
		if err := json.Unmarshal(data, &records); err != nil {
			return err
		}
		seen := make(map[string]struct{}, len(records))
		for _, record := range records {
			key := strings.ToLower(strings.TrimSpace(record.ID))
			if key == "" || strings.TrimSpace(record.Name) == "" || !validDataHost(record.Host) || record.Port < 1 || record.Port > 65535 {
				return errors.New("invalid TCP target fields")
			}
			if _, exists := seen[key]; exists {
				return errors.New("duplicate TCP target id")
			}
			seen[key] = struct{}{}
		}
	case "province-routes.json":
		var records []struct {
			Code     string `json:"code"`
			Name     string `json:"name"`
			Province int    `json:"province"`
			Targets  []struct {
				Carrier string `json:"carrier"`
				IPv4    string `json:"ipv4"`
				IPv6    string `json:"ipv6"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(data, &records); err != nil {
			return err
		}
		seen := make(map[string]struct{}, len(records))
		for _, record := range records {
			code := strings.ToUpper(strings.TrimSpace(record.Code))
			if code == "" || strings.TrimSpace(record.Name) == "" || record.Province <= 0 || len(record.Targets) != 3 {
				return errors.New("invalid province route fields")
			}
			if _, exists := seen[code]; exists {
				return errors.New("duplicate province code")
			}
			seen[code] = struct{}{}
			carriers := make([]string, 0, len(record.Targets))
			for _, target := range record.Targets {
				carrier := strings.ToLower(strings.TrimSpace(target.Carrier))
				if carrier == "" || !validDataHost(target.IPv4) || !validDataHost(target.IPv6) {
					return errors.New("invalid province carrier target")
				}
				carriers = append(carriers, carrier)
			}
			sort.Strings(carriers)
			if carriers[0] == carriers[1] || carriers[1] == carriers[2] {
				return errors.New("duplicate province carrier")
			}
		}
	case "bgp-asn-map.json":
		var records []struct {
			ASN  uint32 `json:"asn"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &records); err != nil {
			return err
		}
		seen := make(map[uint32]struct{}, len(records))
		for _, record := range records {
			if record.ASN == 0 || strings.TrimSpace(record.Name) == "" {
				return errors.New("invalid ASN mapping fields")
			}
			if _, exists := seen[record.ASN]; exists {
				return errors.New("duplicate ASN mapping")
			}
			seen[record.ASN] = struct{}{}
		}
	case "speedtest-servers.json":
		var records []struct {
			ID     string `json:"id"`
			Host   string `json:"host"`
			URL    string `json:"url"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &records); err != nil {
			return err
		}
		seen := make(map[string]struct{}, len(records))
		for _, record := range records {
			parsed, err := url.Parse(strings.TrimSpace(record.URL))
			key := strings.TrimSpace(record.ID)
			if key == "" || !validDataAddress(record.Host) || err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || !validAvailability(record.Status) {
				return errors.New("invalid speedtest server fields")
			}
			if _, exists := seen[key]; exists {
				return errors.New("duplicate speedtest server id")
			}
			seen[key] = struct{}{}
		}
	case "openspeedtest-servers.json":
		var records []struct {
			ID       string `json:"id"`
			Host     string `json:"host"`
			PortFrom int    `json:"port_from"`
			PortTo   int    `json:"port_to"`
			Status   string `json:"status"`
		}
		if err := json.Unmarshal(data, &records); err != nil {
			return err
		}
		seen := make(map[string]struct{}, len(records))
		for _, record := range records {
			key := strings.TrimSpace(record.ID)
			if key == "" || !validDataHost(record.Host) || record.PortFrom < 1 || record.PortTo < record.PortFrom || record.PortTo > 65535 || !validAvailability(record.Status) {
				return errors.New("invalid transfer server fields")
			}
			if _, exists := seen[key]; exists {
				return errors.New("duplicate transfer server id")
			}
			seen[key] = struct{}{}
		}
	case "dnsbl-zones.json":
		var records []struct {
			Zone string `json:"zone"`
			IPv4 bool   `json:"ipv4"`
			IPv6 bool   `json:"ipv6"`
		}
		if err := json.Unmarshal(data, &records); err != nil {
			return err
		}
		seen := make(map[string]struct{}, len(records))
		for _, record := range records {
			zone := strings.ToLower(strings.Trim(strings.TrimSpace(record.Zone), "."))
			if !validDataHost(zone) || (!record.IPv4 && !record.IPv6) {
				return errors.New("invalid DNSBL zone fields")
			}
			if _, exists := seen[zone]; exists {
				return errors.New("duplicate DNSBL zone")
			}
			seen[zone] = struct{}{}
		}
	case "media-providers.json":
		var records []struct {
			ID           string   `json:"id"`
			Name         string   `json:"name"`
			Groups       []string `json:"groups"`
			SupportsIPv6 bool     `json:"supports_ipv6"`
		}
		if err := json.Unmarshal(data, &records); err != nil {
			return err
		}
		seen := make(map[string]struct{}, len(records))
		for _, record := range records {
			key := strings.ToLower(strings.TrimSpace(record.ID))
			if key == "" || strings.TrimSpace(record.Name) == "" || len(record.Groups) == 0 {
				return errors.New("invalid media provider fields")
			}
			for index, group := range record.Groups {
				if strings.TrimSpace(group) == "" || index > 0 && group <= record.Groups[index-1] {
					return errors.New("invalid media provider groups")
				}
			}
			if _, exists := seen[key]; exists {
				return errors.New("duplicate media provider id")
			}
			seen[key] = struct{}{}
		}
	default:
		return fmt.Errorf("unsupported data schema %q", name)
	}
	return nil
}

func validDataAddress(value string) bool {
	value = strings.TrimSpace(value)
	if host, port, err := net.SplitHostPort(value); err == nil {
		parsedPort, portErr := net.LookupPort("tcp", port)
		return validDataHost(host) && portErr == nil && parsedPort >= 1 && parsedPort <= 65535
	}
	return validDataHost(value)
}

func validDataHost(value string) bool {
	value = strings.Trim(strings.TrimSpace(value), ".")
	if value == "" || len(value) > 253 || strings.ContainsAny(value, " /\\\t\r\n") {
		return false
	}
	if net.ParseIP(value) != nil {
		return true
	}
	for _, label := range strings.Split(value, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' {
				return false
			}
		}
	}
	return true
}

func validAvailability(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "candidate", "available", "unavailable":
		return true
	default:
		return false
	}
}
