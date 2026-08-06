package runner

import (
	"sync"

	"github.com/oneclickvirt/ecs/utils"
)

var runnerOutputSanitizer struct {
	sync.RWMutex
	fn func(string) string
}

// SetOutputSanitizer installs the public-output boundary used by live capture
// output as well as runner-owned files and uploads. Passing nil restores the
// identity behavior for embedders that do not configure a sanitizer.
func SetOutputSanitizer(fn func(string) string) {
	runnerOutputSanitizer.Lock()
	runnerOutputSanitizer.fn = fn
	runnerOutputSanitizer.Unlock()
	utils.SetOutputSanitizer(fn)
}

func sanitizeRunnerOutput(value string) (cleaned string) {
	runnerOutputSanitizer.RLock()
	fn := runnerOutputSanitizer.fn
	runnerOutputSanitizer.RUnlock()
	if fn == nil {
		return value
	}
	// A broken privacy boundary must fail closed rather than writing the
	// unsanitized input to a result file or remote endpoint.
	cleaned = "output unavailable"
	defer func() { _ = recover() }()
	return fn(value)
}
