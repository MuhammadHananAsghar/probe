# Task 13: Session Running Totals

## Phase 1 — Proxy Core & Live Traffic

**Status:** Done

## Description

Maintain session-level aggregates: total cost, request count, error count, average latency, average TTFT.

## Requirements

- Thread-safe aggregation (multiple concurrent requests)
- Track: total_cost, request_count, error_count, total_latency, total_ttft, ttft_count
- Expose computed stats: avg_latency, avg_ttft, total_cost formatted
- Reset on new session (each `probe listen` invocation)
- Publish stat updates to TUI and dashboard via event channel

## Files

- `internal/analyze/session.go`
- `internal/store/models.go`
