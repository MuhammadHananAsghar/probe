# Task 12: Homebrew Tap & Scoop Bucket

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Set up Homebrew tap and Scoop bucket so users can install probe with a single command on macOS and Windows.

## Requirements

- Homebrew: goreleaser auto-generates Formula for `probe-dev/tap/probe` tap
- Formula installs the macOS binary and sets up shell completions
- `brew install probe-dev/tap/probe` installs on macOS/Linux
- Scoop: goreleaser auto-generates manifest for `probe-dev/scoop-bucket`
- `scoop bucket add probe-dev https://github.com/probe-dev/scoop-bucket && scoop install probe` installs on Windows
- Both tap/bucket repos created as separate GitHub repos with auto-update on each release

## Files

- `.goreleaser.yml`
- `deployments/homebrew/`
- `deployments/scoop/`
