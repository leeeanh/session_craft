package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type BookmarksConfig struct {
	Bookmarks []string `yaml:"bookmarks"`
}

func LoadBookmarks() ([]string, error) {
	// 1. Check main config
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// We might have already loaded main config in Model, so we could pass it.
	// But spec says `~/.config/sessioncraft/bookmarks.yml` is separate.
	// And also `Can also be managed in separate bookmarks.yml`.
	// Main config `config.yml` has `bookmarks` field too.
	// We should merge them.

	// Strategy: Load separate file if exists.
	var bookmarks []string

	path := filepath.Join(home, ".config", "sessioncraft", "bookmarks.yml")
	data, err := os.ReadFile(path)
	if err == nil {
		var bCfg BookmarksConfig
		if err := yaml.Unmarshal(data, &bCfg); err == nil {
			bookmarks = append(bookmarks, bCfg.Bookmarks...)
		}
	}

	return bookmarks, nil
}
