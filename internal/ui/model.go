package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/user/sessioncraft/internal/config"
	"github.com/user/sessioncraft/internal/tmux"
	"github.com/user/sessioncraft/internal/ui/preview"
	"github.com/user/sessioncraft/internal/ui/search"
	"github.com/user/sessioncraft/internal/ui/sidebar"
	"github.com/user/sessioncraft/internal/ui/styles"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeInput
	ModeConfirm
	ModeSelectTarget
	ModeSearch
)

type InputPurpose int

const (
	InputNone InputPurpose = iota
	InputNewSession
	InputNewWindow
	InputRenameSession
	InputRenameWindow
	InputFork
	InputSearch
)

type ConfirmAction int

const (
	ConfirmNone ConfirmAction = iota
	ConfirmDeleteSession
	ConfirmDeleteWindow
)

type Model struct {
	sessions     []tmux.Session
	nodes        []sidebar.TreeNode
	expanded     map[string]bool
	cursor       int
	offset       int // Scroll offset
	client       *tmux.Client
	err          error
	width        int
	height       int
	preview      preview.Model
	previewReady bool // true if we have fetched data for current selection

	mode          Mode
	input         textinput.Model
	inputPurpose  InputPurpose
	confirmPrompt string
	confirmAction ConfirmAction
	confirmWindow struct {
		sessionName string
		windowIndex int
		windowName  string
	}
	confirmSessionName string
	statusMessage      string

	config     config.Config
	styles     styles.Styles
	bookmarks  []string
	sourceNode *sidebar.TreeNode // For teleportation/fork source
}

type previewMsg struct {
	content  string
	usage    tmux.ResourceUsage
	metadata preview.Metadata
}

const forkScratchWindowName = "__sessioncraft__fork__"

