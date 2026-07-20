//go:build !windows

package runner

import (
	"os"
	"syscall"
)

// forceExit kills the entire process group (so child subprocesses such as
// stream, fio, dd, sysbench, or geekbench are also terminated) and then calls
// os.Exit with the given code.
func forceExit(code int) {
	pgid, err := syscall.Getpgid(os.Getpid())
	if err == nil {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}
	os.Exit(code)
}

// ForceExit terminates this process and any benchmark subprocesses at the
// global hard deadline.
func ForceExit(code int) { forceExit(code) }
