# Task 8: Cost Comparison Display

## Phase 5 — Replay & Compare

**Status:** Not Started

## Description

After a multi-model or cross-provider replay, show a clear cost/performance comparison summary.

## Requirements

- Show table: model | cost | latency | TTFT | tokens out | cost per output token
- Highlight cheapest option in green, most expensive in red
- "Switch to X and save Y% on this request type" recommendation
- Show cumulative savings projection: "If all N requests in this session used gpt-4o-mini: saved $X.XX"
- Export comparison table as markdown (one-click copy)

## Files

- `internal/replay/compare.go`
- `internal/tui/detail.go`
- `internal/dashboard/ui/src/pages/RequestDetail.tsx`
