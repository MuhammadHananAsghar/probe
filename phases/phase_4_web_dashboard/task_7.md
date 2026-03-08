# Task 7: Timeline View (Waterfall Chart)

## Phase 4 — Web Dashboard

**Status:** Not Started

## Description

Build a timeline / waterfall chart view showing all requests' timing in sequence, like Chrome DevTools Network tab.

## Requirements

- Each request shown as a horizontal bar spanning from request-start to response-end
- Color the bar: blue for waiting (TTFT), green for streaming, red for error
- Bars align on a shared time axis (x-axis = wall clock time, relative to session start)
- Click a bar to navigate to that request's detail view
- Show TTFT marker as a vertical line within each bar
- Zoom in/out on the time axis; pan left/right

## Files

- `internal/dashboard/ui/src/pages/Timeline.tsx`
- `internal/dashboard/ui/src/components/LatencyBar.tsx`
