# Task 6: CLI Replay Commands

## Phase 5 — Replay & Compare

**Status:** Not Started

## Description

Implement the full suite of `probe replay` and `probe compare` CLI commands with cobra.

## Requirements

- `probe replay N` — replay request N to same provider
- `probe replay N --model <m>` — replay with different model
- `probe replay N --provider <p>` — translate and replay to different provider
- `probe replay N --temperature <f> --max-tokens <n> --system <s>` — parameter overrides
- `probe compare N M` — diff two captured requests
- `probe compare N --models m1,m2,m3` — multi-model comparison
- All commands show spinner while in-flight and print result summary to stdout

## Files

- `cmd/probe/main.go`
- `internal/replay/replay.go`
- `internal/replay/compare.go`
