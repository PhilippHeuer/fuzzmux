package util

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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

// IsAncestorProcess reports whether pid is the current process or one of its parent processes (e.g. the terminal tmx runs in)
func IsAncestorProcess(pid int) bool {
	current := os.Getpid()
	for i := 0; i < 64 && current > 1; i++ {
		if current == pid {
			return true
		}
		ppid, err := parentPID(current)
		if err != nil {
			return false
		}
		current = ppid
	}
	return false
}

// parentPID reads the parent pid of the given process from /proc
func parentPID(pid int) (int, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, err
	}

	// format: pid (comm) state ppid ..., comm may contain spaces and parentheses
	s := string(data)
	idx := strings.LastIndexByte(s, ')')
	if idx < 0 {
		return 0, fmt.Errorf("unexpected format of /proc/%d/stat", pid)
	}
	fields := strings.Fields(s[idx+1:])
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected format of /proc/%d/stat", pid)
	}

	return strconv.Atoi(fields[1])
}