func NewModel() Model {
	cfg, err := config.LoadConfig()
	if err != nil {
		// Log error or just use default? Default is returned on error usually.
		cfg = config.DefaultConfig()
	}

	extraBookmarks, _ := config.LoadBookmarks()
	allBookmarks := append(cfg.Bookmarks, extraBookmarks...)

	st := styles.NewStyles(cfg.Theme)

	ti := textinput.New()
	ti.Cursor.Style = st.Accent
	ti.Prompt = "" // We render prompt manually or here

	return Model{
		client:    tmux.NewClient(),
		expanded:  make(map[string]bool),
		nodes:     []sidebar.TreeNode{},
		preview:   preview.NewModel(cfg.Theme),
		input:     ti,
		mode:      ModeNormal,
		config:    cfg,
		styles:    st,
		bookmarks: allBookmarks,
		err:       err, // Capture load error if we want to show it
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadSessions
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case ModeSelectTarget:
			switch msg.String() {
			case "esc":
				m.mode = ModeNormal
				m.sourceNode = nil
				m.statusMessage = "Cancelled move"
			case "enter":
				// Confirm move
				targetNode := m.currentNode()
				if targetNode != nil && targetNode.Type == sidebar.SessionNode && m.sourceNode != nil {
					err := m.client.MoveWindow(m.sourceNode.Window.SessionName, m.sourceNode.Window.Index, targetNode.Session.Name)
					if err != nil {
						m.statusMessage = fmt.Sprintf("Move failed: %v", err)
					} else {
						m.statusMessage = fmt.Sprintf("Moved window to %s", targetNode.Session.Name)
						cmd = m.loadSessions
					}
					m.mode = ModeNormal
					m.sourceNode = nil
				} else {
					m.statusMessage = "Select a target session"
				}
			case "j", "down":
				if m.cursor < len(m.nodes)-1 {
					m.cursor++
				}
			case "k", "up":
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case ModeInput:
			switch msg.Type {
			case tea.KeyEnter:
				// Execute action
				val := m.input.Value()
				// For search, Enter might select the first result or quit search?
				// Usually keeps filter active or attaches if single result?
				// User wants "Search mode UI", usually real-time filtering.
				// If Enter, we probably switch to Normal mode but keep filter? Or reset?
				// Let's say Enter -> Normal mode, clear filter (simple). Or keep filter to navigate results.
				// Let's keep filter if InputSearch.

				if m.inputPurpose == InputSearch {
					m.mode = ModeNormal
					m.input.Blur()
					return m, nil
				}

				if m.inputPurpose == InputNewWindow {
					node := m.currentNode()
					if node != nil && node.Type == sidebar.WindowNode {
						err := m.client.CreateWindow(node.Window.SessionName, val)
						if err != nil {
							m.statusMessage = fmt.Sprintf("Create window failed: %v", err)
						} else {
							cmd = m.loadSessions
							if val == "" {
								m.statusMessage = "Created window"
							} else {
								m.statusMessage = "Created window " + val
							}
						}
					} else {
						m.statusMessage = "Select a window to create a sibling"
					}
				} else if val != "" {
					switch m.inputPurpose {
					case InputNewSession:
						m.client.CreateSession(val, "") // Default dir
						cmd = m.loadSessions
						m.statusMessage = "Created session " + val
					case InputRenameSession:
						node := m.currentNode()
						if node != nil && node.Type == sidebar.SessionNode {
							m.client.RenameSession(node.Session.Name, val)
							cmd = m.loadSessions
							m.statusMessage = "Renamed session to " + val
						}
					case InputRenameWindow:
						node := m.currentNode()
						if node != nil && node.Type == sidebar.WindowNode {
							m.client.RenameWindow(node.Window.SessionName, node.Window.Index, val)
							cmd = m.loadSessions
							m.statusMessage = "Renamed window to " + val
						}
					case InputFork:
						srcNode := m.sourceNode
						if srcNode != nil && srcNode.Type == sidebar.WindowNode {
							// Create session
							err := m.client.CreateSessionWithWindowName(val, "", forkScratchWindowName)
							if err != nil {
								m.statusMessage = fmt.Sprintf("Fork failed: %v", err)
								break
							}

							// Link window into the new session.
							if err := m.client.LinkWindow(srcNode.Window.SessionName, srcNode.Window.Index, val); err != nil {
								m.statusMessage = fmt.Sprintf("Fork failed: %v", err)
								break
							}

							// Remove the scratch window created by new-session.
							if err := m.client.KillWindow(val, forkScratchWindowName); err != nil {
								m.statusMessage = fmt.Sprintf("Forked to %s (cleanup failed: %v)", val, err)
							} else {
								m.statusMessage = "Forked to " + val
							}
							cmd = m.loadSessions
						}
						m.sourceNode = nil
					}
				}
				m.mode = ModeNormal
				m.input.Blur()
			case tea.KeyEsc:
				// Cancel search or input
				if m.inputPurpose == InputSearch {
					// Clear search query
					m.input.SetValue("")
					m.updateNodes()
				}
				m.mode = ModeNormal
				m.input.Blur()
				m.sourceNode = nil
			default:
				m.input, cmd = m.input.Update(msg)
				// Real-time search update
				if m.inputPurpose == InputSearch {
					m.updateNodes()
				}
			}

		case ModeConfirm:
			switch msg.String() {
			case "y", "Y":
				switch m.confirmAction {
				case ConfirmDeleteSession:
					if m.confirmSessionName == "" {
						m.statusMessage = "Cannot delete session"
						break
					}
					m.client.KillSession(m.confirmSessionName)
					m.statusMessage = "Deleted session " + m.confirmSessionName
					cmd = m.loadSessions
				case ConfirmDeleteWindow:
					if m.confirmWindow.sessionName == "" {
						m.statusMessage = "Cannot delete window"
						break
					}
					target := fmt.Sprintf("%d", m.confirmWindow.windowIndex)
					if err := m.client.KillWindow(m.confirmWindow.sessionName, target); err != nil {
						m.statusMessage = fmt.Sprintf("Delete failed: %v", err)
					} else {
						m.statusMessage = "Deleted window " + m.confirmWindow.windowName
						cmd = m.loadSessions
					}
				default:
					m.statusMessage = "Nothing to confirm"
				}
				m.mode = ModeNormal
			default:
				// Cancel
				m.mode = ModeNormal
				m.statusMessage = "Cancelled"
			}

		case ModeNormal:
			switch msg.String() {
			case "esc":
				if m.inputPurpose == InputSearch {
					m.input.SetValue("")
					m.updateNodes()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "j", "down":
				if m.cursor < len(m.nodes)-1 {
					m.cursor++
					// Scroll down
					visibleLines := m.height - 5
					if m.cursor >= m.offset+visibleLines {
						m.offset++
					}
					cmd = m.updatePreview()
				}
			case "k", "up":
				if m.cursor > 0 {
					m.cursor--
					// Scroll up
					if m.cursor < m.offset {
						m.offset = m.cursor
					}
					cmd = m.updatePreview()
				}
			case "l":
				// Expand
				node := m.currentNode()
				if node != nil && node.Type == sidebar.SessionNode {
					if !m.expanded[node.Session.Name] {
						m.expanded[node.Session.Name] = true
						m.updateNodes()
					}
				}
			case "h":
				// Collapse
				node := m.currentNode()
				if node != nil {
					if node.Type == sidebar.SessionNode {
						if m.expanded[node.Session.Name] {
							m.expanded[node.Session.Name] = false
							m.updateNodes()
						}
					}
				}
			case "enter":
				// Attach
				node := m.currentNode()
				target := ""
				if node != nil {
					if node.Type == sidebar.SessionNode {
						target = node.Session.Name
					} else if node.Type == sidebar.WindowNode {
						target = fmt.Sprintf("%s:%d", node.Window.SessionName, node.Window.Index)
					} else if node.Type == sidebar.GhostNodeType {
						// Create session from ghost
						name := node.Ghost.Name
						path := node.Ghost.Path
						m.client.CreateSession(name, path)
						// Attach immediately
						target = name
					}
				}
				if target != "" {
					cmd = tea.ExecProcess(m.client.GetAttachCmd(target), func(err error) tea.Msg {
						// On finish (detach or switch success)
						return tea.Quit() // Just quit
					})
					return m, cmd
				}

			case "n":
				node := m.currentNode()
				if node != nil && node.Type == sidebar.WindowNode {
					m.mode = ModeInput
					m.inputPurpose = InputNewWindow
					m.input.Prompt = "New Window Name: "
					m.input.SetValue("")
					m.input.Focus()
					return m, textinput.Blink
				}

				m.mode = ModeInput
				m.inputPurpose = InputNewSession
				m.input.Prompt = "New Session Name: "
				m.input.SetValue("")
				m.input.Focus()
				return m, textinput.Blink

			case "r":
				node := m.currentNode()
				if node != nil {
					m.mode = ModeInput
					if node.Type == sidebar.SessionNode {
						m.inputPurpose = InputRenameSession
						m.input.Prompt = "Rename Session: "
						m.input.SetValue(node.Session.Name)
					} else if node.Type == sidebar.WindowNode {
						m.inputPurpose = InputRenameWindow
						m.input.Prompt = "Rename Window: "
						m.input.SetValue(node.Window.Name)
					}
					m.input.Focus()
					return m, textinput.Blink
				}

			case "d":
				node := m.currentNode()
				if node != nil && node.Type == sidebar.SessionNode {
					m.mode = ModeConfirm
					m.confirmAction = ConfirmDeleteSession
					m.confirmSessionName = node.Session.Name
					m.confirmPrompt = fmt.Sprintf("Delete session '%s'? [y/N]", node.Session.Name)
				} else if node != nil && node.Type == sidebar.WindowNode {
					m.mode = ModeConfirm
					m.confirmAction = ConfirmDeleteWindow
					m.confirmWindow.sessionName = node.Window.SessionName
					m.confirmWindow.windowIndex = node.Window.Index
					m.confirmWindow.windowName = node.Window.Name
					m.confirmPrompt = fmt.Sprintf("Delete window '%s'? [y/N]", node.Window.Name)
				} else {
					m.statusMessage = "Can only delete sessions"
				}

			case "m":
				node := m.currentNode()
				if node != nil && node.Type == sidebar.WindowNode {
					m.mode = ModeSelectTarget
					m.sourceNode = node
					m.statusMessage = fmt.Sprintf("Select target session for %s", node.Window.Name)
				} else {
					m.statusMessage = "Select a window to move"
				}

			case "f":
				node := m.currentNode()
				if node != nil && node.Type == sidebar.WindowNode {
					m.mode = ModeInput
					m.inputPurpose = InputFork
					m.sourceNode = node
					m.input.Prompt = "Fork to new session: "
					m.input.SetValue("")
					m.input.Focus()
					return m, textinput.Blink
				}

			case "/":
				m.mode = ModeInput
				m.inputPurpose = InputSearch
				m.input.Prompt = "/"
				m.input.SetValue("")
				m.input.Focus()
				return m, textinput.Blink

			case "R":
				node := m.currentNode()
				if node != nil && node.Type == sidebar.SessionNode {
					// Restart: Kill and Recreate
					m.client.KillSession(node.Session.Name)
					m.client.CreateSession(node.Session.Name, "")
					m.statusMessage = "Restarted session " + node.Session.Name
					cmd = m.loadSessions
				}
			}
		}
	case []tmux.Session:
		m.sessions = msg
		m.updateNodes()
		cmd = m.updatePreview() // Fetch preview for initial selection
	case previewMsg:
		m.preview.Content = msg.content
		m.preview.ResourceUsage = msg.usage
		m.preview.Metadata = msg.metadata
		m.preview.Active = true
	case error:
		m.err = msg
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		contentWidth := m.width - 2
		if contentWidth < 1 {
			contentWidth = 1
		}
		contentHeight := m.height - 2
		if contentHeight < 1 {
			contentHeight = 1
		}

		sidebarWidth := int(float64(contentWidth) * 0.3)

		m.preview.Width = contentWidth - sidebarWidth - 1
		if m.preview.Width < 0 {
			m.preview.Width = 0
		}

		m.preview.Height = contentHeight
		if m.preview.Height < 3 {
			m.preview.Height = 3
		}

		m.input.Width = sidebarWidth - 6 // Fit in sidebar
		if m.input.Width < 1 {
			m.input.Width = 1
		}
	}
	return m, cmd
}

func (m *Model) updateNodes() {
	var nodes []sidebar.TreeNode

	// Flatten sessions
	ghosts := sidebar.FilterGhostNodes(m.bookmarks, m.sessions)
	allNodes := sidebar.FlattenTree(m.sessions, ghosts, m.expanded)

	// Filter if searching
	if m.inputPurpose == InputSearch && m.input.Value() != "" {
		query := m.input.Value()
		for _, node := range allNodes {
			match := false
			var indices []int
			if node.Type == sidebar.SessionNode {
				if ok, idxs := search.Match(query, node.Session.Name); ok {
					match = true
					indices = idxs
				}
			} else if node.Type == sidebar.WindowNode {
				if ok, idxs := search.Match(query, node.Window.Name); ok {
					match = true
					indices = idxs
				}
				// Also match if specific window index? Or if session matched?
				// If session matched, maybe show all windows?
				// For now simple node-by-node match.
			} else if node.Type == sidebar.GhostNodeType {
				if ok, idxs := search.Match(query, node.Ghost.Name); ok {
					match = true
					indices = idxs
				}
			}

			if match {
				node.Matches = indices
				nodes = append(nodes, node)
			}
		}
	} else {
		nodes = allNodes
	}

	m.nodes = nodes
	// Clamp cursor
	if m.cursor >= len(m.nodes) {
		m.cursor = len(m.nodes) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m Model) currentNode() *sidebar.TreeNode {
	if m.cursor >= 0 && m.cursor < len(m.nodes) {
		return &m.nodes[m.cursor]
	}
	return nil
}

func (m Model) updatePreview() tea.Cmd {
	node := m.currentNode()
	if node == nil {
		return nil
	}

	var sessionName string
	var windowIndex int
	var panePID int
	var windowCount int
	var session *tmux.Session

	if node.Type == sidebar.SessionNode {
		sessionName = node.Session.Name
		windowCount = len(node.Session.Windows)
		session = node.Session
		// Find active window or default to 0
		found := false
		for _, w := range node.Session.Windows {
			if w.Active {
				windowIndex = w.Index
				panePID = w.ActivePanePID
				found = true
				break
			}
		}
		if !found && len(node.Session.Windows) > 0 {
			windowIndex = node.Session.Windows[0].Index
			panePID = node.Session.Windows[0].ActivePanePID
		}
	} else if node.Type == sidebar.WindowNode {
		sessionName = node.Window.SessionName
		windowIndex = node.Window.Index
		panePID = node.Window.ActivePanePID
		session = m.sessionByName(sessionName)
		if session != nil {
			windowCount = len(session.Windows)
		} else {
			windowCount = 1
		}
	} else {
		ghostPath := node.Ghost.Path
		if ghostPath == "" {
			ghostPath = "~"
		}
		return func() tea.Msg {
			return previewMsg{
				content: "[No preview available]",
				usage:   tmux.ResourceUsage{CPU: "-", Memory: "-"},
				metadata: preview.Metadata{
					Path:        ghostPath,
					Uptime:      "—",
					ClientCount: 0,
					WindowCount: 0,
					Tags:        []string{},
				},
			}
		}
	}

	sessionPath := "~"
	if session != nil && session.Path != "" {
		sessionPath = session.Path
	}
	uptime := "—"
	if session != nil && !session.Created.IsZero() {
		uptime = formatUptime(session.Created)
	}
	clientCount := 0
	if session != nil {
		clientCount = session.ClientCount
	}

	return func() tea.Msg {
		// Run CapturePane and GetResourceUsage
		content, err := m.client.CapturePane(sessionName, windowIndex, 20)

		if err != nil {
			content = fmt.Sprintf("Error capturing pane: %v", err)
		}

		usage, err := tmux.GetResourceUsage(panePID)
		if err != nil {
			usage = tmux.ResourceUsage{CPU: "-", Memory: "-"}
		}

		return previewMsg{
			content: content,
			usage:   usage,
			metadata: preview.Metadata{
				Path:        sessionPath,
				Uptime:      uptime,
				ClientCount: clientCount,
				WindowCount: windowCount,
				Tags:        []string{},
			},
		}
	}
}

func (m Model) sessionByName(name string) *tmux.Session {
	for i := range m.sessions {
		if m.sessions[i].Name == name {
			return &m.sessions[i]
		}
	}
	return nil
}

func formatUptime(created time.Time) string {
	if created.IsZero() {
		return "—"
	}
	elapsed := time.Since(created)
	if elapsed < 0 {
		elapsed = 0
	}

	totalMinutes := int(elapsed.Minutes())
	if totalMinutes < 1 {
		seconds := int(elapsed.Seconds())
		if seconds < 1 {
			return "0s"
		}
		return fmt.Sprintf("%ds", seconds)
	}

	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	days := hours / 24
	hours = hours % 24

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", minutes)
}

func (m Model) View() string {
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.config.Theme.Danger)).
			Padding(1, 2)
		return errorStyle.Render(fmt.Sprintf("Error: %v\nPress q to quit.", m.err))
	}

	if len(m.nodes) == 0 && len(m.sessions) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.config.Theme.Border)).
			Padding(1, 2)
		return emptyStyle.Render("No sessions found.\nPress n to create one, q to quit.")
	}

	contentWidth := m.width - 2
	if contentWidth < 1 {
		contentWidth = 1
	}
	contentHeight := m.height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	sidebarWidth := int(float64(contentWidth) * 0.3)
	previewWidth := contentWidth - sidebarWidth - 1
	if previewWidth < 1 {
		previewWidth = 1
	}
	previewContentWidth := previewWidth - 4
	if previewContentWidth < 1 {
		previewContentWidth = 1
	}
	footerHeight := m.preview.InfoCardHeight(previewContentWidth)

	// Layout: Sidebar (Left) | Preview (Right)
	sidebarContent := m.renderSidebar(sidebarWidth, contentHeight, footerHeight)

	previewModel := m.preview
	previewModel.Width = previewWidth
	if previewModel.Width < 1 {
		previewModel.Width = 1
	}
	previewModel.Height = contentHeight
	previewContent := previewModel.View()

	// Join horizontally with top alignment for consistent borders
	inner := lipgloss.JoinHorizontal(lipgloss.Top, sidebarContent, previewContent)

	// Outer frame with rounded border that fits within terminal
	// Use Place to center the content
	frameStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.config.Theme.Border)).
		Width(contentWidth).
		Height(contentHeight)

	return frameStyle.Render(inner)
}

