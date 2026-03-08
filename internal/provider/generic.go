package provider

import (
	"github.com/MuhammadHananAsghar/probe/internal/store"
)

// Generic is an OpenAI-compatible provider that delegates all parsing to the
// OpenAI implementation while reporting a custom provider name.
// Use NewGeneric to construct an instance.
type Generic struct {
	name   store.ProviderName
	openai OpenAI
}

// NewGeneric returns a Generic provider that reports the given name and parses
// requests/responses using the OpenAI format.
func NewGeneric(name store.ProviderName) *Generic {
	return &Generic{name: name}
}

// Name returns the provider identifier supplied at construction time.
func (g *Generic) Name() store.ProviderName {
	return g.name
}

// ParseRequest delegates to the OpenAI parser.
func (g *Generic) ParseRequest(body []byte, req *store.Request) error {
	return g.openai.ParseRequest(body, req)
}

// ParseResponse delegates to the OpenAI parser.
func (g *Generic) ParseResponse(body []byte, req *store.Request) error {
	return g.openai.ParseResponse(body, req)
}

// ParseEvent delegates to the OpenAI streaming parser, enabling Generic
// providers to handle SSE streaming without a dedicated parser.
func (g *Generic) ParseEvent(eventType, data string, req *store.Request) string {
	return g.openai.ParseEvent(eventType, data, req)
}
