package util

import (
	"fmt"
	"os"
	"syscall"
)

// KillProcessByPID kills a process with the given PID.
func KillProcessByPID(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error finding process with PID %d: %w", pid, err)
	}

	err = proc.Signal(syscall.SIGKILL)
	if err != nil {
		return fmt.Errorf("error killing process with PID %d: %w", pid, err)
	}

	return nil
}
