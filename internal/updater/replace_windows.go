//go:build windows

package updater

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

const updateHelperArgument = "--goecs-update-helper"

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	openProcess              = kernel32.NewProc("OpenProcess")
	waitForSingleObject      = kernel32.NewProc("WaitForSingleObject")
	closeHandle              = kernel32.NewProc("CloseHandle")
	moveFileEx               = kernel32.NewProc("MoveFileExW")
	processSynchronizeAccess = uintptr(0x00100000)
	moveFileReplaceExisting  = uintptr(0x00000001)
)

// replaceExecutable copies the current binary to a temporary helper image.
// A running Windows executable cannot be replaced, so that helper waits for
// this process to exit and then performs the replacement from a different
// executable path.
func replaceExecutable(target, candidate string, mode os.FileMode) (bool, error) {
	helper, err := copyHelperBinary(target, mode)
	if err != nil {
		return false, err
	}
	command := exec.Command(helper, updateHelperArgument, candidate, target, strconv.Itoa(os.Getpid()))
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := command.Start(); err != nil {
		_ = os.Remove(helper)
		return false, fmt.Errorf("start Windows update helper: %w", err)
	}
	return true, nil
}

func copyHelperBinary(target string, mode os.FileMode) (string, error) {
	source, err := os.Open(target)
	if err != nil {
		return "", err
	}
	defer source.Close()
	helper, err := os.CreateTemp(filepath.Dir(target), ".goecs-update-helper-*.exe")
	if err != nil {
		return "", err
	}
	helperPath := helper.Name()
	succeeded := false
	defer func() {
		_ = helper.Close()
		if !succeeded {
			_ = os.Remove(helperPath)
		}
	}()
	if _, err := io.Copy(helper, source); err != nil {
		return "", err
	}
	if err := helper.Chmod(mode); err != nil {
		return "", err
	}
	if err := helper.Sync(); err != nil {
		return "", err
	}
	if err := helper.Close(); err != nil {
		return "", err
	}
	succeeded = true
	return helperPath, nil
}

// HandleHelper is called before normal CLI flag parsing. It deliberately does
// not print to the parent's terminal; a successful next invocation exposes the
// new version, while a failed helper leaves the original executable untouched.
func HandleHelper(arguments []string) (bool, error) {
	if len(arguments) == 0 || arguments[0] != updateHelperArgument {
		return false, nil
	}
	if len(arguments) != 4 {
		return true, errors.New("invalid Windows update helper arguments")
	}
	source, target := arguments[1], arguments[2]
	parentPID, err := strconv.Atoi(arguments[3])
	if err != nil || parentPID <= 0 {
		return true, errors.New("invalid Windows update helper parent process")
	}
	if err := waitForProcessExit(uint32(parentPID)); err != nil {
		return true, err
	}
	for attempt := 0; attempt < 60; attempt++ {
		if err := replaceFile(source, target); err == nil {
			return true, nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return true, errors.New("Windows update helper could not replace GoECS")
}

func waitForProcessExit(pid uint32) error {
	handle, _, _ := openProcess.Call(processSynchronizeAccess, 0, uintptr(pid))
	if handle == 0 {
		// The parent may already have exited before the helper starts.
		return nil
	}
	defer closeHandle.Call(handle)
	const infinite = uintptr(0xffffffff)
	status, _, waitErr := waitForSingleObject.Call(handle, infinite)
	if status != 0 {
		return fmt.Errorf("wait for GoECS exit: %v", waitErr)
	}
	return nil
}

func replaceFile(source, target string) error {
	sourcePointer, err := syscall.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	targetPointer, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	ok, _, callErr := moveFileEx.Call(uintptr(unsafe.Pointer(sourcePointer)), uintptr(unsafe.Pointer(targetPointer)), moveFileReplaceExisting)
	if ok == 0 {
		return callErr
	}
	return nil
}
