# Task 3: TLS Interception & Dynamic Cert Generation

## Phase 1 — Proxy Core & Live Traffic

**Status:** Not Started

## Description

Implement TLS man-in-the-middle: intercept HTTPS CONNECT tunnels, decrypt, inspect, re-encrypt, and forward to real provider.

## Requirements

- Generate local CA keypair (`~/.probe/ca-cert.pem`, `~/.probe/ca-key.pem`) on first run
- Dynamically generate per-hostname TLS certs signed by local CA
- Cache generated certs in memory (avoid regenerating per-request)
- Only intercept known LLM provider hostnames; passthrough all others
- Preserve SNI and ensure re-encrypted connection to real upstream

## Files

- `internal/proxy/ca.go`
- `internal/proxy/tls.go`
