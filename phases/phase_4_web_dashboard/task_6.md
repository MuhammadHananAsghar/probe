# Task 6: Cost Dashboard

## Phase 4 — Web Dashboard

**Status:** Not Started

## Description

Build a cost analytics page showing spending breakdown over time, by model, and by provider.

## Requirements

- Line chart: cost over time (per request, cumulative) for the current session
- Bar chart: cost breakdown by model (show top 5 models by spend)
- Bar chart: cost breakdown by provider
- Summary cards: session total cost, most expensive request, most expensive model
- Cost per hour projection: "at this rate, ~$X/hour"
- All charts update live via WebSocket

## Files

- `internal/dashboard/ui/src/pages/CostDashboard.tsx`
- `internal/dashboard/ui/src/components/CostBadge.tsx`
