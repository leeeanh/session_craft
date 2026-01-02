package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func highlightString(text string, matches []int, matchStyle, normalStyle lipgloss.Style) string {
	if len(matches) == 0 {
		return text
	}

	var sb strings.Builder
	runes := []rune(text)
	matchSet := make(map[int]bool)
	for _, m := range matches {
		matchSet[m] = true
	}

	for i, r := range runes {
		if matchSet[i] {
			sb.WriteString(matchStyle.Render(string(r)))
		} else {
			sb.WriteString(string(r))
		}
	}
	return sb.String()
}

// ProcessIcon returns a Nerd Font icon for common process names
func ProcessIcon(processName string) string {
	icons := map[string]string{
		"nvim":    "",
		"vim":     "",
		"vi":      "",
		"nano":    "",
		"code":    "󰨞",
		"node":    "󰎙",
		"npm":     "",
		"yarn":    "",
		"python":  "",
		"python3": "",
		"go":      "",
		"cargo":   "",
		"rustc":   "",
		"docker":  "",
		"git":     "",
		"ssh":     "󰣀",
		"htop":    "",
		"btop":    "",
		"bash":    "",
		"zsh":     "",
		"fish":    "",
		"tmux":    "",
	}

	if icon, ok := icons[processName]; ok {
		return icon
	}
	return "" // Default terminal icon
}

// SessionStateIcon returns icon based on session state
func SessionStateIcon(attached bool) string {
	if attached {
		return "󱫋" // Active/attached
	}
	return "󰏤" // Idle/detached
}

// GhostIcon returns the ghost directory icon
func GhostIcon() string {
	return "" // Ghost/bookmark folder
}

// WindowIcon returns icon for a window based on its name
func WindowIcon(windowName string) string {
	// Could be enhanced to detect actual running process
	return ProcessIcon(windowName)
}
