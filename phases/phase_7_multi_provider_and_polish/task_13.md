# Task 13: probe update Self-Update Command

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Implement a self-update mechanism so users can update probe to the latest version without reinstalling.

## Requirements

- `probe update` checks GitHub Releases API for a newer version than the current binary
- If update available: download the appropriate binary for the current OS/arch, verify checksum, replace current binary
- `probe update --check` — only check without downloading
- Show changelog summary (first 5 lines of release notes) before applying update
- Handle permission error gracefully (if binary is in a read-only location, print `sudo probe update` hint)
- Respect `--no-update-check` flag and `PROBE_NO_UPDATE_CHECK=1` env var to disable

## Files

- `cmd/probe/main.go`
- `pkg/updater/updater.go`
