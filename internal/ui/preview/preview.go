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

// Metadata contains additional info for the preview pane
type Metadata struct {
	Path        string
	Uptime      string
	ClientCount int
	WindowCount int
	Tags        []string
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
		// Render placeholders to keep panel height consistent on first load.
		if m.ResourceUsage.CPU == "" {
			m.ResourceUsage.CPU = "-"
		}
		if m.ResourceUsage.Memory == "" {
			m.ResourceUsage.Memory = "-"
		}
		if m.Metadata.Path == "" {
			m.Metadata.Path = "~"
		}
		if m.Metadata.Uptime == "" {
			m.Metadata.Uptime = "—"
		}
	}

	// Calculate dimensions
	contentWidth := m.Width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Theme colors
	border := lipgloss.Color(m.Theme.Border)
	dimmed := lipgloss.Color(m.Theme.Dimmed)
	accent := lipgloss.Color(m.Theme.Accent)
	surface := lipgloss.Color(m.Theme.Surface)

	// Metrics card
	metricsCard := m.renderMetricsCard(contentWidth, border, dimmed, accent, surface)
	metricsHeight := lipgloss.Height(metricsCard)

	// Info card (will be at bottom)
	infoCard := m.renderInfoCard(contentWidth, border, dimmed, accent)
	infoHeight := lipgloss.Height(infoCard)

	// Calculate available height for preview card
	// Account for metrics, info, and spacing between cards (2 lines)
	baseAvailable := m.Height - metricsHeight - infoHeight - 2
	previewHeight := baseAvailable
	if previewHeight < 3 {
		previewHeight = 3
	}
	previewCard := m.renderPreviewCard(contentWidth, previewHeight, border, dimmed)

	// Build content with info card aligned to bottom
	var content strings.Builder
	content.WriteString(metricsCard)
	content.WriteString("\n")
	content.WriteString(previewCard)

	// Calculate filler lines between preview and info to reach target height
	filler := m.Height - (metricsHeight + 1 + lipgloss.Height(previewCard) + 1 + infoHeight)
	if filler < 0 {
		filler = 0
	}
	for i := 0; i < filler; i++ {
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(infoCard)

	// Wrap in a container
	containerStyle := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		PaddingLeft(2)

	return containerStyle.Render(content.String())
}

func (m Model) InfoCardHeight(width int) int {
	if width < 1 {
		width = 1
	}
	border := lipgloss.Color(m.Theme.Border)
	dimmed := lipgloss.Color(m.Theme.Dimmed)
	accent := lipgloss.Color(m.Theme.Accent)
	infoCard := m.renderInfoCard(width, border, dimmed, accent)
	return lipgloss.Height(infoCard)
}

func (m Model) renderMetricsCard(width int, border, dimmed, accent, surface lipgloss.TerminalColor) string {
	// Parse CPU percentage (returns "3.2%" -> 0.032)
	cpuPct := parsePercent(m.ResourceUsage.CPU)

	// Create CPU progress bar
	barWidth := 10
	cpuBar := components.NewProgressBar(barWidth, accent, surface)

	labelStyle := lipgloss.NewStyle().Foreground(dimmed)

	// CPU as percentage with progress bar
	cpuStr := fmt.Sprintf("CPU %s %s", cpuBar.Render(cpuPct), m.ResourceUsage.CPU)

	// Memory as raw value (KB -> MB conversion for display)
	memStr := fmt.Sprintf("MEM %s", formatMemory(m.ResourceUsage.Memory))

	// Stats
	clientStr := fmt.Sprintf("󰍹 %d", m.Metadata.ClientCount)
	windowStr := fmt.Sprintf("󱂬 %d", m.Metadata.WindowCount)

	metrics := fmt.Sprintf("%s  │  %s  │  %s  │  %s",
		labelStyle.Render(cpuStr),
		labelStyle.Render(memStr),
		labelStyle.Render(clientStr),
		labelStyle.Render(windowStr),
	)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Width(width).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().Foreground(accent).Bold(true)
	return cardStyle.Render(titleStyle.Render("─ Metrics ─") + "\n" + metrics)
}

func (m Model) renderPreviewCard(width, height int, border, dimmed lipgloss.TerminalColor) string {
	// Process content
	content := strings.TrimRight(m.Content, "\n\r")
	lines := strings.Split(content, "\n")

	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		lines = []string{"No preview available"}
	}

	// Take last N lines (tail) so recent output is visible.
	maxLines := height - 3
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
	content = strings.Join(renderedLines, "\n")

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Width(width).
		Height(height).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.Theme.Accent)).Bold(true)
	return cardStyle.Render(titleStyle.Render("─ Preview ─") + "\n" + content)
}

func (m Model) renderInfoCard(width int, border, dimmed, accent lipgloss.TerminalColor) string {
	iconStyle := lipgloss.NewStyle().Foreground(accent)
	textColor := m.Theme.TextMuted
	if textColor == "" {
		textColor = m.Theme.Foreground
	}
	if textColor == "" {
		textColor = m.Theme.Dimmed
	}
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textColor))

	// Calculate available width for content (subtract padding and border)
	contentWidth := width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Get uptime
	uptime := m.Metadata.Uptime
	if uptime == "" {
		uptime = "—"
	}
	uptimeStr := fmt.Sprintf("%s %s", iconStyle.Render("󱎫"), textStyle.Render(uptime))

	// Tags as badges
	var tags []string
	for _, tag := range m.Metadata.Tags {
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(m.Theme.Surface)).
			Foreground(lipgloss.Color(m.Theme.Foreground)).
			Padding(0, 1).
			Render(tag)
		tags = append(tags, badge)
	}
	tagStr := strings.Join(tags, " ")
	if tagStr == "" {
		tagStr = textStyle.Render("—")
	}

	// Calculate fixed-width elements (separators + uptime + tags)
	// "  │  " is 5 characters each, uptime icon + space is ~3, tag section varies
	uptimeWidth := runewidth.StringWidth(uptimeStr)
	tagWidth := runewidth.StringWidth(tagStr)
	separatorWidth := 10 // "  │  " twice
	fixedWidth := uptimeWidth + tagWidth + separatorWidth

	// Calculate available width for path
	pathMaxWidth := contentWidth - fixedWidth
	if pathMaxWidth < 10 {
		pathMaxWidth = 10
	}

	// Get and truncate path
	path := m.Metadata.Path
	if path == "" {
		path = "~"
	}
	path = truncatePath(path, pathMaxWidth)
	pathStr := fmt.Sprintf("%s %s", iconStyle.Render("󰉋"), textStyle.Render(path))

	info := fmt.Sprintf("%s  │  %s  │  %s", pathStr, uptimeStr, tagStr)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Width(width).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().Foreground(accent).Bold(true)
	return cardStyle.Render(titleStyle.Render("─ Info ─") + "\n" + info)
}

// truncatePath intelligently truncates a path to fit within maxWidth
func truncatePath(path string, maxWidth int) string {
	// Icon (󰉋) + space takes ~3 characters
	actualMaxWidth := maxWidth - 3
	if actualMaxWidth < 5 {
		actualMaxWidth = 5
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
