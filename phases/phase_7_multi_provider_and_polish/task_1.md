# Task 1: Google Gemini Parser

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Implement the provider parser for Google's Generative Language API (Gemini), which uses a different request/response format from OpenAI/Anthropic.

## Requirements

- Detect host `generativelanguage.googleapis.com` and path pattern `/v1/models/*/generateContent`
- Parse request: model (from URL path), `contents` array (Gemini's equivalent of messages), `tools`, `generationConfig` (temperature, maxOutputTokens, etc.)
- Parse non-streaming response: `candidates[0].content.parts`, `usageMetadata` (promptTokenCount, candidatesTokenCount)
- Parse streaming response (`streamGenerateContent`): accumulate `candidates[0].content.parts` deltas
- Map Gemini finish reasons (`STOP`, `MAX_TOKENS`, `SAFETY`) to probe's normalized finish reason

## Files

- `internal/provider/google.go`
- `internal/provider/detect.go`
