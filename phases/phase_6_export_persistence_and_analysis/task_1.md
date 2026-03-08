# Task 1: HAR Format Export

## Phase 6 — Export, Persistence & Analysis

**Status:** Not Started

## Description

Export captured requests as an HTTP Archive (HAR) file compatible with Chrome DevTools, Charles Proxy, and other tools.

## Requirements

- Generate valid HAR 1.2 format JSON (`log.entries` array)
- Each entry maps probe's `Request` model to HAR `entry` fields: startedDateTime, time, request (method, url, headers, postData), response (status, headers, content), timings
- Mask sensitive headers (`x-api-key`, `Authorization`) before export — show `sk-...XXXX` truncated form
- `probe export --format har --output session.har`
- `probe export --format har --last 1h` exports only requests from last hour
- Validate output against HAR schema before writing

## Files

- `internal/export/har.go`
- `cmd/probe/main.go`
