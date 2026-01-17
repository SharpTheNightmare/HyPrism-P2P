package mods

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"HyPrism/internal/env"
)

// Mod represents a mod
type Mod struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Author       string `json:"author"`
	Description  string `json:"description"`
	DownloadURL  string `json:"downloadUrl,omitempty"`
	CurseForgeID int    `json:"curseforgeId,omitempty"`
	Enabled      bool   `json:"enabled"`
	InstalledAt  string `json:"installedAt"`  // ISO 8601 format
	UpdatedAt    string `json:"updatedAt"`    // ISO 8601 format
	FilePath     string `json:"filePath"`
	IconURL      string `json:"iconUrl,omitempty"`
	Downloads    int    `json:"downloads,omitempty"`
	Category     string `json:"category,omitempty"`
}

// ModManifest stores installed mods info
type ModManifest struct {
	Mods    []Mod  `json:"mods"`
	Version string `json:"version"`
}

// GetModsDir returns the mods directory path  
// Mods should be in UserData/Mods as that's where the game reads them
func GetModsDir() string {
	return filepath.Join(env.GetDefaultAppDir(), "UserData", "Mods")
}

// GetModManifestPath returns the mod manifest path
func GetModManifestPath() string {
	return filepath.Join(GetModsDir(), "manifest.json")
}

// LoadManifest loads the mod manifest
func LoadManifest() (*ModManifest, error) {
	path := GetModManifestPath()
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ModManifest{Mods: []Mod{}, Version: "1.0"}, nil
		}
		return nil, err
	}

	var manifest ModManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// SaveManifest saves the mod manifest
func SaveManifest(manifest *ModManifest) error {
	path := GetModManifestPath()
	
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetInstalledMods returns all installed mods
func GetInstalledMods() ([]Mod, error) {
	manifest, err := LoadManifest()
	if err != nil {
		return nil, err
	}
	return manifest.Mods, nil
}

// AddMod adds a mod to the manifest
func AddMod(mod Mod) error {
	manifest, err := LoadManifest()
	if err != nil {
		return err
	}

	// Check if already exists
	for i, m := range manifest.Mods {
		if m.ID == mod.ID {
			manifest.Mods[i] = mod
			return SaveManifest(manifest)
		}
	}

	manifest.Mods = append(manifest.Mods, mod)
	return SaveManifest(manifest)
}

// RemoveMod removes a mod from manifest and deletes files
func RemoveMod(modID string) error {
	manifest, err := LoadManifest()
	if err != nil {
		return err
	}

	var newMods []Mod
	var modToRemove *Mod
	for _, m := range manifest.Mods {
		if m.ID == modID {
			modToRemove = &m
		} else {
			newMods = append(newMods, m)
		}
	}

	if modToRemove == nil {
		return fmt.Errorf("mod not found: %s", modID)
	}

	// Delete mod file
	if modToRemove.FilePath != "" {
		if err := os.Remove(modToRemove.FilePath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	manifest.Mods = newMods
	return SaveManifest(manifest)
}

// ToggleMod enables or disables a mod
func ToggleMod(modID string, enabled bool) error {
	manifest, err := LoadManifest()
	if err != nil {
		return err
	}

	for i, m := range manifest.Mods {
		if m.ID == modID {
			manifest.Mods[i].Enabled = enabled
			
			// Rename file to enable/disable
			oldPath := m.FilePath
			newPath := oldPath
			
			if enabled && filepath.Ext(oldPath) == ".disabled" {
				newPath = oldPath[:len(oldPath)-9] // Remove .disabled
			} else if !enabled && filepath.Ext(oldPath) != ".disabled" {
				newPath = oldPath + ".disabled"
			}
			
			if oldPath != newPath {
				if err := os.Rename(oldPath, newPath); err != nil {
					return err
				}
				manifest.Mods[i].FilePath = newPath
			}
			
			return SaveManifest(manifest)
		}
	}

	return fmt.Errorf("mod not found: %s", modID)
}
