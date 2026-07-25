package utils

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestValidatePasteUploadSize(t *testing.T) {
	for _, test := range []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{name: "empty", size: 0},
		{name: "exact limit", size: MaxPasteUploadBytes},
		{name: "one byte over", size: MaxPasteUploadBytes + 1, wantErr: true},
		{name: "negative", size: -1, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validatePasteUploadSize(test.size)
			if (err != nil) != test.wantErr {
				t.Fatalf("validatePasteUploadSize(%d) error = %v, wantErr=%t", test.size, err, test.wantErr)
			}
			if test.wantErr && !strings.Contains(err.Error(), "80 KiB") {
				t.Fatalf("limit error = %v", err)
			}
		})
	}
}

func TestUploadTextContextHonorsCancellationBeforeFileAccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := UploadTextContext(ctx, "/path/that/must/not/be/read")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected upload error: %v", err)
	}
}
