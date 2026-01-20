package config

// Config represents the launcher configuration
type Config struct {
	Version         string `toml:"version" json:"version"`
	MusicEnabled    bool   `toml:"music_enabled" json:"musicEnabled"`
	GameInstallPath string `toml:"game_install_path" json:"gameInstallPath"` // Path to official Hytale installation
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Version:         "1.0.0",
		MusicEnabled:    true,
		GameInstallPath: "",
	}
}
