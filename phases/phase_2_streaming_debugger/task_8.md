# Task 8: Stream Replay at Original Timing

## Phase 2 — Streaming Debugger

**Status:** Not Started

## Description

Allow replaying a captured streaming response to the dashboard/TUI at the same inter-chunk timing as the original, useful for demos and debugging.

## Requirements

- Store per-chunk content + timing in `Request.Chunks` array
- `probe inspect N --stream --replay` re-emits chunks with original delays
- Replay speed multiplier: `--speed 2x` for 2× faster, `--speed 0.5x` for slower
- Works in both TUI (animated) and stdout (plain text with delays)
- Replay is read-only — does not re-send to provider

## Files

- `internal/intercept/streaming.go`
- `internal/store/models.go`
- `internal/tui/stream.go`
