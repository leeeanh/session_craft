package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/user/sessioncraft/internal/config"
)

type Styles struct {
	// Existing
	Accent   lipgloss.Style
	Dimmed   lipgloss.Style
	Border   lipgloss.Style
	Selected lipgloss.Style
	Header   lipgloss.Style
	Footer   lipgloss.Style
	Normal   lipgloss.Style
	Match    lipgloss.Style

	// New for Obsidian Dashboard
	SidebarBg      lipgloss.Style // Tinted background
	ActiveBorder   lipgloss.Style // Green left border
	IdleBorder     lipgloss.Style // Dimmed left border
	GhostBorder    lipgloss.Style // Dashed/dim left border
	SelectedBorder lipgloss.Style // Lavender left border

	// Badges
	BadgeActive  lipgloss.Style
	BadgeProcess lipgloss.Style
	BadgeCount   lipgloss.Style
	BadgeDanger  lipgloss.Style
	BadgeWarning lipgloss.Style

	// Footer pills
	PillKey   lipgloss.Style
	PillLabel lipgloss.Style

	// Cards
	CardBorder lipgloss.Style
	CardTitle  lipgloss.Style

	// Text variants
	TextMuted   lipgloss.Style
	TextBright  lipgloss.Style
	TextDanger  lipgloss.Style
	TextWarning lipgloss.Style
	TextActive  lipgloss.Style

	// Frame
	OuterFrame lipgloss.Style
}

func NewStyles(theme config.ThemeConfig) Styles {
	accent := lipgloss.Color(theme.Accent)
	dimmed := lipgloss.Color(theme.Dimmed)
	border := lipgloss.Color(theme.Border)
	fg := lipgloss.Color(theme.Foreground)
	bg := lipgloss.Color(theme.Background)
	mantle := lipgloss.Color(theme.Mantle)
	active := lipgloss.Color(theme.Active)
	warning := lipgloss.Color(theme.Warning)
	danger := lipgloss.Color(theme.Danger)
	surface := lipgloss.Color(theme.Surface)
	textMuted := lipgloss.Color(theme.TextMuted)

	return Styles{
		// Existing styles
		Accent:   lipgloss.NewStyle().Foreground(accent),
		Dimmed:   lipgloss.NewStyle().Foreground(dimmed),
		Border:   lipgloss.NewStyle().Foreground(border),
		Selected: lipgloss.NewStyle().Foreground(bg).Background(accent).Bold(true),
		Header:   lipgloss.NewStyle().Foreground(accent).Bold(true).Padding(0, 1).MarginBottom(1),
		Footer:   lipgloss.NewStyle().Foreground(dimmed).Italic(true),
		Normal:   lipgloss.NewStyle().Foreground(fg),
		Match:    lipgloss.NewStyle().Foreground(warning).Bold(true).Underline(true),

		// Sidebar background
		SidebarBg: lipgloss.NewStyle().Background(mantle),

		// Accent borders for tree items
		ActiveBorder:   lipgloss.NewStyle().Foreground(active).Bold(true),
		IdleBorder:     lipgloss.NewStyle().Foreground(dimmed),
		GhostBorder:    lipgloss.NewStyle().Foreground(surface),
		SelectedBorder: lipgloss.NewStyle().Foreground(accent).Bold(true),

		// Badges
		BadgeActive:  lipgloss.NewStyle().Background(active).Foreground(bg).Padding(0, 1).Bold(true),
		BadgeProcess: lipgloss.NewStyle().Background(accent).Foreground(bg).Padding(0, 1),
		BadgeCount:   lipgloss.NewStyle().Background(surface).Foreground(textMuted).Padding(0, 1),
		BadgeDanger:  lipgloss.NewStyle().Background(danger).Foreground(bg).Padding(0, 1).Bold(true),
		BadgeWarning: lipgloss.NewStyle().Background(warning).Foreground(bg).Padding(0, 1),

		// Footer pills
		PillKey:   lipgloss.NewStyle().Background(surface).Foreground(accent).Padding(0, 1).Bold(true),
		PillLabel: lipgloss.NewStyle().Foreground(textMuted).MarginRight(2),

		// Cards
		CardBorder: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(border),
		CardTitle:  lipgloss.NewStyle().Foreground(accent).Bold(true),

		// Text variants
		TextMuted:   lipgloss.NewStyle().Foreground(textMuted),
		TextBright:  lipgloss.NewStyle().Foreground(fg).Bold(true),
		TextDanger:  lipgloss.NewStyle().Foreground(danger),
		TextWarning: lipgloss.NewStyle().Foreground(warning),
		TextActive:  lipgloss.NewStyle().Foreground(active),

		// Outer frame
		OuterFrame: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(0, 1),
	}
}
