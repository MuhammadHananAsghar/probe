# Task 15: First-Run Wizard & Error Messages

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Create a guided first-run experience that helps new users set up probe correctly and shows clear, actionable error messages when things go wrong.

## Requirements

- On first run (`~/.probe/` doesn't exist): print a welcome banner explaining proxy mode vs base URL mode
- Detect if no requests seen after 30 seconds: print "No requests captured yet. Make sure your app sets HTTPS_PROXY=http://localhost:8080"
- Detect if CA cert not trusted and a TLS error is seen: print the OS-specific trust command
- Error: "Cannot bind to port 8080" — suggest `probe listen --port 9090`
- Error: rate limit hit — show remaining quota and reset time if headers available
- All error messages include a "See docs: https://probe.dev/docs/..." link

## Files

- `cmd/probe/main.go`
- `internal/proxy/ca.go`
- `internal/tui/app.go`
