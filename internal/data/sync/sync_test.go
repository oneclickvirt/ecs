package datasync

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseTCPTargets(t *testing.T) {
	value, count, err := parseTCPTargets([]byte(`NAMES=("One" "Two")
HOSTS=("one.example" "two.example")`))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d", count)
	}
	targets := value.([]tcpTarget)
	if targets[0].Port != 443 || targets[1].Host == "" {
		t.Fatalf("unexpected targets: %#v", targets)
	}
}

func TestProbeTCPAddresses(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	result := probeTCPAddresses(context.Background(), []string{listener.Addr().String(), "127.0.0.1:1"}, 100*time.Millisecond, 2)
	if !result[0] || result[1] {
		t.Fatalf("unexpected probe result: %v", result)
	}
}

func TestHealthChecksRejectSevereAvailabilityDrop(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	address := listener.Addr().String()
	servers := []map[string]any{{"id": "one", "host": address}, {"id": "two", "host": "127.0.0.1:1"}}
	if _, err := probeSpeedtestServers(context.Background(), servers); err == nil {
		t.Fatal("expected one of two reachable speedtest nodes to be rejected")
	}
	targets := []transferTarget{{ID: "one", Host: "127.0.0.1", PortFrom: listener.Addr().(*net.TCPAddr).Port, PortTo: listener.Addr().(*net.TCPAddr).Port}, {ID: "two", Host: "127.0.0.1", PortFrom: 1, PortTo: 1}}
	if _, err := probeTransferTargets(context.Background(), targets); err == nil {
		t.Fatal("expected one of two reachable transfer nodes to be rejected")
	}
}

func TestSemanticDataEqual(t *testing.T) {
	dir := t.TempDir()
	data := []byte("[]\n")
	hash := sha256.Sum256(data)
	current := manifest{
		Schema: schemaVersion, GeneratedAt: time.Now(),
		Files: map[string]fileManifest{"a.json": {SHA256: hex.EncodeToString(hash[:]), Count: 0, Source: "fixture"}},
	}
	manifestData, err := json.Marshal(current)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifestData, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	staged := map[string]stagedFile{"a.json": {data: data, count: 0, source: "fixture"}}
	if !semanticDataEqual(dir, staged) {
		t.Fatal("expected identical data to be semantic no-op")
	}
	staged["a.json"] = stagedFile{data: []byte("[1]\n"), count: 1, source: "fixture"}
	if semanticDataEqual(dir, staged) {
		t.Fatal("expected data change to be detected")
	}
}

func TestValidateCountDrops(t *testing.T) {
	dir := t.TempDir()
	current := manifest{Schema: schemaVersion, Files: map[string]fileManifest{"a.json": {Count: 100}}}
	manifestData, err := json.Marshal(current)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifestData, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateCountDrops(dir, map[string]stagedFile{"a.json": {count: 70}}, 0.35); err != nil {
		t.Fatalf("expected moderate drop to pass: %v", err)
	}
	if err := validateCountDrops(dir, map[string]stagedFile{"a.json": {count: 50}}, 0.35); err == nil {
		t.Fatal("expected large drop to fail")
	}

	current.Files["a.json"] = fileManifest{Count: 3}
	manifestData, err = json.Marshal(current)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifestData, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateCountDrops(dir, map[string]stagedFile{"a.json": {count: 1}}, 0.35); err == nil {
		t.Fatal("expected rounded small-set drop to fail")
	}
}

