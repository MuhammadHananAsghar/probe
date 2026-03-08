# Task 3: Anthropic SSE Chunk Parser

## Phase 2 — Streaming Debugger

**Status:** Not Started

## Description

Parse Anthropic's streaming event format, which uses named event types (`message_start`, `content_block_delta`, `message_delta`, `message_stop`).

## Requirements

- Parse `event: <type>` + `data: {JSON}` pairs
- Handle all event types: `message_start` (captures input_tokens), `content_block_start`, `content_block_delta` (captures text delta), `content_block_stop`, `message_delta` (captures output_tokens, stop_reason), `message_stop`
- Accumulate full text response across `content_block_delta` events
- Extract tool use blocks from streaming
- Record arrival timestamp per event

## Files

- `internal/provider/anthropic.go`
- `internal/intercept/streaming.go`
