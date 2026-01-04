# SessionCraft

A Go + Bubble Tea TUI for managing `tmux` sessions as a project-centric dashboard: browse sessions/windows, preview pane content, and perform common actions (attach/create/rename/delete/move/fork) without leaving the keyboard.

For the full product/design spec, see `SPEC.md`.

## Features

- Session + window tree with expand/collapse (`h`/`l`)
- Preview pane: `tmux capture-pane` output (ANSI preserved)
- Lightweight metadata + resource usage (`ps` RSS/%CPU for active pane PID)
- Create sessions and windows, rename, delete (with confirmation)
- Move a window to another session
- Fork a window into a new session (links the window; does not duplicate processes)
- Bookmarks (“ghost nodes”): show project directories without a running session; `Enter` creates a session named after the folder and attaches
- Subsequence search over sessions/windows/bookmarks (`/`)

## Requirements

- `tmux` on your `PATH` (inside tmux, SessionCraft uses `tmux switch-client`; outside it uses `tmux attach-session`)
- `ps` on your `PATH` (for resource usage)
- Go toolchain (to build from source; see `go.mod`)
- Optional: a Nerd Font for nicer glyphs/icons

## Install / Build

From the repo root:

```bash
go build -o sessioncraft ./cmd/sessioncraft
```

## Run

```bash
./sessioncraft
```

### tmux popup binding (optional)

Add to your `~/.tmux.conf`:

```tmux
bind-key s display-popup -E -w 80% -h 80% "sessioncraft"
```

## Keybindings

- `j`/`k` (or arrows): move selection
- `h`/`l`: collapse/expand a session
- `Enter`: attach (or create+attach from a bookmark)
- `n`: new session (or new window when a window is selected)
- `r`: rename session/window
- `d`: delete session/window (with confirmation)
- `m`: move a window to another session
- `f`: fork a window into a new session
- `R`: restart a session (kill + recreate)
- `/`: search
- `Esc`: cancel input / clear search
- `q` / `Ctrl+C`: quit

## Configuration

SessionCraft reads config from:

- `~/.config/sessioncraft/config.yml`
- `~/.config/sessioncraft/bookmarks.yml` (merged with `config.yml` bookmarks)

Example `config.yml`:

```yaml
preview:
  lines: 20

theme:
  background: "#1a1b26"
  foreground: "#c0caf5"
  accent: "#7aa2f7"
  border: "#414868"
  dimmed: "#787fa0"
  mantle: "#1f2335"
  active: "#9ece6a"
  warning: "#e0af68"
  danger: "#f7768e"
  surface: "#24283b"
  text_muted: "#787fa0"

bookmarks:
  - ~/projects/webapp
  - ~/dotfiles
```

Example `bookmarks.yml`:

```yaml
bookmarks:
  - ~/projects/api
```

## Development

```bash
go test ./...
```

Note: several tests in `internal/tmux` are integration-style and will call `tmux` (creating/killing sessions). `internal/tmux/client_test.go` currently expects a session named `sessioncraft-test` to already exist.

