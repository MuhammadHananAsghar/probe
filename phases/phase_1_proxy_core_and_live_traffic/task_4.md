# Task 4: First-Run CA Trust Setup

## Phase 1 — Proxy Core & Live Traffic

**Status:** Done

## Description

On first run, detect OS and prompt user to trust the local CA certificate so TLS interception works without browser/app errors.

## Requirements

- Detect macOS / Linux / Windows
- Print OS-specific trust command (macOS: `sudo security add-trusted-cert ...`, Linux: `sudo cp ... && update-ca-certificates`, Windows: `certutil -addstore -f "ROOT" ...`)
- Offer to run the trust command automatically with sudo prompt
- Skip gracefully if CA already trusted or if user declines (falls back to base URL mode hint)
- Store `~/.probe/ca-trusted` flag after successful trust

## Files

- `internal/proxy/ca.go`
- `cmd/probe/main.go`
