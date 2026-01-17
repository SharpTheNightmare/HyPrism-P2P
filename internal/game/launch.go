package game

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"HyPrism/internal/env"
)

// Launch launches the game with the given player name
// Uses a shell script that exactly matches TEMPLATE.sh since that's proven to work
func Launch(playerName string, version string) error {
	baseDir := env.GetDefaultAppDir()
	
	// The actual game files are in release/package/game/{version}
	actualGameDir := filepath.Join(baseDir, "release", "package", "game", version)
	
	// Create symlink: baseDir/game -> release/package/game/latest
	gameDir := filepath.Join(baseDir, "game")
	
	// Remove existing symlink/folder and recreate
	os.Remove(gameDir)
	os.RemoveAll(gameDir)
	
	// Create symlink
	if err := os.Symlink(actualGameDir, gameDir); err != nil {
		fmt.Printf("Warning: Could not create symlink: %v\n", err)
		gameDir = actualGameDir
	}

	// Verify client exists
	var clientPath string
	switch runtime.GOOS {
	case "darwin":
		clientPath = filepath.Join(gameDir, "Client", "Hytale.app", "Contents", "MacOS", "HytaleClient")
	case "windows":
		clientPath = filepath.Join(gameDir, "Client", "HytaleClient.exe")
	default:
		clientPath = filepath.Join(gameDir, "Client", "HytaleClient")
	}

	if _, err := os.Stat(clientPath); err != nil {
		return fmt.Errorf("game client not found at %s: %w", clientPath, err)
	}

	// Create UserData directory
	userDataDir := filepath.Join(baseDir, "UserData")
	_ = os.MkdirAll(userDataDir, 0755)

	// Set up Java path based on OS
	var jrePath string
	jreDir := filepath.Join(baseDir, "jre")
	
	switch runtime.GOOS {
	case "darwin":
		// macOS: Create symlink structure matching working installation
		// Working installation has: java/Contents/Home/bin/java
		javaDir := filepath.Join(baseDir, "java")
		javaHomeBin := filepath.Join(javaDir, "Contents", "Home", "bin")
		
		if _, err := os.Stat(javaHomeBin); err != nil {
			os.RemoveAll(javaDir)
			os.MkdirAll(filepath.Join(javaDir, "Contents", "Home"), 0755)
			os.Symlink(filepath.Join(jreDir, "bin"), filepath.Join(javaDir, "Contents", "Home", "bin"))
			os.Symlink(filepath.Join(jreDir, "lib"), filepath.Join(javaDir, "Contents", "Home", "lib"))
		}
		jrePath = filepath.Join(baseDir, "java", "Contents", "Home", "bin", "java")
	case "windows":
		// Windows: Use direct path to java.exe in jre directory
		jrePath = filepath.Join(jreDir, "bin", "java.exe")
	default:
		// Linux: Use direct path to java in jre directory
		jrePath = filepath.Join(jreDir, "bin", "java")
	}

	// Verify Java exists
	if _, err := os.Stat(jrePath); err != nil {
		return fmt.Errorf("Java not found at %s: %w", jrePath, err)
	}

	// Create and run a shell script - this is PROVEN to work
	scriptPath := filepath.Join(baseDir, "launch.sh")
	
	var scriptContent string
	if runtime.GOOS == "darwin" {
		scriptContent = fmt.Sprintf(`#!/bin/bash
"%s" \
    --app-dir "%s" \
    --user-dir "%s" \
    --java-exec "%s" \
    --auth-mode offline \
    --uuid 00000000-1337-1337-1337-000000000000 \
    --name "%s"
`, clientPath, gameDir, userDataDir, jrePath, playerName)
	} else if runtime.GOOS == "windows" {
		scriptContent = fmt.Sprintf(`@echo off
"%s" ^
    --app-dir "%s" ^
    --user-dir "%s" ^
    --java-exec "%s" ^
    --auth-mode offline ^
    --uuid 00000000-1337-1337-1337-000000000000 ^
    --name "%s"
`, clientPath, gameDir, userDataDir, jrePath, playerName)
		scriptPath = filepath.Join(baseDir, "launch.bat")
	} else {
		// Linux
		clientDir := filepath.Join(gameDir, "Client")
		scriptContent = fmt.Sprintf(`#!/bin/bash
export LD_LIBRARY_PATH="%s:$LD_LIBRARY_PATH"
"%s" \
    --app-dir "%s" \
    --user-dir "%s" \
    --java-exec "%s" \
    --auth-mode offline \
    --uuid 00000000-1337-1337-1337-000000000000 \
    --name "%s"
`, clientDir, clientPath, gameDir, userDataDir, jrePath, playerName)
	}

	// Write the script
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create launch script: %w", err)
	}

	fmt.Printf("=== LAUNCH DEBUG ===\n")
	fmt.Printf("Script path: %s\n", scriptPath)
	fmt.Printf("Base dir: %s\n", baseDir)
	fmt.Printf("Player: %s\n", playerName)
	fmt.Printf("==================\n")

	// On macOS, use 'open' command to launch the .app bundle properly
	// This ensures the app launches with the correct environment and frameworks
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// Get the .app bundle path
		appBundlePath := filepath.Join(gameDir, "Client", "Hytale.app")
		
		// Use open command with arguments
		cmd = exec.Command("open", appBundlePath, 
			"--args",
			"--app-dir", gameDir,
			"--user-dir", userDataDir,
			"--java-exec", jrePath,
			"--auth-mode", "offline",
			"--uuid", "00000000-1337-1337-1337-000000000000",
			"--name", playerName,
		)
	} else if runtime.GOOS == "windows" {
		// Windows: Launch the executable directly without cmd wrapper
		cmd = exec.Command(clientPath,
			"--app-dir", gameDir,
			"--user-dir", userDataDir,
			"--java-exec", jrePath,
			"--auth-mode", "offline",
			"--uuid", "00000000-1337-1337-1337-000000000000",
			"--name", playerName,
		)
		// Detach the process so it runs independently
		cmd.SysProcAttr = getWindowsSysProcAttr()
	} else {
		// Linux: Launch directly with LD_LIBRARY_PATH set
		clientDir := filepath.Join(gameDir, "Client")
		cmd = exec.Command(clientPath,
			"--app-dir", gameDir,
			"--user-dir", userDataDir,
			"--java-exec", jrePath,
			"--auth-mode", "offline",
			"--uuid", "00000000-1337-1337-1337-000000000000",
			"--name", playerName,
		)
		// Set LD_LIBRARY_PATH for Linux
		cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s:%s", clientDir, os.Getenv("LD_LIBRARY_PATH")))
	}
	
	cmd.Dir = baseDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Set up environment - inherit current environment
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start game: %w", err)
	}

	// Store the process for later termination
	gameProcess = cmd.Process
	
	// Let the process run independently
	go func() {
		cmd.Wait()
		gameProcess = nil
	}()

	return nil
}

