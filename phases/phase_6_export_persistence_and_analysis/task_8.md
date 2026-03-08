# Task 8: Cost & Latency Alerting

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Notify the user in the terminal when a configurable cost or latency threshold is exceeded.

## Requirements

- `probe listen --alert-cost 1.00` — terminal bell + banner when session total exceeds $1.00
- `probe listen --alert-latency 5s` — flag individual requests slower than 5 seconds with a warning banner
- `probe listen --alert-error` — terminal bell on any 4xx/5xx response or rate limit hit
- Configurable thresholds in `~/.probe/config.yaml` (persistent defaults)
- Alert appears as a bordered banner in the TUI, dismissible with `d`
- Multiple alerts stack (don't replace each other)

## Files

- `internal/analyze/anomaly.go`
- `internal/tui/app.go`
- `pkg/config/config.go`
