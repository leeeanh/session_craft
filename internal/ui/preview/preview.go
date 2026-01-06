package preview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
	"github.com/user/sessioncraft/internal/config"
	"github.com/user/sessioncraft/internal/tmux"
	"github.com/user/sessioncraft/internal/ui/components"
)

const (
	cardBorderV    = 2 // top + bottom border
	cardPaddingH   = 2 // left + right padding
	cardTitleLines = 1 // title line (for preview card)
	cardGap        = 1 // gap between cards

	headerContentLines = 1 // header content
	footerContentLines = 1 // footer content
)

// Metadata contains additional info for the preview pane
type Metadata struct {
	SessionName  string
	WindowName   string
	WindowIndex  int
	SessionIndex int
	IsActive     bool // Session is attached
	Path         string
	Uptime       string
	ClientCount  int
	WindowCount  int
	Tags         []string
}

type Model struct {
	Content       string
	ResourceUsage tmux.ResourceUsage
	Metadata      Metadata
	Active        bool
	Width         int
	Height        int
	Theme         config.ThemeConfig
}

func NewModel(theme config.ThemeConfig) Model {
	return Model{Theme: theme}
}

func (m Model) View() string {
	if !m.Active {
		return ""
	}

	if m.Width < 20 {
		m.Width = 20
	}
	if m.Height < 15 {
		m.Height = 15
	}

	contentWidth := m.Width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Theme colors
	border := lipgloss.Color(m.Theme.Border)
	accent := lipgloss.Color(m.Theme.Accent)

	// Fixed heights
	headerCardHeight := cardBorderV + headerContentLines
	footerCardHeight := cardBorderV + footerContentLines

	// Preview takes remaining space
	fixedHeight := headerCardHeight + cardGap + cardGap + footerCardHeight
	previewCardHeight := m.Height - fixedHeight
	if previewCardHeight < 5 {
		previewCardHeight = 5
	}

	// Render cards
	headerCard := m.renderHeaderCard(contentWidth, accent)
	previewCard := m.renderPreviewCard(contentWidth, previewCardHeight, m.Metadata.IsActive)
	footerCard := m.renderFooterCard(contentWidth, border)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerCard,
		previewCard,
		footerCard,
	)

	containerStyle := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		PaddingLeft(2)

	return containerStyle.Render(content)
}

func (m Model) renderHeaderCard(width int, accent lipgloss.TerminalColor) string {
	icon := "󰆍"
	if m.Metadata.SessionName == "" {
		icon = ""
	}

	identifier := m.Metadata.SessionName
	if identifier != "" && m.Metadata.SessionIndex > 0 {
		if m.Metadata.WindowName != "" {
			identifier = fmt.Sprintf("%d: %s: %d %s", m.Metadata.SessionIndex, identifier, m.Metadata.WindowIndex, m.Metadata.WindowName)
		} else {
			identifier = fmt.Sprintf("%d: %s", m.Metadata.SessionIndex, identifier)
		}
	} else if m.Metadata.WindowName != "" {
		identifier = fmt.Sprintf("%s:%d %s", identifier, m.Metadata.WindowIndex, m.Metadata.WindowName)
	}
	if identifier == "" {
		identifier = "No session selected"
	}

	var statusBadge string
	activeBg := lipgloss.Color(m.Theme.Active)
	surfaceBg := lipgloss.Color(m.Theme.Surface)
	bgColor := lipgloss.Color(m.Theme.Background)

	if m.Metadata.IsActive {
		statusBadge = lipgloss.NewStyle().
			Background(activeBg).
			Foreground(bgColor).
			Padding(0, 1).
			Bold(true).
			Render("active")
	} else {
		statusBadge = lipgloss.NewStyle().
			Background(surfaceBg).
			Foreground(lipgloss.Color(m.Theme.TextMuted)).
			Padding(0, 1).
			Render("idle")
	}

	iconStyle := lipgloss.NewStyle().Foreground(accent).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(accent).Bold(true)

	headerContent := fmt.Sprintf("%s %s  %s",
		iconStyle.Render(icon),
		nameStyle.Render(identifier),
		statusBadge,
	)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Width(width).
		Padding(0, 1)

	return cardStyle.Render(headerContent)
}

