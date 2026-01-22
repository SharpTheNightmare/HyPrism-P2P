package config

type Config struct {
	Version         string `toml:"version" json:"version"`
	MusicEnabled    bool   `toml:"music_enabled" json:"musicEnabled"`
	GameInstallPath string `toml:"game_install_path" json:"gameInstallPath"`
}

func Default() *Config {
	return &Config{
		Version:         "1.0.0",
		MusicEnabled:    true,
		GameInstallPath: "",
	}
}
