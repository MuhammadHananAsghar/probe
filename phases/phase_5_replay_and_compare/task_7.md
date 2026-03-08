# Task 7: Dashboard Replay UI

## Phase 5 — Replay & Compare

**Status:** Not Started

## Description

Add replay and compare functionality to the web dashboard.

## Requirements

- "Replay" button on every request detail page
- Replay options modal: model override, provider override, temperature, max_tokens, system prompt
- "Compare with..." dropdown on request detail to select another request for diff
- "Multi-model compare" button: opens a panel to select 2-4 models and dispatches simultaneously
- Live progress indicators for in-flight replays in the dashboard
- API endpoints: `POST /api/replay/{id}` and `POST /api/compare` to trigger replays from the dashboard

## Files

- `internal/dashboard/server.go`
- `internal/dashboard/ui/src/pages/RequestDetail.tsx`
