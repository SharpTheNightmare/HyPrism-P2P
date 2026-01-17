package worlds

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"HyPrism/internal/env"
)

// World represents a Hytale world/save
type World struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	CreatedAt    string `json:"createdAt"`  // ISO 8601 format
	LastPlayed   string `json:"lastPlayed"` // ISO 8601 format
	SizeBytes    int64  `json:"sizeBytes"`
	GameMode     string `json:"gameMode"`
	Seed         string `json:"seed,omitempty"`
	IsBackup     bool   `json:"isBackup"`
	BackupOf     string `json:"backupOf,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
}

// WorldManifest stores world info
type WorldManifest struct {
	Worlds  []World `json:"worlds"`
	Version string  `json:"version"`
}

// GetWorldsDir returns the worlds directory path
func GetWorldsDir() string {
	return filepath.Join(env.GetDefaultAppDir(), "UserData", "worlds")
}

// GetBackupsDir returns the backups directory path
func GetBackupsDir() string {
	return filepath.Join(env.GetDefaultAppDir(), "UserData", "backups")
}

// GetManifestPath returns the world manifest path
func GetManifestPath() string {
	return filepath.Join(GetWorldsDir(), "manifest.json")
}

// LoadManifest loads the world manifest
func LoadManifest() (*WorldManifest, error) {
	path := GetManifestPath()
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &WorldManifest{Worlds: []World{}, Version: "1.0"}, nil
		}
		return nil, err
	}

	var manifest WorldManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// SaveManifest saves the world manifest
func SaveManifest(manifest *WorldManifest) error {
	path := GetManifestPath()
	
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ScanWorlds scans the worlds directory for saves
func ScanWorlds() ([]World, error) {
	worldsDir := GetWorldsDir()
	
	if err := os.MkdirAll(worldsDir, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(worldsDir)
	if err != nil {
		return nil, err
	}

	var worlds []World
	manifest, _ := LoadManifest()
	existingWorlds := make(map[string]World)
	for _, w := range manifest.Worlds {
		existingWorlds[w.Path] = w
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		if entry.Name() == "backups" {
			continue
		}

		worldPath := filepath.Join(worldsDir, entry.Name())
		
		// Check if it's a valid world directory
		levelFile := filepath.Join(worldPath, "level.dat")
		if _, err := os.Stat(levelFile); err != nil {
			// Also check for Hytale-specific world files
			worldFile := filepath.Join(worldPath, "world.json")
			if _, err := os.Stat(worldFile); err != nil {
				continue
			}
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Calculate directory size
		size := calculateDirSize(worldPath)

		// Use existing data if available, otherwise create new
		world := World{
			ID:         entry.Name(),
			Name:       entry.Name(),
			Path:       worldPath,
			CreatedAt:  info.ModTime().Format(time.RFC3339),
			LastPlayed: info.ModTime().Format(time.RFC3339),
			SizeBytes:  size,
			GameMode:   "Survival",
		}

		if existing, ok := existingWorlds[worldPath]; ok {
			world.Name = existing.Name
			world.CreatedAt = existing.CreatedAt
			world.GameMode = existing.GameMode
			world.Seed = existing.Seed
		}

		worlds = append(worlds, world)
	}

	// Sort by last played
	sort.Slice(worlds, func(i, j int) bool {
		return worlds[i].LastPlayed > worlds[j].LastPlayed
	})

	// Save updated manifest
	manifest.Worlds = worlds
	_ = SaveManifest(manifest)

	return worlds, nil
}

// GetWorld gets a specific world by ID
func GetWorld(worldID string) (*World, error) {
	worlds, err := ScanWorlds()
	if err != nil {
		return nil, err
	}

	for _, w := range worlds {
		if w.ID == worldID {
			return &w, nil
		}
	}

	return nil, fmt.Errorf("world not found: %s", worldID)
}

// RenameWorld renames a world
func RenameWorld(worldID, newName string) error {
	manifest, err := LoadManifest()
	if err != nil {
		return err
	}

	for i, w := range manifest.Worlds {
		if w.ID == worldID {
			manifest.Worlds[i].Name = newName
			return SaveManifest(manifest)
		}
	}

	return fmt.Errorf("world not found: %s", worldID)
}

// DeleteWorld deletes a world
func DeleteWorld(worldID string) error {
	world, err := GetWorld(worldID)
	if err != nil {
		return err
	}

	// Delete the directory
	if err := os.RemoveAll(world.Path); err != nil {
		return fmt.Errorf("failed to delete world: %w", err)
	}

	// Update manifest
	manifest, err := LoadManifest()
	if err != nil {
		return err
	}

	var newWorlds []World
	for _, w := range manifest.Worlds {
		if w.ID != worldID {
			newWorlds = append(newWorlds, w)
		}
	}
	manifest.Worlds = newWorlds

	return SaveManifest(manifest)
}

// BackupWorld creates a backup of a world
func BackupWorld(worldID string) (*World, error) {
	world, err := GetWorld(worldID)
	if err != nil {
		return nil, err
	}

	backupsDir := GetBackupsDir()
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		return nil, err
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("%s_backup_%s", world.Name, timestamp)
	backupPath := filepath.Join(backupsDir, backupName)

	// Copy world to backup
	if err := copyDir(world.Path, backupPath); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	size := calculateDirSize(backupPath)

	backup := World{
		ID:         backupName,
		Name:       backupName,
		Path:       backupPath,
		CreatedAt:  time.Now().Format(time.RFC3339),
		LastPlayed: world.LastPlayed,
		SizeBytes:  size,
		GameMode:   world.GameMode,
		Seed:       world.Seed,
		IsBackup:   true,
		BackupOf:   worldID,
	}

	return &backup, nil
}

// RestoreBackup restores a backup to the worlds directory
func RestoreBackup(backupID string) (*World, error) {
	backupsDir := GetBackupsDir()
	backupPath := filepath.Join(backupsDir, backupID)
	
	if _, err := os.Stat(backupPath); err != nil {
		return nil, fmt.Errorf("backup not found: %s", backupID)
	}

	// Create a new world from backup
	timestamp := time.Now().Format("20060102150405")
	worldID := fmt.Sprintf("restored_%s", timestamp)
	worldPath := filepath.Join(GetWorldsDir(), worldID)

	if err := copyDir(backupPath, worldPath); err != nil {
		return nil, fmt.Errorf("failed to restore backup: %w", err)
	}

	size := calculateDirSize(worldPath)

	world := World{
		ID:         worldID,
		Name:       fmt.Sprintf("Restored - %s", backupID),
		Path:       worldPath,
		CreatedAt:  time.Now().Format(time.RFC3339),
		LastPlayed: time.Now().Format(time.RFC3339),
		SizeBytes:  size,
		GameMode:   "Survival",
	}

	// Add to manifest
	manifest, err := LoadManifest()
	if err != nil {
		return nil, err
	}
	manifest.Worlds = append(manifest.Worlds, world)
	if err := SaveManifest(manifest); err != nil {
		return nil, err
	}

	return &world, nil
}

// GetBackups returns all backups
func GetBackups() ([]World, error) {
	backupsDir := GetBackupsDir()
	
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return nil, err
	}

	var backups []World
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		backupPath := filepath.Join(backupsDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		size := calculateDirSize(backupPath)

		backup := World{
			ID:         entry.Name(),
			Name:       entry.Name(),
			Path:       backupPath,
			CreatedAt:  info.ModTime().Format(time.RFC3339),
			LastPlayed: info.ModTime().Format(time.RFC3339),
			SizeBytes:  size,
			IsBackup:   true,
		}

		backups = append(backups, backup)
	}

	// Sort by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})

	return backups, nil
}

// DeleteBackup deletes a backup
func DeleteBackup(backupID string) error {
	backupPath := filepath.Join(GetBackupsDir(), backupID)
	return os.RemoveAll(backupPath)
}

// Helper functions

func calculateDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// FormatSize formats bytes to human readable string
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
