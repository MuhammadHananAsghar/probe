# Task 4: Side-by-Side Response Comparison

## Phase 5 — Replay & Compare

**Status:** Not Started

## Description

Display a comparison between the original captured response and a replayed response in the TUI.

## Requirements

- Show two-column layout: original on left, replay on right
- Compare: model, tokens in/out, cost, latency, TTFT, finish reason
- Compute deltas: cost difference (absolute + percentage), latency difference, token count difference
- Show "Cost savings: 95.5% with gpt-4o-mini" style summary
- TUI: press `d` on a replay request to show diff against its original
- Word-level diff of the response text content

## Files

- `internal/replay/compare.go`
- `internal/tui/detail.go`
