package sidebar

import (
	"path/filepath"

	"github.com/user/sessioncraft/internal/tmux"
)

// ... existing code ...

// FilterGhostNodes returns bookmarks that do not have an active session
// Matching is done by name (basename of path) vs session name
func FilterGhostNodes(bookmarks []string, sessions []tmux.Session) []GhostNode {
	activeMap := make(map[string]bool)
	for _, s := range sessions {
		activeMap[s.Name] = true
	}

	var ghosts []GhostNode
	for _, path := range bookmarks {
		name := filepath.Base(path)
		if !activeMap[name] {
			ghosts = append(ghosts, GhostNode{
				Name: name,
				Path: path,
			})
		}
	}
	return ghosts
}

type NodeType int

const (
	SessionNode NodeType = iota
	WindowNode
	GhostNodeType
)

// GhostNode represents a bookmarked directory without a session
type GhostNode struct {
	Name string
	Path string
}

type TreeNode struct {
	Type    NodeType
	Session *tmux.Session
	Window  *tmux.Window
	Ghost   *GhostNode
	Depth   int
	Prefix  string // Visual prefix (e.g. "│  ├─")
	Matches []int  // Indices of matched characters for search highlighting
}

// FlattenTree converts the hierarchy into a flat list for rendering
func FlattenTree(sessions []tmux.Session, ghosts []GhostNode, expanded map[string]bool) []TreeNode {
	var nodes []TreeNode

	// Iterating sessions
	for i := range sessions {
		sess := &sessions[i]

		// Session Node (Root)
		// We can add a root marker if we want, but usually roots are flush left.
		// Or we can treat them as children of a virtual root?
		// Let's keep them flush left for now, or maybe just ` `
		nodes = append(nodes, TreeNode{
			Type:    SessionNode,
			Session: sess,
			Depth:   0,
			Prefix:  "",
		})

		// Windows
		// Always expand? The user requested "always expand".
		// We should respect the expanded map, but we will pre-fill it in model.
		// However, if we want to FORCE it here, we could.
		// Better to respect the map so toggling still works if they really want to collapse.
		if expanded[sess.Name] {
			for j := range sess.Windows {
				win := &sess.Windows[j]
				isLastWindow := j == len(sess.Windows)-1

				// Calculate prefix
				// If session is NOT last, we draw vertical line for next sessions?
				// Actually, roots usually don't connect to each other in file trees, they are separate items.
				// But windows CONNECT to session.

				var prefix string
				if isLastWindow {
					prefix = " └─ "
					// If we were connecting roots, we'd add vertical bar if not last session.
					// But standard is roots are independent.
				} else {
					prefix = " ├─ "
				}

				nodes = append(nodes, TreeNode{
					Type:   WindowNode,
					Window: win,
					Depth:  1, // Depth implies indent, but we use Prefix now
					Prefix: prefix,
				})
			}
		}
	}

	for i := range ghosts {
		ghost := &ghosts[i]
		nodes = append(nodes, TreeNode{
			Type:   GhostNodeType,
			Ghost:  ghost,
			Depth:  0,
			Prefix: "",
		})
	}

	return nodes
}
