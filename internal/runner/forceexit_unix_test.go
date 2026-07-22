//go:build !windows

package runner

import (
	"context"
	"os"
	"os/exec"
	"reflect"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/oneclickvirt/ecs/internal/params"
)

func TestSignalAfterSoftDeadlineForceExits(t *testing.T) {
	const helperEnv = "GOECS_TEST_DEADLINE_SIGNAL_HELPER"
	if os.Getenv(helperEnv) == "1" {
		IsolateProcessGroup()
		if syscall.Getpgrp() != os.Getpid() {
			os.Exit(98)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		sig := make(chan os.Signal, 1)
		sig <- os.Interrupt
		config := params.NewConfig("test")
		start := time.Now()
		var output string
		var outputMutex sync.Mutex
		HandleSignalInterrupt(ctx, cancel, sig, config, &start, &output, "", make(chan bool, 1), &outputMutex)
		os.Exit(99)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSignalAfterSoftDeadlineForceExits$")
	cmd.Env = append(os.Environ(), helperEnv+"=1")
	started := time.Now()
	err := cmd.Run()
	if err == nil {
		t.Fatal("deadline signal helper returned without force-exiting")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("deadline signal helper returned unexpected error: %v", err)
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
		t.Fatalf("deadline signal helper was not killed: %v", err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("deadline signal helper took %s to exit", elapsed)
	}
}

func TestParseProcStatParentHandlesSpacesAndParentheses(t *testing.T) {
	parent, ok := parseProcStatParent([]byte("123 (benchmark worker) name) S 42 123 123 0"))
	if !ok || parent != 42 {
		t.Fatalf("parseProcStatParent() = %d, %t", parent, ok)
	}
}

func TestDescendantsFromParentsReturnsDeepestFirst(t *testing.T) {
	parents := map[int]int{11: 10, 12: 10, 21: 11, 22: 11, 31: 21, 99: 1}
	got := descendantsFromParents(10, parents)
	want := []int{31, 21, 22, 11, 12}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("descendantsFromParents() = %v, want %v", got, want)
	}
}

func TestIsTerminalRejectsRegularFile(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "terminal-check")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if isTerminal(int(file.Fd())) {
		t.Fatal("regular file detected as a terminal")
	}
}
