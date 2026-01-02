package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProgressBar renders a block-based progress indicator
type ProgressBar struct {
	width      int
	fillStyle  lipgloss.Style
	emptyStyle lipgloss.Style
}

// NewProgressBar creates a progress bar with specified width and colors
func NewProgressBar(width int, fill, empty lipgloss.TerminalColor) ProgressBar {
	return ProgressBar{
		width:      width,
		fillStyle:  lipgloss.NewStyle().Foreground(fill),
		emptyStyle: lipgloss.NewStyle().Foreground(empty),
	}
}

// Render returns the progress bar at the given percentage (0.0 to 1.0)
func (p ProgressBar) Render(percent float64) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}

	filled := int(float64(p.width) * percent)
	empty := p.width - filled

	return p.fillStyle.Render(strings.Repeat("█", filled)) +
		p.emptyStyle.Render(strings.Repeat("░", empty))
}

// RenderWithLabel returns "Label [████░░░░] XX%"
func (p ProgressBar) RenderWithLabel(label string, percent float64, labelStyle lipgloss.Style) string {
	bar := p.Render(percent)
	pct := int(percent * 100)
	return fmt.Sprintf("%s [%s] %3d%%", labelStyle.Render(label), bar, pct)
}
