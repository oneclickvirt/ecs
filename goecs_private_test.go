//go:build !ecs_public

package main

import (
	"os"
	"regexp"
	"testing"

	privatepst "github.com/oneclickvirt/privatespeedtest/pst"
)

func TestPrivateSpeedtestDependencyContract(t *testing.T) {
	if got := privatepst.PrivateSpeedTestVersion; got != pinnedPrivateSpeedtestVersion(t) {
		t.Fatalf("private speedtest component version = %q, want %q", got, pinnedPrivateSpeedtestVersion(t))
	}
}

func pinnedPrivateSpeedtestVersion(t *testing.T) string {
	t.Helper()
	contents, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	match := regexp.MustCompile(`(?m)^\s*github\.com/oneclickvirt/privatespeedtest\s+(\S+)`).FindSubmatch(contents)
	if len(match) != 2 {
		t.Fatal("private speedtest is not pinned")
	}
	return string(match[1])
}
