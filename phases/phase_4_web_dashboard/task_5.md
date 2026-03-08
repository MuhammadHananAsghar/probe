# Task 5: JSON Viewer Component

## Phase 4 — Web Dashboard

**Status:** Not Started

## Description

Build a reusable JSON viewer with syntax highlighting, collapsible nodes, and copy-to-clipboard.

## Requirements

- Syntax-highlight JSON: strings in one color, numbers in another, keys in another, booleans/nulls distinct
- Collapsible objects and arrays: click to expand/collapse
- Deep objects collapsed by default beyond depth 3; top-level keys always visible
- Copy button copies the entire JSON or a selected sub-tree
- Search/highlight: type to highlight matching keys/values
- Handle invalid JSON gracefully (show as plain text with error indicator)

## Files

- `internal/dashboard/ui/src/components/JsonViewer.tsx`
