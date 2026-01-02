package tmux

import (
	"os/exec"
	"testing"
)

func TestActions(t *testing.T) {
	client := NewClient()
	sessionName := "action-test"

	// Ensure cleanup
	defer exec.Command("tmux", "kill-session", "-t", sessionName).Run()
	defer exec.Command("tmux", "kill-session", "-t", "renamed-test").Run()

	// 1. Create
	err := client.CreateSession(sessionName, "")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Verify creation
	sessions, _ := client.ListSessions()
	found := false
	for _, s := range sessions {
		if s.Name == sessionName {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Session not found after creation")
	}

	// 2. Rename
	newName := "renamed-test"
	err = client.RenameSession(sessionName, newName)
	if err != nil {
		t.Fatalf("RenameSession failed: %v", err)
	}

	sessions, _ = client.ListSessions()
	found = false
	for _, s := range sessions {
		if s.Name == newName {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Session not found after rename")
	}

	// 3. Kill
	err = client.KillSession(newName)
	if err != nil {
		t.Fatalf("KillSession failed: %v", err)
	}

	sessions, _ = client.ListSessions()
	for _, s := range sessions {
		if s.Name == newName {
			t.Fatal("Session found after kill")
		}
	}
}
