package components

import "github.com/charmbracelet/lipgloss"

// Badge renders a pill-shaped label with background color
type Badge struct {
	text  string
	style lipgloss.Style
}

// NewBadge creates a badge with specified colors
func NewBadge(text string, bg, fg lipgloss.TerminalColor) Badge {
	return Badge{
		text: text,
		style: lipgloss.NewStyle().
			Background(bg).
			Foreground(fg).
			Padding(0, 1).
			Bold(true),
	}
}

// Render returns the styled badge string
func (b Badge) Render() string {
	return b.style.Render(b.text)
}

// BadgeFromStyle creates a badge using a pre-configured style
func BadgeFromStyle(text string, style lipgloss.Style) Badge {
	return Badge{text: text, style: style}
}
