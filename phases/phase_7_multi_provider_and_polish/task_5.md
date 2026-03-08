# Task 5: Azure OpenAI Parser

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Implement the parser for Azure OpenAI Service, which uses a different URL scheme but mostly the same request/response format.

## Requirements

- Detect host pattern `*.openai.azure.com` and path `/openai/deployments/*/chat/completions`
- Extract deployment name from URL path (maps to model in probe's display)
- Reuse OpenAI request/response parsing logic
- Handle Azure-specific headers: `api-key` header (instead of `Authorization: Bearer`)
- Azure API version parameter (`?api-version=2024-02-01`) — preserve in display
- Add Azure OpenAI pricing to the pricing database (same as OpenAI model pricing)

## Files

- `internal/provider/azure.go`
- `internal/provider/detect.go`
