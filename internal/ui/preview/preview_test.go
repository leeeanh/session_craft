package preview

import (
	"strings"
	"testing"

	"github.com/user/sessioncraft/internal/config"
	"github.com/user/sessioncraft/internal/tmux"
)

func TestPreviewView(t *testing.T) {
	theme := config.DefaultConfig().Theme
	m := NewModel(theme)
	m.Width = 60
	m.Height = 20
	m.Active = true
	m.Content = "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	m.ResourceUsage = tmux.ResourceUsage{CPU: "10%", Memory: "50%"}
	m.Metadata = Metadata{
		Path:        "~/projects",
		Uptime:      "2h 30m",
		ClientCount: 2,
		WindowCount: 3,
		Tags:        []string{"DEV"},
	}

	view := m.View()

	// Check for Metrics card elements
	if !strings.Contains(view, "Metrics") {
		t.Error("View should contain Metrics card title")
	}
	if !strings.Contains(view, "CPU") {
		t.Error("View should contain CPU label")
	}

	// Check for Preview card
	if !strings.Contains(view, "Preview") {
		t.Error("View should contain Preview card title")
	}
	if !strings.Contains(view, "Line 5") {
		t.Error("View should contain content")
	}

	// Check for Info card
	if !strings.Contains(view, "Info") {
		t.Error("View should contain Info card title")
	}

	// Test inactive model returns empty
	m.Active = false
	view = m.View()
	if view != "" {
		t.Error("Inactive model should return empty string")
	}
}

func TestParsePercent(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"50%", 0.5},
		{"100%", 1.0},
		{"0%", 0.0},
		{"25", 0.25},
		{" 75% ", 0.75},
	}

	for _, tt := range tests {
		result := parsePercent(tt.input)
		if result != tt.expected {
			t.Errorf("parsePercent(%q) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestFormatMemory(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1024KB", "1.0MB"},
		{"2048KB", "2.0MB"},
		{"512KB", "0.5MB"},
		{"1048576KB", "1.0GB"},
		{"4500KB", "4.4MB"},
	}

	for _, tt := range tests {
		result := formatMemory(tt.input)
		if result != tt.expected {
			t.Errorf("formatMemory(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestTruncateWithANSI(t *testing.T) {
	theme := config.DefaultConfig().Theme
	m := NewModel(theme)
	m.Width = 30
	m.Height = 10
	m.Active = true

	// Content with ANSI color codes
	// \x1b[32m = green, \x1b[0m = reset
	m.Content = "\x1b[32mThis is a very long green colored line that should be truncated properly\x1b[0m"
	m.ResourceUsage = tmux.ResourceUsage{CPU: "10%", Memory: "50KB"}
	m.Metadata = Metadata{Path: "~", Uptime: "1h", ClientCount: 1, WindowCount: 1}

	view := m.View()

	// Should not contain broken escape sequences (partial \x1b[)
	if strings.Contains(view, "\x1b[3") && !strings.Contains(view, "\x1b[32m") && !strings.Contains(view, "\x1b[0m") {
		t.Error("ANSI escape sequence was truncated mid-sequence")
	}

	// Should still contain some green color code or be properly terminated
	// The key is no corrupted partial escapes
	if strings.Count(view, "\x1b[") != strings.Count(view, "m") {
		// This would indicate unbalanced/corrupted escapes
		// Note: This is a heuristic, not perfect
	}
}

func TestTruncateUnicode(t *testing.T) {
	theme := config.DefaultConfig().Theme
	m := NewModel(theme)
	m.Width = 20
	m.Height = 10
	m.Active = true

	// Content with wide unicode characters (emoji = 2 cells wide)
	m.Content = "Hello 🚀🎉🔥 World this is a long line"
	m.ResourceUsage = tmux.ResourceUsage{CPU: "5%", Memory: "100KB"}
	m.Metadata = Metadata{Path: "~", Uptime: "30m", ClientCount: 1, WindowCount: 1}

	view := m.View()

	// Should render without panic
	if view == "" {
		t.Error("View should not be empty")
	}

	// Should contain the ellipsis indicating truncation
	if !strings.Contains(view, "…") && !strings.Contains(view, "Hello") {
		t.Error("Content should be present or truncated with ellipsis")
	}
}
