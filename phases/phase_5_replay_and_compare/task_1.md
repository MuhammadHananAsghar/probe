# Task 1: Request Replay (Same Provider)

## Phase 5 — Replay & Compare

**Status:** Not Started

## Description

Re-send any captured request to the same LLM provider, storing the replayed response alongside the original for comparison.

## Requirements

- `probe replay N` re-sends request #N with identical headers, body, and API key
- Show a progress indicator while the replay is in flight
- Store replayed response as a new request entry linked to the original via `ReplayOf: N`
- Display replay result in TUI immediately after completion
- Handle streaming responses in replays the same as original captures
- Return error if original request is too old (> 24h) or API key is no longer available

## Files

- `internal/replay/replay.go`
- `internal/store/models.go`
- `cmd/probe/main.go`