func (m Model) renderSidebar(sidebarWidth, totalHeight, footerHeightTarget int) string {
	usable := sidebarWidth - 4
	if usable < 1 {
		usable = 1
	}

	// Header with mode indicator
	headerText := "󱫋 SESSIONCRAFT"
	modeText := m.getModeLabel()
	headerStyle := m.styles.Header.Width(sidebarWidth - 4)
	header := headerStyle.Render(headerText + " [" + modeText + "]")

	// Footer dock and status area (bottom section)
	footerDock := m.renderFooterDock(usable)
	statusArea := m.renderStatusArea(usable)

	var footer strings.Builder
	footer.WriteString(footerDock)
	if statusArea != "" {
		footer.WriteString("\n")
		footer.WriteString(statusArea)
	}

	// Calculate heights
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer.String())
	if footerHeightTarget > footerHeight {
		footerHeight = footerHeightTarget
	}

	// Account for border overhead (right border = 1 char width, no height impact)
	// Available height for tree nodes
	availableHeight := totalHeight - headerHeight - 1 - footerHeight
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Render tree nodes
	var treeContent strings.Builder
	end := m.offset + availableHeight
	if end > len(m.nodes) {
		end = len(m.nodes)
	}

	for i := m.offset; i < end; i++ {
		node := m.nodes[i]
		isSelected := m.cursor == i
		line := m.renderTreeNode(node, isSelected, usable)
		treeContent.WriteString(line + "\n")
	}

	treeStr := strings.TrimRight(treeContent.String(), "\n")

	// Build top content (header + tree)
	var topContent strings.Builder
	topContent.WriteString(header)
	topContent.WriteString("\n")
	if treeStr != "" {
		topContent.WriteString(treeStr)
	}

	// Use lipgloss.JoinVertical with bottom placement for footer
	// Calculate inner height (excluding border)
	innerHeight := totalHeight

	// Place top content at top, footer at bottom using lipgloss.Place
	topSection := topContent.String()
	bottomSection := footer.String()

	// Combine with proper spacing using Place
	topHeight := lipgloss.Height(topSection)
	bottomHeight := lipgloss.Height(bottomSection)
	middlePadding := innerHeight - topHeight - bottomHeight
	if middlePadding < 0 {
		middlePadding = 0
	}

	var content strings.Builder
	content.WriteString(topSection)
	for i := 0; i < middlePadding; i++ {
		content.WriteString("\n")
	}
	content.WriteString(bottomSection)

	// Apply sidebar style with fixed height
	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(totalHeight).
		PaddingLeft(1).
		PaddingRight(1).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color(m.config.Theme.Border))

	return sidebarStyle.Render(content.String())
}

