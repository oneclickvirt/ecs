//go:build !ecs_public

package main

import (
	"testing"

	privatepst "github.com/oneclickvirt/privatespeedtest/pst"
)

func TestPrivateSpeedtestDependencyContract(t *testing.T) {
	if got := privatepst.PrivateSpeedTestVersion; got != "v0.0.14" {
		t.Fatalf("private speedtest component version = %q, want v0.0.14", got)
	}
}
