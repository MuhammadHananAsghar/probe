# Task 7: Auto-Cleanup & Retention Config

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Automatically clean up old SQLite data to prevent unbounded disk growth, with configurable retention policy.

## Requirements

- Default retention: 7 days (delete requests older than 7 days)
- Configurable via `~/.probe/config.yaml`: `storage.retention_days: 30`
- Run cleanup on startup (if `--persist` enabled) and once daily while running
- Never delete requests from the current session regardless of age
- Log cleanup: "Deleted N requests older than 7 days (freed ~XMB)"
- `probe history --cleanup` triggers manual cleanup

## Files

- `internal/store/sqlite.go`
- `pkg/config/config.go`
