# Task 11: Embedded Pricing Database

## Phase 1 — Proxy Core & Live Traffic

**Status:** Not Started

## Description

Bundle a JSON pricing database into the binary and implement lookup by model name.

## Requirements

- Embed `pricing/models.json` via `go:embed`
- JSON schema: `{ "models": { "<model-id>": { "provider": "...", "input_cost_per_1m": float, "output_cost_per_1m": float, "context_window": int } } }`
- Include all current major models: GPT-4o, GPT-4o-mini, claude-sonnet-4, claude-haiku, gemini-2.0-flash, etc.
- Support fuzzy model name matching (e.g. `claude-3-5-sonnet-20241022` matches `claude-3-5-sonnet`)
- Return zero cost (not error) for unknown models

## Files

- `internal/cost/pricing.go`
- `internal/cost/pricing_data.json`
- `pricing/models.json`
