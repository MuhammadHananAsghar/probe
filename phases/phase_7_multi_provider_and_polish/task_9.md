# Task 9: GitHub Action for Weekly Pricing Updates

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Set up a GitHub Actions workflow that automatically fetches current model pricing from provider documentation and opens a PR to update `pricing/models.json` weekly.

## Requirements

- Workflow file: `.github/workflows/update-pricing.yml`
- Runs weekly via cron schedule (`0 9 * * 1` — Monday 9am UTC)
- Runs a Go script (`pricing/update.go`) that fetches/parses pricing from provider docs/APIs
- If pricing changed: creates a PR with the diff titled "chore: update model pricing (YYYY-MM-DD)"
- If no changes: workflow exits cleanly with no PR
- Workflow also runs on manual dispatch (`workflow_dispatch`)

## Files

- `.github/workflows/update-pricing.yml`
- `pricing/update.go`
