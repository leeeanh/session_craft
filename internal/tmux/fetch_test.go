package tmux

import (
	"os/exec"
	"testing"
)

func TestFetchState(t *testing.T) {
	// Setup: create session with 2 windows
	exec.Command("tmux", "new-session", "-d", "-s", "fetch-test", "-n", "window1").Run()
	exec.Command("tmux", "new-window", "-t", "fetch-test", "-n", "window2").Run()
	defer exec.Command("tmux", "kill-session", "-t", "fetch-test").Run()

	client := NewClient()
	sessions, err := client.FetchState()
	if err != nil {
		t.Fatalf("FetchState failed: %v", err)
	}

	var foundSession *Session
	for i := range sessions {
		if sessions[i].Name == "fetch-test" {
			foundSession = &sessions[i]
			break
		}
	}

	if foundSession == nil {
		t.Fatal("Session 'fetch-test' not found")
	}

	if len(foundSession.Windows) != 2 {
		t.Errorf("Expected 2 windows, got %d", len(foundSession.Windows))
	}
}
