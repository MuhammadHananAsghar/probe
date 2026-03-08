# Task 4: Build Tool Call Chain

## Phase 3 — Tool Call Inspector

**Status:** Not Started

## Description

Reconstruct the full tool call chain across multiple requests in a session — model calls tool → app executes → app returns result → model continues.

## Requirements

- Group requests by conversation/session context (detect via identical system prompt + message history prefix)
- Build a `ToolCallChain` struct: ordered list of steps (model-decision, tool-result, model-continues)
- Each step records: tool name, arguments, result, timing
- Handle multi-step chains (model calls multiple tools before final response)
- Handle parallel tool calls (OpenAI supports multiple tool_calls in one response)

## Files

- `internal/analyze/session.go`
- `internal/store/models.go`
