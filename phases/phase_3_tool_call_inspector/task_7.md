# Task 7: Multi-Turn Tool Call Tracking

## Phase 3 — Tool Call Inspector

**Status:** Not Started

## Description

Track tool call chains across multiple HTTP requests that belong to the same logical conversation (multi-turn conversations with tool use).

## Requirements

- Correlate requests by matching the growing messages array (each turn appends to previous)
- Assign a `ConversationID` to groups of related requests
- Display conversation grouping in TUI list view (indent follow-up requests under parent)
- Compute conversation-level stats: total turns, total tool calls, total cost across all turns
- Allow filtering request list by conversation: press `c` to show only current conversation

## Files

- `internal/analyze/session.go`
- `internal/store/models.go`
- `internal/tui/list.go`
