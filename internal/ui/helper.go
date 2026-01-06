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
        // Editors & IDEs
        "nvim":    "", "vim":     "", "vi":      "",
        "nano":    "󰏫", "code":    "󰨞", "emacs":   "",
        "subl":    "",

        // Languages & Runtimes
        "node":    "󰎙", "npm":     "", "yarn":    "", "bun":     "",
        "python":  "", "python3": "", "go":      "", "rustc":   "",
        "cargo":   "", "java":     "", "ruby":    "", "php":     "󰌟",
        "lua":     "", "deno":    "", "perl":    "",

        // Infrastructure & Databases
        "docker":  "󰡨", "k8s":     "󱄄", "kubectl": "󱄄", "terraform": "",
        "aws":     "", "gcloud":  "󱇶", "sql":     "", "postgres": "",
        "mysql":   "", "redis":   "", "mongo":   "",

        // Version Control & CLI Tools
        "git":     "", "ssh":     "󰣀", "tmux":    "", "direnv":  "",
        "make":    "", "cmake":   "", "just":    "󱁤",

        // Shells
        "bash":    "", "zsh":     "", "fish":    "󰈺", "sh":      "",
        "powershell": "", "pwsh": "",

        // Monitoring & System
        "htop":    "󱫔", "btop":    "󱫔", "top":     "󱫔", "nvtop":   "󰢮",
        "ping":    "󰓅", "curl":    "", "wget":    "", "grep":    "",
        "ripgrep": "", "rg":      "", "fzf":     "",
    }

    // Normalize input (lowercase) to ensure matches
    icon, ok := icons[strings.ToLower(processName)]
    if ok {
        return icon
    }
    
    return "" // Default terminal/prompt icon
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
	return "󱗜" // Ghost/bookmark folder
}

// WindowIcon returns icon for a window based on its name
func WindowIcon(windowName string) string {
	// Simple mapping: you can expand this to check for 
	// specific suffixes or prefixes if needed.
	icon := ProcessIcon(windowName)
	if icon == "" { // If it's just the default terminal icon
		return "󰖲" // Generic window icon
	}
	return icon
}
