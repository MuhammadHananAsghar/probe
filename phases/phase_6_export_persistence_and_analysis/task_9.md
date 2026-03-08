# Task 9: Token Waste Detector

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Analyze requests and surface patterns of inefficient token usage to help developers reduce costs.

## Requirements

- Detect identical system prompt across all requests in the session — suggest prompt caching if provider supports it
- Detect very high input/output ratio (> 10:1 input tokens to output) — may indicate oversized context
- Detect system prompt that never changes but is re-sent on every request — "consider Anthropic prompt caching: save ~$X per request"
- Detect requests where most of the input is previous conversation history — flag after 5+ turns
- Show warnings as annotations on affected requests in the TUI and dashboard
- `probe analyze --waste` prints a waste report to stdout

## Files

- `internal/analyze/anomaly.go`
- `internal/analyze/session.go`
