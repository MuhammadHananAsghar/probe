# Task 1: HTTP Forward Proxy

## Phase 1 — Proxy Core & Live Traffic

**Status:** Not Started

## Description

Build the HTTP forward proxy server that handles CONNECT tunneling for HTTPS traffic when users set `HTTPS_PROXY=http://localhost:8080`.

## Requirements

- Listen on configurable port (default 8080) via `internal/proxy/proxy.go`
- Handle HTTP CONNECT method for HTTPS tunneling
- Maintain client connections during CONNECT tunnel lifetime
- Return appropriate errors (400 for malformed, 502 for upstream failures)
- Support `--port` flag override

## Files

- `internal/proxy/proxy.go`
- `cmd/probe/main.go`
