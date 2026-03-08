# Task 6: Graceful Shutdown

## Phase 1 — Proxy Core & Live Traffic

**Status:** Done

## Description

Handle OS signals (SIGINT, SIGTERM) to shut down the proxy cleanly — drain in-flight requests, flush stats, close listeners.

## Requirements

- Listen for SIGINT/SIGTERM via `signal.NotifyContext`
- Drain in-flight proxied requests with a 5-second timeout
- Flush any pending TUI/dashboard updates before exit
- Print session summary (total requests, total cost) on exit
- Exit with code 0 on clean shutdown, 1 on forced kill

## Files

- `cmd/probe/main.go`
- `internal/proxy/proxy.go`