func (m Model) renderTreeNode(node sidebar.TreeNode, isSelected bool, maxWidth int) string {
	var parts []string

	// 1. Accent border
	borderChar := m.getAccentBorder(node, isSelected)
	parts = append(parts, borderChar)

	// 2. Icon
	icon := m.getNodeIcon(node)
	parts = append(parts, icon)

	// 3. Name with optional highlighting
	name := m.getNodeName(node)
	styledName := m.styleNodeName(name, node, isSelected)
	parts = append(parts, styledName)

	// 4. Badges
	badges := m.getNodeBadges(node)
	if badges != "" {
		parts = append(parts, badges)
	}

	line := strings.Join(parts, " ")

	// Truncate if needed
	if ansi.StringWidth(line) > maxWidth {
		line = ansi.Truncate(line, maxWidth, "…")
	}

	return line
}

func (m Model) getAccentBorder(node sidebar.TreeNode, isSelected bool) string {
	if isSelected {
		return m.styles.SelectedBorder.Render("▌")
	}

	switch node.Type {
	case sidebar.SessionNode:
		if node.Session.Attached {
			return m.styles.ActiveBorder.Render("┃")
		}
		return m.styles.IdleBorder.Render("┆")
	case sidebar.WindowNode:
		return m.styles.Dimmed.Render(" ")
	case sidebar.GhostNodeType:
		return m.styles.GhostBorder.Render("┊")
	}
	return " "
}

