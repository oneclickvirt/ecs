package api

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFinalizeRunResultWritesPrivateFilesAndUploadsAllowedRun(t *testing.T) {
	temp := t.TempDir()
	textPath := filepath.Join(temp, "result.txt")
	jsonPath := filepath.Join(temp, "result.json")
	config := NewDefaultConfig()
	config.FilePath = textPath
	config.JSONPath = jsonPath
	config.EnableUpload = true
	result := &RunResult{
		Output: "\x1b[31mresult\x1b[0m\n", JSON: []byte(`{"schema_version":"goecs.report/v1"}`),
		Report: &StructuredReport{Status: ReportStatusOK},
	}
	originalUpload := uploadTextContext
	t.Cleanup(func() { uploadTextContext = originalUpload })
	uploadCalls := 0
	uploadTextContext = func(ctx context.Context, path string) (string, string, error) {
		uploadCalls++
		if ctx.Err() != nil || path != textPath {
			t.Fatalf("unexpected upload input: path=%q err=%v", path, ctx.Err())
		}
		return "http://example.test/result", "https://example.test/result", nil
	}
	finalized, err := FinalizeRunResultContext(context.Background(), NetCheckResult{Connected: true}, config, result)
	if err != nil {
		t.Fatal(err)
	}
	if uploadCalls != 1 || finalized.TextPath != textPath || finalized.JSONPath != jsonPath || finalized.HTTPSURL == "" {
		t.Fatalf("unexpected finalize result: %#v calls=%d", finalized, uploadCalls)
	}
	text, err := os.ReadFile(textPath)
	if err != nil || string(text) != "result\n" {
		t.Fatalf("unexpected text file: %q err=%v", text, err)
	}
	info, err := os.Stat(textPath)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected text permissions: %v err=%v", info.Mode().Perm(), err)
	}
}

func TestFinalizeRunResultNeverUploadsPrivateOrCanceledRun(t *testing.T) {
	for _, test := range []struct {
		name    string
		privacy bool
		status  ReportStatus
	}{
		{name: "privacy", privacy: true, status: ReportStatusOK},
		{name: "canceled report", status: ReportStatusCanceled},
		{name: "timeout report", status: ReportStatusTimeout},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := NewDefaultConfig()
			config.FilePath = filepath.Join(t.TempDir(), "result.txt")
			config.EnableUpload = true
			config.PrivacyMode = test.privacy
			originalUpload := uploadTextContext
			t.Cleanup(func() { uploadTextContext = originalUpload })
			uploadTextContext = func(context.Context, string) (string, string, error) {
				t.Fatal("upload must not be called")
				return "", "", nil
			}
			_, _ = FinalizeRunResultContext(context.Background(), NetCheckResult{Connected: true}, config, &RunResult{
				Output: "result", Report: &StructuredReport{Status: test.status},
			})
		})
	}
}

func TestFinalizeRunResultHonorsCanceledContextBeforeWrites(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	path := filepath.Join(t.TempDir(), "result.txt")
	config := NewDefaultConfig()
	config.FilePath = path
	_, err := FinalizeRunResultContext(ctx, NetCheckResult{Connected: true}, config, &RunResult{Output: "result"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected finalize error: %v", err)
	}
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("canceled finalize created a file: %v", statErr)
	}
}
