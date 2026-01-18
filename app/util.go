package app

import (
	"HyPrism/internal/env"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// OpenFolder opens the instances folder in the system file explorer
func (a *App) OpenFolder() error {
	path := env.GetInstancesDir()

	// Verify folder exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return FileSystemError("creating game folder", err)
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // Linux
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		return FileSystemError("opening folder", err)
	}

	return nil
}

// RepairInstallation cleans up corrupted/incomplete installation files for latest instance
// This is useful when butler fails with "Access Denied" errors
func (a *App) RepairInstallation() error {
	gameDir := env.GetInstanceGameDir("release", 0)
	
	// Clean staging directory
	stagingDir := filepath.Join(gameDir, "staging-temp")
	if err := os.RemoveAll(stagingDir); err != nil {
		// On Windows, try harder
		if runtime.GOOS == "windows" {
			filepath.Walk(stagingDir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					os.Remove(path)
				}
				return nil
			})
			os.RemoveAll(stagingDir)
		}
	}
	
	// Clean any .tmp files
	entries, _ := os.ReadDir(gameDir)
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".tmp") || strings.HasPrefix(name, "sf-") || name == ".installing" {
			os.Remove(filepath.Join(gameDir, name))
		}
	}
	
	// Also clean PWR downloads to force re-download
	cacheDir := env.GetCacheDir()
	cacheEntries, _ := os.ReadDir(cacheDir)
	for _, entry := range cacheEntries {
		if strings.HasSuffix(entry.Name(), ".pwr") {
			os.Remove(filepath.Join(cacheDir, entry.Name()))
		}
	}
	
	return nil
}

// DeleteGame deletes the game installation
func (a *App) DeleteGame() error {
	homeDir := env.GetDefaultAppDir()

	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return FileSystemError("reading game directory", err)
	}

	// Track deletion errors
	var deleteErrors []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(homeDir, entry.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				deleteErrors = append(deleteErrors, entry.Name())
			}
		}
	}

	if len(deleteErrors) > 0 {
		return GameError(
			fmt.Sprintf("Failed to delete some folders: %v", deleteErrors),
			nil,
		)
	}

	// Recreate folder structure
	if err := env.CreateFolders(); err != nil {
		return FileSystemError("recreating folder structure", err)
	}

	return nil
}

// GetPlatformInfo returns information about the current platform
func (a *App) GetPlatformInfo() map[string]string {
	return map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}
}
