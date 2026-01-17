package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"HyPrism/internal/config"
	"HyPrism/internal/env"
	"HyPrism/internal/game"
	"HyPrism/internal/mods"
	"HyPrism/internal/news"
	"HyPrism/internal/pwr"
	"HyPrism/internal/skin"
	"HyPrism/internal/worlds"

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

// GetVersions returns current and latest game versions
func (a *App) GetVersions() (currentVersion string, latestVersion string) {
	current := pwr.GetLocalVersion()
	latest := pwr.FindLatestVersion("release")
	return current, strconv.Itoa(latest)
}

// DownloadAndLaunch downloads the game if needed and launches it
func (a *App) DownloadAndLaunch(playerName string) error {
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

	// Ensure game is installed
	if err := game.EnsureInstalled(a.ctx, a.progressCallback); err != nil {
		wrappedErr := GameError("Failed to install or update game", err)
		a.emitError(wrappedErr)
		return wrappedErr
	}

	// Launch the game
	a.progressCallback("launch", 100, "Launching game...", "", "", 0, 0)

	if err := game.Launch(playerName, "latest"); err != nil {
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

// GetSkinPreset returns the current skin preset
func (a *App) GetSkinPreset() (*skin.AvatarPreset, error) {
	gameDir := filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest")
	return skin.LoadPreset(gameDir)
}

// SaveSkinPreset saves a skin preset
func (a *App) SaveSkinPreset(preset skin.AvatarPreset) error {
	gameDir := filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest")
	return skin.SavePreset(gameDir, &preset)
}

// GetCosmeticCategories returns all available cosmetic categories
func (a *App) GetCosmeticCategories() map[string][]skin.CosmeticItem {
	return skin.GetAllCosmetics()
}

// GetAvailableColors returns available colors for cosmetics
func (a *App) GetAvailableColors() []string {
	return skin.GetAvailableColors()
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

// GetInstalledMods returns all installed mods
func (a *App) GetInstalledMods() ([]mods.Mod, error) {
	return mods.GetInstalledMods()
}

// InstallMod downloads and installs a mod from CurseForge
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

// UninstallMod removes an installed mod
func (a *App) UninstallMod(modID string) error {
	return mods.RemoveMod(modID)
}

// ToggleMod enables or disables a mod
func (a *App) ToggleMod(modID string, enabled bool) error {
	return mods.ToggleMod(modID, enabled)
}

// GetModCategories returns available mod categories
func (a *App) GetModCategories() ([]mods.ModCategory, error) {
	return mods.GetCategories(a.ctx)
}

// CheckModUpdates checks for mod updates
func (a *App) CheckModUpdates() ([]mods.Mod, error) {
	return mods.CheckForUpdates(a.ctx)
}

// OpenModsFolder opens the mods folder in file explorer
func (a *App) OpenModsFolder() error {
	modsDir := mods.GetModsDir()
	if err := os.MkdirAll(modsDir, 0755); err != nil {
		return err
	}
	return openFolder(modsDir)
}

// ==================== WORLD MANAGER ====================

// GetWorlds returns all worlds
func (a *App) GetWorlds() ([]worlds.World, error) {
	return worlds.ScanWorlds()
}

// GetWorld returns a specific world
func (a *App) GetWorld(worldID string) (*worlds.World, error) {
	return worlds.GetWorld(worldID)
}

// RenameWorld renames a world
func (a *App) RenameWorld(worldID, newName string) error {
	return worlds.RenameWorld(worldID, newName)
}

// DeleteWorld deletes a world
func (a *App) DeleteWorld(worldID string) error {
	return worlds.DeleteWorld(worldID)
}

// BackupWorld creates a backup of a world
func (a *App) BackupWorld(worldID string) (*worlds.World, error) {
	return worlds.BackupWorld(worldID)
}

// GetBackups returns all world backups
func (a *App) GetBackups() ([]worlds.World, error) {
	return worlds.GetBackups()
}

// RestoreBackup restores a backup
func (a *App) RestoreBackup(backupID string) (*worlds.World, error) {
	return worlds.RestoreBackup(backupID)
}

// DeleteBackup deletes a backup
func (a *App) DeleteBackup(backupID string) error {
	return worlds.DeleteBackup(backupID)
}

// OpenWorldsFolder opens the worlds folder in file explorer
func (a *App) OpenWorldsFolder() error {
	worldsDir := worlds.GetWorldsDir()
	if err := os.MkdirAll(worldsDir, 0755); err != nil {
		return err
	}
	return openFolder(worldsDir)
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

// OpenGameFolder opens the game folder
func (a *App) OpenGameFolder() error {
	gameDir := filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest")
	if err := os.MkdirAll(gameDir, 0755); err != nil {
		return err
	}
	return openFolder(gameDir)
}

// GetGamePath returns the game installation path
func (a *App) GetGamePath() string {
	return filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest")
}

// IsGameInstalled checks if the game is installed
func (a *App) IsGameInstalled() bool {
	gameClient := "HytaleClient"
	if runtime.GOOS == "windows" {
		gameClient += ".exe"
	}
	clientPath := filepath.Join(a.GetGamePath(), "Client", gameClient)
	_, err := os.Stat(clientPath)
	return err == nil
}

// QuickLaunch launches the game with saved nickname
func (a *App) QuickLaunch() error {
	nick := a.cfg.Nick
	if nick == "" {
		nick = "Player"
	}
	return a.DownloadAndLaunch(nick)
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

// GetAvailableVersions returns list of available game versions (release and prerelease)
func (a *App) GetAvailableVersions() map[string]int {
	versions := make(map[string]int)
	versions["release"] = pwr.FindLatestVersion("release")
	versions["prerelease"] = pwr.FindLatestVersion("prerelease")
	return versions
}

// GetCurrentVersion returns the currently installed game version with formatted date
func (a *App) GetCurrentVersion() string {
	return pwr.GetLocalVersionFull()
}

// InstalledVersion represents an installed game version
type InstalledVersion struct {
	Version     int    `json:"version"`
	VersionType string `json:"versionType"`
	InstallDate string `json:"installDate"`
}

// GetInstalledVersions returns all installed game versions
func (a *App) GetInstalledVersions() []InstalledVersion {
	versions := pwr.GetInstalledVersions()
	result := make([]InstalledVersion, 0, len(versions))
	for _, v := range versions {
		result = append(result, InstalledVersion{
			Version:     v.Version,
			VersionType: v.VersionType,
			InstallDate: v.InstallDate,
		})
	}
	return result
}

// SwitchVersion switches to a different installed version
func (a *App) SwitchVersion(version int) error {
	return pwr.SwitchVersion(version)
}

// DownloadVersion downloads a specific version type
func (a *App) DownloadVersion(versionType string, playerName string) error {
	if versionType != "release" && versionType != "prerelease" {
		return fmt.Errorf("invalid version type: %s", versionType)
	}
	
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

	// Install specific version
	if err := game.EnsureInstalledVersion(a.ctx, versionType, a.progressCallback); err != nil {
		wrappedErr := GameError("Failed to install game version", err)
		a.emitError(wrappedErr)
		return wrappedErr
	}

	// Launch the game
	a.progressCallback("launch", 100, "Launching game...", "", "", 0, 0)

	if err := game.Launch(playerName, "latest"); err != nil {
		wrappedErr := GameError("Failed to launch game", err)
		a.emitError(wrappedErr)
		return wrappedErr
	}

	return nil
}

// ==================== NEWS ====================

// GetNews fetches news from hytale.com
func (a *App) GetNews(limit int) ([]news.NewsItem, error) {
	if limit <= 0 {
		limit = 5
	}
	return a.newsService.GetNews(limit)
}
