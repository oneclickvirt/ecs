package utils

import (
	"context"
	"errors"
	"testing"
)

func TestUploadTextContextHonorsCancellationBeforeFileAccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := UploadTextContext(ctx, "/path/that/must/not/be/read")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected upload error: %v", err)
	}
}
