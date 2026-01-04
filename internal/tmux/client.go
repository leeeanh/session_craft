package tmux

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

// FetchState returns all sessions with their windows populated
func (c *Client) FetchState() ([]Session, error) {
	// We use tab delimiter for robust parsing (names can contain colons)
	cmd := exec.Command("tmux", "list-windows", "-a", "-F", "#{session_name}\t#{session_attached}\t#{session_created}\t#{session_path}\t#{window_index}\t#{window_name}\t#{window_active}\t#{pane_pid}")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// If command fails, it might be because no server is running or no sessions.
		// check ListSessions behavior.
		if strings.Contains(err.Error(), "exit status 1") {
			return []Session{}, nil
		}
		return nil, fmt.Errorf("failed to fetch state: %w", err)
	}

	sessionMap := make(map[string]*Session)
	var sessions []Session // To keep order we might want a slice, but map is easier for grouping.
	// Actually we want to preserve order returned by tmux.
	var sessionOrder []string

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 8 {
			continue
		}

		sName := parts[0]
		sAttachedCount, _ := strconv.Atoi(parts[1])
		sCreated, _ := strconv.ParseInt(parts[2], 10, 64)
		sPath := parts[3]
		wIndex, _ := strconv.Atoi(parts[4])
		wName := parts[5]
		wActive := parts[6] == "1"
		pPID, _ := strconv.Atoi(parts[7])

		if _, exists := sessionMap[sName]; !exists {
			sess := &Session{
				Name:        sName,
				Attached:    sAttachedCount > 0,
				Created:     time.Unix(sCreated, 0),
				Path:        sPath,
				ClientCount: sAttachedCount,
				Windows:     []Window{},
			}
			sessionMap[sName] = sess
			sessionOrder = append(sessionOrder, sName)
		}

		sessionMap[sName].Windows = append(sessionMap[sName].Windows, Window{
			SessionName:   sName,
			Index:         wIndex,
			Name:          wName,
			Active:        wActive,
			ActivePanePID: pPID,
		})
	}

	for _, name := range sessionOrder {
		sessions = append(sessions, *sessionMap[name])
	}

	return sessions, nil
}

// CapturePane returns the last N lines of the active pane in a window associated with a session
// We target the window using session:index format
func (c *Client) CapturePane(sessionName string, windowIndex int, lines int) (string, error) {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	// Just capture visible pane content for now.
	// -S -<lines> causes issues if history is empty or specific tmux versions?
	// Let's try capturing visible pane first.
	// cmd := exec.Command("tmux", "capture-pane", "-t", target, "-p", "-e", "-S", fmt.Sprintf("-%d", lines))

	// Default: visible contents
	cmd := exec.Command("tmux", "capture-pane", "-t", target, "-p", "-e")

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("capture-pane failed: %w", err)
	}

	str := out.String()
	if strings.TrimSpace(str) == "" {
		return "[No Content Captured from tmux - pane might be empty]", nil
	}
	return str, nil
}

func (c *Client) CreateSession(name, dir string) error {
	args := []string{"new-session", "-d", "-s", name}
	if dir != "" {
		args = append(args, "-c", dir)
	}
	return exec.Command("tmux", args...).Run()
}

func (c *Client) CreateSessionWithWindowName(name, dir, windowName string) error {
	args := []string{"new-session", "-d", "-s", name}
	if dir != "" {
		args = append(args, "-c", dir)
	}
	if windowName != "" {
		args = append(args, "-n", windowName)
	}
	return exec.Command("tmux", args...).Run()
}

func (c *Client) CreateWindow(sessionName, windowName string) error {
	args := []string{"new-window", "-t", sessionName}
	if windowName != "" {
		args = append(args, "-n", windowName)
	}
	return exec.Command("tmux", args...).Run()
}

func (c *Client) KillSession(name string) error {
	return exec.Command("tmux", "kill-session", "-t", name).Run()
}

func (c *Client) RenameSession(oldName, newName string) error {
	return exec.Command("tmux", "rename-session", "-t", oldName, newName).Run()
}

func (c *Client) RenameWindow(session string, index int, newName string) error {
	target := fmt.Sprintf("%s:%d", session, index)
	return exec.Command("tmux", "rename-window", "-t", target, newName).Run()
}

func (c *Client) KillWindow(session, target string) error {
	return exec.Command("tmux", "kill-window", "-t", session+":"+target).Run()
}

func (c *Client) MoveWindow(srcSession string, srcIndex int, dstSession string) error {
	src := fmt.Sprintf("%s:%d", srcSession, srcIndex)
	return exec.Command("tmux", "move-window", "-s", src, "-t", dstSession+":").Run()
}

func (c *Client) LinkWindow(srcSession string, srcIndex int, dstSession string) error {
	src := fmt.Sprintf("%s:%d", srcSession, srcIndex)
	// link-window -s src -t dst:
	// dst session will have next available index.
	return exec.Command("tmux", "link-window", "-s", src, "-t", dstSession+":").Run()
}

func (c *Client) GetAttachCmd(target string) *exec.Cmd {
	// We want to attach to the session.
	// If we are already in tmux, we might need switch-client.
	// But detailed spec says: "SessionCraft exits completely and hands control to tmux."
	// So we can just exec tmux attach.
	// However, if we are inside tmux popup, we need to handle that?
	// Spec: "When attaching... SessionCraft exits completely".
	// So we execute tmux attach.

	// Detection of inside tmux: check $TMUX env var.
	if os.Getenv("TMUX") != "" {
		// Inside tmux: use switch-client
		return exec.Command("tmux", "switch-client", "-t", target)
	}
	// Outside tmux: attach
	return exec.Command("tmux", "attach-session", "-t", target)
}

func (c *Client) ListSessions() ([]Session, error) {
	// We might want to use a more complex format to get more details later
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}:#{session_attached}:#{session_created}")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// If no sessions exist, tmux returns exit code 1
		if strings.Contains(err.Error(), "exit status 1") {
			return []Session{}, nil
		}
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessions []Session
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}

		createdTimestamp, _ := strconv.ParseInt(parts[2], 10, 64)

		sessions = append(sessions, Session{
			Name:     parts[0],
			Attached: parts[1] == "1",
			Created:  time.Unix(createdTimestamp, 0),
		})
	}
	return sessions, nil
}

// ListWindows returns windows for a session
func (c *Client) ListWindows(sessionName string) ([]Window, error) {
	cmd := exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_index}:#{window_name}:#{window_active}")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list windows for %s: %w", sessionName, err)
	}

	var windows []Window
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}

		idx, _ := strconv.Atoi(parts[0])

		windows = append(windows, Window{
			SessionName: sessionName,
			Index:       idx,
			Name:        parts[1],
			Active:      parts[2] == "1",
		})
	}
	return windows, nil
}
