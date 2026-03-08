# Task 10: Duplicate Request Detector

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Detect when the same (or nearly identical) request is made multiple times, suggesting the app is missing a cache.

## Requirements

- Hash each request by: model + messages array content (normalized, whitespace-stripped)
- Detect exact duplicates: same hash appears 2+ times in the session
- Detect near-duplicates: same model + system prompt, only user message differs by < 10 tokens
- Annotate duplicate requests in the TUI list with a "DUP" badge
- `probe analyze --duplicates` prints a deduplication report: "Requests #3, #7, #12 are identical — adding a cache would save $0.027"
- Track cache hit rate projection: "X% of requests could be cached"

## Files

- `internal/analyze/anomaly.go`
- `internal/store/models.go`
