package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestProgressBar_Render(t *testing.T) {
	tests := []struct {
		name       string
		percent    float64
		width      int
		wantFilled int
	}{
		{"half", 0.5, 10, 5},
		{"full", 1.0, 10, 10},
		{"empty", 0.0, 10, 0},
		{"quarter", 0.25, 8, 2},
	}

	fill := lipgloss.Color("#cba6f7")
	empty := lipgloss.Color("#313244")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewProgressBar(tt.width, fill, empty)
			result := pb.Render(tt.percent)

			// Count filled blocks
			filled := strings.Count(result, "█")
			if filled != tt.wantFilled {
				t.Errorf("got %d filled, want %d", filled, tt.wantFilled)
			}
		})
	}
}

func TestProgressBar_Clamp(t *testing.T) {
	fill := lipgloss.Color("#cba6f7")
	empty := lipgloss.Color("#313244")
	pb := NewProgressBar(10, fill, empty)

	// Test over 100%
	result := pb.Render(1.5)
	filled := strings.Count(result, "█")
	if filled != 10 {
		t.Errorf("expected 10 filled for >100%%, got %d", filled)
	}

	// Test negative
	result = pb.Render(-0.5)
	filled = strings.Count(result, "█")
	if filled != 0 {
		t.Errorf("expected 0 filled for negative, got %d", filled)
	}
}
