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
			Background: "#1e1e2e",
			Foreground: "#cdd6f4",
			Accent:     "#cba6f7", // Mauve
			Border:     "#45475a", // Surface1
			Dimmed:     "#6c7086", // Overlay0
			Mantle:     "#181825",
			Active:     "#a6e3a1",
			Warning:    "#f9e2af",
			Danger:     "#f38ba8",
			Surface:    "#313244",
			TextMuted:  "#a6adc8",
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
