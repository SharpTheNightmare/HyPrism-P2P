//go:build !windows

package game

import "syscall"

// getWindowsSysProcAttr returns nil on non-Windows platforms
func getWindowsSysProcAttr() *syscall.SysProcAttr {
	return nil
}
