# Task 8: OpenAI Request/Response Parser

## Phase 1 — Proxy Core & Live Traffic

**Status:** Not Started

## Description

Parse OpenAI API requests and responses to extract structured data (model, tokens, messages, tool calls, cost inputs).

## Requirements

- Parse request: model, messages array, tools array, temperature, max_tokens, stream flag
- Parse non-streaming response: usage (prompt_tokens, completion_tokens), finish_reason, content, tool_calls
- Parse streaming response: accumulate chunks, extract same fields from `data: {...}` SSE lines
- Extract error details from non-2xx responses
- Handle both `/v1/chat/completions` and `/v1/completions` (legacy)

## Files

- `internal/provider/openai.go`
- `internal/provider/provider.go`
