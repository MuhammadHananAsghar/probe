# Task 5: Non-LLM Traffic Passthrough

## Phase 1 — Proxy Core & Live Traffic

**Status:** Not Started

## Description

Ensure probe does not intercept or break traffic destined for non-LLM hosts — it should be invisible to all other HTTPS connections.

## Requirements

- Maintain allowlist of LLM provider hostnames (`api.openai.com`, `api.anthropic.com`, `generativelanguage.googleapis.com`, etc.)
- For non-LLM CONNECT tunnels: pass through raw TCP without decryption
- For non-LLM base URL requests: return 404 with helpful message
- Log passthrough decisions at debug level only
- Zero latency impact on non-LLM traffic

## Files

- `internal/proxy/passthrough.go`
- `internal/provider/detect.go`
