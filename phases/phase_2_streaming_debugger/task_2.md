# Task 2: OpenAI SSE Chunk Parser

## Phase 2 — Streaming Debugger

**Status:** Done

## Description

Parse OpenAI's Server-Sent Events streaming format, extracting delta tokens from each `data: {...}` line.

## Requirements

- Parse lines matching `data: {JSON}` format; skip `data: [DONE]`
- Extract from each chunk: delta content text, tool call delta, finish_reason, usage (if present in final chunk)
- Accumulate full response across all chunks
- Handle multi-line JSON (though OpenAI typically sends one JSON per line)
- Record exact arrival timestamp for each chunk

## Files

- `internal/provider/openai.go`
- `internal/intercept/streaming.go`
