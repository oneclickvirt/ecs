package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/oneclickvirt/ecs/utils"
)

var ansiOutputPattern = regexp.MustCompile("\x1b\\[[0-9;]+[A-Za-z]")

var uploadTextContext = utils.UploadTextContext

type FinalizeResult struct {
	TextPath string `json:"text_path,omitempty"`
	JSONPath string `json:"json_path,omitempty"`
	HTTPURL  string `json:"http_url,omitempty"`
	HTTPSURL string `json:"https_url,omitempty"`
}

// FinalizeRunResultContext performs explicitly requested file and upload side
// effects after report collection. Upload is fail-closed for private,
// canceled, timed-out, erroneous, or offline runs.
func FinalizeRunResultContext(ctx context.Context, preCheck NetCheckResult, config *Config, result *RunResult) (FinalizeResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if config == nil || result == nil {
		return FinalizeResult{}, errors.New("config and run result are required")
	}
	var finalized FinalizeResult
	var finalErr error

	if config.JSONPath != "" && config.JSONPath != "-" && len(result.JSON) > 0 {
		if err := writeResultFile(ctx, config.JSONPath, append(append([]byte(nil), result.JSON...), '\n')); err != nil {
			finalErr = errors.Join(finalErr, fmt.Errorf("write JSON report: %w", err))
		} else {
			finalized.JSONPath = config.JSONPath
		}
	}

	textPath := strings.TrimSpace(config.FilePath)
	if textPath != "" && !config.PrivacyMode && result.Output != "" {
		cleaned := ansiOutputPattern.ReplaceAllString(result.Output, "")
		if err := writeResultFile(ctx, textPath, []byte(cleaned)); err != nil {
			finalErr = errors.Join(finalErr, fmt.Errorf("write text report: %w", err))
		} else {
			finalized.TextPath = textPath
		}
	}

	if !uploadAllowed(ctx, preCheck, config, result) {
		return finalized, finalErr
	}
	if finalized.TextPath == "" {
		finalErr = errors.Join(finalErr, errors.New("upload requested but no text report was written"))
		return finalized, finalErr
	}
	absolute, err := filepath.Abs(finalized.TextPath)
	if err != nil {
		return finalized, errors.Join(finalErr, fmt.Errorf("resolve upload path: %w", err))
	}
	progressStarted(ctx, "upload")
	httpURL, httpsURL, err := uploadTextContext(ctx, absolute)
	if err != nil {
		progressCompleted(ctx, "upload", ReportStatusError, err.Error())
		return finalized, errors.Join(finalErr, fmt.Errorf("upload result: %w", err))
	}
	finalized.HTTPURL, finalized.HTTPSURL = httpURL, httpsURL
	progressCompleted(ctx, "upload", ReportStatusOK, "")
	return finalized, finalErr
}

func writeResultFile(ctx context.Context, path string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func uploadAllowed(ctx context.Context, preCheck NetCheckResult, config *Config, result *RunResult) bool {
	if ctx.Err() != nil || !preCheck.Connected || !config.EnableUpload || config.PrivacyMode {
		return false
	}
	if result.Report == nil {
		return true
	}
	switch result.Report.Status {
	case ReportStatusTimeout, ReportStatusCanceled, ReportStatusError:
		return false
	default:
		return true
	}
}
