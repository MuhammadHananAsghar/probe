# Task 15: Terminal UI — Session Stats Bar

## Phase 1 — Proxy Core & Live Traffic

**Status:** Done

## Description

Show a persistent status bar at the bottom of the TUI with live session stats.

## Requirements

- Show: `Session: N requests | $X.XX total | Avg TTFT: Xms | N errors`
- Update in real-time as each request completes
- Highlight error count in red if > 0
- Show `Proxy: http://localhost:8080  Dashboard: http://localhost:4041` on startup line

## Files

- `internal/tui/stats.go`
- `internal/tui/app.go`
