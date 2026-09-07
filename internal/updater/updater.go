// Package updater implements the self-update path used by the interactive
// GoECS menu. It intentionally depends only on the Go standard library so a
// release binary can repair itself even when local module caches are absent.
package updater

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultReleaseAPI = "https://api.github.com/repos/oneclickvirt/ecs/releases/latest"
	checksumAssetName = "checksums.txt"
	maxAssetBytes     = 512 << 20
)

var (
	stableVersionPattern = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	checksumLinePattern  = regexp.MustCompile(`^([A-Fa-f0-9]{64})[ \t]+\*?([^ \t]+)[ \t]*$`)
)

// Status is the outcome class callers can render without matching errors.
type Status string

const (
	StatusUpToDate Status = "up_to_date"
	StatusUpdated  Status = "updated"
)

// Result describes a completed update check. Deferred is true only on
// Windows, where the replacement is completed by a short-lived helper after
// the current executable exits.
type Result struct {
	Status         Status
	CurrentVersion string
	LatestVersion  string
	Asset          string
	Deferred       bool
}

// Options exposes the small set of runtime seams needed by the updater and
// its local tests. Zero values select the official repository and host values.
type Options struct {
	CurrentVersion string
	APIURL         string
	Executable     string
	GOOS           string
	GOARCH         string
	HTTPClient     *http.Client
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type release struct {
	TagName    string         `json:"tag_name"`
	Draft      bool           `json:"draft"`
	Prerelease bool           `json:"prerelease"`
	Assets     []releaseAsset `json:"assets"`
}

type semanticVersion struct {
	major uint64
	minor uint64
	patch uint64
}

func (v semanticVersion) compare(other semanticVersion) int {
	for _, pair := range [][2]uint64{{v.major, other.major}, {v.minor, other.minor}, {v.patch, other.patch}} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	return 0
}

// Update downloads the matching release asset, verifies it against the
// release checksum, validates the ZIP layout, and atomically replaces the
// resolved executable. It never uses a shell or embeds credentials in URLs.
func Update(ctx context.Context, options Options) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	current, err := parseStableVersion(options.CurrentVersion)
	if err != nil {
		return Result{}, fmt.Errorf("invalid current GoECS version: %w", err)
	}
	options = normalizeOptions(options)

	target, info, err := resolveExecutable(options.Executable)
	if err != nil {
		return Result{}, err
	}
	if !info.Mode().IsRegular() {
		return Result{}, errors.New("current GoECS executable is not a regular file")
	}

	lock, err := createUpdateLock(target)
	if err != nil {
		return Result{}, err
	}
	defer func() {
		_ = lock.Close()
		_ = os.Remove(lock.Name())
	}()

	latestRelease, err := fetchRelease(ctx, options.HTTPClient, options.APIURL)
	if err != nil {
		return Result{}, err
	}
	if latestRelease.Draft || latestRelease.Prerelease {
		return Result{}, errors.New("latest GoECS release is not a stable release")
	}
	latest, err := parseStableVersion(latestRelease.TagName)
	if err != nil {
		return Result{}, fmt.Errorf("latest GoECS release has an invalid version: %w", err)
	}
	result := Result{CurrentVersion: normalizeVersion(options.CurrentVersion), LatestVersion: normalizeVersion(latestRelease.TagName)}
	if current.compare(latest) >= 0 {
		result.Status = StatusUpToDate
		return result, nil
	}

	assetName := fmt.Sprintf("goecs_%s_%s.zip", options.GOOS, options.GOARCH)
	asset, ok := findAsset(latestRelease.Assets, assetName)
	if !ok || strings.TrimSpace(asset.BrowserDownloadURL) == "" {
		return Result{}, fmt.Errorf("release %s has no asset for %s/%s", normalizeVersion(latestRelease.TagName), options.GOOS, options.GOARCH)
	}
	checksums, ok := findAsset(latestRelease.Assets, checksumAssetName)
	if !ok || strings.TrimSpace(checksums.BrowserDownloadURL) == "" {
		return Result{}, fmt.Errorf("release %s has no %s", normalizeVersion(latestRelease.TagName), checksumAssetName)
	}

	checksumBody, err := downloadBytes(ctx, options.HTTPClient, checksums.BrowserDownloadURL, 4<<20)
	if err != nil {
		return Result{}, fmt.Errorf("download release checksums: %w", err)
	}
	expectedDigest, err := checksumForAsset(string(checksumBody), assetName)
	if err != nil {
		return Result{}, err
	}

	temporaryDir, err := os.MkdirTemp("", "goecs-update-")
	if err != nil {
		return Result{}, fmt.Errorf("create update workspace: %w", err)
	}
	defer os.RemoveAll(temporaryDir)
	archivePath := filepath.Join(temporaryDir, assetName)
	if err := downloadFile(ctx, options.HTTPClient, asset.BrowserDownloadURL, archivePath, maxAssetBytes); err != nil {
		return Result{}, fmt.Errorf("download GoECS update: %w", err)
	}
	if err := verifySHA256(archivePath, expectedDigest); err != nil {
		return Result{}, err
	}

	candidate, err := extractExecutable(archivePath, filepath.Dir(target), filepath.Base(target), info.Mode().Perm())
	if err != nil {
		return Result{}, err
	}
	replaced := false
	defer func() {
		if !replaced {
			_ = os.Remove(candidate)
		}
	}()
	deferred, err := replaceExecutable(target, candidate, info.Mode().Perm())
	if err != nil {
		return Result{}, err
	}
	replaced = true
	result.Status = StatusUpdated
	result.Asset = assetName
	result.Deferred = deferred
	return result, nil
}

