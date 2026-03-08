# Task 14: Terminal UI — Live Request List

## Phase 1 — Proxy Core & Live Traffic

**Status:** Done

## Description

Build the bubbletea TUI that shows a scrollable live list of intercepted LLM requests.

## Requirements

- Each row: `#N  METHOD provider /path  model  Tokens: X in → Y out  Cost: $Z  Latency: Xs  Status: 200 ✓`
- Color coding: green for 200, yellow for 4xx/retries, red for 5xx/errors
- New requests append at bottom; list auto-scrolls unless user has scrolled up
- Arrow keys / vim keys (j/k) to navigate; Enter to open detail view
- Show spinner on in-flight requests

## Files

- `internal/tui/app.go`
- `internal/tui/list.go`
- `internal/tui/styles.go`
