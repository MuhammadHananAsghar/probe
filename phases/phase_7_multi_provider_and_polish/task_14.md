# Task 14: CLI Flags & Filter Options Polish

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Implement all remaining CLI flags documented in the project plan that haven't been implemented in previous phases.

## Requirements

- `probe listen --port 9090` — custom proxy port (already planned but ensure flag is wired)
- `probe listen --dashboard-port 4042` — custom dashboard port
- `probe listen --filter anthropic` — only intercept and display Anthropic calls (pass through all others silently)
- `probe listen --filter model=gpt-4o` — only intercept requests to a specific model
- `probe listen --no-tls` — skip CA cert setup, only work in base URL mode
- `probe listen --no-browser` — don't auto-open browser (already planned, ensure it's wired)
- `probe listen --quiet` — suppress TUI, only log errors to stderr

## Files

- `cmd/probe/main.go`
- `pkg/config/config.go`
- `internal/proxy/proxy.go`
