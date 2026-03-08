# Task 6: probe history Command

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Browse and query past sessions stored in SQLite, enabling post-hoc analysis of historical LLM API usage.

## Requirements

- `probe history` — interactive TUI browser of past sessions (date, request count, total cost)
- `probe history --cost --last 7d` — print cost summary table for last 7 days
- `probe history --errors` — show only requests with error status codes
- `probe history --model gpt-4o` — filter by model
- `probe history --session SESSION_ID` — replay a specific past session in the TUI
- Output to stdout (non-TUI) when piped: `probe history --last 7d | grep anthropic`

## Files

- `cmd/probe/main.go`
- `internal/store/sqlite.go`
- `internal/tui/app.go`
