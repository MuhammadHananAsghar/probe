# Task 10: User Pricing Overrides

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Allow users to define custom pricing for fine-tuned, private, or enterprise models not in the embedded pricing database.

## Requirements

- Config file: `~/.probe/config.yaml` with `pricing.custom` section
- Format: `pricing.custom.<model-id>: { input: 6.00, output: 12.00 }` (per 1M tokens)
- User overrides take priority over embedded pricing and over fetched pricing
- `probe config set pricing.custom.my-model "6.00/12.00"` CLI shorthand
- Show "(custom pricing)" note in the cost display for custom-priced models
- Validate input: reject negative prices, warn if prices seem implausible (> $1000/1M)

## Files

- `pkg/config/config.go`
- `internal/cost/pricing.go`
- `cmd/probe/main.go`
