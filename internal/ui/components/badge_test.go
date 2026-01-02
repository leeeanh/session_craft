package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestBadge_Render(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		bg      string
		fg      string
		wantLen int // Approximate rendered length
	}{
		{"simple", "active", "#a6e3a1", "#1e1e2e", 8},
		{"short", "3", "#cba6f7", "#1e1e2e", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBadge(tt.text, lipgloss.Color(tt.bg), lipgloss.Color(tt.fg))
			result := b.Render()
			if len(result) == 0 {
				t.Error("expected non-empty render")
			}
		})
	}
}
