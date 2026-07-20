package data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVerify(t *testing.T) {
	if err := verify([]byte("ok"), "2689367b205c16ce8f6b8e5c7d5b4f2c5b2c3d0a9fbf0a3c2c0e0a5a7e5a8d9"); err == nil {
		t.Fatal("expected mismatch")
	}
}

func TestLoaderFallsBackToEmbeddedSnapshot(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	loader := NewLoader(server.Client(), server.URL)
	loader.RawBase = server.URL
	result, err := loader.Load(context.Background(), "tcp-targets.json")
	if err != nil {
		t.Fatal(err)
	}
	if result.Fallback != "embedded" || result.UsedRemote {
		t.Fatalf("unexpected fallback: %#v", result)
	}
}

func TestLoaderValidatesRemoteManifestHashAndCount(t *testing.T) {
	payload := []byte("[\n  {\"id\":\"one\",\"name\":\"One\",\"host\":\"one.example\",\"port\":443}\n]\n")
	hash := sha256.Sum256(payload)
	manifest := fmt.Sprintf(`{"schema":"ecs-data/v1","generated_at":"2026-07-19T00:00:00Z","files":{"tcp-targets.json":{"sha256":"%s","count":1}}}`, hex.EncodeToString(hash[:]))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			_, _ = w.Write([]byte(manifest))
		case "/tcp-targets.json":
			_, _ = w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	loader := NewLoader(server.Client(), server.URL)
	loader.RawBase = ""
	result, err := loader.Load(context.Background(), "tcp-targets.json")
	if err != nil {
		t.Fatal(err)
	}
	if !result.UsedRemote || result.Source != "cdn" || result.Fallback != "" {
		t.Fatalf("unexpected source metadata: %#v", result)
	}
}

func TestLoaderRejectsRemoteCountMismatch(t *testing.T) {
	payload := []byte("[]\n")
	hash := sha256.Sum256(payload)
	manifest := fmt.Sprintf(`{"schema":"ecs-data/v1","generated_at":"2026-07-19T00:00:00Z","files":{"tcp-targets.json":{"sha256":"%s","count":1}}}`, hex.EncodeToString(hash[:]))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manifest.json" {
			_, _ = w.Write([]byte(manifest))
			return
		}
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	loader := NewLoader(server.Client(), server.URL)
	loader.RawBase = ""
	result, err := loader.Load(context.Background(), "tcp-targets.json")
	if err != nil {
		t.Fatal(err)
	}
	if result.Source != "embedded" || result.Fallback != "embedded" {
		t.Fatalf("expected embedded fallback, got %#v", result)
	}
}

func TestLoaderSkipsCorruptCDNAndUsesRaw(t *testing.T) {
	payload := []byte("[{\"id\":\"raw\",\"name\":\"Raw\",\"host\":\"raw.example\",\"port\":443}]\n")
	hash := sha256.Sum256(payload)
	manifest := fmt.Sprintf(`{"schema":"ecs-data/v1","generated_at":"2026-07-19T00:00:00Z","files":{"tcp-targets.json":{"sha256":"%s","count":1}}}`, hex.EncodeToString(hash[:]))
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			_, _ = w.Write([]byte("not-json"))
		default:
			http.Error(w, "missing", http.StatusNotFound)
		}
	}))
	defer cdn.Close()
	raw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			_, _ = w.Write([]byte(manifest))
		case "/tcp-targets.json":
			_, _ = w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer raw.Close()
	loader := NewLoader(raw.Client(), cdn.URL)
	loader.RawBase = raw.URL
	result, err := loader.Load(context.Background(), "tcp-targets.json")
	if err != nil {
		t.Fatal(err)
	}
	if result.Source != "raw" || result.Fallback != "raw" || !result.UsedRemote {
		t.Fatalf("expected raw fallback after corrupt CDN manifest, got %#v", result)
	}
}

