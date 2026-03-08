# Task 10: Generic OpenAI-Compatible Fallback Parser

## Phase 1 — Proxy Core & Live Traffic

**Status:** Done

## Description

Fallback parser for any OpenAI-compatible endpoint (Together, Fireworks, OpenRouter, self-hosted) that uses the `/v1/chat/completions` format.

## Requirements

- Reuse OpenAI parser logic
- Mark provider as `openai-compatible` with detected hostname
- Cost calculation shows `unknown` pricing with note to add custom pricing
- Log which host was detected as compatible

## Files

- `internal/provider/generic.go`
- `internal/provider/detect.go`
