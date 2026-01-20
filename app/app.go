package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"HyPrism/internal/auth"
	"HyPrism/internal/config"
	"HyPrism/internal/env"
	"HyPrism/internal/game"
	"HyPrism/internal/mods"
	"HyPrism/internal/news"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx         context.Context
	cfg         *config.Config
	newsService *news.NewsService
}

// ProgressUpdate represents download/install progress
type ProgressUpdate struct {
	Stage       string  `json:"stage"`
	Progress    float64 `json:"progress"`
	Message     string  `json:"message"`
	CurrentFile string  `json:"currentFile"`
	Speed       string  `json:"speed"`
	Downloaded  int64   `json:"downloaded"`
	Total       int64   `json:"total"`
}

// NewApp creates a new App instance
func NewApp() *App {
	cfg, _ := config.Load()
	if cfg == nil {
		cfg = config.Default()
	}
	return &App{
		cfg:         cfg,
		newsService: news.NewNewsService(),
	}
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           HyPrism - Hytale Launcher Starting...             ║")
	fmt.Printf("║           Version: %-43s║\n", AppVersion)
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")


	// Initialize environment
	if err := env.CreateFolders(); err != nil {
		fmt.Printf("Warning: Failed to create folders: %v\n", err)
	}

	// Check for launcher updates in background
	go func() {
		fmt.Println("Starting background update check...")
		a.checkUpdateSilently()
	}()
}

// Shutdown is called when the app closes
func (a *App) Shutdown(ctx context.Context) {
	fmt.Println("HyPrism shutting down...")
}

// SelectInstanceDirectory opens a folder picker dialog and saves the selected directory
func (a *App) SelectInstanceDirectory() (string, error) {
	// On Windows, start at "This PC" (empty string) to show all drives
	// This allows easy navigation to different drives
	defaultDir := ""
	if runtime.GOOS == "windows" {
		// Empty string on Windows opens at "This PC" level showing all drives
		defaultDir = ""
	}
	
	selectedDir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title:            "Select Instances Directory",
		DefaultDirectory: defaultDir,
	})
	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}
	
	if selectedDir == "" {
		// User cancelled the dialog
		return "", nil
	}
	
	// This function is no longer used since we use the official Hytale installation
	return "", fmt.Errorf("custom instance directories are not supported - use official Hytale installation")
}

// SelectGameInstallDirectory opens a folder picker to select official Hytale installation
func (a *App) SelectGameInstallDirectory() (string, error) {
	defaultDir := ""
	if runtime.GOOS == "windows" {
		defaultDir = "" // Start at "This PC" to show all drives
	}
	
	selectedDir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title:            "Select Hytale Installation Directory",
		DefaultDirectory: defaultDir,
	})
	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}
	
	if selectedDir == "" {
		return "", nil // User cancelled
	}
	
	// Validate that this looks like a Hytale installation
	gameDir := filepath.Join(selectedDir, "install", "release", "package", "game", "latest")
	if _, err := os.Stat(gameDir); os.IsNotExist(err) {
		return "", fmt.Errorf("invalid Hytale installation: game directory not found at %s", gameDir)
	}
	
	// Save to config
	a.cfg.GameInstallPath = selectedDir
	if err := config.Save(a.cfg); err != nil {
		return "", fmt.Errorf("failed to save config: %w", err)
	}
	
	fmt.Printf("Game install directory set to: %s\n", selectedDir)
	return selectedDir, nil
}

// progressCallback sends progress updates to frontend
func (a *App) progressCallback(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64) {
	wailsRuntime.EventsEmit(a.ctx, "progress-update", ProgressUpdate{
		Stage:       stage,
		Progress:    progress,
		Message:     message,
		CurrentFile: currentFile,
		Speed:       speed,
		Downloaded:  downloaded,
		Total:       total,
	})
}

// emitError sends structured errors to frontend
func (a *App) emitError(err error) {
	if appErr, ok := err.(*AppError); ok {
		wailsRuntime.EventsEmit(a.ctx, "error", appErr)
	} else {
		wailsRuntime.EventsEmit(a.ctx, "error", NewAppError(ErrorTypeUnknown, err.Error(), err))
	}
}

// AppVersion is the current launcher version - set at build time via ldflags
var AppVersion string = "dev"

// AppTitle is the app window title - set at build time via ldflags
var AppTitle string = "HyPrism - Hytale Launcher"

// GetLauncherVersion returns the current launcher version
func (a *App) GetLauncherVersion() string {
	return AppVersion
}

