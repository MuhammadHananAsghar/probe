# Task 6: Tool Call Issue Detection

## Phase 3 — Tool Call Inspector

**Status:** Not Started

## Description

Automatically detect common tool call problems and surface them as warnings in the TUI and dashboard.

## Requirements

- Detect: tool call with no matching result in follow-up (orphaned tool call)
- Detect: malformed JSON in tool arguments (parse error)
- Detect: tool result that doesn't match the expected schema (if schema available)
- Detect: tool call loop (same tool called 5+ times in a single chain)
- Detect: very slow tool execution (> 5 seconds) — flag with warning
- Show detected issues as warning badges in the request list and detail view

## Files

- `internal/analyze/anomaly.go`
- `internal/store/models.go`
