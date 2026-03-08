# Task 5: Inter-Chunk Timing & Stall Detection

## Phase 2 — Streaming Debugger

**Status:** Done

## Description

Record the time between every successive streaming chunk, and flag stalls where the gap exceeds a threshold.

## Requirements

- Record `arrived_at` for each chunk
- Compute inter-chunk gap: `gap[i] = chunk[i].arrived_at - chunk[i-1].arrived_at`
- Flag a stall when gap > 500ms (configurable via `--stall-threshold`)
- Store per-chunk timing array on the `Request` model
- Count total stall events per request

## Files

- `internal/intercept/streaming.go`
- `internal/store/models.go`
