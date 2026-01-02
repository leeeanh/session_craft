package tmux

import "time"

// Session represents a tmux session
type Session struct {
	Name        string
	Windows     []Window
	Attached    bool
	Created     time.Time
	Path        string
	ClientCount int
}

// Window represents a tmux window
type Window struct {
	Name          string
	Index         int
	Active        bool
	SessionName   string
	Layout        string
	ActivePanePID int
}
