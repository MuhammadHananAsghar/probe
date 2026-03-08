# Task 6: AWS Bedrock Parser

## Phase 7 — Multi-Provider Intelligence & Polish

**Status:** Not Started

## Description

Implement the parser for AWS Bedrock, which wraps multiple model providers in a unified AWS API format.

## Requirements

- Detect host pattern `bedrock-runtime.*.amazonaws.com` and path `/model/*/invoke` or `/model/*/invoke-with-response-stream`
- Extract model ID from URL path (e.g. `anthropic.claude-3-5-sonnet-20241022-v2:0`)
- Detect underlying model provider from model ID prefix (`anthropic.*`, `amazon.*`, `meta.*`, `mistral.*`)
- Parse request: body format varies by model provider — delegate to appropriate sub-parser
- Parse response: outer Bedrock wrapper + inner provider-specific response
- Handle Bedrock streaming (chunk event format differs from SSE)

## Files

- `internal/provider/bedrock.go`
- `internal/provider/detect.go`
