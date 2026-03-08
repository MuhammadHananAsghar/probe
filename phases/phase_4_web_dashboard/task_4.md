# Task 4: Request Detail View with Tabs

## Phase 4 — Web Dashboard

**Status:** Not Started

## Description

Build the request detail page with tabs for each aspect of a captured request.

## Requirements

- Tabs: Overview, Messages, Tools, Stream, Headers, Raw
- Overview tab: provider, model, status, timing breakdown (latency, TTFT), token counts, cost, finish reason
- Messages tab: render each message with role label, full content, token count per message
- Tools tab: tool definitions list + tool call chain (from Phase 3 data)
- Stream tab: chunk timeline visualization (from Phase 2 data)
- Headers tab: request headers (with API key masked) + response headers
- Raw tab: raw request JSON + raw response JSON with syntax highlighting

## Files

- `internal/dashboard/ui/src/pages/RequestDetail.tsx`
- `internal/dashboard/ui/src/components/JsonViewer.tsx`