func normalizeOptions(options Options) Options {
	if strings.TrimSpace(options.APIURL) == "" {
		options.APIURL = defaultReleaseAPI
	}
	if strings.TrimSpace(options.GOOS) == "" {
		options.GOOS = runtime.GOOS
	}
	if strings.TrimSpace(options.GOARCH) == "" {
		options.GOARCH = runtime.GOARCH
	}
	if options.HTTPClient == nil {
		options.HTTPClient = &http.Client{Timeout: 45 * time.Second}
	}
	return options
}

func normalizeVersion(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "v") {
		return value
	}
	return "v" + value
}

func parseStableVersion(value string) (semanticVersion, error) {
	matches := stableVersionPattern.FindStringSubmatch(strings.TrimSpace(value))
	if matches == nil {
		return semanticVersion{}, fmt.Errorf("%q is not a stable semantic version", value)
	}
	parts := [3]uint64{}
	for index := range parts {
		parsed, err := strconv.ParseUint(matches[index+1], 10, 64)
		if err != nil {
			return semanticVersion{}, err
		}
		parts[index] = parsed
	}
	return semanticVersion{major: parts[0], minor: parts[1], patch: parts[2]}, nil
}

func resolveExecutable(executable string) (string, os.FileInfo, error) {
	if strings.TrimSpace(executable) == "" {
		resolved, err := os.Executable()
		if err != nil {
			return "", nil, fmt.Errorf("locate current GoECS executable: %w", err)
		}
		executable = resolved
	}
	target, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", nil, fmt.Errorf("resolve current GoECS executable: %w", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		return "", nil, fmt.Errorf("inspect current GoECS executable: %w", err)
	}
	return target, info, nil
}

func createUpdateLock(target string) (*os.File, error) {
	lockPath := target + ".update.lock"
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, errors.New("another GoECS update is already running")
		}
		return nil, fmt.Errorf("create update lock: %w", err)
	}
	return lock, nil
}

func fetchRelease(ctx context.Context, client *http.Client, endpoint string) (release, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return release{}, fmt.Errorf("create GitHub release request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "goecs-self-updater")
	response, err := client.Do(request)
	if err != nil {
		return release{}, fmt.Errorf("request latest GoECS release: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return release{}, fmt.Errorf("latest GoECS release request returned HTTP %d", response.StatusCode)
	}
	var value release
	decoder := json.NewDecoder(io.LimitReader(response.Body, 8<<20))
	if err := decoder.Decode(&value); err != nil {
		return release{}, fmt.Errorf("decode latest GoECS release: %w", err)
	}
	return value, nil
}

func findAsset(assets []releaseAsset, name string) (releaseAsset, bool) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset, true
		}
	}
	return releaseAsset{}, false
}

func downloadBytes(ctx context.Context, client *http.Client, location string, limit int64) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "goecs-self-updater")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned HTTP %d", response.StatusCode)
	}
	reader := io.LimitReader(response.Body, limit+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("download exceeds allowed size")
	}
	return data, nil
}

