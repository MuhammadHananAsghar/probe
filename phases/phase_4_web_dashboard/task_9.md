# Task 9: Diff View

## Phase 4 — Web Dashboard

**Status:** Not Started

## Description

Build a side-by-side diff view for comparing two requests — useful for spotting prompt changes, model behavior changes, or A/B testing results.

## Requirements

- Select two requests to compare from the request list (checkbox or right-click context menu)
- Side-by-side layout: left = request A, right = request B
- Diff the messages arrays: highlight added/removed/changed messages
- Diff the response content: word-level diff highlighting
- Show delta stats: cost difference, latency difference, token count difference
- "Compare with..." button on request detail page

## Files

- `internal/dashboard/ui/src/pages/RequestDetail.tsx`
- `internal/dashboard/ui/src/components/DiffView.tsx`
