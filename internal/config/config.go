package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Preview   PreviewConfig `yaml:"preview"`
	Theme     ThemeConfig   `yaml:"theme"`
	Bookmarks []string      `yaml:"bookmarks"`
}

type PreviewConfig struct {
	Lines int `yaml:"lines"`
}

type ThemeConfig struct {
	Background string `yaml:"background"`
	Foreground string `yaml:"foreground"`
	Accent     string `yaml:"accent"`
	Border     string `yaml:"border"`
	Dimmed     string `yaml:"dimmed"`
	// New fields for Obsidian Dashboard
	Mantle    string `yaml:"mantle"`     // Sidebar tint
	Active    string `yaml:"active"`     // Green for active sessions
	Warning   string `yaml:"warning"`    // Yellow for warnings
	Danger    string `yaml:"danger"`     // Red for destructive actions
	Surface   string `yaml:"surface"`    // Progress bar empty
	TextMuted string `yaml:"text_muted"` // Secondary text
}

func DefaultConfig() Config {
	return Config{
		Preview: PreviewConfig{
			Lines: 20,
		},
		Theme: ThemeConfig{
			Background: "#1a1b26",
			Foreground: "#c0caf5",
			Accent:     "#7aa2f7",
			Border:     "#414868",
			Dimmed:     "#787fa0",
			Mantle:     "#1f2335",
			Active:     "#9ece6a",
			Warning:    "#e0af68",
			Danger:     "#f7768e",
			Surface:    "#24283b",
			TextMuted:  "#787fa0",
		},
	}
}

func LoadConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig(), err
	}

	path := filepath.Join(home, ".config", "sessioncraft", "config.yml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	if err != nil {
		return DefaultConfig(), fmt.Errorf("failed to read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
