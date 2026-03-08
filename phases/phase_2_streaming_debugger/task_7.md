# Task 7: Terminal UI — Stream Visualization

## Phase 2 — Streaming Debugger

**Status:** Done

## Description

Add a stream view in the TUI that shows chunk arrival as an animated timeline while the stream is in progress.

## Requirements

- Show live progress while streaming: a dot or token appears for each chunk in real-time
- After stream completes, show full timeline: timestamp, chunk number, content preview, gap to previous
- Display summary: TTFT, total chunks, duration, throughput, stall count
- Navigate to stream view by pressing `s` on a request in the detail view
- Format matches the ASCII art in the project plan

## Files

- `internal/tui/stream.go`
- `internal/tui/detail.go`
- `internal/tui/app.go`
