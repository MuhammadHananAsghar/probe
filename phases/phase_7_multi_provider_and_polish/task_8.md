# Task 8: probe update-pricing Command

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Add a command to fetch the latest model pricing from a maintained source and update the embedded pricing database.

## Requirements

- `probe update-pricing` fetches latest pricing JSON from the probe GitHub repo's `pricing/models.json`
- Merges fetched pricing with user's custom overrides (custom overrides win)
- Shows diff: "Updated 3 models: gpt-4o ($2.50→$2.00), ..."
- Updates `~/.probe/pricing.json` (user-local override, not the binary's embedded copy)
- Probe loads user-local pricing file at startup if it exists, falling back to embedded
- Requires internet access; fails gracefully with clear error if offline

## Files

- `cmd/probe/main.go`
- `internal/cost/pricing.go`
- `pkg/config/config.go`
