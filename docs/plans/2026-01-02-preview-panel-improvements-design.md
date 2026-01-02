# Preview Panel & Sidebar Footer Improvements

**Date:** 2026-01-02
**Status:** Approved

---

## Overview

This design addresses rendering bugs and visual improvements for SessionCraft's preview panel and sidebar footer alignment.

---

## Phase 1: Bug Fixes

### 1.1 Sidebar Footer Alignment

**Problem:** Footer floats in the middle with empty lines between it and the bottom border.

**Root Cause:** Manual padding calculation in `renderSidebar()` doesn't account for border overhead (2 lines for top/bottom). The redundant `footerPadding` loop adds extra blank lines.

**Solution:**
- Replace manual newline padding with `lipgloss.Place(width, height, lipgloss.Left, lipgloss.Bottom, content)`
- This anchors content to bottom-left automatically
- Remove the redundant `footerPadding` loop

**Files:** `internal/ui/model.go` (renderSidebar function)

---

### 1.2 Line Truncation (ANSI-aware)

**Problem:** Wide characters (unicode, emojis, CJK) and ANSI escape sequences get cut mid-character, causing display corruption.

**Root Cause:** `runewidth.Truncate()` doesn't handle ANSI escape sequences. Color codes like `\x1b[32m` have zero display width but consume string bytes. Truncation can cut mid-sequence.

**Solution:**
- Replace `runewidth.Truncate()` with `ansi.Truncate()` from `github.com/charmbracelet/x/ansi`
- This is ANSI-aware and handles escape sequences correctly
- Already imported in model.go

**Code Change:**
```go
// Before
lines[i] = runewidth.Truncate(l, usable, "...")

// After
lines[i] = ansi.Truncate(l, usable, "...")
```

**Files:** `internal/ui/preview/preview.go` (renderPreviewCard function)

---

### 1.3 Height/Width Calculation

**Problem:** Content overflows card boundaries or leaves too much empty space. Preview doesn't adapt correctly to terminal size.

**Root Cause:**
- Magic number `-2` doesn't account for all spacing
- `contentWidth := m.Width - 4` assumes fixed padding
- `maxLines` calculation doesn't match actual card chrome

**Solution:**
Define explicit constants for card chrome:
```go
const (
    cardBorderHeight = 2  // top + bottom border
    cardPaddingHeight = 0 // vertical padding (if any)
    cardTitleHeight = 1   // title line
    cardGap = 1           // gap between cards
)

cardChrome := cardBorderHeight + cardPaddingHeight + cardTitleHeight
totalChrome := (cardChrome * 3) + (cardGap * 2)  // 3 cards, 2 gaps
availableForContent := m.Height - totalChrome
```

Height distribution:
- Metrics card: fixed 3 lines content
- Info card: fixed 2 lines content
- Preview card: remaining space (availableForContent - 5)

**Files:** `internal/ui/preview/preview.go` (View function)

---

## Phase 2: Visual Enhancement

### 2.1 New Layout Structure

Restructure from current stacked cards to **Header -> Hero -> Footer**:

```
+--------------------------------------------+
|  Header Card (Identity & Status)           |
+--------------------------------------------+
                    |
                    v
+--------------------------------------------+
|                                            |
|  Hero Card (Preview Content)               |
|  - Maximized space                         |
|  - Terminal buffer display                 |
|                                            |
+--------------------------------------------+
                    |
                    v
+--------------------------------------------+
|  Footer Card (Technical Foundation)        |
+--------------------------------------------+
```

---

### 2.2 Header Card Design

**Purpose:** Identity nameplate - instant recognition of current session/window

**Content:**
- Session icon (terminal or folder)
- Session name : window index + window name (e.g., `development:2 backend-api`)
- Status badges: `[active]`, `[running]`, custom tags from `.sessioncraft.yml`

**Styling:**
- High-contrast background (slightly lighter than app background)
- Lavender rounded border
- Bold session name in accent color
- Running indicator dot (green pulse effect via color)

**Height:** Fixed, 1 line of content + card chrome

---

### 2.3 Hero Card Design (Preview)

**Purpose:** The "work context" - see terminal content without switching sessions

**Content:**
- Last N lines of pane capture (tail, so recent output visible)
- ANSI colors preserved
- Overflow hint when content exceeds visible area

**Styling:**
- Largest card, takes remaining vertical space
- Subtle dimmed border (content is the focus)
- Dynamic border: green when session is active/attached
- Minimal chrome to maximize content area

**Height:** Dynamic, fills available space

---

### 2.4 Footer Card Design

**Purpose:** Technical foundation - dense utility information

**Content (left to right):**
- Path: `folder-icon ~/dev/project` (truncated intelligently)
- Uptime: `clock-icon 3h 20m`
- CPU: `cpu-icon` + mini progress bar + percentage
- Memory: `mem-icon` + formatted value (e.g., `128MB`)

**Styling:**
- Wide, thin profile
- Dimmed text color
- Nerd Font icons for density
- Capsule/pill badges for metrics
- Separator: `|` between sections

**Height:** Fixed, 1 line of content + card chrome

---

### 2.5 Spacing & Gaps

- **Between cards:** 1 empty line
- **Card internal padding:** 0 vertical, 1 horizontal
- **Border style:** Rounded (`lipgloss.RoundedBorder()`)

---

### 2.6 Color Scheme (Obsidian Theme)

| Element | Color | Hex |
|---------|-------|-----|
| Header border | Lavender (accent) | `#b4befe` |
| Hero border (default) | Dimmed | `#6c7086` |
| Hero border (active) | Green (active) | `#a6e3a1` |
| Footer border | Border | `#313244` |
| Header background | Surface | `#313244` |
| Text (primary) | Foreground | `#cdd6f4` |
| Text (muted) | TextMuted | `#a6adc8` |
| Text (dimmed) | Dimmed | `#6c7086` |

---

## Implementation Order

1. **Fix sidebar footer alignment** (model.go)
2. **Fix line truncation** (preview.go)
3. **Fix height/width calculations** (preview.go)
4. **Restructure to Header/Hero/Footer** (preview.go)
5. **Apply new card styling** (preview.go, styles.go)
6. **Add dynamic active border** (preview.go)

---

## Testing Checklist

- [ ] Footer sticks to bottom at various terminal sizes
- [ ] Unicode/emoji/CJK characters truncate cleanly
- [ ] ANSI colors don't bleed or corrupt
- [ ] Preview fills available space without overflow
- [ ] Cards maintain gaps at min/max terminal sizes
- [ ] Active session shows green hero border
- [ ] Header badges display correctly