func TestSynchronizeFirstRunAndSemanticNoOp(t *testing.T) {
	payload := []byte(`[{"id":"one"},{"id":"two"}]`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	dir := filepath.Join(t.TempDir(), "data")
	specs := []sourceSpec{{name: "fixture", file: "fixture.json", url: server.URL, minimum: 2, transform: passJSONArray}}
	changed, err := synchronize(context.Background(), dir, specs)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected first run to write a snapshot")
	}
	firstManifest, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var generated manifest
	if err := json.Unmarshal(firstManifest, &generated); err != nil {
		t.Fatal(err)
	}
	if generated.Schema != schemaVersion || generated.Files["fixture.json"].Count != 2 {
		t.Fatalf("unexpected manifest: %+v", generated)
	}
	data, err := os.ReadFile(filepath.Join(dir, "fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(data)
	if generated.Files["fixture.json"].SHA256 != hex.EncodeToString(hash[:]) {
		t.Fatal("manifest SHA-256 does not match generated data")
	}

	changed, err = synchronize(context.Background(), dir, specs)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("expected identical data to be a semantic no-op")
	}
	secondManifest, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstManifest, secondManifest) {
		t.Fatal("semantic no-op rewrote the manifest timestamp")
	}
}

func TestSynchronizeRemovesSourceNoLongerInManifest(t *testing.T) {
	payload := []byte(`[{"id":"one"}]`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	dir := t.TempDir()
	initial := []sourceSpec{
		{name: "current", file: "current.json", url: server.URL, minimum: 1, transform: passJSONArray},
		{name: "obsolete", file: "obsolete.json", url: server.URL, minimum: 1, transform: passJSONArray},
	}
	if changed, err := synchronize(context.Background(), dir, initial); err != nil || !changed {
		t.Fatalf("write initial snapshot: changed=%v err=%v", changed, err)
	}
	untouched := []string{"local-config.yaml", "local-config.json"}
	for _, name := range untouched {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("keep\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	payload = []byte(`[{"id":"updated"}]`)
	if changed, err := synchronize(context.Background(), dir, initial[:1]); err != nil || !changed {
		t.Fatalf("write reduced snapshot: changed=%v err=%v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "obsolete.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("obsolete generated file was not removed: %v", err)
	}
	for _, name := range untouched {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(got) != "keep\n" {
			t.Fatalf("untracked file %s changed: data=%q err=%v", name, got, err)
		}
	}
	manifestData, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var current manifest
	if err := json.Unmarshal(manifestData, &current); err != nil {
		t.Fatal(err)
	}
	if len(current.Files) != 1 {
		t.Fatalf("manifest has %d files, want 1", len(current.Files))
	}
	if _, ok := current.Files["current.json"]; !ok {
		t.Fatal("manifest is missing current.json")
	}
}

func TestSynchronizeDoesNotOverwriteOnQuantityDrop(t *testing.T) {
	payload := []byte(`[{"id":1},{"id":2},{"id":3},{"id":4},{"id":5},{"id":6},{"id":7},{"id":8},{"id":9},{"id":10}]`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	dir := t.TempDir()
	specs := []sourceSpec{{name: "fixture", file: "fixture.json", url: server.URL, minimum: 1, transform: passJSONArray}}
	if _, err := synchronize(context.Background(), dir, specs); err != nil {
		t.Fatal(err)
	}
	beforeData, err := os.ReadFile(filepath.Join(dir, "fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	beforeManifest, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}

	payload = []byte(`[{"id":1},{"id":2},{"id":3},{"id":4},{"id":5},{"id":6}]`)
	if _, err := synchronize(context.Background(), dir, specs); err == nil {
		t.Fatal("expected a 40 percent quantity drop to be rejected")
	}
	afterData, err := os.ReadFile(filepath.Join(dir, "fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	afterManifest, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeData, afterData) || !bytes.Equal(beforeManifest, afterManifest) {
		t.Fatal("rejected update changed the current valid snapshot")
	}
}

func TestSynchronizeRejectsEncodedSchemaOrCountMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"one"}]`))
	}))
	defer server.Close()
	tests := []struct {
		name      string
		transform func([]byte) (any, int, error)
	}{
		{name: "not-array", transform: func([]byte) (any, int, error) { return map[string]string{"id": "one"}, 1, nil }},
		{name: "wrong-count", transform: func([]byte) (any, int, error) { return []map[string]string{{"id": "one"}}, 2, nil }},
		{name: "null-record", transform: func([]byte) (any, int, error) { return []any{nil}, 1, nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			specs := []sourceSpec{{name: test.name, file: "fixture.json", url: server.URL, minimum: 1, transform: test.transform}}
			if _, err := synchronize(context.Background(), dir, specs); err == nil {
				t.Fatal("expected invalid generated schema to fail")
			}
			if _, err := os.Stat(filepath.Join(dir, "manifest.json")); !os.IsNotExist(err) {
				t.Fatalf("invalid update wrote a manifest: %v", err)
			}
		})
	}
}

func TestValidateManifestChecksSHAAndMetadata(t *testing.T) {
	data := []byte("[]\n")
	staged := map[string]stagedFile{"fixture.json": {data: data, count: 0, source: "fixture"}}
	hash := sha256.Sum256(data)
	candidate := manifest{
		Schema: schemaVersion, GeneratedAt: time.Now().UTC(),
		Files: map[string]fileManifest{"fixture.json": {SHA256: hex.EncodeToString(hash[:]), Count: 0, Source: "fixture"}},
	}
	encoded, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateManifest(encoded, staged); err != nil {
		t.Fatal(err)
	}
	candidate.Files["fixture.json"] = fileManifest{SHA256: "bad", Count: 0, Source: "fixture"}
	encoded, err = json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateManifest(encoded, staged); err == nil {
		t.Fatal("expected invalid SHA-256 to fail")
	}
}

func TestCommittedDataMatchesManifest(t *testing.T) {
	dir := filepath.Join("..", "snapshot")
	manifestData, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var current manifest
	if err := json.Unmarshal(manifestData, &current); err != nil {
		t.Fatal(err)
	}
	specs := defaultSourceSpecs()
	if len(current.Files) != len(specs) {
		t.Fatalf("manifest has %d files, want %d", len(current.Files), len(specs))
	}
	staged := make(map[string]stagedFile, len(specs))
	for _, spec := range specs {
		meta, ok := current.Files[spec.file]
		if !ok {
			t.Fatalf("manifest is missing %s", spec.file)
		}
		if meta.Source != spec.url || meta.Count < spec.minimum {
			t.Fatalf("invalid metadata for %s: %+v", spec.file, meta)
		}
		data, err := os.ReadFile(filepath.Join(dir, spec.file))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateJSONArray(data, meta.Count); err != nil {
			t.Fatalf("validate %s: %v", spec.file, err)
		}
		if spec.validateOutput == nil {
			t.Fatalf("%s has no output schema validator", spec.file)
		}
		if err := spec.validateOutput(data); err != nil {
			t.Fatalf("validate output schema %s: %v", spec.file, err)
		}
		staged[spec.file] = stagedFile{data: data, count: meta.Count, source: meta.Source}
	}
	if err := validateManifest(manifestData, staged); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" || entry.Name() == "manifest.json" {
			continue
		}
		if _, ok := current.Files[entry.Name()]; !ok {
			t.Fatalf("data file %s is not tracked by manifest", entry.Name())
		}
	}
}

func TestDefaultOutputSchemasRejectEmptyFieldsAndTypeDrift(t *testing.T) {
	invalid := map[string][]byte{
		"tcp-targets.json":           []byte(`[{"id":"one","name":"","host":"one.example","port":443,"category":"global"}]`),
		"province-routes.json":       []byte(`[{"code":"BJ","name":"Beijing","province":"11","short":"BJ","targets":[]}]`),
		"speedtest-servers.json":     []byte(`[{"cc":"CN","city":"City","cityzh":"City","code":1,"country":"China","distance":0,"force_ping_select":0,"host":"speed.example:80","https_functional":1,"id":"one","lat":"0","lon":"0","name":"One","preferred":0,"provider":"Provider","providerzh":"Provider","sponsor":"Sponsor","status":"available","url":"http://speed.example/upload","unexpected":true}]`),
		"openspeedtest-servers.json": []byte(`[{"id":"one","host":"one.example","port_from":"5201","port_to":5210,"provider":"Provider","country":"US","city":"City","status":"available"}]`),
		"dnsbl-zones.json":           []byte(`[{"zone":"dnsbl.example","ipv6":true}]`),
		"bgp-asn-map.json":           []byte(`[{"asn":64500,"name":""}]`),
		"media-providers.json":       []byte(`[{"id":"one","name":"One","unexpected":"field"}]`),
	}
	for _, spec := range defaultSourceSpecs() {
		t.Run(spec.file, func(t *testing.T) {
			if spec.validateOutput == nil {
				t.Fatal("missing output validator")
			}
			if err := spec.validateOutput(invalid[spec.file]); err == nil {
				t.Fatalf("invalid schema for %s was accepted", spec.file)
			}
		})
	}
}

func TestSynchronizeDoesNotOverwriteOnOutputSchemaDrift(t *testing.T) {
	payload := []byte(`[{"id":"one","name":"One","host":"one.example","port":443,"category":"global"}]`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	transform := func(data []byte) (any, int, error) {
		var records []tcpTarget
		if err := json.Unmarshal(data, &records); err != nil {
			return nil, 0, err
		}
		return records, len(records), nil
	}
	dir := t.TempDir()
	specs := []sourceSpec{{
		name: "fixture", file: "tcp-targets.json", url: server.URL, minimum: 1,
		transform: transform, validateOutput: validateTCPTargetSchema,
	}}
	if _, err := synchronize(context.Background(), dir, specs); err != nil {
		t.Fatal(err)
	}
	beforeData, err := os.ReadFile(filepath.Join(dir, "tcp-targets.json"))
	if err != nil {
		t.Fatal(err)
	}
	beforeManifest, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	payload = []byte(`[{"id":"one","name":"","host":"one.example","port":443,"category":"global"}]`)
	changed, err := synchronize(context.Background(), dir, specs)
	if err != nil || changed {
		t.Fatalf("schema drift should reuse the current valid snapshot: changed=%v err=%v", changed, err)
	}
	afterData, _ := os.ReadFile(filepath.Join(dir, "tcp-targets.json"))
	afterManifest, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if !bytes.Equal(beforeData, afterData) || !bytes.Equal(beforeManifest, afterManifest) {
		t.Fatal("schema-rejected update changed the current valid snapshot")
	}
}

func TestSynchronizeUpdatesHealthySourceWhileReusingFailedSource(t *testing.T) {
	payloads := map[string][]byte{
		"/healthy": []byte(`[{"id":"one","name":"One","host":"one.example","port":443,"category":"global"}]`),
		"/failed":  []byte(`[{"id":"two","name":"Two","host":"two.example","port":443,"category":"global"}]`),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) { _, _ = w.Write(payloads[request.URL.Path]) }))
	defer server.Close()
	transform := func(data []byte) (any, int, error) {
		var records []tcpTarget
		if err := json.Unmarshal(data, &records); err != nil {
			return nil, 0, err
		}
		return records, len(records), nil
	}
	specs := []sourceSpec{
		{name: "healthy", file: "healthy.json", url: server.URL + "/healthy", minimum: 1, transform: transform, validateOutput: validateTCPTargetSchema},
		{name: "failed", file: "failed.json", url: server.URL + "/failed", minimum: 1, transform: transform, validateOutput: validateTCPTargetSchema},
	}
	dir := t.TempDir()
	if changed, err := synchronize(context.Background(), dir, specs); err != nil || !changed {
		t.Fatalf("initial sync failed: changed=%v err=%v", changed, err)
	}
	failedBefore, _ := os.ReadFile(filepath.Join(dir, "failed.json"))
	payloads["/healthy"] = []byte(`[{"id":"updated","name":"Updated","host":"updated.example","port":443,"category":"global"}]`)
	payloads["/failed"] = []byte(`[{"id":"two","name":"","host":"two.example","port":443,"category":"global"}]`)
	if changed, err := synchronize(context.Background(), dir, specs); err != nil || !changed {
		t.Fatalf("partial source refresh failed: changed=%v err=%v", changed, err)
	}
	failedAfter, _ := os.ReadFile(filepath.Join(dir, "failed.json"))
	healthyAfter, _ := os.ReadFile(filepath.Join(dir, "healthy.json"))
	if !bytes.Equal(failedBefore, failedAfter) || !bytes.Contains(healthyAfter, []byte(`"updated"`)) {
		t.Fatalf("failed source was replaced or healthy source was not updated")
	}
}

func TestParseDNSBLDeduplicates(t *testing.T) {
	value, count, err := parseDNSBL([]byte("b.example.\na.example\nb.example\ninvalid value\n-bad.example\n"))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d", count)
	}
	zones := value.([]dnsblZone)
	if zones[0].Zone != "a.example" {
		t.Fatalf("zones not sorted: %#v", zones)
	}
	if zones[1].Zone != "b.example" {
		t.Fatalf("trailing-dot zone was not normalized: %#v", zones)
	}
}

func TestParseDNSBLRejectsConcatenatedAndSingleLabelEntries(t *testing.T) {
	value, count, err := parseDNSBL([]byte("dul.dnsbl.sorbs.net\ndul.ru\ndul.dnsbl.sorbs.netdul.ru\npaidaccessviarsync\n"))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d, zones=%#v", count, value)
	}
	for _, zone := range value.([]dnsblZone) {
		if zone.Zone == "dul.dnsbl.sorbs.netdul.ru" || zone.Zone == "paidaccessviarsync" {
			t.Fatalf("malformed zone survived: %#v", zone)
		}
	}
}

func TestParseDNSBLDoesNotTreatAllV6AsDualStack(t *testing.T) {
	value, count, err := parseDNSBL([]byte("all.v6.example\nregular.example\n"))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d", count)
	}
	zones := value.([]dnsblZone)
	if zones[0].Zone != "all.v6.example" || zones[0].IPv4 || !zones[0].IPv6 {
		t.Fatalf("unexpected IPv6 capability: %#v", zones[0])
	}
	if !zones[1].IPv4 || zones[1].IPv6 {
		t.Fatalf("general zone must not claim IPv6 support: %#v", zones[1])
	}
}

func TestProbeDNSBLZonesKeepsOnlyReachableZones(t *testing.T) {
	original := lookupDNSBLZone
	lookupDNSBLZone = func(_ context.Context, zone string) error {
		if zone == "dead.example" {
			return fmt.Errorf("%w: NXDOMAIN", errDNSBLZoneInvalid)
		}
		return nil
	}
	defer func() { lookupDNSBLZone = original }()
	value, err := probeDNSBLZones(context.Background(), []dnsblZone{{Zone: "dead.example"}, {Zone: "live.example"}})
	if err != nil {
		t.Fatal(err)
	}
	zones := value.([]dnsblZone)
	if len(zones) != 1 || zones[0].Zone != "live.example" {
		t.Fatalf("unexpected health-filtered zones: %#v", zones)
	}
}

func TestProbeDNSBLZonesAbortsOnTransientFailure(t *testing.T) {
	original := lookupDNSBLZone
	lookupDNSBLZone = func(context.Context, string) error { return errors.New("temporary DoH failure") }
	defer func() { lookupDNSBLZone = original }()
	if _, err := probeDNSBLZones(context.Background(), []dnsblZone{{Zone: "live.example"}}); err == nil {
		t.Fatal("expected transient health failure to abort the update")
	}
}

func TestPassJSONArrayRejectsMalformedData(t *testing.T) {
	if _, _, err := passJSONArray([]byte(`{"broken":`)); err == nil {
		t.Fatal("expected malformed JSON to fail")
	}
	value, count, err := passJSONArray([]byte(`[{"ok":true}]`))
	if err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	if _, err := json.Marshal(value); err != nil {
		t.Fatal(err)
	}
}

func TestParseProvinceRoutesBuildsCarrierDualStackTargets(t *testing.T) {
	value, count, err := parseProvinceRoutes([]byte(`[{"code":"BJ","name":"北京市","province":11,"short":"京"},{"code":"TW","name":"台湾省","province":71,"short":"台"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count=%d", count)
	}
	routes := value.([]provinceRoute)
	if len(routes[0].Targets) != 3 || routes[0].Targets[0].IPv4 != "bj-ct-v4.ip.zstaticcdn.com" || routes[0].Targets[2].IPv6 != "bj-cm-v6.ip.zstaticcdn.com" {
		t.Fatalf("unexpected routes: %+v", routes)
	}
}

func TestParseProvinceRoutesDeduplicatesCodesAndProvinceNumbers(t *testing.T) {
	value, count, err := parseProvinceRoutes([]byte(`[
		{"code":"BJ","name":"北京市","province":11},
		{"code":"bj","name":"Duplicate code","province":12},
		{"code":"SH","name":"Duplicate province","province":11},
		{"code":"SH","name":"上海市","province":31}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count=%d", count)
	}
	if len(value.([]provinceRoute)) != 2 {
		t.Fatalf("unexpected routes: %#v", value)
	}
}

func TestParseSpeedtestServersDeduplicatesAndSorts(t *testing.T) {
	value, count, err := parseSpeedtestServers([]byte(`[
		{"id":"2","host":"two.example:8080","url":"http://two.example/upload"},
		{"id":"1","host":"one.example:8080","url":"http://one.example/upload"},
		{"id":"2","host":"duplicate.example:8080","url":"http://duplicate.example/upload"},
		{"id":"3","host":"","url":"http://invalid.example/upload"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	servers := value.([]map[string]any)
	if count != 2 || servers[0]["id"] != "1" || servers[1]["id"] != "2" {
		t.Fatalf("unexpected servers: %+v", servers)
	}
}

func TestParseTransferTargetsDeduplicatesAndValidatesPorts(t *testing.T) {
	value, count, err := parseTransferTargets([]byte(`[
		{"code":1,"server":"one.example","portl":5201,"portu":5210,"city":"Tokyo"},
		{"code":1,"server":"duplicate.example","portl":5201,"portu":5210,"city":"Tokyo"},
		{"code":2,"server":"bad.example","portl":0,"portu":5210,"city":"Osaka"},
		{"code":3,"server":"three.example","portl":5201,"portu":65536,"city":"Seoul"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count=%d value=%#v", count, value)
	}
	targets := value.([]transferTarget)
	if targets[0].ID != "1-tokyo" || targets[0].Host != "one.example" {
		t.Fatalf("unexpected targets: %#v", targets)
	}
}

func TestParseProviderNamesIgnoresCommentsAndUsesAST(t *testing.T) {
	value, count, err := parseProviderNames([]byte(`package providers
	type TestItem struct{}
	var GlobeTests = []TestItem{
		{"Enabled TV", EnabledFunc, true},
		{"Heading", nil, true},
		// {"Disabled TV", nil},
	}
	`))
	if err != nil {
		t.Fatal(err)
	}
	provider := value.([]providerName)[0]
	if count != 1 || provider.Name != "Enabled TV" || !provider.SupportsIPv6 || len(provider.Groups) != 1 || provider.Groups[0] != "global" {
		t.Fatalf("unexpected providers: %#v", value)
	}
	if _, _, err := parseProviderNames([]byte("package providers\nvar broken =")); err == nil {
		t.Fatal("expected malformed Go source to fail")
	}
}

func TestValidateAvailabilityDropsProtectsHealthyBaseline(t *testing.T) {
	dir := t.TempDir()
	previous := make([]map[string]string, 15)
	for index := range previous {
		status := "unavailable"
		if index < 8 {
			status = "available"
		}
		previous[index] = map[string]string{"status": status}
	}
	encodedPrevious, err := json.Marshal(previous)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "speedtest-servers.json"), encodedPrevious, 0o644); err != nil {
		t.Fatal(err)
	}
	candidate := make([]map[string]string, 15)
	for index := range candidate {
		status := "unavailable"
		if index < 2 {
			status = "available"
		}
		candidate[index] = map[string]string{"status": status}
	}
	encodedCandidate, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	staged := map[string]stagedFile{"speedtest-servers.json": {data: encodedCandidate}}
	if err := validateAvailabilityDrops(dir, staged, 0.35); err == nil {
		t.Fatal("expected severe availability drop to be rejected")
	}
	for index := 2; index < 15; index++ {
		candidate[index]["status"] = "available"
	}
	encodedCandidate, err = json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	staged["speedtest-servers.json"] = stagedFile{data: encodedCandidate}
	if err := validateAvailabilityDrops(dir, staged, 0.35); err != nil {
		t.Fatalf("moderate availability change was rejected: %v", err)
	}
}

func TestCommitSnapshotRollsBackWriteFailure(t *testing.T) {
	dir := t.TempDir()
	paths := map[string][]byte{
		"a.json":        []byte("old-a\n"),
		"obsolete.json": []byte("old-obsolete\n"),
		"manifest.json": mustEncodeManifest(t, "a.json", "obsolete.json"),
	}
	for name, data := range paths {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	staged := map[string]stagedFile{
		"a.json": {data: []byte("new-a\n")},
		"b.json": {data: []byte("new-b\n")},
	}
	writer := func(path string, data []byte) error {
		if filepath.Base(path) == "manifest.json" {
			return errors.New("injected write failure")
		}
		return atomicWrite(path, data)
	}
	if err := commitSnapshot(dir, staged, []byte("new-manifest\n"), writer); err == nil {
		t.Fatal("expected injected write failure")
	}
	for name, want := range paths {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%s was not rolled back: got %q want %q", name, got, want)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "b.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("new file was not removed during rollback: %v", err)
	}
}

func TestCommitSnapshotRollsBackRemoveFailure(t *testing.T) {
	dir := t.TempDir()
	paths := map[string][]byte{
		"a.json":          []byte("old-a\n"),
		"obsolete-a.json": []byte("old-obsolete-a\n"),
		"obsolete-b.json": []byte("old-obsolete-b\n"),
		"manifest.json":   mustEncodeManifest(t, "a.json", "obsolete-a.json", "obsolete-b.json"),
	}
	for name, data := range paths {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	staged := map[string]stagedFile{"a.json": {data: []byte("new-a\n")}}
	remover := func(path string) error {
		if filepath.Base(path) == "obsolete-b.json" {
			return errors.New("injected remove failure")
		}
		return os.Remove(path)
	}
	if err := commitSnapshotWithRemove(dir, staged, []byte("new-manifest\n"), nil, remover); err == nil {
		t.Fatal("expected injected remove failure")
	}
	for name, want := range paths {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%s was not rolled back: got %q want %q", name, got, want)
		}
	}
}

func mustEncodeManifest(t *testing.T, names ...string) []byte {
	t.Helper()
	files := make(map[string]fileManifest, len(names))
	for _, name := range names {
		files[name] = fileManifest{}
	}
	encoded, err := json.Marshal(manifest{Schema: schemaVersion, GeneratedAt: time.Now().UTC(), Files: files})
	if err != nil {
		t.Fatal(err)
	}
	return append(encoded, '\n')
}
