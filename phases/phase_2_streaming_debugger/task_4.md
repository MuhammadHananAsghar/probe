# Task 4: Time to First Token (TTFT) Measurement

## Phase 2 — Streaming Debugger

**Status:** Done

## Description

Measure the time from when the request was sent to when the first content token arrived in the streaming response.

## Requirements

- Record `request_sent_at` timestamp when request leaves probe toward provider
- Record `first_token_at` timestamp when the first non-empty content chunk arrives
- `TTFT = first_token_at - request_sent_at`
- Distinguish first token from first chunk (a chunk may arrive with empty delta)
- Store TTFT on the `Request` model for display and aggregation

## Files

- `internal/intercept/streaming.go`
- `internal/store/models.go`
- `internal/analyze/session.go`
