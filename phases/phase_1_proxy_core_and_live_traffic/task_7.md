# Task 7: Provider Auto-Detection

## Phase 1 — Proxy Core & Live Traffic

**Status:** Not Started

## Description

Given a request's hostname (or URL path for base URL mode), identify which LLM provider it targets.

## Requirements

- Detect by hostname: `api.openai.com` → OpenAI, `api.anthropic.com` → Anthropic, `generativelanguage.googleapis.com` → Google, `api.mistral.ai` → Mistral, `api.cohere.com` → Cohere, `api.groq.com` → Groq, `localhost:11434` → Ollama, `*.openai.azure.com` → Azure, `bedrock-runtime.*.amazonaws.com` → Bedrock
- Detect OpenAI-compatible fallback: any host with path `/v1/chat/completions`
- Return a `Provider` enum/struct with name and base URL
- Return `Unknown` (passthrough) for unrecognized hosts

## Files

- `internal/provider/detect.go`
- `internal/provider/provider.go`
