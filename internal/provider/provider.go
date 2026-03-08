// Package provider parses LLM API requests and responses for specific providers.
package provider

import (
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// Provider parses LLM API requests and responses for a specific provider.
type Provider interface {
	// Name returns the provider's identifier.
	Name() store.ProviderName
	// ParseRequest extracts structured data from a raw request body.
	ParseRequest(body []byte, req *store.Request) error
	// ParseResponse extracts structured data from a raw response body.
	ParseResponse(body []byte, req *store.Request) error
}

// StreamParser is implemented by providers that support SSE streaming.
// Providers implement both Provider and StreamParser.
type StreamParser interface {
	// ParseEvent processes a single SSE event and returns the content delta.
	// eventType is the SSE "event:" line value (empty string for OpenAI which
	// has no event type). data is the raw JSON string from the "data:" line.
	// Implementations must update req fields (tokens, tool calls, finish reason)
	// as needed. They return the text delta for TTFT/chunk content tracking.
	ParseEvent(eventType, data string, req *store.Request) string
}
