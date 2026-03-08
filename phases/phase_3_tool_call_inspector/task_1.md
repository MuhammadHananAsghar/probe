# Task 1: Parse Tool Definitions from Request

## Phase 3 — Tool Call Inspector

**Status:** Not Started

## Description

Extract the tool/function definitions from an LLM API request body so probe knows what tools the model has available.

## Requirements

- OpenAI format: parse `tools` array (each has `type`, `function.name`, `function.description`, `function.parameters`)
- Anthropic format: parse `tools` array (each has `name`, `description`, `input_schema`)
- Store parsed tool definitions on `Request.Tools` as a normalized struct
- Handle both `tools` (new) and `functions` (legacy OpenAI) fields
- Count number of tools defined; flag if > 20 tools (known to affect model behavior)

## Files

- `internal/provider/openai.go`
- `internal/provider/anthropic.go`
- `internal/store/models.go`