func TestLoaderSkipsStaleCDNDataAndUsesRaw(t *testing.T) {
	payload := []byte("[{\"id\":\"raw\",\"name\":\"Raw\",\"host\":\"raw.example\",\"port\":443}]\n")
	cdnPayload := []byte("[{\"id\":\"stale\",\"name\":\"Stale\",\"host\":\"stale.example\",\"port\":443}]\n")
	hash := sha256.Sum256(payload)
	manifest := fmt.Sprintf(`{"schema":"ecs-data/v1","generated_at":"2026-07-19T00:00:00Z","files":{"tcp-targets.json":{"sha256":"%s","count":1}}}`, hex.EncodeToString(hash[:]))
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			_, _ = w.Write([]byte(manifest))
		case "/tcp-targets.json":
			_, _ = w.Write(cdnPayload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer cdn.Close()
	raw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			_, _ = w.Write([]byte(manifest))
		case "/tcp-targets.json":
			_, _ = w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer raw.Close()
	loader := NewLoader(raw.Client(), cdn.URL)
	loader.RawBase = raw.URL
	result, err := loader.Load(context.Background(), "tcp-targets.json")
	if err != nil {
		t.Fatal(err)
	}
	if result.Source != "raw" || result.Fallback != "raw" || string(result.Data) != string(payload) {
		t.Fatalf("expected raw data after stale CDN data, got %#v", result)
	}
}

func TestLoaderUsesCDNPayloadAgainstCanonicalRawManifest(t *testing.T) {
	payload := []byte("[{\"id\":\"cdn\",\"name\":\"CDN\",\"host\":\"cdn.example\",\"port\":443}]\n")
	hash := sha256.Sum256(payload)
	manifest := fmt.Sprintf(`{"schema":"ecs-data/v1","generated_at":"2026-07-19T00:00:00Z","files":{"tcp-targets.json":{"sha256":"%s","count":1}}}`, hex.EncodeToString(hash[:]))
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tcp-targets.json" {
			_, _ = w.Write(payload)
			return
		}
		http.NotFound(w, r)
	}))
	defer cdn.Close()
	raw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manifest.json" {
			_, _ = w.Write([]byte(manifest))
			return
		}
		http.NotFound(w, r)
	}))
	defer raw.Close()
	loader := NewLoader(raw.Client(), cdn.URL)
	loader.RawBase = raw.URL
	result, err := loader.Load(context.Background(), "tcp-targets.json")
	if err != nil {
		t.Fatal(err)
	}
	if result.Source != "cdn" || result.Fallback != "" || string(result.Data) != string(payload) {
		t.Fatalf("canonical Raw manifest incorrectly marked CDN data as fallback: %#v", result)
	}
}

func TestLoaderReportsRawFallback(t *testing.T) {
	payload := []byte("[]\n")
	hash := sha256.Sum256(payload)
	manifest := fmt.Sprintf(`{"schema":"ecs-data/v1","generated_at":"2026-07-19T00:00:00Z","files":{"tcp-targets.json":{"sha256":"%s","count":0}}}`, hex.EncodeToString(hash[:]))
	cdn := httptest.NewServer(http.NotFoundHandler())
	defer cdn.Close()
	raw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manifest.json" {
			_, _ = w.Write([]byte(manifest))
			return
		}
		_, _ = w.Write(payload)
	}))
	defer raw.Close()
	loader := NewLoader(raw.Client(), cdn.URL)
	loader.RawBase = raw.URL
	result, err := loader.Load(context.Background(), "tcp-targets.json")
	if err != nil {
		t.Fatal(err)
	}
	if result.Source != "raw" || result.Fallback != "raw" {
		t.Fatalf("unexpected fallback: %#v", result)
	}
}

func TestVerifyCount(t *testing.T) {
	if err := verifyCount([]byte(`[1,2]`), 2); err != nil {
		t.Fatal(err)
	}
	if err := verifyCount([]byte(`[1]`), 2); err == nil {
		t.Fatal("expected count mismatch")
	}
	if err := verifyCount([]byte(`{}`), 0); err == nil {
		t.Fatal("expected non-array payload to fail")
	}
}

func TestVerifySchemaRejectsFieldDrift(t *testing.T) {
	if err := verifySchema("tcp-targets.json", []byte(`[{"id":"missing-fields"}]`)); err == nil {
		t.Fatal("expected TCP target schema drift to fail")
	}
	if err := verifySchema("media-providers.json", []byte(`[{"id":"one","name":"One","groups":["global"],"supports_ipv6":true}]`)); err != nil {
		t.Fatal(err)
	}
}

func TestAllEmbeddedFilesValidate(t *testing.T) {
	manifestData, err := embedded.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	loader := NewLoader(http.DefaultClient, "")
	results, err := loader.loadEmbeddedMany(KnownFiles())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != len(manifest.Files) {
		t.Fatalf("loaded %d embedded files, manifest has %d", len(results), len(manifest.Files))
	}
	for name, result := range results {
		if result.Source != "embedded" || result.Fallback != "embedded" {
			t.Fatalf("%s: unexpected result %#v", name, result)
		}
	}
}

