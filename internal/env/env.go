package env

import (
	"os"
	"path/filepath"
	"runtime"
)

// IsFlatpak returns true if running inside a Flatpak sandbox
func IsFlatpak() bool {
	// Flatpak sets FLATPAK_ID environment variable
	if os.Getenv("FLATPAK_ID") != "" {
		return true
	}
	// Also check for /.flatpak-info file which exists in Flatpak sandboxes
	if _, err := os.Stat("/.flatpak-info"); err == nil {
		return true
	}
	return false
}

// GetDefaultAppDir returns the default application directory
// Uses APPDATA (roaming) to store only config and session files
func GetDefaultAppDir() string {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("USERPROFILE")
		}
	case "darwin":
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, "Library", "Application Support")
	default:
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".config")
	}

	return filepath.Join(baseDir, "HyPrism")
}

// CreateFolders creates the required folder structure
// Only creates the base HyPrism directory for config and session storage
func CreateFolders() error {
	appDir := GetDefaultAppDir()
	return os.MkdirAll(appDir, 0755)
}

// Legacy stubs - no longer used but kept for compatibility
func GetCacheDir() string { return filepath.Join(GetDefaultAppDir(), "cache") }
func GetJREDir() string { return filepath.Join(GetDefaultAppDir(), "jre") }
func GetInstanceDir(branch string, version int) string { return "" }
func GetInstanceGameDir(branch string, version int) string { return "" }
func GetInstanceUserDataDir(branch string, version int) string { return "" }

