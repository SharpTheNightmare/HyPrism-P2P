package app

import (
	"runtime"
)


// GetPlatformInfo returns information about the current platform
func (a *App) GetPlatformInfo() map[string]string {
	return map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}
}
