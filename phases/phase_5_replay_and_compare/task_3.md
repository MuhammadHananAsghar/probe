# Task 3: Cross-Provider Replay (Request Format Translation)

## Phase 5 — Replay & Compare

**Status:** Not Started

## Description

Translate a captured request from one provider's format to another and replay to the target provider.

## Requirements

- `probe replay N --provider openai` translates an Anthropic request to OpenAI format (and vice versa)
- Translation map: Anthropic messages → OpenAI messages (system prompt extraction, content block normalization)
- OpenAI → Anthropic: map `messages` with `role: system` to Anthropic's top-level `system` field
- Preserve tool definitions in the translated format (OpenAI tools ↔ Anthropic tools schema)
- Warn if translation is lossy (e.g., Anthropic `cache_control` has no OpenAI equivalent)
- Target provider's API key must be set in config or environment

## Files

- `internal/replay/modify.go`
- `internal/provider/openai.go`
- `internal/provider/anthropic.go`
