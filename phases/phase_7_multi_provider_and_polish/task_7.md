# Task 7: OpenRouter Parser

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Implement the parser for OpenRouter, a proxy that provides access to many models via a unified OpenAI-compatible API.

## Requirements

- Detect host `openrouter.ai` and path `/api/v1/chat/completions`
- Reuse OpenAI parser for request/response format
- Extract model name (OpenRouter uses `provider/model` naming like `anthropic/claude-3-5-sonnet`)
- Parse OpenRouter-specific headers: `X-RateLimit-Limit-Requests`, `X-RateLimit-Remaining-Requests`
- OpenRouter passes through provider's token usage in `usage` field — use it for cost
- Look up pricing by the underlying model name (strip provider prefix)

## Files

- `internal/provider/openrouter.go`
- `internal/provider/detect.go`
