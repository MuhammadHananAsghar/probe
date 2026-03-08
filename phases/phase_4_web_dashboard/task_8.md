# Task 8: Full-Text Search

## Phase 4 — Web Dashboard

**Status:** Not Started

## Description

Implement full-text search across request and response bodies within the current session.

## Requirements

- Search input in the header bar (keyboard shortcut: `/` or `Cmd+K`)
- Search across: model name, provider, message content, tool names, tool arguments, error messages
- Show matched requests in the list with the matching text highlighted
- Real-time search (debounced 150ms) — no submit button needed
- Search is in-memory (no server round-trip) since data is already in the browser via WebSocket
- Show "N results" count

## Files

- `internal/dashboard/ui/src/App.tsx`
- `internal/dashboard/ui/src/pages/RequestList.tsx`
