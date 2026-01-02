package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestCard_Render(t *testing.T) {
	borderColor := lipgloss.Color("#45475a")
	titleColor := lipgloss.Color("#6c7086")

	card := NewCard("Metrics", 40, borderColor, titleColor)
	result := card.Render("CPU: 50%")

	// Should contain rounded border characters
	if !strings.Contains(result, "╭") {
		t.Error("expected rounded top-left corner")
	}
	if !strings.Contains(result, "Metrics") {
		t.Error("expected title in output")
	}
	if !strings.Contains(result, "CPU: 50%") {
		t.Error("expected content in output")
	}
}

func TestCard_RenderWithoutTitle(t *testing.T) {
	borderColor := lipgloss.Color("#45475a")
	titleColor := lipgloss.Color("#6c7086")

	card := NewCard("", 40, borderColor, titleColor)
	result := card.Render("Some content")

	if !strings.Contains(result, "╭") {
		t.Error("expected rounded border even without title")
	}
	if !strings.Contains(result, "Some content") {
		t.Error("expected content in output")
	}
}
