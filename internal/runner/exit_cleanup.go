package runner

import "sync"

// exitCleanup is the last-chance boundary for process-local artifacts (for
// example, the component log) when a path uses os.Exit and therefore skips
// ordinary deferred cleanup. The callback is deliberately process-global:
// runner exits terminate the process, so there is no useful per-run lifetime.
var exitCleanup struct {
	sync.Mutex
	fn   func()
	done bool
}

// SetExitCleanup installs a callback that should run before an immediate
// runner exit. Passing nil clears the callback and resets its one-shot state.
func SetExitCleanup(fn func()) {
	exitCleanup.Lock()
	exitCleanup.fn = fn
	exitCleanup.done = false
	exitCleanup.Unlock()
}

// runExitCleanup executes the currently installed callback at most once. A
// cleanup failure must never prevent the process from terminating.
func runExitCleanup() {
	exitCleanup.Lock()
	if exitCleanup.done || exitCleanup.fn == nil {
		exitCleanup.Unlock()
		return
	}
	fn := exitCleanup.fn
	exitCleanup.done = true
	exitCleanup.Unlock()

	func() {
		defer func() { _ = recover() }()
		fn()
	}()
}