var gameProcess *os.Process

// KillGame terminates the running game process
func KillGame() error {
	if gameProcess == nil {
		return fmt.Errorf("no game process running")
	}
	
	err := gameProcess.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill game process: %w", err)
	}
	
	gameProcess = nil
	fmt.Println("Game process terminated")
	return nil
}

// IsGameRunning checks if the game process is still running
func IsGameRunning() bool {
	if gameProcess == nil {
		return false
	}
	
	// Check if process is still alive by sending signal 0
	process, err := os.FindProcess(gameProcess.Pid)
	if err != nil {
		gameProcess = nil
		return false
	}
	
	// On Unix, Signal(0) checks if process exists
	err = process.Signal(os.Signal(nil))
	if err != nil {
		gameProcess = nil
		return false
	}
	
	return true
}

// WaitForGameExit waits for the game to exit and returns
func WaitForGameExit() {
	if gameProcess == nil {
		return
	}
	
	// Wait for the process to exit
	gameProcess.Wait()
	gameProcess = nil
}

// GetGameLogs returns the game log file content
func GetGameLogs() (string, error) {
	baseDir := env.GetDefaultAppDir()
	
	// Try multiple log paths based on typical Hytale log locations
	paths := []string{
		// UserData logs
		filepath.Join(baseDir, "UserData", "logs", "latest.log"),
		filepath.Join(baseDir, "UserData", "logs", "game.log"),
		filepath.Join(baseDir, "UserData", "logs", "client.log"),
		// Game directory logs
		filepath.Join(baseDir, "release", "package", "game", "latest", "logs", "latest.log"),
		filepath.Join(baseDir, "release", "package", "game", "latest", "logs", "game.log"),
		filepath.Join(baseDir, "release", "package", "game", "latest", "Client", "logs", "latest.log"),
		// HyPrism specific log
		filepath.Join(baseDir, "logs", "game.log"),
	}
	
	var allLogs strings.Builder
	foundAny := false
	
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil && len(data) > 0 {
			foundAny = true
			allLogs.WriteString(fmt.Sprintf("=== %s ===\n", filepath.Base(p)))
			
			content := string(data)
			// Return last 30KB of each log
			if len(content) > 30*1024 {
				content = content[len(content)-30*1024:]
			}
			allLogs.WriteString(content)
			allLogs.WriteString("\n\n")
		}
	}
	
	if foundAny {
		return allLogs.String(), nil
	}
	
	// List what directories exist to help debug
	var debug strings.Builder
	debug.WriteString("No game logs found. Checking directories:\n\n")
	
	checkDirs := []string{
		filepath.Join(baseDir, "UserData"),
		filepath.Join(baseDir, "UserData", "logs"),
		filepath.Join(baseDir, "release", "package", "game", "latest"),
		filepath.Join(baseDir, "release", "package", "game", "latest", "logs"),
	}
	
	for _, dir := range checkDirs {
		if _, err := os.Stat(dir); err == nil {
			debug.WriteString(fmt.Sprintf("✓ %s exists\n", dir))
			// List files in the directory
			if entries, err := os.ReadDir(dir); err == nil {
				for _, e := range entries {
					debug.WriteString(fmt.Sprintf("   - %s\n", e.Name()))
				}
			}
		} else {
			debug.WriteString(fmt.Sprintf("✗ %s not found\n", dir))
		}
	}
	
	return debug.String(), nil
}

// UUID represents a UUID
type UUID [16]byte

// OfflineUUID generates an offline UUID from a player name
func OfflineUUID(name string) UUID {
	data := "OfflinePlayer:" + name
	hash := md5.Sum([]byte(data))
	
	// Set version to 3 (name-based)
	hash[6] = (hash[6] & 0x0f) | 0x30
	// Set variant to 1
	hash[8] = (hash[8] & 0x3f) | 0x80
	
	return UUID(hash)
}

// String returns the UUID as a string
func (u UUID) String() string {
	hex := hex.EncodeToString(u[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex[0:8], hex[8:12], hex[12:16], hex[16:20], hex[20:32])
}

// setSDLVideoDriver sets the SDL_VIDEODRIVER environment variable for Linux
func setSDLVideoDriver(cmd *exec.Cmd) {
	if runtime.GOOS != "linux" {
		return
	}

	// Check if running under Wayland
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	xdgSession := os.Getenv("XDG_SESSION_TYPE")

	if waylandDisplay != "" || strings.ToLower(xdgSession) == "wayland" {
		cmd.Env = append(os.Environ(), "SDL_VIDEODRIVER=wayland,x11")
	}
}