func (m Model) getNodeIcon(node sidebar.TreeNode) string {
	switch node.Type {
	case sidebar.SessionNode:
		return SessionStateIcon(node.Session.Attached)
	case sidebar.WindowNode:
		icon := WindowIcon(node.Window.Name)
		if icon == "" {
			return node.Prefix
		}
		return node.Prefix + icon
	case sidebar.GhostNodeType:
		return GhostIcon()
	}
	return ""
}

func (m Model) getNodeName(node sidebar.TreeNode) string {
	switch node.Type {
	case sidebar.SessionNode:
		return node.Session.Name
	case sidebar.WindowNode:
		return node.Window.Name
	case sidebar.GhostNodeType:
		return node.Ghost.Name
	}
	return ""
}

func (m Model) styleNodeName(name string, node sidebar.TreeNode, isSelected bool) string {
	var baseStyle lipgloss.Style

	if isSelected {
		baseStyle = m.styles.Selected
	} else {
		switch node.Type {
		case sidebar.SessionNode:
			if node.Session.Attached {
				baseStyle = m.styles.Normal
			} else {
				baseStyle = m.styles.Dimmed
			}
		case sidebar.WindowNode:
			if node.Window.Active {
				baseStyle = m.styles.TextMuted
			} else {
				baseStyle = m.styles.Dimmed
			}
		case sidebar.GhostNodeType:
			baseStyle = m.styles.GhostBorder
		default:
			baseStyle = m.styles.Normal
		}
	}

	// Apply match highlighting
	if len(node.Matches) > 0 {
		return highlightString(name, node.Matches, m.styles.Match, baseStyle)
	}

	return baseStyle.Render(name)
}

