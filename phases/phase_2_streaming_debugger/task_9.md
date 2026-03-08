# Task 9: Streaming Edge Case Handling

## Phase 2 — Streaming Debugger

**Status:** Not Started

## Description

Handle all known edge cases in SSE streaming to prevent probe from crashing or blocking the caller's response.

## Requirements

- Handle interrupted streams: connection drop mid-stream → mark request as `interrupted`, record how many chunks arrived
- Handle empty chunks: `data: {}` or `data: {"choices":[{"delta":{}}]}` — skip without recording as content
- Handle error mid-stream: some providers send an error event partway through → parse and record
- Handle concurrent streams: multiple simultaneous streaming requests must not interfere
- Handle very large chunks (e.g. tool result embedded in stream) without memory spike

## Files

- `internal/intercept/streaming.go`
- `internal/proxy/proxy.go`