// Launch launches the game using the official Hytale installation
func (a *App) Launch(playerName string) error {
	// Validate nickname
	if len(playerName) == 0 {
		err := ValidationError("Please enter a nickname")
		a.emitError(err)
		return err
	}

	if len(playerName) > 16 {
		err := ValidationError("Nickname is too long (max 16 characters)")
		a.emitError(err)
		return err
	}

	// Check if game install path is configured
	if a.cfg.GameInstallPath == "" {
		err := GameError("Game not configured", fmt.Errorf("please set the Hytale install directory in settings"))
		a.emitError(err)
		return err
	}

	// Launch the game
	a.progressCallback("launch", 100, "Launching game...", "", "", 0, 0)

	if err := game.LaunchInstance(a.cfg.GameInstallPath); err != nil {
		wrappedErr := GameError("Failed to launch game", err)
		a.emitError(wrappedErr)
		return wrappedErr
	}

	return nil
}

// GetLogs returns launcher logs
func (a *App) GetLogs() (string, error) {
	logPath := filepath.Join(env.GetDefaultAppDir(), "logs", "launcher.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ==================== MOD MANAGER ====================

// SearchMods searches for mods on CurseForge
func (a *App) SearchMods(query string, categoryID int, page int) (*mods.SearchResult, error) {
	return mods.SearchMods(a.ctx, mods.SearchModsParams{
		Query:      query,
		CategoryID: categoryID,
		SortField:  "2", // Popularity
		SortOrder:  "desc",
		PageSize:   20,
		Index:      page * 20,
	})
}

// GetInstalledMods returns all installed mods (legacy)
func (a *App) GetInstalledMods() ([]mods.Mod, error) {
	return mods.GetInstalledMods()
}

// GetInstanceInstalledMods returns installed mods for a specific instance
func (a *App) GetInstanceInstalledMods(branch string, version int) ([]mods.Mod, error) {
	return mods.GetInstanceInstalledMods(branch, version)
}

// GetModDetails returns detailed info about a specific mod from CurseForge
func (a *App) GetModDetails(modID int) (*mods.CurseForgeMod, error) {
	return mods.GetModDetails(a.ctx, modID)
}

// GetModFiles returns available files/versions for a mod from CurseForge
func (a *App) GetModFiles(modID int) ([]mods.ModFile, error) {
	return mods.GetModFiles(a.ctx, modID)
}

// InstallMod downloads and installs a mod from CurseForge (legacy)
func (a *App) InstallMod(modID int) error {
	cfMod, err := mods.GetModDetails(a.ctx, modID)
	if err != nil {
		return err
	}

	return mods.DownloadMod(a.ctx, *cfMod, func(progress float64, message string) {
		wailsRuntime.EventsEmit(a.ctx, "mod-progress", map[string]interface{}{
			"progress": progress,
			"message":  message,
		})
	})
}

// InstallModToInstance downloads and installs a mod to a specific instance
func (a *App) InstallModToInstance(modID int, branch string, version int) error {
	cfMod, err := mods.GetModDetails(a.ctx, modID)
	if err != nil {
		return err
	}

	return mods.DownloadModToInstance(a.ctx, *cfMod, branch, version, func(progress float64, message string) {
		wailsRuntime.EventsEmit(a.ctx, "mod-progress", map[string]interface{}{
			"progress": progress,
			"message":  message,
		})
	})
}

// InstallModFile downloads and installs a specific mod file version from CurseForge (legacy)
func (a *App) InstallModFile(modID int, fileID int) error {
	return mods.DownloadModFile(a.ctx, modID, fileID, func(progress float64, message string) {
		wailsRuntime.EventsEmit(a.ctx, "mod-progress", map[string]interface{}{
			"progress": progress,
			"message":  message,
		})
	})
}

// InstallModFileToInstance downloads and installs a specific mod file version to an instance
func (a *App) InstallModFileToInstance(modID int, fileID int, branch string, version int) error {
	return mods.DownloadModFileToInstance(a.ctx, modID, fileID, branch, version, func(progress float64, message string) {
		wailsRuntime.EventsEmit(a.ctx, "mod-progress", map[string]interface{}{
			"progress": progress,
			"message":  message,
		})
	})
}

// UninstallMod removes an installed mod (legacy)
func (a *App) UninstallMod(modID string) error {
	return mods.RemoveMod(modID)
}

// UninstallInstanceMod removes an installed mod from an instance
func (a *App) UninstallInstanceMod(modID string, branch string, version int) error {
	return mods.RemoveInstanceMod(modID, branch, version)
}

// ToggleMod enables or disables a mod (legacy)
func (a *App) ToggleMod(modID string, enabled bool) error {
	return mods.ToggleMod(modID, enabled)
}

// ToggleInstanceMod enables or disables a mod in an instance
func (a *App) ToggleInstanceMod(modID string, enabled bool, branch string, version int) error {
	return mods.ToggleInstanceMod(modID, enabled, branch, version)
}

// GetModCategories returns available mod categories
func (a *App) GetModCategories() ([]mods.ModCategory, error) {
	return mods.GetCategories(a.ctx)
}

// CheckModUpdates checks for mod updates (legacy)
func (a *App) CheckModUpdates() ([]mods.Mod, error) {
	return mods.CheckForUpdates(a.ctx)
}

// CheckInstanceModUpdates checks for mod updates in an instance
func (a *App) CheckInstanceModUpdates(branch string, version int) ([]mods.Mod, error) {
	return mods.CheckInstanceForUpdates(a.ctx, branch, version)
}

// OpenModsFolder opens the mods folder in file explorer (legacy)
func (a *App) OpenModsFolder() error {
	modsDir := mods.GetModsDir()
	if err := os.MkdirAll(modsDir, 0755); err != nil {
		return err
	}
	return openFolder(modsDir)
}

// OpenInstanceModsFolder opens the mods folder for a specific instance
func (a *App) OpenInstanceModsFolder(branch string, version int) error {
	modsDir := mods.GetInstanceModsDir(branch, version)
	if err := os.MkdirAll(modsDir, 0755); err != nil {
		return err
	}
	return openFolder(modsDir)
}

// ==================== UTILITY ====================

// openFolder opens a folder in the system file explorer
func openFolder(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

// OpenGameFolder opens the configured Hytale game folder
func (a *App) OpenGameFolder() error {
	if a.cfg.GameInstallPath == "" {
		return fmt.Errorf("game install path not configured")
	}
	gameDir := filepath.Join(a.cfg.GameInstallPath, "UserData")
	if err := os.MkdirAll(gameDir, 0755); err != nil {
		return err
	}
	return openFolder(gameDir)
}

// QuickLaunch launches the game
func (a *App) QuickLaunch() error {
	return a.Launch("Player")
}

// ExitGame terminates the running game process
func (a *App) ExitGame() error {
	return game.KillGame()
}

// IsGameRunning returns whether the game is currently running
func (a *App) IsGameRunning() bool {
	return game.IsGameRunning()
}

// GetGameLogs returns the game log content
func (a *App) GetGameLogs() (string, error) {
	return game.GetGameLogs()
}


// ==================== AUTHENTICATION ====================

// LoginWithHytaleAccount initiates the Hytale account login flow
func (a *App) LoginWithHytaleAccount() error {
	fmt.Println("Starting Hytale account login...")
	
	// Emit progress updates to frontend
	progressCallback := func(message string) {
		wailsRuntime.EventsEmit(a.ctx, "auth-progress", message)
		fmt.Printf("[AUTH] %s\n", message)
	}
	
	// Browser opener function
	openBrowser := func(url string) error {
		return auth.OpenBrowser(url)
	}
	
	session, err := auth.LoginWithAuthCodeFlow(a.ctx, progressCallback, openBrowser)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	
	// Save session
	if err := auth.SaveSession(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	
	wailsRuntime.EventsEmit(a.ctx, "auth-success", map[string]string{
		"username": session.Username,
		"uuid":     session.UUID,
	})
	
	fmt.Printf("Successfully logged in as %s (UUID: %s)\n", session.Username, session.UUID)
	return nil
}

// LogoutHytaleAccount logs out the current Hytale account
func (a *App) LogoutHytaleAccount() error {
	if err := auth.ClearSession(); err != nil {
		return err
	}
	
	wailsRuntime.EventsEmit(a.ctx, "auth-logout", nil)
	fmt.Println("Logged out successfully")
	return nil
}

// GetAuthStatus returns the current authentication status
func (a *App) GetAuthStatus() map[string]interface{} {
	session, err := auth.LoadSession()
	if err != nil || session == nil {
		return map[string]interface{}{
			"logged_in": false,
		}
	}
	
	// Check if session is expired
	if session.IsExpired() {
		return map[string]interface{}{
			"logged_in": false,
			"expired":   true,
		}
	}
	
	return map[string]interface{}{
		"logged_in": true,
		"username":  session.Username,
		"uuid":      session.UUID,
	}
}

// OpenAuthURL opens the authentication URL in the default browser
func (a *App) OpenAuthURL(url string) error {
	return auth.OpenBrowser(url)
}

// GetUserProfile returns the current user's profile information
func (a *App) GetUserProfile() (map[string]string, error) {
	session, err := auth.GetValidSession()
	if err != nil {
		return nil, fmt.Errorf("not logged in: %w", err)
	}
	
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}
	
	return map[string]string{
		"username":         session.Username,
		"uuid":             session.UUID,
		"account_owner_id": session.AccountOwnerID,
	}, nil
}

// GetUserUUID returns just the UUID of the current user
func (a *App) GetUserUUID() (string, error) {
	session, err := auth.GetValidSession()
	if err != nil {
		return "", fmt.Errorf("not logged in: %w", err)
	}
	
	if session == nil {
		return "", fmt.Errorf("no active session")
	}
	
	return session.UUID, nil
}

// ==================== NEWS ====================

// GetNews fetches news from hytale.com
func (a *App) GetNews(limit int) ([]news.NewsItem, error) {
	if limit <= 0 {
		limit = 5
	}
	return a.newsService.GetNews(limit)
}
