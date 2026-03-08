# Task 2: Replay with Parameter Modifications

## Phase 5 — Replay & Compare

**Status:** Not Started

## Description

Allow modifying request parameters before replaying — change model, temperature, max_tokens, or system prompt.

## Requirements

- `probe replay N --model gpt-4o-mini` overrides the model field in the request body
- `probe replay N --temperature 0` overrides temperature
- `probe replay N --max-tokens 1000` overrides max_tokens
- `probe replay N --system "You are concise."` replaces the system message
- All modifications are applied to a deep copy; original stored request is untouched
- Show parameter diff in TUI result: "Replayed with: model=gpt-4o-mini (was claude-sonnet-4-20250514)"

## Files

- `internal/replay/modify.go`
- `internal/replay/replay.go`
