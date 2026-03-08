# Task 10: Dark/Light Theme & Auto-Open Browser

## Phase 4 — Web Dashboard

**Status:** Not Started

## Description

Implement dark and light themes for the dashboard, and auto-open the browser on first `probe listen`.

## Requirements

- Default to system preference via `prefers-color-scheme` media query
- Toggle button in the header to switch between dark and light
- Persist theme preference to localStorage
- On `probe listen` first run: open `http://localhost:4041` in default browser (`open` on macOS, `xdg-open` on Linux, `start` on Windows)
- `--no-browser` flag disables auto-open
- Dashboard shows connection status indicator (green dot = connected to probe, grey = disconnected)

## Files

- `internal/dashboard/ui/src/App.tsx`
- `internal/dashboard/server.go`
- `cmd/probe/main.go`