func (m Model) renderPreviewCard(width, height int, isActive bool) string {
	// Process content
	content := strings.TrimRight(m.Content, "\n\r")
	lines := strings.Split(content, "\n")

	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		lines = []string{"No preview available"}
	}

	// Use constants for maxLines calculation
	maxLines := height - cardBorderV - cardTitleLines
	if maxLines < 1 {
		maxLines = 1
	}

	textMuted := lipgloss.Color(m.Theme.TextMuted)
	if m.Theme.TextMuted == "" {
		textMuted = lipgloss.Color(m.Theme.Dimmed)
	}
	dimmedColor := lipgloss.Color(m.Theme.Dimmed)
	if m.Theme.Dimmed == "" {
		dimmedColor = textMuted
	}
	mutedStyle := lipgloss.NewStyle().Foreground(dimmedColor)
	contentStyle := lipgloss.NewStyle().Foreground(textMuted)

	overflow := 0
	if len(lines) > maxLines {
		overflow = len(lines) - maxLines
		// Reserve a line for the overflow hint when space allows.
		if maxLines > 2 {
			maxLines--
		}
	}
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	// Truncate long lines using ANSI-aware truncation
	usable := width - 4
	if usable < 1 {
		usable = 1
	}
	for i, l := range lines {
		if ansi.StringWidth(l) > usable {
			lines[i] = ansi.Truncate(l, usable, "…")
		}
	}

	renderedLines := make([]string, 0, len(lines)+1)
	if overflow > 0 && height > 4 {
		hint := fmt.Sprintf("… %d more above", overflow)
		if ansi.StringWidth(hint) > usable {
			hint = ansi.Truncate(hint, usable, "…")
		}
		renderedLines = append(renderedLines, mutedStyle.Render(hint))
	}

	for _, line := range lines {
		renderedLines = append(renderedLines, contentStyle.Render(line))
	}
	contentStr := strings.Join(renderedLines, "\n")

	// Dynamic border color
	var borderColor lipgloss.TerminalColor
	if isActive {
		borderColor = lipgloss.Color(m.Theme.Active)
	} else {
		borderColor = lipgloss.Color(m.Theme.Border)
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Height(height).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.Theme.Accent)).Bold(true)
	return cardStyle.Render(titleStyle.Render("─ Preview ─") + "\n" + contentStr)
}

func (m Model) renderFooterCard(width int, border lipgloss.TerminalColor) string {
	dimmed := lipgloss.Color(m.Theme.Dimmed)
	accent := lipgloss.Color(m.Theme.Accent)
	surface := lipgloss.Color(m.Theme.Surface)

	labelStyle := lipgloss.NewStyle().Foreground(dimmed)
	iconStyle := lipgloss.NewStyle().Foreground(accent)
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.Theme.TextMuted))

	// Uptime
	uptime := m.Metadata.Uptime
	if uptime == "" {
		uptime = "—"
	}
	uptimeStr := fmt.Sprintf("%s %s", iconStyle.Render("󱎫"), textStyle.Render(uptime))

	// CPU
	cpuPct := parsePercent(m.ResourceUsage.CPU)
	barWidth := 6
	cpuBar := components.NewProgressBar(barWidth, accent, surface)
	cpuStr := fmt.Sprintf("%s %s", labelStyle.Render("CPU"), cpuBar.Render(cpuPct))

	// Memory
	memStr := fmt.Sprintf("%s %s", labelStyle.Render("MEM"), textStyle.Render(formatMemory(m.ResourceUsage.Memory)))

	separator := labelStyle.Render(" │ ")
	usable := width - 4
	staticWidth := ansi.StringWidth(uptimeStr) + ansi.StringWidth(cpuStr) + ansi.StringWidth(memStr) + (ansi.StringWidth(separator) * 3)
	pathMaxWidth := usable - staticWidth
	if pathMaxWidth < 1 {
		pathMaxWidth = 1
	}

	// Path
	path := m.Metadata.Path
	if path == "" {
		path = "~"
	}
	path = truncatePath(path, pathMaxWidth)
	pathStr := fmt.Sprintf("%s %s", iconStyle.Render("󰉋"), textStyle.Render(path))

	footerContent := pathStr + separator + uptimeStr + separator + cpuStr + separator + memStr
	if ansi.StringWidth(footerContent) > usable {
		footerContent = ansi.Truncate(footerContent, usable, "…")
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Width(width).
		Padding(0, 1)

	return cardStyle.Render(footerContent)
}

// truncatePath intelligently truncates a path to fit within maxWidth
func truncatePath(path string, maxWidth int) string {
	// Icon (󰉋) + space takes ~3 characters
	actualMaxWidth := maxWidth - 3
	if actualMaxWidth < 1 {
		actualMaxWidth = 1
	}

	currentWidth := runewidth.StringWidth(path)
	if currentWidth <= actualMaxWidth {
		return path
	}

	// Split path into components
	parts := strings.Split(path, "/")

	// For paths starting with ~, preserve it
	prefix := ""
	if strings.HasPrefix(path, "~/") {
		prefix = "~/"
		parts = parts[1:] // Skip empty first element and ~
	} else if strings.HasPrefix(path, "/") {
		prefix = "/"
		parts = parts[1:] // Skip empty first element
	}

	if len(parts) == 0 {
		return path
	}

	// Try to preserve the last 2-3 components
	// Start from the end and build backwards
	result := parts[len(parts)-1]
	for i := len(parts) - 2; i >= 0; i-- {
		candidate := parts[i] + "/" + result
		if runewidth.StringWidth(prefix+"…/"+candidate) > actualMaxWidth {
			break
		}
		result = candidate
	}

	// Add ellipsis if we truncated
	if result != path[len(prefix):] {
		result = "…/" + result
	}

	return prefix + result
}

func parsePercent(s string) float64 {
	// Parse "42%" or "42" to 0.42
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	var pct float64
	fmt.Sscanf(s, "%f", &pct)
	return pct / 100.0
}

func formatMemory(s string) string {
	// Convert "4500KB" to "4.5MB" or similar
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "KB")
	s = strings.TrimSuffix(s, "kb")

	var kb float64
	fmt.Sscanf(s, "%f", &kb)

	if kb <= 0 {
		return s // Return original if parsing failed
	}

	mb := kb / 1024.0
	if mb >= 1024 {
		return fmt.Sprintf("%.1fGB", mb/1024.0)
	}
	return fmt.Sprintf("%.1fMB", mb)
}