func (m Model) getNodeBadges(node sidebar.TreeNode) string {
	var badges []string

	switch node.Type {
	case sidebar.SessionNode:
		if node.Session.Attached {
			badges = append(badges, m.styles.BadgeActive.Render("active"))
		}
		// Window count badge
		countBadge := fmt.Sprintf("%d", len(node.Session.Windows))
		badges = append(badges, m.styles.BadgeCount.Render(countBadge))

	case sidebar.WindowNode:
		if node.Window.Active {
			badges = append(badges, m.styles.BadgeProcess.Render("󰌽"))
		}
	}

	return strings.Join(badges, " ")
}

func (m Model) getModeLabel() string {
	switch m.mode {
	case ModeSearch:
		return "Search"
	case ModeConfirm:
		return "Confirm"
	case ModeSelectTarget:
		return "Select"
	case ModeInput:
		return "Input"
	default:
		return "Normal"
	}
}

func (m Model) renderFooterDock(maxWidth int) string {
	var pills []string

	switch m.mode {
	case ModeNormal:
		pills = []string{
			m.renderPill("Enter", "Attach"),
			m.renderPill("n", "New"),
			m.renderPill("m", "Move"),
			m.renderPill("d", "Delete"),
			m.renderPill("/", "Search"),
			m.renderPill("q", "Quit"),
		}
	case ModeSearch:
		pills = []string{
			m.renderPill("Esc", "Cancel"),
			m.renderPill("Enter", "Select"),
			m.renderPill("↑↓", "Navigate"),
		}
	case ModeConfirm:
		// Danger styling
		pills = []string{
			m.styles.BadgeDanger.Render("y") + " " + m.styles.TextDanger.Render("Confirm"),
			m.renderPill("n", "Cancel"),
		}
	case ModeSelectTarget:
		pills = []string{
			m.renderPill("↑↓", "Choose"),
			m.renderPill("Enter", "Confirm"),
			m.renderPill("Esc", "Cancel"),
		}
	case ModeInput:
		pills = []string{
			m.renderPill("Enter", "Submit"),
			m.renderPill("Esc", "Cancel"),
		}
	}

	dock := strings.Join(pills, " ")

	// Truncate if too long
	if ansi.StringWidth(dock) > maxWidth {
		dock = ansi.Truncate(dock, maxWidth, "…")
	}

	return dock
}

