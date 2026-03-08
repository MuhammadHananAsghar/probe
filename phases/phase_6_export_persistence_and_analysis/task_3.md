# Task 3: curl Command Generation

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Generate a copy-pasteable `curl` command for any captured request, allowing developers to reproduce the exact LLM API call outside of probe.

## Requirements

- Generate `curl -X POST 'https://api.openai.com/v1/chat/completions' -H 'Content-Type: application/json' -H 'Authorization: Bearer $OPENAI_API_KEY' -d '{"model":...}'`
- Always use `$ENV_VAR` placeholder for API keys, never the actual key value
- Handle streaming: add `-N` flag for SSE streaming requests
- `probe inspect N --curl` prints curl to stdout
- Available as action in TUI detail view (`c` key) and dashboard "Copy as curl" button
- Optionally output as Python `requests` or Node.js `fetch` via `--lang python|node`

## Files

- `internal/export/curl.go`
- `internal/tui/detail.go`
