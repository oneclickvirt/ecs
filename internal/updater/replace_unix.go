//go:build !windows

package updater

import (
	"os"
)

// replaceExecutable is atomic when the candidate and target share a
// directory. Unix kernels permit replacement of a running executable; the
// already-running image remains valid until this process exits.
func replaceExecutable(target, candidate string, mode os.FileMode) (bool, error) {
	if err := os.Chmod(candidate, mode); err != nil {
		return false, err
	}
	if err := os.Rename(candidate, target); err != nil {
		return false, err
	}
	return false, nil
}

// HandleHelper is present on every platform so main can dispatch before flag
// parsing. Only Windows needs an out-of-process replacement helper.
func HandleHelper([]string) (bool, error) { return false, nil }