func (m Model) renderPill(key, label string) string {
	return m.styles.PillKey.Render(key) + " " + m.styles.PillLabel.Render(label)
}

func (m Model) renderStatusArea(maxWidth int) string {
	switch m.mode {
	case ModeInput:
		return m.input.View()
	case ModeConfirm:
		style := lipgloss.NewStyle().
			Background(lipgloss.Color(m.config.Theme.Surface)).
			Foreground(lipgloss.Color(m.config.Theme.Danger)).
			Bold(true).
			Padding(0, 1)
		prompt := m.confirmPrompt
		if ansi.StringWidth(prompt) > maxWidth {
			prompt = ansi.Truncate(prompt, maxWidth, "…")
		}
		return style.Render("⚠ " + prompt)
	default:
		if m.statusMessage != "" {
			msg := m.statusMessage
			if ansi.StringWidth(msg) > maxWidth {
				msg = ansi.Truncate(msg, maxWidth, "…")
			}
			return m.styles.TextActive.Render(msg)
		}
	}
	return ""
}

func (m Model) loadSessions() tea.Msg {
	sessions, err := m.client.FetchState()
	if err != nil {
		return err
	}

	// Auto-expand all sessions by default
	for _, s := range sessions {
		m.expanded[s.Name] = true
	}

	return sessions
}
