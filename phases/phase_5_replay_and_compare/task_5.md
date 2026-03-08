# Task 5: Multi-Model Simultaneous Comparison

## Phase 5 — Replay & Compare

**Status:** Not Started

## Description

Send the same prompt to multiple models concurrently and compare all responses side-by-side.

## Requirements

- `probe compare N --models claude-sonnet-4,gpt-4o,gemini-2.0-flash` sends request N to all three models simultaneously
- Concurrent dispatch using goroutines; display results as each completes
- Show a comparison table: model | tokens | cost | latency | TTFT | finish_reason
- Show response quality comparison in TUI (user must manually judge; no auto-scoring)
- Compute "cheapest", "fastest", "highest token output" badges
- Store all comparison responses in the ring buffer linked to the source request

## Files

- `internal/replay/compare.go`
- `internal/replay/replay.go`
- `internal/tui/detail.go`
