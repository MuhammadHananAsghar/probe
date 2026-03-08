# Task 6: Chunk Count & Token-Per-Chunk Statistics

## Phase 2 — Streaming Debugger

**Status:** Done

## Description

Compute streaming throughput statistics: chunks per second, tokens per second, and distribution of tokens per chunk.

## Requirements

- Count total chunks received
- Count tokens per chunk (from delta content length approximation or SSE usage field)
- Compute `throughput_tokens_per_sec = total_output_tokens / stream_duration_seconds`
- Compute average tokens per chunk
- Store all stats on `Request.StreamStats` for display

## Files

- `internal/intercept/streaming.go`
- `internal/store/models.go`
