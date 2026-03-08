# Task 3: Parse Tool Results from Follow-Up Messages

## Phase 3 — Tool Call Inspector

**Status:** Not Started

## Description

When a follow-up request arrives with tool results (the app sending tool execution outputs back to the model), correlate them with the original tool calls.

## Requirements

- OpenAI format: identify messages with `role: "tool"` and match by `tool_call_id`
- Anthropic format: identify content blocks of type `tool_result` and match by `tool_use_id`
- Link tool results to their originating tool call by ID
- Record tool result content (may be string, JSON object, or array)
- Compute tool execution latency: time between the response containing the tool call and the follow-up request containing the result

## Files

- `internal/provider/openai.go`
- `internal/provider/anthropic.go`
- `internal/store/models.go`
