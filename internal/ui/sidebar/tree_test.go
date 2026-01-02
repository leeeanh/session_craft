package sidebar

import (
	"testing"

	"github.com/user/sessioncraft/internal/tmux"
)

func TestFlattenTree(t *testing.T) {
	sessions := []tmux.Session{
		{
			Name: "session1",
			Windows: []tmux.Window{
				{Name: "win1", Index: 1},
				{Name: "win2", Index: 2},
			},
		},
		{
			Name: "session2",
			Windows: []tmux.Window{
				{Name: "win3", Index: 1},
			},
		},
	}
	ghosts := []GhostNode{}

	// Test collapsed
	expanded := make(map[string]bool)
	nodes := FlattenTree(sessions, ghosts, expanded)

	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes (sessions only), got %d", len(nodes))
	}
	if nodes[0].Type != SessionNode || nodes[0].Session.Name != "session1" {
		t.Error("First node should be session1")
	}

	// Test expanded
	expanded["session1"] = true
	nodes = FlattenTree(sessions, ghosts, expanded)

	// session1 + 2 windows + session2 = 4 nodes
	if len(nodes) != 4 {
		t.Errorf("Expected 4 nodes, got %d", len(nodes))
	}
	if nodes[1].Type != WindowNode || nodes[1].Window.Name != "win1" {
		t.Error("Second node should be win1")
	}
	if nodes[3].Type != SessionNode || nodes[3].Session.Name != "session2" {
		t.Error("Fourth node should be session2")
	}
}
