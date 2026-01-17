//go:build windows

package game

import "syscall"

// getWindowsSysProcAttr returns Windows-specific process attributes
func getWindowsSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
