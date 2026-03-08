# Task 16: Terminal UI — Request Detail View

## Phase 1 — Proxy Core & Live Traffic

**Status:** Done

## Description

Show full detail for a selected request (activated by pressing Enter on a list item).

## Requirements

- Show: Provider, Endpoint, Model, Status, Timing (total latency, TTFT), Tokens & Cost (input/output/total), Messages array (role + content preview + token count), Tool Calls summary, Rate limit headers if present
- Press `q` or `Esc` to return to list
- Press `r` to replay request (Phase 5 feature — show "not yet implemented" for now)
- Press `e` to export as curl (Phase 6 feature — stub)
- Scrollable for long message content

## Files

- `internal/tui/detail.go`
- `internal/tui/app.go`
