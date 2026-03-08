# Task 2: Reverse Proxy / Base URL Mode

## Phase 1 — Proxy Core & Live Traffic

**Status:** Not Started

## Description

Build the reverse proxy mode where users point their SDK's `base_url` at `http://localhost:8080` instead of setting `HTTPS_PROXY`.

## Requirements

- Accept requests to `http://localhost:8080/v1/...` and forward to detected provider
- Map URL paths to provider base URLs (e.g. `/v1/chat/completions` → `https://api.openai.com/v1/chat/completions`)
- Preserve all headers except `Host`
- Forward response (including streaming) back to caller
- No TLS cert setup required for this mode

## Files

- `internal/proxy/proxy.go`
- `internal/proxy/passthrough.go`
