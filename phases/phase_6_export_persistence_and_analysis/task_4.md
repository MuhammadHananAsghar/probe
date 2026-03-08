# Task 4: Markdown Report Export

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Generate a human-readable Markdown report summarizing a session or individual request, suitable for bug reports and team sharing.

## Requirements

- Session report: summary table (total requests, total cost, error count, avg latency), cost breakdown by model, top 5 most expensive requests, any anomalies detected
- Per-request report: provider, model, status, tokens, cost, timing, messages (truncated), tool calls, any issues detected
- `probe export --format markdown --output report.md`
- `probe export --format markdown --request N` for single-request report
- Generated markdown renders well on GitHub (uses standard GFM syntax)

## Files

- `internal/export/markdown.go`
- `cmd/probe/main.go`
