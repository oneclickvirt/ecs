package updater

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateReplacesOnlyVerifiedArchive(t *testing.T) {
	target := filepath.Join(t.TempDir(), "goecs")
	if err := os.WriteFile(target, []byte("old executable"), 0o751); err != nil {
		t.Fatal(err)
	}
	archive := testZIP(t, map[string]string{"goecs": "new executable"})
	digest := sha256.Sum256(archive)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/releases/latest":
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(writer, `{"tag_name":"v0.2.1","assets":[{"name":"goecs_linux_amd64.zip","browser_download_url":%q},{"name":"checksums.txt","browser_download_url":%q}]}`,
				serverURL(request)+"/goecs_linux_amd64.zip", serverURL(request)+"/checksums.txt")
		case "/checksums.txt":
			fmt.Fprintf(writer, "%s  goecs_linux_amd64.zip\n", hex.EncodeToString(digest[:]))
		case "/goecs_linux_amd64.zip":
			_, _ = writer.Write(archive)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	result, err := Update(context.Background(), Options{
		CurrentVersion: "v0.2.0", APIURL: server.URL + "/releases/latest", Executable: target,
		GOOS: "linux", GOARCH: "amd64", HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != StatusUpdated || result.LatestVersion != "v0.2.1" || result.Asset != "goecs_linux_amd64.zip" || result.Deferred {
		t.Fatalf("unexpected update result: %#v", result)
	}
	contents, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "new executable" {
		t.Fatalf("updated executable = %q", contents)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o751); got != want {
		t.Fatalf("replacement permissions = %o, want %o", got, want)
	}
}

func TestUpdateRejectsBadChecksumWithoutReplacingExecutable(t *testing.T) {
	target := filepath.Join(t.TempDir(), "goecs")
	if err := os.WriteFile(target, []byte("old executable"), 0o755); err != nil {
		t.Fatal(err)
	}
	archive := testZIP(t, map[string]string{"goecs": "new executable"})

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/releases/latest":
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(writer, `{"tag_name":"v0.2.1","assets":[{"name":"goecs_linux_amd64.zip","browser_download_url":%q},{"name":"checksums.txt","browser_download_url":%q}]}`,
				serverURL(request)+"/goecs_linux_amd64.zip", serverURL(request)+"/checksums.txt")
		case "/checksums.txt":
			fmt.Fprintln(writer, "0000000000000000000000000000000000000000000000000000000000000000  goecs_linux_amd64.zip")
		case "/goecs_linux_amd64.zip":
			_, _ = writer.Write(archive)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	_, err := Update(context.Background(), Options{
		CurrentVersion: "v0.2.0", APIURL: server.URL + "/releases/latest", Executable: target,
		GOOS: "linux", GOARCH: "amd64", HTTPClient: server.Client(),
	})
	if err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("bad checksum error = %v", err)
	}
	contents, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(contents) != "old executable" {
		t.Fatalf("bad checksum replaced executable: %q", contents)
	}
}

func TestUpdateReportsNoNewRelease(t *testing.T) {
	target := filepath.Join(t.TempDir(), "goecs")
	if err := os.WriteFile(target, []byte("current executable"), 0o755); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/releases/latest" {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"tag_name":"v0.2.0","assets":[]}`))
	}))
	defer server.Close()

	result, err := Update(context.Background(), Options{
		CurrentVersion: "v0.2.0", APIURL: server.URL + "/releases/latest", Executable: target,
		GOOS: "linux", GOARCH: "amd64", HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != StatusUpToDate || result.CurrentVersion != "v0.2.0" || result.LatestVersion != "v0.2.0" {
		t.Fatalf("unexpected up-to-date result: %#v", result)
	}
}

func TestExtractExecutableRejectsUnsafeArchivePath(t *testing.T) {
	directory := t.TempDir()
	archivePath := filepath.Join(directory, "unsafe.zip")
	if err := os.WriteFile(archivePath, testZIP(t, map[string]string{"../goecs": "not allowed"}), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := extractExecutable(archivePath, directory, "goecs", 0o755)
	if err == nil || !strings.Contains(err.Error(), "unsafe path") {
		t.Fatalf("unsafe archive error = %v", err)
	}
}

func TestParseStableVersionAndChecksums(t *testing.T) {
	for _, value := range []string{"v0.2.0", "1.2.3", "v12.34.56"} {
		if _, err := parseStableVersion(value); err != nil {
			t.Fatalf("parseStableVersion(%q): %v", value, err)
		}
	}
	for _, value := range []string{"v0.2", "v0.2.0-rc1", "latest", "v01.2.3"} {
		if _, err := parseStableVersion(value); err == nil {
			t.Fatalf("parseStableVersion(%q) unexpectedly succeeded", value)
		}
	}
	checksum := strings.Repeat("a", 64)
	digest, err := checksumForAsset(checksum+"  goecs_linux_amd64.zip\n", "goecs_linux_amd64.zip")
	if err != nil || hex.EncodeToString(digest) != checksum {
		t.Fatalf("checksum parse = %x, %v", digest, err)
	}
}

func testZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for name, contents := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func serverURL(request *http.Request) string {
	return "http://" + request.Host
}
