package provider

import (
	"strings"

	"github.com/MuhammadHananAsghar/probe/internal/store"
	"github.com/tidwall/gjson"
)

// OpenRouter implements Provider for OpenRouter's OpenAI-compatible API.
// OpenRouter uses the provider/model naming convention (e.g. "anthropic/claude-3-5-sonnet").
type OpenRouter struct {
	openai OpenAI
}

// Name returns the provider identifier.
func (or *OpenRouter) Name() store.ProviderName { return store.ProviderOpenRouter }

// ParseRequest delegates to OpenAI parser, then normalises the model name.
func (or *OpenRouter) ParseRequest(body []byte, req *store.Request) error {
	if err := or.openai.ParseRequest(body, req); err != nil {
		return err
	}
	// Preserve the full OpenRouter model string (provider/model).
	// The cost lookup will strip the prefix if needed.
	if m := gjson.GetBytes(body, "model").String(); m != "" {
		req.Model = m
	}
	return nil
}

// ParseResponse delegates to the OpenAI parser.
func (or *OpenRouter) ParseResponse(body []byte, req *store.Request) error {
	return or.openai.ParseResponse(body, req)
}

// ParseEvent delegates to the OpenAI streaming parser.
func (or *OpenRouter) ParseEvent(eventType, data string, req *store.Request) string {
	return or.openai.ParseEvent(eventType, data, req)
}

// OpenRouterBaseModel strips the provider prefix from an OpenRouter model name.
// "anthropic/claude-3-5-sonnet-20241022" → "claude-3-5-sonnet-20241022"
func OpenRouterBaseModel(model string) string {
	if idx := strings.Index(model, "/"); idx >= 0 {
		return model[idx+1:]
	}
	return model
}
