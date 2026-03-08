# Task 5: Terminal UI — Tool Call Tree View

## Phase 3 — Tool Call Inspector

**Status:** Not Started

## Description

Display the tool call chain as a navigable tree in the TUI, matching the ASCII art shown in the project plan.

## Requirements

- Show each tool call as a numbered step with arrows connecting them
- Display: step number, tool name, arguments (formatted JSON, truncated if long), result preview, execution latency
- Show final model response as the terminal node
- Navigate with arrow keys; press Enter on a step to expand full arguments/result
- Press `t` in request detail view to switch to tool call tree view
- Show summary: "N tool calls | Xms total execution time"

## Files

- `internal/tui/tools.go`
- `internal/tui/detail.go`
- `internal/tui/app.go`
