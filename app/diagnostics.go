package app

import (
	"HyPrism/internal/env"
	"HyPrism/internal/java"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// DiagnosticReport contains system diagnostic information
type DiagnosticReport struct {
	Platform      PlatformInfo      `json:"platform"`
	Connectivity  ConnectivityInfo  `json:"connectivity"`
	GameStatus    GameStatusInfo    `json:"gameStatus"`
	Dependencies  DependenciesInfo  `json:"dependencies"`
	Timestamp     string            `json:"timestamp"`
}

type PlatformInfo struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Version string `json:"version"`
}

type ConnectivityInfo struct {
	HytalePatches bool   `json:"hytalePatches"`
	GitHub        bool   `json:"github"`
	ItchIO        bool   `json:"itchIO"`
	Error         string `json:"error,omitempty"`
}

type GameStatusInfo struct {
	Installed       bool   `json:"installed"`
	Version         string `json:"version"`
	ClientExists    bool   `json:"clientExists"`
	OnlineFixApplied bool   `json:"onlineFixApplied"`
}

type DependenciesInfo struct {
	JavaInstalled bool   `json:"javaInstalled"`
	JavaPath      string `json:"javaPath"`
}

// RunDiagnostics runs system diagnostics
func (a *App) RunDiagnostics() DiagnosticReport {
	report := DiagnosticReport{
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Platform info
	report.Platform = PlatformInfo{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: AppVersion,
	}

	// Connectivity checks
	report.Connectivity = checkConnectivity()

	// Game status
	report.GameStatus = checkGameStatus()

	// Dependencies
	report.Dependencies = checkDependencies()

	return report
}

func checkConnectivity() ConnectivityInfo {
	info := ConnectivityInfo{}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Check Hytale patches
	resp, err := client.Head("https://game-patches.hytale.com")
	if err == nil && resp.StatusCode < 500 {
		info.HytalePatches = true
	}

	// Check GitHub
	resp, err = client.Head("https://api.github.com")
	if err == nil && resp.StatusCode < 500 {
		info.GitHub = true
	}


	// DNS check
	_, err = net.LookupHost("game-patches.hytale.com")
	if err != nil {
		info.Error = "DNS resolution failed: " + err.Error()
	}

	return info
}

func checkGameStatus() GameStatusInfo {
	info := GameStatusInfo{}
	// Game status not applicable for official installation launcher
	info.OnlineFixApplied = true
	return info
}

func checkDependencies() DependenciesInfo {
	info := DependenciesInfo{}

	// Check Java
	javaPath, err := java.GetJavaExec()
	if err == nil {
		info.JavaInstalled = true
		info.JavaPath = javaPath
	}

	return info
}

// SaveDiagnosticReport saves diagnostics to a file
func (a *App) SaveDiagnosticReport() (string, error) {
	report := a.RunDiagnostics()
	
	logsDir := filepath.Join(env.GetDefaultAppDir(), "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("hyprism_diagnostic_%s.txt", time.Now().Format("2006-01-02_15-04-05"))
	filepath := filepath.Join(logsDir, filename)

	content := fmt.Sprintf(`HyPrism Diagnostic Report
Generated: %s

=== PLATFORM ===
OS: %s
Arch: %s
Launcher Version: %s

=== CONNECTIVITY ===
Hytale Patches Server: %v
GitHub API: %v
itch.io (Butler): %v
Error: %s

=== GAME STATUS ===
Installed: %v
Version: %s
Client Exists: %v
Online Fix Applied: %v

=== DEPENDENCIES ===
Java Installed: %v
Java Path: %s
`,
		report.Timestamp,
		report.Platform.OS, report.Platform.Arch, report.Platform.Version,
		report.Connectivity.HytalePatches, report.Connectivity.GitHub, report.Connectivity.ItchIO, report.Connectivity.Error,
		report.GameStatus.Installed, report.GameStatus.Version, report.GameStatus.ClientExists, report.GameStatus.OnlineFixApplied,
		report.Dependencies.JavaInstalled, report.Dependencies.JavaPath,
	)

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return "", err
	}

	return filepath, nil
}

// CrashReport represents a crash report
type CrashReport struct {
	Filename  string `json:"filename"`
	Timestamp string `json:"timestamp"`
	Preview   string `json:"preview"`
}

// GetCrashReports returns available crash reports
func (a *App) GetCrashReports() ([]CrashReport, error) {
	// Crash reports not applicable for official installation launcher
	return []CrashReport{}, nil
}
