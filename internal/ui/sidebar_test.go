package ui

import (
	"strings"
	"testing"
)

func TestSidebarFooterAlignment(t *testing.T) {
	m := NewModel()
	m.width = 40
	m.height = 20

	// Render sidebar
	sidebar := m.renderSidebar(30, 18, 2)

	// Count lines
	lines := strings.Split(sidebar, "\n")

	// The footer should be near the bottom of the sidebar
	// Find the line containing footer content (keybinding pills like "Attach")
	footerLineIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "Attach") {
			footerLineIdx = i
			break
		}
	}

	if footerLineIdx == -1 {
		t.Fatal("Could not find footer line containing 'Attach'")
	}

	// Footer should be at or near the last line (allowing for 1-2 lines of padding/border)
	// The footer should be within the last 3 lines of the sidebar
	totalLines := len(lines)
	if footerLineIdx < totalLines-3 {
		t.Errorf("Footer not at bottom. Footer at line %d, total lines %d (expected within last 3 lines)", footerLineIdx, totalLines)
	}
}
