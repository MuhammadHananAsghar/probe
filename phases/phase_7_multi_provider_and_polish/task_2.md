# Task 2: Cohere V2 Parser

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Implement the provider parser for Cohere's V2 Chat API.

## Requirements

- Detect host `api.cohere.com` and path `/v2/chat`
- Parse request: model, messages array (Cohere's role format), tools, temperature
- Parse non-streaming response: `message.content`, `usage` (billed_units: input_tokens, output_tokens), finish_reason
- Parse streaming response: handle `content-delta` events, `message-end` event for usage
- Map Cohere finish reasons to probe's normalized finish reason
- Include Cohere pricing in the embedded pricing database

## Files

- `internal/provider/cohere.go`
- `internal/provider/detect.go`
- `pricing/models.json`
