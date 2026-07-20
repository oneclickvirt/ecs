//go:build windows

package runner

import "os"

// forceExit on Windows simply calls os.Exit; there is no process-group kill
// mechanism equivalent to POSIX SIGKILL on this platform.
func forceExit(code int) {
	os.Exit(code)
}

// ForceExit terminates the process at the global hard deadline.
func ForceExit(code int) { forceExit(code) }
