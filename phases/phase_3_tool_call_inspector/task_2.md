# Task 2: Parse Tool Call Decisions from Response

## Phase 3 — Tool Call Inspector

**Status:** Not Started

## Description

Extract tool call decisions the model made in a response — which tools it chose to call and with what arguments.

## Requirements

- OpenAI format: parse `choices[0].message.tool_calls` array (each has `id`, `type`, `function.name`, `function.arguments` as JSON string)
- Anthropic format: parse content blocks of type `tool_use` (each has `id`, `name`, `input` as object)
- Parse streaming tool calls: accumulate argument JSON string deltas across chunks
- Store on `Request.ToolCalls` as normalized struct with `id`, `name`, `arguments_json`
- Detect malformed JSON in arguments and flag as parse error

## Files

- `internal/provider/openai.go`
- `internal/provider/anthropic.go`
- `internal/store/models.go`