func downloadFile(ctx context.Context, client *http.Client, location, destination string, limit int64) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "goecs-self-updater")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", response.StatusCode)
	}
	if response.ContentLength > limit {
		return errors.New("download exceeds allowed size")
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	succeeded := false
	defer func() {
		_ = file.Close()
		if !succeeded {
			_ = os.Remove(destination)
		}
	}()
	written, err := io.Copy(file, io.LimitReader(response.Body, limit+1))
	if err != nil {
		return err
	}
	if written > limit {
		return errors.New("download exceeds allowed size")
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	succeeded = true
	return nil
}

func checksumForAsset(contents, assetName string) ([]byte, error) {
	var digest []byte
	for _, line := range strings.Split(contents, "\n") {
		matches := checksumLinePattern.FindStringSubmatch(strings.TrimSpace(line))
		if matches == nil || matches[2] != assetName {
			continue
		}
		if digest != nil {
			return nil, fmt.Errorf("checksums contain multiple entries for %s", assetName)
		}
		parsed, err := hex.DecodeString(matches[1])
		if err != nil || len(parsed) != sha256.Size {
			return nil, fmt.Errorf("invalid SHA-256 entry for %s", assetName)
		}
		digest = parsed
	}
	if digest == nil {
		return nil, fmt.Errorf("checksums do not contain %s", assetName)
	}
	return digest, nil
}

func verifySHA256(fileName string, expected []byte) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	if !equalBytes(hash.Sum(nil), expected) {
		return errors.New("downloaded GoECS archive failed SHA-256 verification")
	}
	return nil
}

func equalBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	var difference byte
	for index := range left {
		difference |= left[index] ^ right[index]
	}
	return difference == 0
}

func extractExecutable(archivePath, destinationDir, executableName string, mode os.FileMode) (string, error) {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("open GoECS update archive: %w", err)
	}
	defer archive.Close()

	archiveExecutableName := "goecs"
	if strings.HasSuffix(strings.ToLower(executableName), ".exe") {
		archiveExecutableName = "goecs.exe"
	}
	var executable *zip.File
	for _, entry := range archive.File {
		if err := validateArchiveName(entry.Name); err != nil {
			return "", err
		}
		if entry.FileInfo().IsDir() {
			continue
		}
		if entry.Mode()&os.ModeType != 0 {
			return "", fmt.Errorf("update archive contains a non-regular file: %s", entry.Name)
		}
		if filepath.Base(entry.Name) != archiveExecutableName || normalizeArchiveName(entry.Name) != archiveExecutableName {
			continue
		}
		if executable != nil {
			return "", errors.New("update archive contains multiple GoECS executables")
		}
		if entry.UncompressedSize64 > maxAssetBytes {
			return "", errors.New("GoECS executable in archive exceeds allowed size")
		}
		executable = entry
	}
	if executable == nil {
		return "", errors.New("update archive does not contain the GoECS executable")
	}

	reader, err := executable.Open()
	if err != nil {
		return "", err
	}
	defer reader.Close()
	temporary, err := os.CreateTemp(destinationDir, "."+executableName+".update-")
	if err != nil {
		return "", fmt.Errorf("create replacement executable: %w", err)
	}
	candidate := temporary.Name()
	succeeded := false
	defer func() {
		_ = temporary.Close()
		if !succeeded {
			_ = os.Remove(candidate)
		}
	}()
	if mode == 0 {
		mode = 0o755
	}
	if err := temporary.Chmod(mode); err != nil {
		return "", err
	}
	written, err := io.Copy(temporary, io.LimitReader(reader, maxAssetBytes+1))
	if err != nil {
		return "", err
	}
	if written > maxAssetBytes {
		return "", errors.New("GoECS executable in archive exceeds allowed size")
	}
	if err := temporary.Sync(); err != nil {
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	succeeded = true
	return candidate, nil
}

func normalizeArchiveName(value string) string {
	return strings.TrimPrefix(strings.ReplaceAll(value, "\\", "/"), "./")
}

func validateArchiveName(value string) error {
	value = normalizeArchiveName(value)
	if value == "" || strings.HasPrefix(value, "/") || path.IsAbs(value) {
		return fmt.Errorf("update archive contains an unsafe path: %q", value)
	}
	cleaned := path.Clean(value)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || cleaned != value {
		return fmt.Errorf("update archive contains an unsafe path: %q", value)
	}
	return nil
}
