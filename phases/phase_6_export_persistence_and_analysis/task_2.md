# Task 2: JSON Export

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Export captured requests as a probe-native JSON format for custom analysis, scripting, or archiving.

## Requirements

- Export full `Request` struct array as JSON (or NDJSON for streaming)
- Include all parsed fields: provider, model, messages, tokens, cost, timing, chunks, tool calls
- `probe export --format json --output session.json`
- `probe export --format json --filter provider=anthropic` exports subset
- `probe export --format ndjson` for newline-delimited JSON (one object per line, streamable)
- Pretty-print by default; `--compact` for minified output

## Files

- `internal/export/json.go`
- `cmd/probe/main.go`
