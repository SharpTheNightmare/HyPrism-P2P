package game

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"HyPrism/internal/auth"
)

// LaunchInstance launches using official Hytale installation
func LaunchInstance(gameInstallPath string) error {
	// Require authentication - no offline mode
	session, err := auth.GetValidSession()
	if err != nil || session == nil {
		return fmt.Errorf("authentication required: please log in with your Hytale account")
	}
	
	fmt.Printf("Launching with authenticated account: %s (UUID: %s)\n", session.Username, session.UUID)
	
	// Use official Hytale installation paths
	gameDir := filepath.Join(gameInstallPath, "install", "release", "package", "game", "latest")
	
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

// UserData under official installation by default
userDataDir := filepath.Join(gameInstallPath, "UserData")
_ = os.MkdirAll(userDataDir, 0755)

// Set up Java path from official installation
var jrePath string
jreRoot := filepath.Join(gameInstallPath, "install", "release", "package", "jre", "latest")

switch runtime.GOOS {
case "darwin":
	jrePath = filepath.Join(jreRoot, "bin", "java")
case "windows":
	jrePath = filepath.Join(jreRoot, "bin", "java.exe")
default:
	jrePath = filepath.Join(jreRoot, "bin", "java")
}

if _, err := os.Stat(jrePath); err != nil {
	return fmt.Errorf("Java not found at %s: %w", jrePath, err)
}

fmt.Printf("=== LAUNCH ===\n")
fmt.Printf("Game dir: %s\n", gameDir)
fmt.Printf("UserData: %s\n", userDataDir)
fmt.Printf("JRE: %s\n", jrePath)
fmt.Printf("================\n")

// Prepare auth args
playerName := session.Username
uuid := session.UUID
sessionToken := session.SessionToken
identityToken := session.IdentityToken

authMode := "authenticated"

var cmd *exec.Cmd
if runtime.GOOS == "darwin" {
	appBundlePath := filepath.Join(gameDir, "Client", "Hytale.app")
	args := []string{
		"--args",
		"--app-dir", gameDir,
		"--user-dir", userDataDir,
		"--java-exec", jrePath,
		"--auth-mode", authMode,
		"--uuid", uuid,
	}
	args = append(args, "--name", playerName, "--session-token", sessionToken, "--identity-token", identityToken)
	cmd = exec.Command("open", appBundlePath)
	cmd.Args = append(cmd.Args, args...)
} else if runtime.GOOS == "windows" {
	// Windows - launch directly without special working directory
	args := []string{
		"--app-dir", gameDir,
		"--user-dir", userDataDir,
		"--java-exec", jrePath,
		"--auth-mode", authMode,
		"--uuid", uuid,
	}
	args = append(args, "--name", playerName, "--session-token", sessionToken, "--identity-token", identityToken)
	cmd = exec.Command(clientPath, args...)
	cmd.SysProcAttr = getWindowsSysProcAttr()
} else {
	// Linux - must set LD_LIBRARY_PATH to find SDL3_image and other native libraries
	clientDir := filepath.Join(gameDir, "Client")
	args := []string{
		"--app-dir", gameDir,
		"--user-dir", userDataDir,
		"--java-exec", jrePath,
		"--auth-mode", authMode,
		"--uuid", uuid,
	}
	if authMode == "authenticated" && sessionToken != "" {
		args = append(args, "--name", playerName, "--session-token", sessionToken, "--identity-token", identityToken)
	} else {
		args = append(args, "--name", playerName)
	}
	cmd = exec.Command(clientPath, args...)
	// Preserve LD_LIBRARY_PATH with Client directory first
	existingLdPath := os.Getenv("LD_LIBRARY_PATH")
	newLdPath := clientDir
	if existingLdPath != "" {
		newLdPath = fmt.Sprintf("%s:%s", clientDir, existingLdPath)
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", newLdPath))
	}
	
cmd.Dir = gameInstallPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start game: %w", err)
	}

	gameProcess = cmd.Process
	gameRunning = true
	
	go func() {
		cmd.Wait()
		gameProcess = nil
		gameRunning = false
	}()

	return nil
}

var gameProcess *os.Process
var gameRunning bool

// KillGame terminates the running game process
func KillGame() error {
	if !gameRunning {
		return fmt.Errorf("no game process running")
	}
	
	// Try to kill by process reference first
	if gameProcess != nil {
		err := gameProcess.Kill()
		if err == nil {
			gameProcess = nil
			gameRunning = false
			fmt.Println("Game process terminated")
			return nil
		}
	}
	
	// If that fails, try to find and kill by name
	if runtime.GOOS == "darwin" {
		exec.Command("pkill", "-f", "Hytale").Run()
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/IM", "HytaleClient.exe")
		cmd.SysProcAttr = getWindowsSysProcAttr()
		cmd.Run()
	} else {
		exec.Command("pkill", "-f", "HytaleClient").Run()
	}
	
	gameProcess = nil
	gameRunning = false
	fmt.Println("Game process terminated")
	return nil
}

// IsGameRunning checks if the game process is still running
func IsGameRunning() bool {
	// Active check for the game process by name
	var isRunning bool
	
	if runtime.GOOS == "darwin" {
		// Check for Hytale process on macOS
		out, err := exec.Command("pgrep", "-f", "Hytale").Output()
		isRunning = err == nil && len(out) > 0
	} else if runtime.GOOS == "windows" {
		// Check for HytaleClient.exe on Windows using WMI (no window flash)
		isRunning = isWindowsProcessRunning("HytaleClient.exe")
	} else {
		// Linux
		out, err := exec.Command("pgrep", "-f", "HytaleClient").Output()
		isRunning = err == nil && len(out) > 0
	}
	
	// Update global state
	gameRunning = isRunning
	if !isRunning {
		gameProcess = nil
	}
	
	return isRunning
}

// WaitForGameExit waits for the game to exit and returns
func WaitForGameExit() {
	if !gameRunning {
		return
	}
	
	// Poll until the game is no longer running
	for IsGameRunning() {
		// Sleep briefly to avoid busy-waiting
		// The IsGameRunning function already checks the actual process
	}
	
	gameProcess = nil
	gameRunning = false
}

// GetGameLogs returns the game log file content
func GetGameLogs() (string, error) {
	// Game logs are stored in the official Hytale installation's UserData folder
	return "Game logs are stored in your official Hytale installation's UserData folder.\nPlease check your Hytale installation directory for logs.", nil
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
