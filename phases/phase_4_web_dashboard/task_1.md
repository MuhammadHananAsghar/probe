# Task 1: React SPA Scaffolding & go:embed

## Phase 4 — Web Dashboard

**Status:** Not Started

## Description

Set up the React/TypeScript SPA project inside `internal/dashboard/ui/` and wire it into the Go binary via `go:embed` so the dashboard is served as a single self-contained binary.

## Requirements

- Initialize Vite + React + TypeScript project in `internal/dashboard/ui/`
- Build outputs to `internal/dashboard/ui/dist/`
- Go `dashboard/embed.go` uses `//go:embed ui/dist` to bundle built assets
- Dashboard HTTP server serves the SPA from embedded FS at `/` and the API at `/api/`
- Makefile target `make dashboard` builds the UI before building the Go binary
- `Content-Security-Policy` header set; no external CDN calls

## Files

- `internal/dashboard/embed.go`
- `internal/dashboard/server.go`
- `internal/dashboard/ui/package.json`
- `Makefile`
