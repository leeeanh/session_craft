package components

import "github.com/charmbracelet/lipgloss"

// Card renders content in a rounded border box with optional title
type Card struct {
	title       string
	width       int
	borderStyle lipgloss.Style
	titleStyle  lipgloss.Style
}

// NewCard creates a card with rounded borders
func NewCard(title string, width int, borderColor, titleColor lipgloss.TerminalColor) Card {
	return Card{
		title: title,
		width: width,
		borderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Width(width),
		titleStyle: lipgloss.NewStyle().
			Foreground(titleColor).
			Bold(true),
	}
}

// Render returns the card with content inside
func (c Card) Render(content string) string {
	if c.title != "" {
		titleLine := c.titleStyle.Render("─ " + c.title + " ")
		return c.borderStyle.Render(titleLine + "\n" + content)
	}
	return c.borderStyle.Render(content)
}

// RenderWithHeight creates a card with fixed height
func (c Card) RenderWithHeight(content string, height int) string {
	style := c.borderStyle.Height(height)
	if c.title != "" {
		titleLine := c.titleStyle.Render("─ " + c.title + " ")
		return style.Render(titleLine + "\n" + content)
	}
	return style.Render(content)
}
