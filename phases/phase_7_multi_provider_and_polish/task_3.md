# Task 3: Groq, Together, and Fireworks Parsers

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Implement parsers for Groq, Together AI, and Fireworks AI — all use the OpenAI-compatible format but have distinct hosts and pricing.

## Requirements

- Groq: detect `api.groq.com`, reuse OpenAI parser, add Groq models to pricing DB (llama-3.3-70b-versatile, mixtral-8x7b, etc.)
- Together: detect `api.together.xyz`, reuse OpenAI parser, add Together models to pricing DB
- Fireworks: detect `api.fireworks.ai`, reuse OpenAI parser, add Fireworks models to pricing DB
- Each provider has its own `Provider` enum value (not just "openai-compatible")
- Show provider-specific badges in TUI and dashboard (different color per provider)

## Files

- `internal/provider/groq.go`
- `internal/provider/together.go`
- `internal/provider/fireworks.go`
- `pricing/models.json`
