# Task 1: SSE Stream Interception (Tee Reader)

## Phase 2 — Streaming Debugger

**Status:** Not Started

## Description

Intercept SSE streaming responses without buffering — the response must flow to the caller in real-time while probe reads it simultaneously via a tee reader.

## Requirements

- Use `io.TeeReader` so response bytes flow to client AND to probe's parser simultaneously
- Zero buffering: each chunk must reach the caller before probe processes it
- Work for both proxy mode (CONNECT tunnel) and base URL mode
- Handle `Transfer-Encoding: chunked` correctly
- Attach timing metadata (arrival time in nanoseconds) to each chunk as it's read

## Files

- `internal/intercept/streaming.go`
- `internal/intercept/body.go`