func TestLoadManyRejectsPartialStaleCDNGeneration(t *testing.T) {
	oldTCP := []byte(`[{"id":"old","name":"Old","host":"old.example","port":443}]`)
	oldMedia := []byte(`[{"id":"old","name":"Old","groups":["global"],"supports_ipv6":false}]`)
	newTCP := []byte(`[{"id":"new","name":"New","host":"new.example","port":443}]`)
	newMedia := []byte(`[{"id":"new","name":"New","groups":["global"],"supports_ipv6":true}]`)
	manifestFor := func(generated string, tcp, media []byte) string {
		tcpHash := sha256.Sum256(tcp)
		mediaHash := sha256.Sum256(media)
		return fmt.Sprintf(`{"schema":"ecs-data/v1","generated_at":"%s","files":{"tcp-targets.json":{"sha256":"%s","count":1},"media-providers.json":{"sha256":"%s","count":1}}}`,
			generated, hex.EncodeToString(tcpHash[:]), hex.EncodeToString(mediaHash[:]))
	}
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			_, _ = w.Write([]byte(manifestFor("2026-07-18T00:00:00Z", oldTCP, oldMedia)))
		case "/tcp-targets.json":
			_, _ = w.Write(oldTCP)
		case "/media-providers.json":
			_, _ = w.Write([]byte(`[{"broken":true}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer cdn.Close()
	raw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			_, _ = w.Write([]byte(manifestFor("2026-07-19T00:00:00Z", newTCP, newMedia)))
		case "/tcp-targets.json":
			_, _ = w.Write(newTCP)
		case "/media-providers.json":
			_, _ = w.Write(newMedia)
		default:
			http.NotFound(w, r)
		}
	}))
	defer raw.Close()
	loader := NewLoader(raw.Client(), cdn.URL)
	loader.RawBase = raw.URL
	results, err := loader.LoadMany(context.Background(), []string{"tcp-targets.json", "media-providers.json"})
	if err != nil {
		t.Fatal(err)
	}
	for name, result := range results {
		if result.Source != "raw" || result.Fallback != "raw" || result.Manifest.GeneratedAt.Format(time.RFC3339) != "2026-07-19T00:00:00Z" {
			t.Fatalf("%s mixed a stale generation: %#v", name, result)
		}
	}
}

func TestLoadManyRejectsSelfConsistentStaleCDNGeneration(t *testing.T) {
	oldTCP := []byte(`[{"id":"old","name":"Old","host":"old.example","port":443}]`)
	newTCP := []byte(`[{"id":"new","name":"New","host":"new.example","port":443}]`)
	manifestFor := func(generated string, payload []byte) string {
		hash := sha256.Sum256(payload)
		return fmt.Sprintf(`{"schema":"ecs-data/v1","generated_at":"%s","files":{"tcp-targets.json":{"sha256":"%s","count":1}}}`,
			generated, hex.EncodeToString(hash[:]))
	}
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			_, _ = w.Write([]byte(manifestFor("2026-07-18T00:00:00Z", oldTCP)))
		case "/tcp-targets.json":
			_, _ = w.Write(oldTCP)
		default:
			http.NotFound(w, r)
		}
	}))
	defer cdn.Close()
	raw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			_, _ = w.Write([]byte(manifestFor("2026-07-19T00:00:00Z", newTCP)))
		case "/tcp-targets.json":
			_, _ = w.Write(newTCP)
		default:
			http.NotFound(w, r)
		}
	}))
	defer raw.Close()
	loader := NewLoader(raw.Client(), cdn.URL)
	loader.RawBase = raw.URL
	result, err := loader.Load(context.Background(), "tcp-targets.json")
	if err != nil {
		t.Fatal(err)
	}
	if result.Source != "raw" || result.Fallback != "raw" || string(result.Data) != string(newTCP) || result.Manifest.GeneratedAt.Format(time.RFC3339) != "2026-07-19T00:00:00Z" {
		t.Fatalf("accepted a self-consistent stale CDN generation: %#v", result)
	}
}

func TestLoadManyRejectsDuplicateNames(t *testing.T) {
	loader := NewLoader(http.DefaultClient, "")
	if _, err := loader.LoadMany(context.Background(), []string{"tcp-targets.json", "tcp-targets.json"}); err == nil {
		t.Fatal("expected duplicate file names to fail")
	}
}

func TestValidDataAddressRejectsInvalidPorts(t *testing.T) {
	for _, address := range []string{"server.example:notaport", "server.example:0", "server.example:65536"} {
		if validDataAddress(address) {
			t.Fatalf("invalid address %q was accepted", address)
		}
	}
	for _, address := range []string{"server.example:443", "[2001:db8::1]:8080"} {
		if !validDataAddress(address) {
			t.Fatalf("valid address %q was rejected", address)
		}
	}
}
