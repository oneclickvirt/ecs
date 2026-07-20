//go:build !windows

package runner

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// IsolateProcessGroup keeps hard-deadline cleanup from killing the invoking
// shell while ensuring ordinary benchmark children inherit our process group.
func IsolateProcessGroup() {
	if syscall.Getpgrp() != os.Getpid() {
		_ = syscall.Setpgid(0, 0)
	}
}

// forceExit kills both descendants and the process group. Some third-party
// benchmarks create their own process group, so group cleanup alone can leave
// a process holding an SSH stdout pipe open after goecs exits.
func forceExit(code int) {
	pid := os.Getpid()
	for _, child := range descendantPIDs(pid) {
		_ = syscall.Kill(child, syscall.SIGKILL)
	}
	pgid, err := syscall.Getpgid(os.Getpid())
	if err == nil {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}
	os.Exit(code)
}

func descendantPIDs(root int) []int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	parents := make(map[int]int, len(entries))
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		data, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "stat"))
		if err != nil {
			continue
		}
		parent, ok := parseProcStatParent(data)
		if ok {
			parents[pid] = parent
		}
	}
	return descendantsFromParents(root, parents)
}

func parseProcStatParent(data []byte) (int, bool) {
	closing := strings.LastIndex(string(data), ") ")
	if closing < 0 {
		return 0, false
	}
	fields := strings.Fields(string(data[closing+2:]))
	if len(fields) < 2 {
		return 0, false
	}
	parent, err := strconv.Atoi(fields[1])
	return parent, err == nil && parent >= 0
}

func descendantsFromParents(root int, parents map[int]int) []int {
	children := make(map[int][]int)
	for pid, parent := range parents {
		children[parent] = append(children[parent], pid)
	}
	for parent := range children {
		sort.Ints(children[parent])
	}
	result := make([]int, 0)
	var visit func(int)
	visit = func(parent int) {
		for _, child := range children[parent] {
			visit(child)
			result = append(result, child)
		}
	}
	visit(root)
	return result
}

// ForceExit terminates this process and any benchmark subprocesses at the
// global hard deadline.
func ForceExit(code int) { forceExit(code) }
