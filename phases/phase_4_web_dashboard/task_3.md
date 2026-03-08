# Task 3: Request List with Filtering

## Phase 4 — Web Dashboard

**Status:** Not Started

## Description

Build the main request list page in the React dashboard with live updates and filter/sort controls.

## Requirements

- Show all requests in a table: #, method, provider badge, model, tokens (in/out), cost, latency, TTFT, status badge
- Live updates via WebSocket: new rows appear at top without page reload
- Filter bar: filter by provider (OpenAI/Anthropic/Google/...), model (text search), status code, cost range (min/max)
- Sort by: time (default), cost desc, latency desc, tokens desc
- Click a row to navigate to request detail
- Show session summary bar: total requests, total cost, avg latency

## Files

- `internal/dashboard/ui/src/pages/RequestList.tsx`
- `internal/dashboard/ui/src/App.tsx`
