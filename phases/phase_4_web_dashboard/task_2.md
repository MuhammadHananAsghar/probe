# Task 2: WebSocket Live Request Streaming

## Phase 4 — Web Dashboard

**Status:** Not Started

## Description

Implement a WebSocket endpoint that pushes new requests to the browser dashboard in real-time as probe intercepts them.

## Requirements

- WebSocket endpoint at `/api/ws` using gorilla/websocket
- On connect: send all existing requests in the ring buffer as a batch
- On each new request captured: broadcast `{"type":"request","data":{...}}` JSON message to all connected clients
- On request update (stream completes, TTFT recorded): broadcast `{"type":"update","id":"...","data":{...}}`
- Handle client disconnect gracefully (remove from broadcast set)
- Support multiple simultaneous dashboard clients

## Files

- `internal/dashboard/ws.go`
- `internal/dashboard/server.go`
