# Task 11: Rate Limit Tracker

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Parse rate limit headers from LLM API responses and surface remaining quota in the TUI so developers can see when they're approaching limits before hitting errors.

## Requirements

- Parse OpenAI rate limit headers: `x-ratelimit-remaining-requests`, `x-ratelimit-remaining-tokens`, `x-ratelimit-reset-requests`, `x-ratelimit-reset-tokens`
- Parse Anthropic rate limit headers: `anthropic-ratelimit-requests-remaining`, `anthropic-ratelimit-tokens-remaining`, `anthropic-ratelimit-requests-reset`, `anthropic-ratelimit-tokens-reset`
- Show rate limit status in the TUI stats bar: "Requests: 47/500 remaining | Tokens: 120k/1M remaining"
- Alert (yellow warning) when remaining < 10% of limit
- Alert (red) when a 429 rate limit error is received
- Store rate limit snapshots per provider for the dashboard rate limit timeline chart

## Files

- `internal/analyze/anomaly.go`
- `internal/tui/stats.go`
- `internal/store/models.go`
