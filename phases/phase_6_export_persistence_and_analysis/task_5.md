# Task 5: SQLite Persistent Storage

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Add opt-in SQLite persistence so probe can survive restarts and maintain request history across sessions.

## Requirements

- `probe listen --persist` enables SQLite storage at `~/.probe/history.db`
- Schema: `requests` table with all `Request` fields serialized; `sessions` table for session metadata
- Ring buffer still used as in-memory cache; SQLite is the backing store
- On startup with `--persist`: load last 1000 requests from SQLite into ring buffer
- Auto-migrate schema on version upgrade (additive migrations only)
- WAL mode enabled for concurrent read performance

## Files

- `internal/store/sqlite.go`
- `internal/store/store.go`
- `internal/store/models.go`
