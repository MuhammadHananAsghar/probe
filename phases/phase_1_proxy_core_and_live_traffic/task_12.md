# Task 12: Cost Per Request Calculator

## Phase 1 — Proxy Core & Live Traffic

**Status:** Not Started

## Description

Given parsed request/response data, calculate the dollar cost of an LLM API call.

## Requirements

- `cost = (input_tokens / 1_000_000 * input_price) + (output_tokens / 1_000_000 * output_price)`
- Look up prices from embedded pricing database
- Handle missing prices gracefully (return `$?` in UI)
- Support user-defined custom pricing from `~/.probe/config.yaml`
- Format cost for display: `$0.0089`, `$0.47`, `<$0.0001`

## Files

- `internal/cost/calculator.go`
- `internal/cost/pricing.go`
