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
