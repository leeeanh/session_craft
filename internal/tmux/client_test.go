package tmux

import (
	"testing"
)

func TestListSessions(t *testing.T) {
	client := NewClient()
	sessions, err := client.ListSessions()
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	found := false
	for _, s := range sessions {
		if s.Name == "sessioncraft-test" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find session 'sessioncraft-test', but found %v", sessions)
	}
}
