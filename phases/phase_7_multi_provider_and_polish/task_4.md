# Task 4: Ollama Parser (Local Models)

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Implement the parser for Ollama, which runs local models and uses a different API format with no cost.

## Requirements

- Detect host `localhost:11434` (and `127.0.0.1:11434`) and path `/api/chat`
- Parse request: model, messages array, stream flag, options (temperature, num_predict, etc.)
- Parse non-streaming response: `message.content`, `eval_count` (output tokens), `prompt_eval_count` (input tokens), `total_duration` (nanoseconds)
- Parse streaming response: accumulate `message.content` deltas; capture `done: true` final stats
- Cost is always $0.00 for local models — display "local" instead of a cost
- Include popular Ollama model names in detection (llama3.2, mistral, qwen2.5, deepseek-r1, etc.)

## Files

- `internal/provider/ollama.go`
- `internal/provider/detect.go`
