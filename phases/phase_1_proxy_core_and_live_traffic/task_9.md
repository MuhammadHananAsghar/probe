# Task 9: Anthropic Request/Response Parser

## Phase 1 — Proxy Core & Live Traffic

**Status:** Done

## Description

Parse Anthropic Messages API requests and responses.

## Requirements

- Parse request: model, messages array, system prompt, tools array, max_tokens, stream flag
- Parse non-streaming response: usage (input_tokens, output_tokens), stop_reason, content blocks
- Parse streaming response: handle `event: message_start`, `event: content_block_delta`, `event: message_delta`, `event: message_stop` SSE event types
- Extract tool use blocks and tool result blocks
- Handle error responses (4xx/5xx) with Anthropic error format

## Files

- `internal/provider/anthropic.go`
- `internal/provider/provider.go`
